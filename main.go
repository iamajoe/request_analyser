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

// stats checks the general statistics
func stats(srcRaw string) error {
	filePath, err := sourceToFilePath(srcRaw)
	if err != nil {
		return err
	}

	// TODO: need to read one by one so that we dont screw ourselves
	log.Println("file path", filePath)

	return nil
}

func main() {
	statsFs := flag.NewFlagSet("stats", flag.ExitOnError)
	srcRaw := statsFs.String("s", "", "source of the records")

	switch os.Args[1] {
	case "stats":
		if err := statsFs.Parse(os.Args[2:]); err != nil {
			statsFs.PrintDefaults()
			log.Fatal(err)
		}

		if err := stats(*srcRaw); err != nil {
			log.Fatal(err)
		}
	default:
		help()
	}
}
