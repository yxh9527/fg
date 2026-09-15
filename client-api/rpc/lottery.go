package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"micro_service/services"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LotteryClient 仅用于客户端余额展示，不承接游戏结算。
type LotteryClient struct {
	mu   sync.Mutex
	addr string
	conn *grpc.ClientConn
	cli  services.LotteryServiceClient
}

func NewLotteryClient(addr string) *LotteryClient {
	return &LotteryClient{addr: addr}
}

func (c *LotteryClient) ensure() (services.LotteryServiceClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cli != nil {
		return c.cli, nil
	}
	if c.addr == "" {
		return nil, fmt.Errorf("lottery_grpc empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		zap.L().Error("dial lottery failed", zap.String("addr", c.addr), zap.Error(err))
		return nil, err
	}
	c.conn = conn
	c.cli = services.NewLotteryServiceClient(conn)
	return c.cli, nil
}

func (c *LotteryClient) GetBalance(ctx context.Context, userId uint32) (*services.GetBalanceResp, error) {
	cli, err := c.ensure()
	if err != nil {
		return nil, err
	}
	return cli.GetBalance(ctx, &services.GetBalanceReq{UserId: userId})
}

func (c *LotteryClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
		c.cli = nil
	}
}

func (c *LotteryClient) Ping(ctx context.Context) error {
	if c.addr == "" {
		return fmt.Errorf("lottery_grpc empty")
	}
	_, err := c.ensure()
	return err
}
