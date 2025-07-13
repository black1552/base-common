package token

import (
	"context"

	"github.com/goflyfox/gtoken/gtoken"
	"github.com/gogf/gf/v2/os/gctx"
)

type sToken struct {
	ctx           context.Context
	auth          gtoken.GfToken
	GenerateToken func(ctx context.Context, autoPath, skipPath []string)
}

func init() {
	s := &sToken{}
	s.ctx = gctx.New()
	s.auth = gtoken.GfToken{
		CacheMode:  1,
		CacheKey:   "server",
		Timeout:    6 * 24 * 60 * 60 * 1000,
		MaxRefresh: 5 * 24 * 60 * 60 * 1000,
	}
}
