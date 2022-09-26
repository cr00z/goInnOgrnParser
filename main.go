package main

import (
	"github.com/cr00z/goInnOgrnParser/crawler"
	"github.com/cr00z/goInnOgrnParser/csvproc"
	"github.com/cr00z/goInnOgrnParser/innogrn"
	"github.com/cr00z/goInnOgrnParser/options"
	"github.com/cr00z/goInnOgrnParser/urlqueue"
	"github.com/cr00z/goInnOgrnParser/workerpool"

	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var httpClient *http.Client

var processFn = func(args interface{}) (interface{}, error) {
	work, ok := args.(*urlqueue.Work)
	if !ok {
		return work, errors.New("wrong argument type")
	}
	body, err := crawler.Retrieve(httpClient, work.Url, work.Path)
	work.Body = body
	return work, err
}

func ProcessNumbers(numbers map[string]int, info *map[string]csvproc.Info, id string, primary bool) int {
	founded := 0
	infoRecord, isOk := (*info)[id]
	if !isOk {
		infoRecord = csvproc.Info{
			AddINN:  []string{},
			AddOGRN: []string{},
		}
	}
	for number := range numbers {
		if innogrn.CheckINN(number) {
			if primary && infoRecord.MainINN == "" {
				infoRecord.MainINN = number
			} else {
				infoRecord.AddINN = append(infoRecord.AddINN, number)
			}
			founded++
		}
		if innogrn.CheckOGRN(number) {
			if primary && infoRecord.MainOGRN == "" {
				infoRecord.MainOGRN = number
			} else {
				infoRecord.AddOGRN = append(infoRecord.AddOGRN, number)
			}
			founded++
		}
	}
	(*info)[id] = infoRecord
	return founded
}

func ProcessError(err error, info *map[string]csvproc.Info, id string) {
	infoRecord, isOk := (*info)[id]
	if !isOk {
		infoRecord = csvproc.Info{
			AddINN:  []string{},
			AddOGRN: []string{},
		}
	}
	infoRecord.Error = crawler.HTTPErrorToString(err)
	(*info)[id] = infoRecord
}

func processQueue(allUrls *urlqueue.UrlQueue, interChan chan<- workerpool.Task, queueChan chan<- bool,
	resultFn workerpool.ResultFunction) {

	taskID := 1
	for {
		work := allUrls.GetNextUrl()
		if work == nil {
			queueChan <- true
			break
		}
		options.ILog.Printf("Process url '%s'", work.Url+work.Path)
		task := workerpool.Task{
			ID:        taskID,
			Args:      work,
			ProcessFn: processFn,
			ResultFn:  resultFn,
		}
		interChan <- task
		taskID++
	}
}

func getResultFn(allUrls *urlqueue.UrlQueue, info *map[string]csvproc.Info) workerpool.ResultFunction {
	return func(args interface{}) {
		removeRecord := false
		result := args.(workerpool.Result)
		work := result.Value.(*urlqueue.Work)
		urlRecord := work.Position.Value.(*urlqueue.UrlRecord)
		if result.Error != nil {
			options.ELog.Println(result.Error)
			if work.Path == "/" {
				ProcessError(result.Error, info, urlRecord.Id)
				removeRecord = true
			}
		} else {
			processNumbers := options.AllNumbers ||
				strings.Contains(work.Body, "ИНН") ||
				strings.Contains(work.Body, "ОГРН")
			if processNumbers {
				numbers := crawler.ParseNumbers(work.Body)
				pnResult := ProcessNumbers(numbers, info, urlRecord.Id, urlRecord.Primary)
				if pnResult > 0 {
					options.ILog.Printf("Result '%s' numbers: %d", work.Url, pnResult)
					if options.StopAtFirst {
						removeRecord = true
					}
				}
			}
			if !removeRecord && work.Level < options.MaxParseDeepLvl {
				links := crawler.CollectLinks(work.Url, work.Body)
				added := allUrls.AddUrls(work, links)
				options.TotalLinks += added
			}
		}
		if removeRecord {
			allUrls.Mux.Lock()
			urlRecord.HighPriority = map[string]int{}
			urlRecord.NormalPriority = map[string]int{}
			urlRecord.Visited = map[string]bool{}
			allUrls.Mux.Unlock()
		}
		allUrls.Commit(work)
		options.TotalLinks--
		options.ILog.Printf(
			"Finished work %d, left %d urls [ %d paths ]\n",
			result.ID, allUrls.Queue.Len(), options.TotalLinks,
		)
		removeRecord = false
	}
}

func main() {
	options.ParseOptions()
	start := time.Now()
	httpClient = crawler.NewHttpClient()

	records := make([]csvproc.Record, 0)
	info := map[string]csvproc.Info{}

	err := csvproc.ReadAll(options.InputCSVFile, &records)
	if err != nil {
		log.Fatalf("ReadAll error: %s", err)
	}
	options.PLog.Printf("Read file '%s', read %d records\n", options.InputCSVFile, len(records))

	var allUrls urlqueue.UrlQueue
	allUrls.MakeUrlQueue(&records)
	options.TotalLinks = allUrls.Queue.Len()
	options.PLog.Printf("Processed file '%s', read %d urls\n", options.InputCSVFile, options.TotalLinks)

	ctx, cancelFn := context.WithCancel(context.Background())
	pool := workerpool.MakePool(options.WorkerPoolSize)
	go pool.StartInterChan(ctx)
	go pool.ProcessResults()
	pool.RunWorkers()
	time.Sleep(time.Second)

	queueChan := make(chan bool, 1)
	go processQueue(&allUrls, pool.GetInterChan(), queueChan, getResultFn(&allUrls, &info))

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-queueChan:
		options.PLog.Println("Queue is empty")
	case <-termChan:
		options.PLog.Println()
		options.PLog.Println("******************************************")
		options.PLog.Println("*        Shutdown signal received        *")
		options.PLog.Println("******************************************")
	case <-time.After(time.Second * time.Duration(options.ExecTimeLimit)):
		options.PLog.Println("Time exceed")
	}

	cancelFn()
	pool.WaitWorkers()
	pool.ResultDone()
	pool.WaitResults()
	options.PLog.Println("All workers done, shutting done!")

	count, err := csvproc.WriteAll(options.OutputCSVFile, &records, &info)
	if err != nil {
		log.Fatalf("WriteAll error: %s", err)
	} else {
		options.PLog.Printf("Founded %d results (%d %%)", count, count*100/len(records))
	}

	elapsed := time.Since(start)
	options.PLog.Printf("Took ================> %s\n", elapsed)
}
