package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"client-api/cache"
	"client-api/common"
	"client-api/rpc"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"micro_service/services"
)

type BalanceHandler struct {
	Lottery *rpc.LotteryClient
	Guard   *cache.Guard
}

type balanceCache struct {
	Currency     string `json:"currency"`
	CurrencyCent int64  `json:"currencyCent"`
}

// GetBalance 权威余额（分）；供 CommonRpcProd / 记录页使用。
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userId64, _ := strconv.ParseUint(c.Query("userId"), 10, 32)
	userId := uint32(userId64)
	if userId == 0 {
		common.Fail(c, http.StatusOK, common.CodeBadRequest, "参数错误")
		return
	}
	if h.Lottery == nil {
		common.Fail(c, http.StatusOK, common.CodeSystemError, "余额服务未配置")
		return
	}

	payload := &balanceCache{}
	key := cache.BuildKey("balance", fmt.Sprintf("%d", userId))
	err := h.Guard.Do(key, payload, func() (interface{}, error) {
		resp, err := h.Lottery.GetBalance(c.Request.Context(), userId)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, fmt.Errorf("empty balance resp")
		}
		if resp.Code != services.ErrorCode_OK {
			return nil, fmt.Errorf("balance code=%v", resp.Code)
		}
		return &balanceCache{
			Currency:     resp.Currency,
			CurrencyCent: resp.CurrencyCent,
		}, nil
	})
	if err != nil {
		zap.L().Error("GetBalance failed", zap.Uint32("userId", userId), zap.Error(err))
		common.Fail(c, http.StatusOK, common.CodeSystemError, "余额查询失败")
		return
	}
	common.OK(c, gin.H{
		"userId":       userId,
		"currency":     payload.Currency,
		"currencyCent": payload.CurrencyCent,
	})
}
