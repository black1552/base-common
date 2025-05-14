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
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
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
		r.Response.Status = http.StatusInternalServerError
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
		"accessLogPattern":"access-{Ymd}.log",
        "fileServerEnabled": true
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
"skipUrl":"/dist",
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
 	"rotateSize":"2M",
  	"rotateBackupLimit":50
	"stdoutColorDisabled":false,
	"writerColorEnable":true
}
}`

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

var ConfigPath = gfile.Pwd() + "/manifest/config/config.yaml"
var uploadPath = fmt.Sprintf("%s%vresource", gfile.Pwd(), gfile.Separator)

// ConfigInit 初始化配置文件
func ConfigInit() {
	json, err := gjson.Decode(BaseConfig)
	if err != nil {
		g.Log().Error(gctx.New(), "配置模板解析失败", err)
	}
	yaml, err := gyaml.Encode(json)
	if err != nil {
		g.Log().Error(gctx.New(), "转换yaml失败", err)
	}

	if !gfile.IsDir(uploadPath) {
		_ = gfile.Mkdir(uploadPath)
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vtemplate", gfile.Pwd(), gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vscripts", gfile.Pwd(), gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vhtml", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vresource%vcss", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vresource%vimage", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vresource%vjs", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator, gfile.Separator))
	}
	g.Log().Info(gctx.New(), "正在检查配置文件", gfile.IsFile(ConfigPath))
	if !gfile.IsFile(ConfigPath) {
		g.Log().Info(gctx.New(), "正在创建配置文件", ConfigPath)
		_, _ = gfile.Create(ConfigPath)
		g.Log().Info(gctx.New(), "正在写入配置文件", ConfigPath)
		_ = gfile.PutContents(ConfigPath, gconv.String(yaml))
		g.Log().Info(gctx.New(), "配置文件创建成功！")
	} else {
		gcfg.Instance().GetAdapter().(*gcfg.AdapterFile).SetFileName(ConfigPath)
	}
}

// CreateDB 创建数据库配置
/*
 * @param ctx context.Context
 * @param sqlHost string 数据库地址
 * @param sqlPort string 数据库端口
 * @param sqlRoot string 数据库用户名
 * @param sqlPass string 数据库密码
 * @param baseName string 数据库名
 * @param debug bool 是否开启调试模式
 */
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
	err := s.SetLogPath(fmt.Sprintf("%s%vresource%vlog", gfile.Pwd(), gfile.Separator, gfile.Separator))
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
		"sessionPath": fmt.Sprintf("%s%vresource%vsession", gfile.Pwd(), gfile.Separator, gfile.Separator),
		"serverAgent": agent,
	})
	if err != nil {
		fmt.Println(err)
	}
	s.Use(MiddlewareError)
	skipUrl, _ := g.Cfg().Get(gctx.New(), "skipUrl", "")
	if gconv.String(skipUrl) != "" {
		isFile := gfile.IsFile(gfile.Pwd() + gfile.Separator + "resource" + gfile.Separator + gconv.String(skipUrl))
		if isFile {
			s.AddStaticPath(gconv.String(skipUrl), gfile.Pwd()+gfile.Separator+"resource"+gfile.Separator+gconv.String(skipUrl))
			s.BindHandler("/", func(r *ghttp.Request) {
				r.Response.RedirectTo(gconv.String(skipUrl) + "/index.html")
			})
		}
	} else {
		isFile := gfile.IsFile(gfile.Pwd() + gfile.Separator + "resource" + gfile.Separator + "dist/index.html")
		if isFile {
			s.AddStaticPath("/dist/index.html", gfile.Pwd()+gfile.Separator+"resource"+gfile.Separator+"dist/index.html")
			s.BindHandler("/", func(r *ghttp.Request) {
				r.Response.RedirectTo("/dist/index.html")
			})
		}
	}
	enhanceOpenAPIDoc(s)
	return s
}
