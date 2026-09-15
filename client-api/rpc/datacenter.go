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

type DataCenterClient struct {
	mu   sync.Mutex
	addr string
	conn *grpc.ClientConn
	cli  services.DataCenterServiceClient
}

func NewDataCenterClient(addr string) *DataCenterClient {
	return &DataCenterClient{addr: addr}
}

func (c *DataCenterClient) ensure() (services.DataCenterServiceClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cli != nil {
		return c.cli, nil
	}
	if c.addr == "" {
		return nil, fmt.Errorf("datacenter_grpc empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		zap.L().Error("dial data-center failed", zap.String("addr", c.addr), zap.Error(err))
		return nil, err
	}
	c.conn = conn
	c.cli = services.NewDataCenterServiceClient(conn)
	return c.cli, nil
}

func (c *DataCenterClient) Authenticate(ctx context.Context, token string, gameId, entryUserId uint32) (*services.AuthenticateResp, error) {
	cli, err := c.ensure()
	if err != nil {
		return nil, err
	}
	return cli.Authenticate(ctx, &services.AuthenticateReq{
		Token:       token,
		GameId:      gameId,
		EntryUserId: entryUserId,
	})
}

func (c *DataCenterClient) ValidateToken(ctx context.Context, token string, userId uint32) (*services.ValidateTokenResp, error) {
	cli, err := c.ensure()
	if err != nil {
		return nil, err
	}
	return cli.ValidateToken(ctx, &services.ValidateTokenReq{
		Token:  token,
		UserId: userId,
	})
}

func (c *DataCenterClient) GetLoginData(ctx context.Context, userId uint32) (*services.GetLoginDataResp, error) {
	cli, err := c.ensure()
	if err != nil {
		return nil, err
	}
	return cli.GetLoginData(ctx, &services.GetLoginDataReq{UserId: userId})
}

func (c *DataCenterClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
		c.cli = nil
	}
}

// Ping 启动自检：尝试建立连接。
func (c *DataCenterClient) Ping(ctx context.Context) error {
	cli, err := c.ensure()
	if err != nil {
		return err
	}
	// 用一个轻量非法参数调用，确认通道可用；业务失败也说明 RPC 通了。
	_, err = cli.ValidateToken(ctx, &services.ValidateTokenReq{Token: "__ping__", UserId: 1})
	if err != nil {
		return err
	}
	return nil
}
