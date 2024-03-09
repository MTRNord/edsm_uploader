package datatypes

// The journal has some standard fields on each line that are common. Timestamp and Event. The rest of the fields are specific to the event.
// Example line: `{ "timestamp":"2024-03-09T10:49:40Z", "event":"Fileheader", "part":1, "language":"German/DE", "Odyssey":true, "gameversion":"4.0.0.1801", "build":"r300472/r0 " }`
type JournalLine struct {
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
}

// The fileheader is special and we need to parse it each time to make sure we know the right version.
// Example header: `{ "timestamp":"2024-03-09T10:49:40Z", "event":"Fileheader", "part":1, "language":"German/DE", "Odyssey":true, "gameversion":"4.0.0.1801", "build":"r300472/r0 " }`
type FileHeader struct {
	Timestamp   string `json:"timestamp"`
	Event       string `json:"event"`
	Part        int    `json:"part"`
	Language    string `json:"language"`
	Odyssey     bool   `json:"Odyssey"`
	GameVersion string `json:"gameversion"`
	Build       string `json:"build"`
}
