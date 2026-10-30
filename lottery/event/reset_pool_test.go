package event

import (
	"app/entity"
	"reflect"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// Exercise the producer encoder and the real reflection/jsoniter subscriber path.
func TestPublishedResetBatchSurvivesEventDispatch(t *testing.T) {
	input := entity.ResetPool{BatchID: "1793265000-false", Publisher: "web-api.crontab.clearPool", PublisherInstance: "web-1:42", PublishedAt: 1793265000}
	raw, err := entity.NewResetPoolMessage(input)
	if err != nil {
		t.Fatal(err)
	}
	var message entity.Msg
	if err = jsoniter.UnmarshalFromString(raw, &message); err != nil {
		t.Fatal(err)
	}
	received := make(chan *entity.ResetPool, 1)
	manager := EventMgr{Events: map[string]*Event{}}
	manager.Register("resetPool", reflect.TypeOf(entity.ResetPool{}), func(data interface{}) { received <- data.(*entity.ResetPool) })
	manager.OnEvent(&message)
	select {
	case result := <-received:
		if *result != input {
			t.Fatalf("batch or publisher lost in subscriber: %+v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("reset notification was not dispatched")
	}
}
