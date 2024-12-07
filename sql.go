package v2

import (
	"context"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
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

// PostResult 建立POST请求并返回结果
func PostResult(ctx context.Context, url string, data g.Map, header string, class string) (string, error) {
	if url == "" {
		return "", gerror.New("请求地址不可为空")
	}
	client := g.Client()
	if header != "" {
		client = client.HeaderRaw(header)
	}
	switch class {
	case "json":
		client = client.ContentJson()
	case "xml":
		client = client.ContentXml()
	default:
	}
	result, err := client.Post(ctx, url, data)
	if err != nil {
		return "", err
	}
	return result.ReadAllString(), nil
}

func GetResult(ctx context.Context, url string, data g.Map) (string, error) {
	client := g.Client()
	if url == "" {
		return "", gerror.New("请求地址不可为空")
	}
	resutl, err := client.Get(ctx, url, data)
	if err != nil {
		return "", err
	}
	return resutl.ReadAllString(), nil
}
