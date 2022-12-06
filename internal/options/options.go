package options

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	InputCSVFile    string
	OutputCSVFile   string
	WorkerPoolSize  int
	ExecTimeLimit   int
	MaxParseDeepLvl int
	HttpTimeout     int
	verbose         bool
	silence         bool
	AllNumbers      bool
	StopAtFirst     bool

	TotalLinks = 0

	PLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime) // process Log
	ILog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime) // info log
	ELog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
)

func ParseOptions() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "    -%s    %s", f.Name, f.Usage)
			if f.DefValue != "false" {
				fmt.Fprintf(os.Stderr, " [default %s]\n", f.DefValue)
			} else {
				fmt.Fprintf(os.Stderr, "\n")
			}
		})
		fmt.Fprintf(os.Stderr, "Use ^C for shutdown and save results.\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "%s -a -d 3 -i in.csv -o out.csv -t 3600 -v -w 1000\n", os.Args[0])
	}

	flag.StringVar(&InputCSVFile, "i", "data/part.csv", "input csv file")
	flag.StringVar(&OutputCSVFile, "o", "data/out.csv",
		"output CSV file, added result columns ('main_inn', 'main_ogrn', 'add_inn', 'add_ogrn')")

	flag.IntVar(&WorkerPoolSize, "w", 10, "worker pool size (number of threads)")
	flag.IntVar(&ExecTimeLimit, "t", 12*60*60, "script execution limit in seconds")
	flag.IntVar(&MaxParseDeepLvl, "d", 1, "maximum parsing depth")
	flag.IntVar(&HttpTimeout, "h", 30, "http timeout, maximum server response time")

	flag.BoolVar(&verbose, "v", false, "verbose output, added messages about processed urls, results, queue size, etc")
	flag.BoolVar(&silence, "s", false, "silence (ignored by verbose)")
	flag.BoolVar(&AllNumbers, "a", false, "get all numbers similar to inn and ogrn (lots of result = lots of garbage)")
	flag.BoolVar(&StopAtFirst, "f", false, "do not continue processing the link when the first result is received")

	flag.Parse()

	if !verbose {
		ILog.SetOutput(io.Discard)
		ELog.SetOutput(io.Discard)
		if silence {
			PLog.SetOutput(io.Discard)
		}
	}
}
