package rpc

import "micro_service/services"

// 以下为门面内部复用的本地请求结构（不再出现在 gRPC 协议中）。

type slotsLotteryReq struct {
	PlayerId      uint32
	CurrencyType  string
	AgentId       int64
	GameId        uint32
	ProfitLoss    string
	Bet           string
	State         string
	RoundID       string
	MaxProfitLoss string
	Complete      bool
	Account       string
}

type slotsLotteryResp struct {
	NewCurrency string
	Result      bool
	Pay         string
	Code        services.ErrorCode
}

type singleBetReq struct {
	UserId       uint32
	GameId       uint32
	Win          string
	RoundID      string
	Result       string
	Complete     bool
	Bet          string
	AgentId      uint32
	CurrencyType string
}

type singleBetResp struct {
	Code     services.ErrorCode
	Currency string
}

type multiBetReq struct {
	UserId       uint32
	GameId       uint32
	RoundID      string
	AreaId       uint32
	InitBet      string
	AgentId      uint32
	CurrencyType string
}

type multiBetResp struct {
	Code     services.ErrorCode
	Currency string
}

type multiRefundReq struct {
	UserId       uint32
	GameId       uint32
	Bet          string
	RoundID      string
	AgentId      uint32
	CurrencyType string
}

type multiRefundResp struct {
	Code     services.ErrorCode
	Currency string
}

type settleRecord struct {
	UserId       uint32
	GameId       uint32
	Win          string
	RoundID      string
	Log          string
	Bet          string
	CurrencyType string
	AgentId      uint32
	Account      string
}

type multiSettleReq struct {
	Records  []*settleRecord
	TotalWin string
}

type newCurrencyItem struct {
	UserId   uint32
	Currency string
}

type multiSettleResp struct {
	Code      services.ErrorCode
	Currencys []*newCurrencyItem
}
