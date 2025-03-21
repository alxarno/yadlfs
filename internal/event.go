//nolint:wrapcheck
package internal

import "encoding/json"

type EventName string

const (
	EventNameInit      EventName = "init"
	EventNameUpload    EventName = "upload"
	EventNameComplete  EventName = "complete"
	EventNameDownload  EventName = "download"
	EventNameProgress  EventName = "progress"
	EventNameTerminate EventName = "terminate"
)

type OperationName string

const (
	OperationNameDownload OperationName = "download"
	OperationNameUpload   OperationName = "upload"
)

type Init struct {
	Event               EventName     `json:"event"`
	Operation           OperationName `json:"operation"`
	Remote              string        `json:"remote"`
	Concurrent          bool          `json:"concurrent"`
	ConcurrentTransfers int64         `json:"concurrenttransfers"`
}

type Transfer struct {
	Event EventName `json:"event"`
	OID   string    `json:"oid"`
	Size  int64     `json:"size"`
	Path  string    `json:"path"`
}

type ProgressMessage struct {
	Event          EventName `json:"event"`
	OID            string    `json:"oid"`
	BytesSoFar     int64     `json:"bytesSoFar"`
	BytesSinceLast int64     `json:"bytesSinceLast"`
}

func (m ProgressMessage) Marshal() ([]byte, error) {
	m.Event = EventNameProgress

	return json.Marshal(m)
}

type CompleteMessage struct {
	Event EventName `json:"event"`
	OID   string    `json:"oid"`
	Path  *string   `json:"path,omitempty"`
}

func (m CompleteMessage) Marshal() ([]byte, error) {
	m.Event = EventNameComplete

	return json.Marshal(m)
}

type CompleteErrorMessageContent struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type CompleteErrorMessage struct {
	Event EventName                   `json:"event"`
	OID   string                      `json:"oid"`
	Error CompleteErrorMessageContent `json:"error"`
}

func (m CompleteErrorMessage) Marshal() ([]byte, error) {
	m.Event = EventNameComplete

	return json.Marshal(m)
}

type ConfirmMessage struct {
}

func (m ConfirmMessage) Marshal() ([]byte, error) {
	return []byte("{ }"), nil
}
