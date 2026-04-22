package view

import "github.com/shopspring/decimal"

type Bill struct {
	AgentId        int64           `json:"agentId"`
	UserId         int64           `json:"userId"`
	GameId         int             `json:"gameId"`
	Symbol         string          `json:"symbol"`
	OfficeNumber   string          `json:"roundId"`
	Bets           decimal.Decimal `json:"bet"`
	UserScore      decimal.Decimal `json:"currentScore"`
	FlowingWaterOn string          `json:"flowingWaterOn"`
	CreatTime      int64           `json:"createTime"`
	CurrencyType   string          `json:"currency"`
	GameName       string          `json:"gameName"`
	Msg            string          `json:"desc"`
}
