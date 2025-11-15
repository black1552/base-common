package server

import (
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gtoml"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
)

type Config struct {
	Server             ServiceConfig   `json:"server"`
	Database           *DatabaseConfig `json:"database"`
	SkipUrl            string          `json:"skipUrl"`
	OpenAPITitle       string          `json:"openAPITitle"`
	OpenAPIDescription string          `json:"openAPIDescription"`
	OpenAPIUrl         string          `json:"openAPIUrl"`
	OpenAPIName        string          `json:"openAPIName"`
	DoMain             []string        `json:"doMain"`
	OpenAPIVersion     string          `json:"openAPIVersion"`
	Logger             LoggerConfig    `json:"logger"`
}

type ServiceConfig struct {
	Default ServiceDefault `json:"default"`
}

type ServiceDefault struct {
	Address           string `json:"address"`
	LogPath           string `json:"logPath"`
	LogStdout         bool   `json:"logStdout"`
	ErrorStack        bool   `json:"errorStack"`
	ErrorLogEnabled   bool   `json:"errorLogEnabled"`
	ErrorLogPattern   string `json:"errorLogPattern"`
	AccessLogEnable   bool   `json:"accessLogEnable"`
	AccessLogPattern  string `json:"accessLogPattern"`
	FileServerEnabled bool   `json:"fileServerEnabled"`
}

type DatabaseConfig struct {
	Default DatabaseDefault `json:"default"`
}

type DatabaseDefault struct {
	Host      string `json:"host"`
	Link      string `json:"link" dc:"数据库连接字符串"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Pass      string `json:"pass"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Timezone  string `json:"timezone"`
	Debug     bool   `json:"debug"`
	Charset   string `json:"charset"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type LoggerConfig struct {
	Path                string   `json:"path"`
	File                string   `json:"file"`
	Level               string   `json:"level"`
	TimeFormat          string   `json:"timeFormat"`
	CtxKeys             []string `json:"ctxKeys"`
	Header              bool     `json:"header"`
	Stdout              bool     `json:"stdout"`
	RotateSize          string   `json:"rotateSize"`
	RotateBackupLimit   int      `json:"rotateBackupLimit"`
	StdoutColorDisabled bool     `json:"stdoutColorDisabled"`
	WriterColorEnable   bool     `json:"writerColorEnable"`
}

var DefaultConfig = Config{
	Server: ServiceConfig{
		Default: ServiceDefault{
			Address:           ":8080",
			LogPath:           "./log/",
			LogStdout:         true,
			ErrorStack:        true,
			ErrorLogEnabled:   true,
			ErrorLogPattern:   "error-{Ymd}.log",
			AccessLogEnable:   false,
			FileServerEnabled: true,
		},
	},
	OpenAPITitle:       "",
	OpenAPIDescription: "Api列表 包含各端接口信息 字段注释 枚举说明",
	OpenAPIUrl:         "https://panel.magicany.cc:8888/btpanel",
	OpenAPIName:        "",
	DoMain:             []string{"localhost", "127.0.0.1"},
	OpenAPIVersion:     "v1.0",
	Logger: LoggerConfig{
		Path:              "./log/",
		File:              "access-{Ymd}.log",
		Level:             "all",
		TimeFormat:        "2006-01-02 15:04:05",
		CtxKeys:           []string{},
		Header:            true,
		Stdout:            true,
		RotateSize:        "1M",
		RotateBackupLimit: 10,
	},
}

func DefaultConfigInit() {
	database := &DatabaseConfig{Default: DatabaseDefault{
		Host:      "127.0.0.1",
		Port:      "3306",
		User:      "root",
		Pass:      "123456",
		Name:      "database",
		Type:      "mysql",
		Timezone:  "Local",
		Debug:     true,
		Charset:   "utf8",
		CreatedAt: "create_time",
		UpdatedAt: "update_time",
	}}
	DefaultConfig.Database = database
	toml, err := gtoml.Encode(DefaultConfig)
	if err != nil {
		g.Log().Error(gctx.New(), "转换toml失败", err)
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
		_ = gfile.PutContents(ConfigPath, gconv.String(toml))
		g.Log().Info(gctx.New(), "配置文件创建成功！")
	} else {
		gcfg.Instance().GetAdapter().(*gcfg.AdapterFile).SetFileName(ConfigPath)
	}
}

// DefaultSqliteConfigInit 创建默认的sqlite数据库配置 不会再生成配置文件
// @param path sqlite数据库路径
// @param autoTime 自动时间字段[]string{"create_time","update_time"}
// @param debug 数据库调试模式
// @param prefix 表前缀可空
func DefaultSqliteConfigInit(path string, autoTime []string, debug bool, prefix ...string) {
	glog.Info(gctx.New(), "正在检查文件夹", gfile.IsFile(uploadPath))
	if !gfile.IsDir(uploadPath) {
		_ = gfile.Mkdir(uploadPath)
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vtemplate", gfile.Pwd(), gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vscripts", gfile.Pwd(), gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vhtml", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vresource%vcss", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vresource%vimage", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator, gfile.Separator))
		_ = gfile.Mkdir(fmt.Sprintf("%s%vresource%vpublic%vresource%vjs", gfile.Pwd(), gfile.Separator, gfile.Separator, gfile.Separator, gfile.Separator))
	}
	g.Log().Info(gctx.New(), "正在设置数据库配置")
	node := gdb.ConfigNode{
		Link:      fmt.Sprintf("sqlite::@file(%s)", path),
		Timezone:  "Local",
		Role:      "master",
		Charset:   "utf8",
		CreatedAt: autoTime[0],
		UpdatedAt: autoTime[1],
		Debug:     debug,
	}
	if len(prefix) > 0 {
		node.Prefix = prefix[0]
	}
	err := gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			node,
		}})
	if err != nil {
		g.Log().Error(gctx.New(), "设置数据库配置失败", err)
	}
	g.Log().Info(gctx.New(), "设置数据库配置成功")
}
