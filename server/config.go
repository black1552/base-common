package server

import (
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gyaml"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/gconv"
)

type Config struct {
	Server             ServiceConfig   `yaml:"server"`
	Database           *DatabaseConfig `yaml:"database"`
	SkipUrl            string          `yaml:"skipUrl"`
	OpenAPITitle       string          `yaml:"openAPITitle"`
	OpenAPIDescription string          `yaml:"openAPIDescription"`
	OpenAPIUrl         string          `yaml:"openAPIUrl"`
	OpenAPIName        string          `yaml:"openAPIName"`
	DoMain             []string        `yaml:"doMain"`
	OpenAPIVersion     string          `yaml:"openAPIVersion"`
	Logger             LoggerConfig    `yaml:"logger"`
	Dns                string          `yaml:"dns"`
}

type ServiceConfig struct {
	Default ServiceDefault `yaml:"default"`
}

type ServiceDefault struct {
	Address           string `yaml:"address"`
	LogPath           string `yaml:"logPath"`
	LogStdout         bool   `yaml:"logStdout"`
	ErrorStack        bool   `yaml:"errorStack"`
	ErrorLogEnabled   bool   `yaml:"errorLogEnabled"`
	ErrorLogPattern   string `yaml:"errorLogPattern"`
	AccessLogEnable   bool   `yaml:"accessLogEnable"`
	AccessLogPattern  string `yaml:"accessLogPattern"`
	FileServerEnabled bool   `yaml:"fileServerEnabled"`
}

type DatabaseConfig struct {
	Default DatabaseDefault `yaml:"default"`
}

type DatabaseDefault struct {
	Host      string `yaml:"host" json:"host"`
	Link      string `yaml:"link" dc:"数据库连接字符串" json:"link"`
	Port      string `yaml:"port" json:"port"`
	User      string `yaml:"user" json:"user"`
	Pass      string `yaml:"pass" json:"pass"`
	Name      string `yaml:"name" json:"name"`
	Type      string `yaml:"type" json:"type"`
	Timezone  string `yaml:"timezone" json:"timezone"`
	Debug     bool   `yaml:"debug" json:"debug"`
	Charset   string `yaml:"charset" json:"charset"`
	CreatedAt string `yaml:"createdAt" json:"createdAt"`
	UpdatedAt string `yaml:"updatedAt" json:"updatedAt"`
}

type LoggerConfig struct {
	Path                string   `yaml:"path" json:"path"`
	File                string   `yaml:"file" json:"file"`
	Level               string   `yaml:"level" json:"level"`
	TimeFormat          string   `yaml:"timeFormat" json:"timeFormat"`
	CtxKeys             []string `yaml:"ctxKeys" json:"ctxKeys"`
	Header              bool     `yaml:"header" json:"header"`
	Stdout              bool     `yaml:"stdout" json:"stdout"`
	RotateSize          string   `yaml:"rotateSize" json:"rotateSize"`
	RotateBackupLimit   int      `yaml:"rotateBackupLimit" json:"rotateBackupLimit"`
	StdoutColorDisabled bool     `yaml:"stdoutColorDisabled" json:"stdoutColorDisabled"`
	WriterColorEnable   bool     `yaml:"writerColorEnable" json:"writerColorEnable"`
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
	Dns:                "root:123456@tcp(127.0.0.1:3306)/test",
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
	yaml, err := gyaml.Encode(DefaultConfig)
	if err != nil {
		g.Log().Error(gctx.New(), "转换yaml失败", err)
	}

	if !gfile.IsDir(uploadPath) {
		_ = gfile.Mkdir(uploadPath)
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "template"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "scripts"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "html"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "resource", "css"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "resource", "image"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "resource", "js"))
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

// DefaultSqliteConfigInit 创建默认的sqlite数据库配置 不会再生成配置文件
// @param path sqlite数据库路径
// @param autoTime 自动时间字段[]string{"create_time","update_time"}
// @param debug 数据库调试模式
// @param prefix 表前缀可空
func DefaultSqliteConfigInit(path string, autoTime []string, debug bool, prefix ...string) {
	glog.Info(gctx.New(), "正在检查文件夹", gfile.IsFile(uploadPath))
	if !gfile.IsDir(uploadPath) {
		_ = gfile.Mkdir(uploadPath)
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "template"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "scripts"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "html"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "resource", "css"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "resource", "image"))
		_ = gfile.Mkdir(gfile.Join(gfile.Pwd(), "resource", "public", "resource", "js"))
	}
	g.Log().Info(gctx.New(), "正在设置数据库配置")
	node := gdb.ConfigNode{
		Link:      fmt.Sprintf("sqlite::@file(%s)", path),
		Timezone:  "Local",
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
