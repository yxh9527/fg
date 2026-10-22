package common

import (
	"app/entity/view"
	"testing"
)

func TestPutAnalysisItemWritesWebId(t *testing.T) {
	result := map[string]*view.DataAnalysisItem{}
	putAnalysisItem(result, &view.DataAnalysisItem{WebId: 0, AgentId: 0, GameId: 7012, EffectiveBetsTotal: 1})
	putAnalysisItem(result, &view.DataAnalysisItem{WebId: 23, AgentId: 0, GameId: 7012, EffectiveBetsTotal: 2})
	item := result["0-7012"]
	if item == nil || item.WebId != 23 || item.EffectiveBetsTotal != 3 {
		t.Fatalf("settlement webId must be written onto the analysis document: %#v", item)
	}
}
