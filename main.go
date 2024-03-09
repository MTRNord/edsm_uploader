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

func main() {
	// Take as a first positional argument the path to the folder where the journal files are located.
	// Take as a second positional argument the commander name.
	// Take as a third positional argument the EDSM API key.

	jounnalPath := os.Args[1]
	commanderName := os.Args[2]
	apiKey := os.Args[3]

	// Create a new EDSM object.
	edsm := edsm.NewEDSM(commanderName, apiKey)

	// Find all the journal files in the folder and parse them after sorting them by date.
	files := make(map[string]string)
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
				date := parts[1]
				files[date] = path
			}

		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking the path: %+v", err)
	}

	// Sort the files by date.
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	journal_obj := journal.NewJournal(edsm)
	for _, k := range keys {
		// Parse the journal file.
		err := journal_obj.ParseJournal(files[k])
		if err != nil {
			log.Fatalf("Error parsing journal file: %+v", err)
		}
		// sleep for 1 second to not overload the EDSM API.
		time.Sleep(1 * time.Second)
	}

	log.Println("Done parsing journal files.")
}
