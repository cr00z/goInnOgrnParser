package urlqueue

import (
	"github.com/cr00z/goInnOgrnParser/csvproc"

	"container/list"
	"runtime"
	"strings"
	"sync"
)

type UrlRecord struct {
	Url            string
	Id             string
	Primary        bool
	NormalPriority map[string]int
	HighPriority   map[string]int
	Visited        map[string]bool
	WaitCommit     bool
}

type UrlQueue struct {
	Queue    *list.List
	Position *list.Element
	Mux      sync.Mutex
}

type Work struct {
	Url      string
	Path     string
	Level    int
	High     bool
	Position *list.Element
	Body     string
}

func (allUrls *UrlQueue) MakeUrlQueue(records *[]csvproc.Record) {
	allUrls.Queue = list.New()
	for _, record := range *records {
		primary := true
		for _, url := range strings.Split(record.Web, ",") {
			url := strings.TrimSuffix(url, "/")
			// TODO: whether the secondary site is synonymous with the primary?
			// TODO: some records have same urls. Same urls may be primary or secondary for diff urls?
			allUrls.Queue.PushBack(&UrlRecord{
				Url:            url,
				Id:             record.Id,
				Primary:        primary,
				NormalPriority: map[string]int{},
				HighPriority:   map[string]int{"/": 1},
				Visited:        map[string]bool{},
				WaitCommit:     false,
			})
			primary = false
		}
	}
	allUrls.Position = allUrls.Queue.Back()
}

func (allUrls *UrlQueue) SetNextPosition() *list.Element {
	allUrls.Mux.Lock()
	defer allUrls.Mux.Unlock()

	pos := allUrls.Position
	if pos == nil {
		return nil
	}
	allUrls.Position = allUrls.Position.Next()
	if allUrls.Position == nil {
		allUrls.Position = allUrls.Queue.Front()
	}
	return allUrls.Position
}

func (allUrls *UrlQueue) GetNextUrl() *Work {
	pos := allUrls.SetNextPosition()
	for pos != nil {
		allUrls.Mux.Lock()
		wait := pos.Value.(*UrlRecord).WaitCommit
		allUrls.Mux.Unlock()
		if wait {
			pos = allUrls.SetNextPosition()
			runtime.Gosched()
		} else {
			break
		}
	}
	if pos == nil {
		return nil
	}

	var path string
	var lvl int
	var high, founded bool

	allUrls.Mux.Lock()
	defer allUrls.Mux.Unlock()

	urlRecord := pos.Value.(*UrlRecord)

	for path, lvl = range urlRecord.HighPriority {
		founded, high = true, true
		break
	}
	for path, lvl = range urlRecord.NormalPriority {
		founded = true
		break
	}
	if founded {
		urlRecord.WaitCommit = true
		return &Work{urlRecord.Url, path, lvl, high, pos, ""}
	}
	return nil
}

func (allUrls *UrlQueue) AddUrls(work *Work, paths []string) int {
	allUrls.Mux.Lock()
	defer allUrls.Mux.Unlock()

	added := 0
	urlRecord := work.Position.Value.(*UrlRecord)

	for _, path := range paths {
		_, isAvail := urlRecord.Visited[path]
		if !isAvail {
			// TODO: make insert for high priority

			_, isAvail := urlRecord.NormalPriority[path]
			if !isAvail {
				urlRecord.NormalPriority[path] = work.Level + 1
				added++
			}
		}
	}
	return added
}

func (allUrls *UrlQueue) Commit(work *Work) {
	allUrls.Mux.Lock()
	defer allUrls.Mux.Unlock()

	urlRecord := work.Position.Value.(*UrlRecord)
	if work.High {
		delete(urlRecord.HighPriority, work.Path)
	} else {
		delete(urlRecord.NormalPriority, work.Path)
	}
	urlRecord.WaitCommit = false
	if len(urlRecord.HighPriority) == 0 && len(urlRecord.NormalPriority) == 0 {
		if allUrls.Position == work.Position {
			allUrls.Position = allUrls.SetNextPosition()
		}
		allUrls.Queue.Remove(work.Position)
	} else {
		urlRecord.Visited[work.Path] = true
	}
}
