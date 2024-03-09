package edsm

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/MTRNord/edsm_uploader/datatypes"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

type EDSM struct {
	commanderName string
	apiKey        string
	client        *retryablehttp.Client
}

func NewEDSM(commanderName string, apiKey string) *EDSM {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 10
	return &EDSM{
		commanderName: commanderName,
		apiKey:        apiKey,
		client:        retryClient,
	}
}

type RequestJSON struct {
	CommanderName   string `json:"commanderName"`
	ApiKey          string `json:"apiKey"`
	Software        string `json:"fromSoftware"`
	SoftwareVersion string `json:"fromSoftwareVersion"`
	GameVersion     string `json:"fromGameVersion"`
	GameBuild       string `json:"fromGameBuild"`
	Message         string `json:"message"`
}

func (e *EDSM) SendJournalLine(fileHeader *datatypes.FileHeader, journalLine string) error {
	// Send HTTP Post request to https://www.edsm.net/api-journal-v1
	// If the response is less then 200 then return an error.

	// Build the JSON request.
	requestJSON := RequestJSON{
		CommanderName:   e.commanderName,
		ApiKey:          e.apiKey,
		Software:        "github.com/MTRNord/edsm_uploader",
		SoftwareVersion: "0.1.0",
		GameVersion:     fileHeader.GameVersion,
		GameBuild:       fileHeader.Build,
		Message:         journalLine,
	}
	// Convert struct to json
	b, err := json.Marshal(requestJSON)
	if err != nil {
		return errors.WithStack(err)
	}

	resp, err := e.client.Post("https://www.edsm.net/api-journal-v1", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 {
		// Create error from response
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		bodyString := string(bodyBytes)
		return &EDSMError{
			Status:  resp.StatusCode,
			Message: bodyString,
		}
	} /* else {
		// Log the response
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		bodyString := string(bodyBytes)
		log.Printf("Response from EDSM: %s", bodyString)
	} */

	return nil
}

type EDSMError struct {
	Status  int
	Message string
}

func (e *EDSMError) Error() string {
	return e.Message
}
