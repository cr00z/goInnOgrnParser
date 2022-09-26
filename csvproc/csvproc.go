package csvproc

import (
	"encoding/csv"
	"os"
	"reflect"
	"strings"
)

type Record struct {
	Id string
	/*
		...
		truncated
		...
	*/
	Web string
	/*
		...
		truncated
		...
	*/
}

type Info struct {
	MainINN  string
	MainOGRN string
	AddINN   []string
	AddOGRN  []string
	Error    string
}

func ReadAll(fileName string, records *[]Record) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	// skip first
	_, err = reader.Read()
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		*records = append(*records, Record{
			Id: record[0],
			/*
				...
				truncated
				...
			*/
			Web: record[1],
			/*
				...
				truncated
				...
			*/
		})
	}
	return nil
}

func WriteAll(fileName string, records *[]Record, info *map[string]Info) (int, error) {
	title := []string{
		// main
		"id" /* ...truncated... */, "web", /* ...truncated... */
		// additional
		"main_inn", "main_ogrn", "add_inn", "add_ogrn", "http_error",
	}

	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		return 0, err
	}

	writer := csv.NewWriter(file)
	writer.Comma = ','
	defer writer.Flush()

	writer.Write(title)

	count := 0
	for _, record := range *records {
		writeStr := make([]string, 0)
		recordVal := reflect.ValueOf(record)
		for field := 0; field < recordVal.NumField(); field++ {
			writeStr = append(writeStr, recordVal.Field(field).String())
		}

		addInfo := make([]string, 5)
		if val, isOk := (*info)[record.Id]; isOk {
			addInfo = []string{
				val.MainINN,
				val.MainOGRN,
				strings.Join(val.AddINN, ","),
				strings.Join(val.AddOGRN, ","),
				val.Error,
			}
			if addInfo[0] != "" || addInfo[1] != "" || addInfo[2] != "" || addInfo[3] != "" {
				count++
			}
		}
		writeStr = append(writeStr, addInfo...)

		if err := writer.Write(writeStr); err != nil {
			// TODO: write to log and continue
			return count, err
		}
	}
	return count, nil
}
