package baseGrpc

import (
	"context"
	v2 "github.com/black1552/base-common"
	"github.com/duke-git/lancet/v2/slice"
	"githu
	"github.com/gogf/gf/contrib/registry/etcd/v2"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gsvc"
	"github.com/gogf/gf/v2/os/gctx"
	"google.golang.org/grpc"
	"time"
)

var (
	name string
	open bool
)

type SGrpc struct {
}

func New() *SGrpc {
	name = v2.GenerateString(20)
	config, _ := g.Config().Get(gctx.New(), "grpc.open", false)
	open = config.Bool()
	return &SGrpc{}
}

func (s *SGrpc) clientTimeout(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

// InitServer 初始化服务 初始化后使用需要服务的grpc可以利用其进行注册
func (s *SGrpc) InitServer() *grpcx.GrpcServer {
	s.RegisterResolver(gctx.New())
	config := grpcx.Server.NewConfig()
	config.Options = append(config.Options, []grpc.ServerOption{
		grpcx.Server.ChainUnary(
			grpcx.Server.UnaryError,
		)}...,
	)
	config.Name = s.GetServerName()
	return grpcx.Server.New(config)
}

// Client 获取对应服务CONN 需要客户端使用它进行初始化
func (s *SGrpc) Client(ctx context.Context, name string) *grpc.ClientConn {
	servers, err := s.GetServers(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "%+v", err)
		return nil
	}
	_, ok := slice.FindBy(servers, func(index int, item gsvc.Service) bool {
		return item.GetName() == name
	})
	if !ok {
		return nil
	}
	var conn = grpcx.Client.MustNewGrpcClientConn(name, grpcx.Client.ChainUnary(
		s.clientTimeout,
	))
	return conn
}

// GetServers 获取所有服务
func (s *SGrpc) GetServers(ctx context.Context, inputConfig ...gsvc.SearchInput) ([]gsvc.Service, error) {
	input := gsvc.SearchInput{}
	if len(inputConfig) > 0 {
		input = inputConfig[0]
	}
	servers, err := gsvc.GetRegistry().Search(ctx, input)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// RegisterResolver 注册服务发现
func (s *SGrpc) RegisterResolver(ctx context.Context) {
	etcdConfig, err := g.Config().Get(ctx, "etcd.host")
	if err != nil {
		panic(err)
	}
	grpcx.Resolver.Register(etcd.New(etcdConfig.String()))
}

// GetServerName 获取服务名
func (s *SGrpc) GetServerName() string {
	return name
}

func (s *SGrpc) IsOpen() bool {
	return open
}
