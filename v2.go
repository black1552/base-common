package v2

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gsession"
	"github.com/gogf/gf/v2/text/gstr"
	"time"
)

type Json struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type ApiRes struct {
	ctx  context.Context
	json *Json
}

func Success(ctx context.Context) *ApiRes {
	json := Json{
		Code: 1,
	}

	var a = ApiRes{
		ctx:  ctx,
		json: &json,
	}
	return &a
}

func Error(ctx context.Context) *ApiRes {
	json := Json{
		Code: 0,
	}

	var a = ApiRes{
		ctx:  ctx,
		json: &json,
	}
	return &a
}

func (a *ApiRes) SetCode(code int) *ApiRes {
	a.json.Code = code
	return a
}

func (a *ApiRes) SetData(data interface{}) *ApiRes {
	a.json.Data = data
	return a
}

func (a *ApiRes) SetMsg(msg string) *ApiRes {
	a.json.Msg = msg
	return a
}

func (a *ApiRes) End() {
	from := g.RequestFromCtx(a.ctx)
	from.Response.Status = 200
	err := from.Response.WriteJson(a.json)
	if err != nil {
		panic(err.Error())
	}
	return
}

func ResponseJson(ctx context.Context, data interface{}) {
	err := g.RequestFromCtx(ctx).Response.WriteJson(data)
	if err != nil {
		panic(err.Error())
	}
	return
}

type PageSize struct {
	CurrentPage int         `json:"current_page"`
	Data        interface{} `json:"data"`
	LastPage    int         `json:"last_page"`
	PerPage     int         `json:"per_page"`
	Total       int         `json:"total"`
}

// SetPage 设置分页
func SetPage(page, limit, total int, data interface{}) *PageSize {
	var size = new(PageSize)
	if page == 1 {
		size.LastPage = 1
	} else {
		size.LastPage = page - 1
	}
	size.PerPage = limit
	size.Total = total
	size.CurrentPage = page
	size.Data = data
	return size
}

// MiddlewareError 异常处理中间件
func MiddlewareError(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
	if err := r.GetError(); err != nil {
		r.Response.ClearBuffer()
		json := Json{
			Code: 0,
			Msg:  gstr.SubStrFromEx(err.Error(), ": "),
		}
		r.Response.Status = 200
		err := r.Response.WriteJson(json)
		if err != nil {
			panic(err.Error())
		}
	}
}

// AuthBase 鉴权中间件，只有前端或者后端登录成功之后才能通过
func AuthBase(r *ghttp.Request, name string) {
	r.Response.CORSDefault()
	info, err := r.Session.Get(name, nil)
	if err != nil {
		panic(err.Error())
	}
	if !info.IsEmpty() {
		r.Middleware.Next()
	} else {
		NoLogin(r)
	}
}

func AuthAdmin(r *ghttp.Request) {
	AuthBase(r, "admin")
}
func AuthIndex(r *ghttp.Request) {
	AuthBase(r, "user")
}

func NoLogin(r *ghttp.Request) {

	r.Response.Status = 200
	_ = r.Response.WriteJsonExit(Json{
		Code: 401,
		Data: nil,
		Msg:  "请登录后操作",
	})
}

func Start(address, agent string, time time.Duration, maxBody ...int64) *ghttp.Server {
	s := g.Server()
	s.SetAddr(address)
	s.SetServerRoot(gfile.Pwd() + "/resource")
	path := gfile.Pwd() + "/resource/public/upload"
	s.AddSearchPath(path)
	s.AddStaticPath("/upload", path)
	_ = s.SetLogPath(gfile.Pwd() + "/resource/log")
	s.SetLogLevel("all")
	s.SetLogStdout(false)
	if len(maxBody) > 0 {
		s.SetClientMaxBodySize(maxBody[0])
	} else {
		s.SetClientMaxBodySize(200 * 1024 * 1024)
	}
	s.SetMaxHeaderBytes(1024 * 20)
	s.SetDumpRouterMap(true)
	s.SetErrorStack(true)
	s.SetAccessLogEnabled(true)
	s.SetSessionMaxAge(time)
	_ = s.SetConfigWithMap(g.Map{
		"sessionPath":    gfile.Pwd() + "/resource/session",
		"serverAgent":    agent,
		"SessionStorage": gsession.NewStorageMemory(),
	})
	s.Use(MiddlewareError)
	return s
}
