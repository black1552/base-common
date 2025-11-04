package server

import "fmt"

type Config struct {
	Server             *Server   `json:"server"`
	Database           *Database `json:"database"`
	SkipUrl            string    `json:"skipUrl"`
	OpenAPITitle       string    `json:"openAPITitle"`
	OpenAPIDescription string    `json:"openAPIDescription"`
	OpenAPIUrl         string    `json:"openAPIUrl"`
	OpenAPIName        string    `json:"openAPIName"`
	DoMain             []string  `json:"doMain"`
	OpenAPIVersion     string    `json:"openAPIVersion"`
	Logger             *Logger   `json:"logger"`
}

func (c *Config) SetDefaultLoggerConfig() {
	c.Logger.Path = "./log/"
	c.Logger.File = "access-{Ymd}.log"
	c.Logger.Level = "info"
	c.Logger.TimeFormat = "2006-01-02 15:04:05"
	c.Logger.CtxKeys = []string{}
	c.Logger.Header = true
	c.Logger.Stdout = true
	c.Logger.RotateSize = "2M"
	c.Logger.RotateBackupLimit = 10
	c.Logger.WriterColorEnable = true
	c.Logger.StdoutColorDisabled = false
}

func (c *Config) SetDefaultDatabaseConfig() {
	c.Database.Default.Host = "127.0.0.1"
	c.Database.Default.Port = "3306"
	c.Database.Default.User = ""
	c.Database.Default.Pass = ""
	c.Database.Default.Name = ""
	c.Database.Default.Type = "mysql"
	c.Database.Default.Timezone = "Asia/Shanghai"
	c.Database.Default.Debug = true
	c.Database.Default.Charset = "utf8mb4"
	c.Database.Default.CreatedAt = "created_at"
	c.Database.Default.UpdatedAt = "updated_at"
	c.Database.Default.Prefix = "gf_"
	c.Database.Default.Extra = "parseTime=True"
}

type Logger struct {
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

type Database struct {
	Default struct {
		Host      string `json:"host"`
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
		Prefix    string `json:"prefix"`
		Extra     string `json:"extra"`
	} `json:"default"`
}

type Server struct {
	Default struct {
		Address           string `json:"address"`
		LogPath           string `json:"logPath"`
		LogStdout         bool   `json:"logStdout"`
		ErrorStack        bool   `json:"errorStack"`
		ErrorLogEnabled   bool   `json:"errorLogEnabled"`
		ErrorLogPattern   string `json:"errorLogPattern"`
		AccessLogEnable   bool   `json:"accessLogEnable"`
		AccessLogPattern  string `json:"accessLogPattern"`
		FileServerEnabled bool   `json:"fileServerEnabled"`
		CookieDomain      string `json:"cookieDomain" dc:"http://localhost:80"`
	} `json:"default"`
}

func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.SetDefaultDatabaseConfig()
	cfg.SetDefaultLoggerConfig()
	cfg.SetDefaultServerConfig()
	cfg.SkipUrl = ""
	cfg.OpenAPITitle = ""
	cfg.OpenAPIDescription = "Api列表 包含各端接口信息 字段注释 枚举说明"
	cfg.OpenAPIUrl = "https://panel.magicany.cc:8888/btpanel"
	cfg.OpenAPIName = ""
	cfg.DoMain = []string{"localhost", "127.0.0.1"}
	cfg.OpenAPIVersion = "v1.0"
	return cfg
}

func (c *Config) SetDefaultServerConfig() {
	c.Server.Default.Address = "127.0.0.1:8080"
	c.Server.Default.LogPath = "./log/"
	c.Server.Default.LogStdout = true
	c.Server.Default.ErrorStack = true
	c.Server.Default.ErrorLogEnabled = true
	c.Server.Default.ErrorLogPattern = "error-{Ymd}.log"
	c.Server.Default.AccessLogEnable = false
	c.Server.Default.AccessLogPattern = "access-{Ymd}.log"
	c.Server.Default.FileServerEnabled = true
	c.Server.Default.CookieDomain = fmt.Sprintf("http://%s", c.Server.Default.Address)
}
