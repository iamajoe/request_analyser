package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
)

func help() {
	// TODO: should probably setup an help
	log.Fatal(errors.New("help! i need somebody"))
}

func main() {
	parseFs := flag.NewFlagSet("parse", flag.ExitOnError)
	parseSrcRaw := parseFs.String("s", "", "source of the records")
	parseOutputRaw := parseFs.String("o", "", "output of the parsed records")

	statsFs := flag.NewFlagSet("stats", flag.ExitOnError)
	statsInputRaw := statsFs.String("i", "", "input with parsed records")

	runFs := flag.NewFlagSet("run", flag.ExitOnError)
	runInputRaw := runFs.String("i", "", "input with parsed records")
	runBaseRaw := runFs.String("b", "", "base url to use when no http|https provided")
	runConcurrRaw := runFs.Int("c", 1, "number of concurrent requests")
	runUnixRaw := runFs.Int("t", 1, "ms unix between requests")
	runSpeedRaw := runFs.Int("s", 1, "speed up requests by factor")
	runFilterRaw := runFs.String("f", "[]", "filters an array of patterns")

	switch os.Args[1] {
	case "run":
		if err := runFs.Parse(os.Args[2:]); err != nil {
			runFs.PrintDefaults()
			log.Fatal(err)
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
			*runSpeedRaw,
			filter,
		); err != nil {
			log.Fatal(err)
		}
		break
	case "parse":
		if err := parseFs.Parse(os.Args[2:]); err != nil {
			parseFs.PrintDefaults()
			log.Fatal(err)
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
