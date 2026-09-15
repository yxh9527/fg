package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"client-api/cache"
	"client-api/common"
	"client-api/rpc"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"micro_service/services"
)

type AuthHandler struct {
	DC    *rpc.DataCenterClient
	Guard *cache.Guard
}

type authenticateReq struct {
	Token       string `json:"token"`
	GameId      uint32 `json:"gameId"`
	EntryUserId uint32 `json:"entryUserId"`
}

type validateReq struct {
	Token  string `json:"token"`
	UserId uint32 `json:"userId"`
}

type authResultCache struct {
	Success      bool   `json:"success"`
	UserId       uint32 `json:"userId"`
	IsReEnter    bool   `json:"isReEnter"`
	AttemptCount int64  `json:"attemptCount"`
	Message      string `json:"message"`
	Unauthorized bool   `json:"unauthorized"`
}

// Authenticate Gateway/客户端登录鉴权；相同参数防重入并短缓存。
func (h *AuthHandler) Authenticate(c *gin.Context) {
	var req authenticateReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		common.Fail(c, http.StatusOK, common.CodeBadRequest, "参数错误")
		return
	}
	payload := &authResultCache{}
	key := cache.BuildKey("authAuthenticate", req.Token, fmt.Sprintf("%d", req.GameId), fmt.Sprintf("%d", req.EntryUserId))
	err := h.Guard.Do(key, payload, func() (interface{}, error) {
		resp, err := h.DC.Authenticate(c.Request.Context(), req.Token, req.GameId, req.EntryUserId)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, fmt.Errorf("empty auth resp")
		}
		if resp.Code == services.ErrorCode_SYSTEM_ERROR {
			return nil, fmt.Errorf("auth system error")
		}
		out := &authResultCache{
			Success:      resp.Success,
			UserId:       resp.UserId,
			IsReEnter:    resp.IsReEnter,
			AttemptCount: resp.AttemptCount,
			Message:      resp.Message,
		}
		if resp.Code != services.ErrorCode_OK || !resp.Success {
			out.Unauthorized = true
			out.Message = firstNonEmpty(resp.Message, "未登录")
		}
		return out, nil
	})
	if err != nil {
		zap.L().Error("Authenticate rpc failed", zap.Error(err))
		common.Fail(c, http.StatusOK, common.CodeSystemError, "鉴权服务不可用")
		return
	}
	if payload.Unauthorized || !payload.Success {
		common.Fail(c, http.StatusOK, common.CodeUnauthorized, firstNonEmpty(payload.Message, "未登录"))
		return
	}
	common.OK(c, gin.H{
		"success":      payload.Success,
		"userId":       payload.UserId,
		"isReEnter":    payload.IsReEnter,
		"attemptCount": payload.AttemptCount,
		"message":      payload.Message,
	})
}

// ValidateToken token 与 userId 绑定校验。
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req validateReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Token) == "" || req.UserId == 0 {
		common.Fail(c, http.StatusOK, common.CodeBadRequest, "参数错误")
		return
	}
	type validateCache struct {
		Valid bool `json:"valid"`
	}
	payload := &validateCache{}
	key := cache.BuildKey("authValidate", req.Token, fmt.Sprintf("%d", req.UserId))
	err := h.Guard.Do(key, payload, func() (interface{}, error) {
		resp, err := h.DC.ValidateToken(c.Request.Context(), req.Token, req.UserId)
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Code == services.ErrorCode_SYSTEM_ERROR {
			return nil, fmt.Errorf("validate system error")
		}
		return &validateCache{Valid: resp.Valid}, nil
	})
	if err != nil {
		zap.L().Error("ValidateToken rpc failed", zap.Error(err))
		common.Fail(c, http.StatusOK, common.CodeSystemError, "鉴权服务不可用")
		return
	}
	common.OK(c, gin.H{"valid": payload.Valid})
}

// GetLoginData 登录资料（余额不在此返回）。
func (h *AuthHandler) GetLoginData(c *gin.Context) {
	userId64, _ := strconv.ParseUint(c.Query("userId"), 10, 32)
	userId := uint32(userId64)
	if userId == 0 {
		common.Fail(c, http.StatusOK, common.CodeBadRequest, "参数错误")
		return
	}
	type loginCache struct {
		UserId     uint32  `json:"userId"`
		UserName   string  `json:"userName"`
		AgentId    uint32  `json:"agentId"`
		TopAgentId uint32  `json:"topAgentId"`
		UserIcon   int32   `json:"userIcon"`
		LoginType  int32   `json:"loginType"`
		Gm         int32   `json:"gm"`
		Rate       float64 `json:"rate"`
		CurrencyId uint32  `json:"currencyId"`
		Symbol     string  `json:"symbol"`
		CurrRate   float64 `json:"currencyRate"`
		NotFound   bool    `json:"notFound"`
	}
	payload := &loginCache{}
	key := cache.BuildKey("authLoginData", fmt.Sprintf("%d", userId))
	err := h.Guard.Do(key, payload, func() (interface{}, error) {
		resp, err := h.DC.GetLoginData(c.Request.Context(), userId)
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Code == services.ErrorCode_SYSTEM_ERROR {
			return nil, fmt.Errorf("login data system error")
		}
		if resp.Code != services.ErrorCode_OK || resp.Profile == nil {
			return &loginCache{NotFound: true}, nil
		}
		out := &loginCache{
			UserId:     resp.Profile.UserId,
			UserName:   resp.Profile.UserName,
			AgentId:    resp.Profile.AgentId,
			TopAgentId: resp.Profile.TopAgentId,
			UserIcon:   resp.Profile.UserIcon,
			LoginType:  resp.Profile.LoginType,
			Gm:         resp.Profile.Gm,
			Rate:       resp.Profile.Rate,
		}
		if resp.Profile.Currency != nil {
			out.CurrencyId = resp.Profile.Currency.CurrencyId
			out.Symbol = resp.Profile.Currency.Symbol
			out.CurrRate = resp.Profile.Currency.CurrencyRate
		}
		return out, nil
	})
	if err != nil {
		zap.L().Error("GetLoginData rpc failed", zap.Error(err))
		common.Fail(c, http.StatusOK, common.CodeSystemError, "资料服务不可用")
		return
	}
	if payload.NotFound {
		common.Fail(c, http.StatusOK, common.CodeBadRequest, "用户不存在")
		return
	}
	common.OK(c, gin.H{
		"userId":     payload.UserId,
		"userName":   payload.UserName,
		"agentId":    payload.AgentId,
		"topAgentId": payload.TopAgentId,
		"userIcon":   payload.UserIcon,
		"loginType":  payload.LoginType,
		"gm":         payload.Gm,
		"rate":       payload.Rate,
		"currency": gin.H{
			"currencyId":   payload.CurrencyId,
			"symbol":       payload.Symbol,
			"currencyRate": payload.CurrRate,
		},
	})
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
