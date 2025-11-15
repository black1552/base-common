package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gogf/gf/v2/os/glog"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

type Json struct {
	Code int    `json:"code" d:"1"`
	Data any    `json:"data"`
	Msg  string `json:"msg" d:"操作成功"`
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
	from.Header.Set("Access-Control-Expose-Headers", "Set-Cookie")
	from.Response.Status = 200
	from.Response.WriteJson(a.json)
	return
}

func (a *ApiRes) FileDownload(path, name string) {
	from := g.RequestFromCtx(a.ctx)
	from.Response.ServeFileDownload(path)
	return
}

func (a *ApiRes) FileSelect(path string) {
	from := g.RequestFromCtx(a.ctx)
	from.Response.ServeFile(path)
	return
}

// LoginJson 返回登录json数据
/*
 * @param ctx 上下文
 * @param msg 返回信息
 * @param data 返回数据
 */
func LoginJson(r *ghttp.Request, msg string, data ...interface{}) {
	var info interface{}
	if len(data) > 0 {
		info = data[0]
	} else {
		info = nil
	}
	r.Response.WriteJsonExit(Json{
		Code: 1,
		Data: info,
		Msg:  msg,
	})
}

// ResponseJson 返回json数据
/*
 * @param ctx 上下文
 * @param data 返回数据
 */
func ResponseJson(ctx context.Context, data interface{}) {
	g.RequestFromCtx(ctx).Response.WriteJson(data)
	return
}

type PageSize struct {
	CurrentPage int         `json:"currentPage"`
	Data        interface{} `json:"data"`
	LastPage    int         `json:"lastPage"`
	PerPage     int         `json:"per_page"`
	Total       int         `json:"total"`
}

// SetPage 设置分页
/*
 * @param page 当前页
 * @param limit 每页显示条数
 * @param total 总条数
 * @param data 返回数据
 * @return PageSize
 */
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
	// r.Response.CORSDefault()
	r.Session.RegenerateId(true)
	r.Middleware.Next()
	var (
		msg    string
		err    = r.GetError()
		res    = r.GetHandlerResponse()
		status = r.Response.Status
	)
	json := new(Json)
	json.Data = res
	if err == nil {
		json.Code = 1
		if r.Response.BufferLength() > 0 {
			glog.Infof(r.Context(), "Buffer:%s", r.Response.BufferString())
			if gjson.Valid(r.Response.Buffer()) {
				js, _ := gjson.DecodeToJson(r.Response.Buffer())
				json.Data = js
			} else {
				msg = r.Response.BufferWriter.BufferString()
			}
			r.Response.BufferWriter.ClearBuffer()
		} else {
			msg = "操作成功"
		}
	} else {
		bo := gstr.Contains(err.Error(), ": ")
		if bo {
			msg = gstr.SubStrFromEx(err.Error(), ": ")
		} else {
			msg = err.Error()
		}
	}
	if err := r.GetError(); err != nil {
		bo := gstr.Contains(err.Error(), ": ")
		msg := ""
		if bo {
			msg = gstr.SubStrFromEx(err.Error(), ": ")
		} else {
			msg = err.Error()
		}
		r.Response.ClearBuffer()
		json.Code = 0
		json.Msg = msg
		r.Response.Status = http.StatusInternalServerError
	} else {
		json.Msg = msg
		if status == 401 {
			json.Code = 0
			json.Msg = "请登录后操作"
		}
	}

	r.Response.WriteJson(json)
}

// AuthBase 鉴权中间件，只有前端或者后端登录成功之后才能通过
func AuthBase(r *ghttp.Request, name string) {
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

// AuthAdmin 鉴权中间件，只有后端登录成功之后才能通过
func AuthAdmin(r *ghttp.Request) {
	AuthBase(r, "admin")
}

// AuthIndex 鉴权中间件，只有前端登录成功之后才能通过
func AuthIndex(r *ghttp.Request) {
	AuthBase(r, "user")
}

// NoLogin 未登录返回
func NoLogin(r *ghttp.Request) {
	r.Response.Status = 401
	r.Response.WriteJsonExit(Json{
		Code: 401,
		Data: nil,
		Msg:  "请登录后操作",
	})
}

// CreateFileDir 创建文件目录
func CreateFileDir() error {
	path := gfile.Pwd() + "/resource"
	if !gfile.IsDir(path) {
		if err := gfile.Mkdir(path); err != nil {
			return err
		}
		if err := gfile.Mkdir(path + "/public/upload"); err != nil {
			return err
		}
	}
	return nil
}

func AuthLoginSession(ctx context.Context, sessionKey string) {
	ti, err := g.RequestFromCtx(ctx).Session.Get(sessionKey+"LoginTime", "")
	if err != nil {
		panic(err.Error())
	}
	if !ti.IsEmpty() {
		now := gtime.Now().Timestamp()
		if now-gconv.Int64(ti) <= 300 {
			number, err := g.RequestFromCtx(ctx).Session.Get(sessionKey+"LoginNum", 0)
			if err != nil {
				panic(err.Error())
			}
			if !number.IsEmpty() {
				count := gconv.Int(number)
				if count == 3 {
					panic("请等待5分钟后再次尝试或修改后尝试登录")
				}
			}
		}
	}
}

func LoginCountSession(ctx context.Context, sessionKey string) {
	ti, err := g.RequestFromCtx(ctx).Session.Get(sessionKey+"LoginTime", "")
	if err != nil {
		panic(err.Error())
	}
	if ti.IsEmpty() {
		_ = g.RequestFromCtx(ctx).Session.Set(sessionKey+"LoginTime", gtime.Now().Timestamp())
	}
	now := gtime.Now().Timestamp()
	if now-gconv.Int64(ti) <= 300 {
		number, err := g.RequestFromCtx(ctx).Session.Get(sessionKey+"LoginNum", 0)
		if err != nil {
			panic(err.Error())
		}
		if number.IsEmpty() {
			_ = g.RequestFromCtx(ctx).Session.Set(sessionKey+"LoginNum", 1)
		} else {
			count := gconv.Int(number)
			if count == 3 {
				panic("尝试登录已超过限制，请等待5分钟后再次尝试或修改后尝试登录")
			}
			_ = g.RequestFromCtx(ctx).Session.Set(sessionKey+"LoginNum", count+1)
		}
	} else {
		_ = g.RequestFromCtx(ctx).Session.Set(sessionKey+"LoginTime", gtime.Now().Timestamp())
		_ = g.RequestFromCtx(ctx).Session.Set(sessionKey+"LoginNum", 1)
	}
}

func enhanceOpenAPIDoc(s *ghttp.Server) {
	cfg := gcfg.Instance()
	apiTitle, err := cfg.Get(gctx.New(), "openAPITitle", "Api列表")
	if err != nil {
		panic(err)
	}
	apiDes, err := cfg.Get(gctx.New(), "openAPIDescription", "Api列表 包含各端接口信息 字段注释 枚举说明")
	if err != nil {
		panic(err)
	}
	apiUrl, err := cfg.Get(gctx.New(), "openAPIUrl", "https://panel.magicany.cc:8888/btpanel")
	if err != nil {
		panic(err)
	}
	apiName, err := cfg.Get(gctx.New(), "openAPIName", "Api列表")
	if err != nil {
		panic(err)
	}
	version, err := cfg.Get(gctx.New(), "openAPIVersion", "Api列表")
	if err != nil {
		panic(err)
	}
	openapi := s.GetOpenApi()
	openapi.Config.CommonResponse = ghttp.DefaultHandlerResponse{}
	openapi.Config.CommonResponseDataField = `Data`

	// API description.
	openapi.Info = goai.Info{
		Title:       gconv.String(apiTitle),
		Description: gconv.String(apiDes),
		Contact: &goai.Contact{
			Name: gconv.String(apiName),
			URL:  gconv.String(apiUrl),
		},
		License: &goai.License{
			Name: "马国栋",
			URL:  "https://panel.magicany.cc:8888/btpanel",
		},
		Version: gconv.String(version),
	}
}

var ConfigPath = filepath.Join(gfile.Pwd(), "manifest", "config", "config.yaml")
var uploadPath = filepath.Join(gfile.Pwd(), "resource")

// Start 启动服务
/*
 * @param agent string 浏览器标识
 * @param maxSessionTime time.Duration session最大时间
 * @param isApi bool 是否开启api
 * @param maxBody ...int64 最大上传文件大小 默认200M
 * @return *ghttp.Server 服务实例
 */
func Start(agent string, maxSessionTime time.Duration, isApi bool, maxBody ...int64) *ghttp.Server {
	// var s *ghttp.Server
	s := g.Server()
	s.SetDumpRouterMap(false)
	s.AddStaticPath(fmt.Sprintf("%vstatic", gfile.Separator), uploadPath)
	err := s.SetLogPath(gfile.Join(gfile.Pwd(), "resource", "log"))
	if err != nil {
		fmt.Println(err)
	}
	s.SetLogLevel("all")
	s.SetLogStdout(false)
	if len(maxBody) > 0 {
		s.SetClientMaxBodySize(maxBody[0])
	} else {
		s.SetClientMaxBodySize(200 * 1024 * 1024)
	}
	s.SetFormParsingMemory(50 * 1024 * 1024)
	if isApi {
		s.SetOpenApiPath("/api.json")
		s.SetSwaggerPath("/swagger")
	}
	s.SetMaxHeaderBytes(1024 * 20)
	s.SetErrorStack(true)
	s.SetSessionIdName("zrSession")
	s.SetAccessLogEnabled(true)
	s.SetSessionMaxAge(maxSessionTime)
	err = s.SetConfigWithMap(g.Map{
		"sessionPath": gfile.Join(gfile.Pwd(), "resource", "session"),
		"serverAgent": agent,
	})
	if err != nil {
		fmt.Println(err)
	}
	s.Use(MiddlewareError)
	enhanceOpenAPIDoc(s)
	return s
}

// SetConfigAndRun 设置配置并运行服务
// @param s *ghttp.Server 服务实例
// @param address string 监听地址
func SetConfigAndRun(s *ghttp.Server, address string) {
	logCfg := glog.Config{
		File:              "{Y-m-d}.log",
		Path:              gfile.Join(gfile.Pwd(), "log"),
		RotateBackupLimit: 10,
		RotateSize:        1024 * 1024 * 2,
		StdoutPrint:       true,
		TimeFormat:        "2006-01-02 15:04:05",
		WriterColorEnable: true,
		Level:             glog.LEVEL_ALL,
	}
	_ = glog.SetConfig(logCfg)
	log := glog.New()
	logCfg.Level = glog.LEVEL_ERRO
	logCfg.StdoutPrint = false
	logCfg.Path = gfile.Join(gfile.Pwd(), "resource", "log")
	_ = log.SetConfig(logCfg)
	s.SetAccessLogEnabled(false)
	s.SetErrorLogEnabled(true)
	_ = s.SetConfig(ghttp.ServerConfig{
		ErrorLogPattern: "error-{Ymd}.log",
	})
	s.SetLogger(log)
	s.SetAddr(address)
	s.SetFileServerEnabled(true)
	s.SetCookieDomain(fmt.Sprintf("http://%s", address))
	s.Run()
}

func CORSMiddleware(r *ghttp.Request) {
	corsOptions := r.Response.DefaultCORSOptions()
	cfg, _ := gcfg.Instance().Get(r.Context(), "doMain", nil)
	if !cfg.IsNil() {
		corsOptions.AllowDomain = cfg.Strings()
	}
	r.Response.CORS(corsOptions)
	r.Middleware.Next()
}
