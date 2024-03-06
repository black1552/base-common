package v2

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/encoding/gyaml"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/sirupsen/logrus"
	"os"
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

type Logs struct {
	logs *glog.Logger
	ctx  context.Context
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
	//r.Response.CORSDefault()
	r.Middleware.Next()
	if err := r.GetError(); err != nil {
		bo := gstr.Contains(err.Error(), ": ")
		msg := ""
		if bo {
			msg = gstr.SubStrFromEx(err.Error(), ": ")
		} else {
			msg = err.Error()
		}
		r.Response.ClearBuffer()
		json := Json{
			Code: 0,
			Msg:  msg,
		}
		r.Response.Status = 200
		r.Response.WriteJson(json)
	}
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

func AuthAdmin(r *ghttp.Request) {
	AuthBase(r, "admin")
}
func AuthIndex(r *ghttp.Request) {
	AuthBase(r, "user")
}

func NoLogin(r *ghttp.Request) {
	r.Response.Status = 401
	r.Response.WriteJsonExit(Json{
		Code: 401,
		Data: nil,
		Msg:  "请登录后操作",
	})
}

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

const BaseConfig = `{
"server":{
	"default":{
		"address":"127.0.0.1:8080",
		"logPath":"./log/",
		"logStdout":true,
		"errorStack":true,
		"errorLogEnabled":true,
		"errorLogPattern":"error-{Ymd}.log",
		"accessLogEnable":true,
		"accessLogPattern":"access-{Ymd}.log"
	}
},
"database":{
	"default":{
		"host":"127.0.0.1",
		"port":"3306",
		"user":"",
		"pass":"",
		"name":"",
		"type":"mysql",
		"debug":false,
		"charset":"utf8mb4",
		"createdAt":"create_time",
		"updatedAt":"update_time"
	}
},
"skipUrl":"/dist/index.html",
"openAPITitle": "",
"openAPIDescription": "Api列表 包含各端接口信息 字段注释 枚举说明",
"openAPIUrl": "https://panel.magicany.cc:8888/btpanel",
"openAPIName": "",
"openAPIVersion":"v1.0",
"logger":{
	"path":"./log/",
	"file":"{Y-m-d}.log",
	"level":"all",
	"timeFormat":"2006-01-02 15:04:05",
	"ctxKeys":[],
	"header":true,
	"stdout":true,
	"stdoutColorDisabled":false,
	"writerColorEnable":true
},
"gfcli":{
	"build":{
		"name":"checkRisk",
		"arch":"amd64",
		"system":"linux",
		"mod":"none",
		"packSrc":"manifest",
		"packDst":"internal/packed/packed.go",
		"version":"v1.0.0001",
		"output":"./bin"
	},
	"gen":{
		"dao":{
			"- link":"mysql:root:123456@tcp(127.0.0.1:3306)/check_risk",
			"jsonCase":"CamelLower"
		}
	}
}
}`

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

var ConfigPath = gfile.Pwd() + "/manifest/config/config.yaml"
var log *glog.Logger

func ConfigInit() {
	json, err := gjson.Decode(BaseConfig)
	if err != nil {
		g.Log().Error(gctx.New(), "配置模板解析失败", err)
	}
	yaml, err := gyaml.Encode(json)
	if err != nil {
		g.Log().Error(gctx.New(), "转换yaml失败", err)
	}

	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infoln("正在检查配置文件", gfile.IsFile(ConfigPath))
	if !gfile.IsFile(ConfigPath) {
		logrus.Infoln("正在创建配置文件", ConfigPath)
		_, _ = gfile.Create(ConfigPath)
		logrus.Infoln("正在写入配置文件", ConfigPath)
		_ = gfile.PutContents(ConfigPath, gconv.String(yaml))
		logrus.Infoln("配置文件创建成功！")
	} else {
		gcfg.Instance().GetAdapter().(*gcfg.AdapterFile).SetFileName(ConfigPath)
		log = g.Log()
	}
}

func LogDeBug(content ...interface{}) {
	log.Debug(gctx.New(), content)
}
func LogInfo(content ...interface{}) {
	log.Info(gctx.New(), content)
}
func LogError(content ...interface{}) {
	log.Error(gctx.New(), content)
}

func CreateDB(ctx context.Context, sqlHost, sqlPort, sqlRoot, sqlPass, baseName string, debug bool) {
	cfg := gcfg.Instance()
	for {
		cfgBase, _ := cfg.Get(ctx, "database")
		if cfgBase == nil {
			gdb.SetConfig(gdb.Config{
				"default": gdb.ConfigGroup{
					gdb.ConfigNode{
						Host:      sqlHost,
						Port:      sqlPort,
						User:      sqlRoot,
						Pass:      sqlPass,
						Name:      baseName,
						Timezone:  "Local",
						Type:      "mysql",
						Role:      "master",
						CreatedAt: "create_time",
						UpdatedAt: "update_time",
						Debug:     debug,
					},
				}})
		}
		time.Sleep(time.Minute * 10)
	}
}

func Start(agent string, maxSessionTime time.Duration, isApi bool, maxBody ...int64) *ghttp.Server {
	//var s *ghttp.Server
	s := g.Server()
	s.SetDumpRouterMap(false)
	path := gfile.Pwd() + "/resource/public/upload"
	if !gfile.IsDir(path) {
		_ = os.Mkdir(path, os.ModePerm)
		_ = os.Mkdir(gfile.Pwd()+"/resource/template", os.ModePerm)
		_ = os.Mkdir(gfile.Pwd()+"/resource/scripts", os.ModePerm)
		_ = os.Mkdir(gfile.Pwd()+"/resource/public/html", os.ModePerm)
		_ = os.Mkdir(gfile.Pwd()+"/resource/public/resource/css", os.ModePerm)
		_ = os.Mkdir(gfile.Pwd()+"/resource/public/resource/image", os.ModePerm)
		_ = os.Mkdir(gfile.Pwd()+"/resource/public/resource/js", os.ModePerm)
	}
	s.SetServerRoot(gfile.Pwd() + "/resource")
	s.AddSearchPath(path)
	s.AddStaticPath("/upload", path)
	err := s.SetLogPath(gfile.Pwd() + "/resource/log")
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
		"sessionPath": gfile.Pwd() + "/resource/session",
		"serverAgent": agent,
	})
	if err != nil {
		fmt.Println(err)
	}
	s.Use(MiddlewareError)
	skipUrl, _ := g.Cfg().Get(gctx.New(), "skipUrl")
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.RedirectTo(gconv.String(skipUrl))
	})
	enhanceOpenAPIDoc(s)
	return s
}
