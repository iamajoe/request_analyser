package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
)

func help() {
	log.Println(
		"Usage:\n./request_analyser <parse|stats|run> [options...]\n\nCheck documentation for more information",
	)
}

type runnerWriter struct {
	output string
}

func (w *runnerWriter) Write(v []byte) (int, error) {
	// send to result tmp file
	f, err := os.OpenFile(w.output,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return 1, err
	}

	if _, err := f.Write(v); err != nil {
		return 1, err
	}

	// send to stdout
	arr := strings.Split(string(v), ";;")
	log.Println("===============================")
	for _, v := range arr {
		log.Println("-", v)
	}
	log.Println("===============================")

	return 0, nil
}

func main() {
	parseFs := flag.NewFlagSet("parse", flag.ExitOnError)
	parseSrcRaw := parseFs.String("s", "", "source of the records")
	parseOutputRaw := parseFs.String("o", "tmp_parse", "output of the parsed records")
	parseHelpRaw := parseFs.Bool("h", false, "help manual")

	statsFs := flag.NewFlagSet("stats", flag.ExitOnError)
	statsInputRaw := statsFs.String("i", "", "input with parsed records")
	statsHelpRaw := statsFs.Bool("h", false, "help manual")

	runFs := flag.NewFlagSet("run", flag.ExitOnError)
	runInputRaw := runFs.String("i", "", "input with parsed records")
	runOutputRaw := runFs.String("o", "tmp_run_result", "output with stat results")
	runBaseRaw := runFs.String(
		"b",
		"http://localhost:4040",
		"base url to use when no http|https provided",
	)
	runConcurrRaw := runFs.Int("c", 1, "number of concurrent requests")
	runUnixRaw := runFs.Int("t", 500, "ms unix between requests")
	runFilterRaw := runFs.String("f", "[]", "filters an array of patterns")
	runHelpRaw := runFs.Bool("h", false, "help manual")

	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "run":
		if err := runFs.Parse(os.Args[2:]); err != nil {
			runFs.PrintDefaults()
			log.Fatal(err)
		}

		if *runHelpRaw {
			runFs.PrintDefaults()
			return
		}

		// parse the filter
		filter := []string{}
		if runFilterRaw != nil && len(*runFilterRaw) > 0 {
			err := json.Unmarshal([]byte(*runFilterRaw), &filter)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err := run(
			*runInputRaw,
			*runBaseRaw,
			*runConcurrRaw,
			*runUnixRaw,
			filter,
			&runnerWriter{*runOutputRaw},
		); err != nil {
			log.Fatal(err)
		}
		break
	case "parse":
		if err := parseFs.Parse(os.Args[2:]); err != nil {
			parseFs.PrintDefaults()
			log.Fatal(err)
		}

		if *parseHelpRaw {
			parseFs.PrintDefaults()
			return
		}

		if err := parse(*parseSrcRaw, *parseOutputRaw); err != nil {
			log.Fatal(err)
		}
		break
	case "stats":
		if err := statsFs.Parse(os.Args[2:]); err != nil {
			statsFs.PrintDefaults()
			log.Fatal(err)
		}

		if *statsHelpRaw {
			statsFs.PrintDefaults()
			return
		}

		res, err := stats(*statsInputRaw, 20)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("count:", res.count, "\n")

		// methods
		for k, v := range res.requestMethodCount {
			log.Println("method count:", k, v)
		}

		log.Println("\n")

		// used requests
		for i, v := range res.mostUsed {
			log.Println(
				"most used request (",
				len(res.mostUsed)-i,
				") :",
				v.method,
				v.url,
				v.count,
			)
		}
		break
	default:
		help()
	}
}
