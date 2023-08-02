package main

import (
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

	switch os.Args[1] {
	case "parse":
		if err := parseFs.Parse(os.Args[2:]); err != nil {
			parseFs.PrintDefaults()
			log.Fatal(err)
		}

		if err := parse(*parseSrcRaw, *parseOutputRaw); err != nil {
			log.Fatal(err)
		}
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
	default:
		help()
	}
}
