package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"log"

	"github.com/MTRNord/edsm_uploader/edsm"
	"github.com/MTRNord/edsm_uploader/journal"
	"github.com/pkg/errors"
)

type JournalFile struct {
	path string
	date time.Time
}

func main() {
	// Take as a first positional argument the path to the folder where the journal files are located.
	// Take as a second positional argument the commander name.
	// Take as a third positional argument the EDSM API key.

	logger := log.New(os.Stderr, "edsm_uploader: ", log.LstdFlags)

	jounnalPath := os.Args[1]
	commanderName := os.Args[2]
	apiKey := os.Args[3]

	// Create a new EDSM object.
	edsm := edsm.NewEDSM(commanderName, apiKey, logger)

	// Find all the journal files in the folder and parse them after sorting them by date.
	files := make(map[string]JournalFile)
	err := filepath.WalkDir(jounnalPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}
		if !d.IsDir() {
			// If the file starts with "Journal." then we add it to the files map.
			// The date is after "Journal." and before the first ".".
			// Example: "Journal.2024-03-09T104947.01.log"
			// The date is "2024-03-09T104947.01"
			if d.Name()[:8] == "Journal." {
				name := d.Name()
				parts := strings.Split(name, ".")
				date := parts[1] + "." + parts[2]
				parsedDate, err := time.Parse("2006-01-02T150405.00", date)
				if err != nil {
					return errors.WithStack(err)
				}
				files[date] = JournalFile{
					path: path,
					date: parsedDate,
				}
			}

		}
		return nil
	})
	if err != nil {
		logger.Fatalf("Error walking the path: %+v", err)
	}

	// Sort the files by date.
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	journal_obj := journal.NewJournal(edsm, logger)
	for _, k := range keys {
		// Parse the journal file.
		// Parse date of the journal file and ignore if older than journal_obj.lastDate.

		currentDate := *journal_obj.LastDate
		startOfDay := Bod(currentDate)
		if journal_obj.LastDate != nil && files[k].date.Before(startOfDay) {
			continue
		}

		err := journal_obj.ParseJournal(files[k].path)
		if err != nil {
			logger.Fatalf("Error parsing journal file: %+v", err)
		}
		// sleep for 1 second to not overload the EDSM API.
		time.Sleep(1 * time.Second)
	}

	logger.Println("Done parsing journal files.")
}

func Bod(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
