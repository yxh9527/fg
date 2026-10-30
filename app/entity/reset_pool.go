package entity

import (
	"encoding/json"
	"errors"
	"strings"
)

// Reset batch IDs become request IDs (reset:<batch>) in an 80-byte SQL column.
func ValidateResetBatchID(batchID string) error {
	if batchID == "" || len(batchID) > 74 || strings.TrimSpace(batchID) != batchID {
		return errors.New("水池重置批次编号无效，必须是1至74字节且无首尾空白")
	}
	return nil
}

// NewResetPoolMessage keeps the batch identity in the nested event payload.
func NewResetPoolMessage(reset ResetPool) (string, error) {
	if err := ValidateResetBatchID(reset.BatchID); err != nil {
		return "", err
	}
	data, err := json.Marshal(reset)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(Msg{Event: "resetPool", Data: string(data)})
	return string(raw), err
}
