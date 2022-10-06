package jsonhandling

import (
	"encoding/json"
	"log"
	"time"
)

func HandleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Message struct {
	Sender      string    `json:"sender,omitempty" binding:"required"`
	Receiver    string    `json:"receiver,omitempty" binding:"required"`
	MessageBody string    `json:"message,omitempty" binding:"required"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func JsonMarshal(msg Message) []byte {
	json, err := json.Marshal(msg)
	HandleError(err, "Error encoding JSON")
	return json
}

func JsonUnmarshal(data []byte, msg Message) Message {
	err := json.Unmarshal(data, &msg)
	HandleError(err, "Error decoding JSON")
	return msg
}

func RadisJsonMarshal(msgs []Message) []byte {
	json, err := json.Marshal(msgs)
	HandleError(err, "Error encoding JSON")
	return json
}

func RadisJsonUnmarshal(data []byte, msgs []Message) []Message {
	err := json.Unmarshal(data, &msgs)
	HandleError(err, "Error decoding JSON")
	return msgs
}
