package entity

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResetPoolMessagePreservesBatchAndPublisher(t *testing.T) {
	input := ResetPool{BatchID: "1793265000-false", Publisher: "web-api.crontab.clearPool", PublisherInstance: "server-1:123", PublishedAt: 1793265000}
	raw, err := NewResetPoolMessage(input)
	if err != nil {
		t.Fatal(err)
	}
	var envelope Msg
	if err = json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Event != "resetPool" {
		t.Fatal(envelope)
	}
	var output ResetPool
	if err = json.Unmarshal([]byte(envelope.Data), &output); err != nil {
		t.Fatal(err)
	}
	if output != input {
		t.Fatalf("event lost batch identity: %+v", output)
	}
}

func TestResetBatchValidationRejectsMissingOrUnstorableIdentity(t *testing.T) {
	for _, id := range []string{"", " ", "	", " batch", "batch ", strings.Repeat("x", 75)} {
		if err := ValidateResetBatchID(id); err == nil {
			t.Fatal("invalid batch accepted", id)
		}
		if raw, err := NewResetPoolMessage(ResetPool{BatchID: id}); err == nil || raw != "" {
			t.Fatal("invalid reset was serialized", id, raw, err)
		}
	}
	if err := ValidateResetBatchID(strings.Repeat("x", 74)); err != nil {
		t.Fatal(err)
	}
}
