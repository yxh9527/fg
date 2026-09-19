package view

import "github.com/shopspring/decimal"

type Bill struct {
	AgentId        int64           `json:"agentId"`
	UserId         int64           `json:"userId"`
	GameId         int             `json:"gameId"`
	Symbol         string          `json:"symbol"`
	OfficeNumber   string          `json:"roundId"`
	Bets           decimal.Decimal `json:"bet"`          // 下注（负）
	Award          decimal.Decimal `json:"award"`        // 返奖/到账（>=0）
	UserScore      decimal.Decimal `json:"currentScore"` // 账变后余额
	FlowingWaterOn string          `json:"flowingWaterOn"`
	CreatTime      int64           `json:"createTime"`
	CurrencyType   string          `json:"currency"`
	GameName       string          `json:"gameName"`
	Msg            string          `json:"desc"`
}
