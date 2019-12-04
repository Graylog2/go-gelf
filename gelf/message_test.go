package gelf

import (
	"encoding/json"
	"testing"
)

func TestWrongFieldTypes(t *testing.T) {
	msgData := map[string]string{
		"version":       `{"version": 1.1}`,
		"host":          `{"host": ["a", "b"]}`,
		"short_message": `{"short_message": {"a": "b"}}`,
		"full_message":  `{"full_message": null}`,
		"timestamp":     `{"timestamp": "12345"}`,
		"level":         `{"level": false}`,
		"facility":      `{"facility": true}`,
	}

	for k, j := range msgData {
		var msg Message
		err := json.Unmarshal([]byte(j), &msg)
		if err == nil {
			t.Errorf("expected type error on field %s", k)
		}
	}

}

func Test_constructMessage_WritesMultilineMessageToBothMessageFields(t *testing.T) {
	msgText := "hello\nthere"
	msg := constructMessage([]byte(msgText), "", "", "", 0)
	if msg.Short != msgText || msg.Full != msgText {
		t.Errorf("Short field of message \"%s\" does not coincide with message text \"%s\"", msg.Short, msgText)
	}
}
