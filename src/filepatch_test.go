package onebotfilter

import (
	"encoding/json"
	"testing"
)

func TestNormalizeFileSegmentsAddsFileID(t *testing.T) {
	raw := []byte(`{"post_type":"message","message_type":"private","user_id":123,"message":[{"type":"file","data":{"id":"abc123","name":"x.har","size":100}}]}`)
	out := NormalizeFileSegments(raw)
	var event struct {
		Message []struct {
			Type string                 `json:"type"`
			Data map[string]interface{} `json:"data"`
		} `json:"message"`
	}
	if err := json.Unmarshal(out, &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(event.Message) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(event.Message))
	}
	seg := event.Message[0]
	if seg.Type != "file" {
		t.Fatalf("expected file segment, got %s", seg.Type)
	}
	if seg.Data["file_id"] != "abc123" {
		t.Fatalf("file_id not added: %v", seg.Data)
	}
	if seg.Data["file"] != "abc123" {
		t.Fatalf("file not added: %v", seg.Data)
	}
}

func TestNormalizeFileSegmentsLeavesImageAlone(t *testing.T) {
	raw := []byte(`{"post_type":"message","message_type":"group","group_id":1,"user_id":2,"message":[{"type":"image","data":{"file":"res1","url":"http://x/img"}}]}`)
	out := NormalizeFileSegments(raw)
	if string(out) != string(raw) {
		t.Fatalf("image message should pass through unchanged:\n%s\n%s", raw, out)
	}
}

func TestNormalizeFileSegmentsKeepsExistingFileID(t *testing.T) {
	raw := []byte(`{"post_type":"message","message_type":"private","user_id":123,"message":[{"type":"file","data":{"id":"abc","file_id":"keep","name":"x"}}]}`)
	out := NormalizeFileSegments(raw)
	var event struct {
		Message []struct {
			Data map[string]interface{} `json:"data"`
		} `json:"message"`
	}
	if err := json.Unmarshal(out, &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if event.Message[0].Data["file_id"] != "keep" {
		t.Fatalf("existing file_id should not be overwritten: %v", event.Message[0].Data)
	}
	if event.Message[0].Data["file"] != "abc" {
		t.Fatalf("missing file field should be filled from id: %v", event.Message[0].Data)
	}
}

func TestIsMessageEventOnlyForMessages(t *testing.T) {
	msg := []byte(`{"post_type":"message","message_type":"group","group_id":1,"user_id":2,"message":[{"type":"text","data":{"text":"hi"}}]}`)
	notice := []byte(`{"post_type":"notice","notice_type":"friend_recall","user_id":2}`)
	request := []byte(`{"post_type":"request","request_type":"friend","user_id":2}`)
	meta := []byte(`{"post_type":"meta_event","meta_event_type":"heartbeat","status":{}}`)
	if !isMessageEvent(msg) {
		t.Fatalf("message event should be detected")
	}
	for name, raw := range map[string][]byte{
		"notice":  notice,
		"request": request,
		"meta":    meta,
	} {
		if isMessageEvent(raw) {
			t.Fatalf("%s event should NOT be treated as a message: %s", name, raw)
		}
	}
	if isMessageEvent(nil) || isMessageEvent([]byte(`not json`)) {
		t.Fatalf("unparsable payload should not be treated as a message")
	}
}
