package utils

import (
	"context"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// GetCapitalPass MD5化并转换为大写
func GetCapitalPass(val string) string {
	md5, err := gmd5.Encrypt(val)
	if err != nil {
		panic(err.Error())
	}
	return gstr.CaseCamel(md5)
}

// Transaction 简单封装事务操作
func Transaction(function func() error) {
	err := g.DB().Transaction(context.TODO(), func(ctx context.Context, tx gdb.TX) error {
		return function()
	})
	if err != nil {
		panic(err.Error())
	}
}

type SClient[R any] struct {
	client  *gclient.Client
	request any
	header  map[string]string
	url     string
}

func NewClient[R any](request any, url string, header map[string]string) *SClient[R] {
	s := &SClient[R]{}
	if header != nil {
		s.client = g.Client().ContentJson().SetHeaderMap(header)
	} else {
		s.client = g.Client().ContentJson()
	}
	s.header = header
	s.url = url
	s.request = request
	return s
}
func (w *SClient[R]) Post(ctx context.Context) (res *R) {
	g.Log().Infof(ctx, "请求Url:%s,请求头:%v,请求方法：%s,请求内容：%s", w.url, w.header, "post", w.request)
	resp := w.client.PostVar(ctx, w.url, w.request)
	err := gconv.Struct(resp, &res)
	if err != nil {
		g.Log().Errorf(ctx, "解析响应体异常：%s", err)
		return nil
	}
	return
}
func (w *SClient[R]) Get(ctx context.Context) (res *R) {
	g.Log().Infof(ctx, "请求Url:%s,请求头:%v,请求方法：%s,请求内容：%s", w.url, w.header, "get", w.request)
	resp := w.client.GetVar(ctx, w.url, w.request)
	err := gconv.Struct(resp, &res)
	if err != nil {
		g.Log().Errorf(ctx, "解析响应体异常：%s", err)
		return nil
	}
	return
}
func (w *SClient[R]) Put(ctx context.Context) (res *R) {
	g.Log().Infof(ctx, "请求Url:%s,请求头:%v,请求方法：%s,请求内容：%s", w.url, w.header, "put", w.request)
	resp := w.client.PutVar(ctx, w.url, w.request)
	err := gconv.Struct(resp, &res)
	if err != nil {
		g.Log().Errorf(ctx, "解析响应体异常：%s", err)
		return nil
	}
	return
}
func (w *SClient[R]) Delete(ctx context.Context) (res *R) {
	g.Log().Infof(ctx, "请求Url:%s,请求头:%v,请求方法：%s,请求内容：%s", w.url, w.header, "delete", w.request)
	resp := w.client.DeleteVar(ctx, w.url, w.request)
	err := gconv.Struct(resp, &res)
	if err != nil {
		g.Log().Errorf(ctx, "解析响应体异常：%s", err)
		return nil
	}
	return
}
func (w *SClient[R]) Patch(ctx context.Context) (res *R) {
	g.Log().Infof(ctx, "请求Url:%s,请求头:%v,请求方法：%s,请求内容：%s", w.url, w.header, "patch", w.request)
	resp := w.client.PatchVar(ctx, w.url, w.request)
	err := gconv.Struct(resp, &res)
	if err != nil {
		g.Log().Errorf(ctx, "解析响应体异常：%s", err)
		return nil
	}
	return
}
