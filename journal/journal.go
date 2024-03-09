package journal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MTRNord/edsm_uploader/datatypes"
	"github.com/MTRNord/edsm_uploader/edsm"
	"github.com/pkg/errors"

	_ "github.com/mattn/go-sqlite3"
)

type FileHeader struct {
	// Define the fields of the FileHeader struct here.
}

type Journal struct {
	edsm       *edsm.EDSM
	lastDate   *time.Time
	fileHeader *datatypes.FileHeader
	logger     *log.Logger
}

func NewJournal(edsm *edsm.EDSM, logger *log.Logger) *Journal {
	// Read lastDate from latest.txt
	file, err := os.OpenFile("latest.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		logger.Printf("Error opening latest.txt: %s", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var lastDate *time.Time
	if scanner.Scan() {
		text := scanner.Text()
		// Check if the latest.txt is empty.
		// If it is empty then we need to set the lastDate to nil.
		if text == "" {
			lastDate = nil
		} else {
			lastDateL, err := time.Parse(time.RFC3339, text)
			if err != nil {
				logger.Printf("Error parsing latest.txt: %s", err)
			}
			lastDate = &lastDateL
		}
	}

	return &Journal{
		// TODO: Parse the last date from the sqlite database.
		// If there is no last date, then we need to start from the beginning and define it as nil.
		edsm:       edsm,
		lastDate:   lastDate,
		fileHeader: nil,
		logger:     logger,
	}
}

func (j *Journal) ParseJournal(journalPath string) error {
	splitPath := strings.Split(journalPath, "/")
	j.logger.Printf("Parsing journal file: %s", splitPath[len(splitPath)-1])
	file, err := os.Open(journalPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Parse first line first
	if scanner.Scan() {
		err := j.parseLine(scanner.Text())
		if err != nil {
			return errors.WithStack(err)
		}
	}
	var wg sync.WaitGroup
	for scanner.Scan() {
		wg.Add(1)
		go func(line string) error {
			defer wg.Done()
			err := j.parseLine(line)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}(scanner.Text())
		time.Sleep(1 * time.Millisecond)
	}
	if err := scanner.Err(); err != nil {
		return errors.WithStack(err)
	}
	wg.Wait()

	return nil
}

func (j *Journal) storeLastDate(timestamp time.Time) {
	if j.lastDate == nil || j.lastDate.Before(timestamp) {
		j.lastDate = &timestamp
		file, err := os.OpenFile("latest.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			j.logger.Printf("Error opening latest.txt: %s", err)
		}
		defer file.Close()
		_, err = file.WriteString(timestamp.Format(time.RFC3339))
		if err != nil {
			j.logger.Printf("Error writing to latest.txt: %s", err)
		}
	}
}

// This parses the journal line of the elite dangerous journal.
// TODO: We need to store where we are in the journal so we can pick up where we left off.
func (j *Journal) parseLine(line string) error {
	// Parse json line as a JournalLine.
	var journalLine datatypes.JournalLine
	bytesString := []byte(line)
	bytesString = bytes.Trim(bytesString, "\x00")
	// Exit early if the line is empty
	if len(bytesString) == 0 {
		return nil
	}
	err := json.Unmarshal(bytesString, &journalLine)
	if err != nil {
		return errors.WithStack(err)
	}

	if journalLine.Event == "Fileheader" {
		// Parse the json line as a FileHeader.
		var fileHeader datatypes.FileHeader
		json.Unmarshal([]byte(line), &fileHeader)

		// Store the file header.
		j.fileHeader = &fileHeader
	}

	// If we have a startdate make sure new lines are newer than the startdate.
	parsedDate, err := time.Parse(time.RFC3339, journalLine.Timestamp)
	if err != nil {
		return errors.WithStack(err)
	}
	if j.lastDate != nil {
		lastDate, err := time.Parse(time.RFC3339, j.lastDate.Format(time.RFC3339))
		if err != nil {
			return errors.WithStack(err)
		}
		if parsedDate.Before(lastDate) {
			return nil
		}
	}

	// Send the line to edsm.
	err = j.edsm.SendJournalLine(j.fileHeader, line)
	if err != nil {
		return errors.WithStack(err)
	}

	// Store the last date.
	j.storeLastDate(parsedDate)

	return nil
}
