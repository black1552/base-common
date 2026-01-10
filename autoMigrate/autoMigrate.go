package autoMigrate

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	ctx context.Context
	db  *gorm.DB
)

func NewAutoMigrate() {
	ctx = gctx.New()
	err := g.DB().PingMaster()
	if err != nil {
		g.Log().Error(ctx, "数据库连接失败", err)
		return
	}
	dns, err := gcfg.Instance().Get(ctx, "dns", "")
	if err != nil {
		g.Log().Error(ctx, "获取配置失败", err)
		return
	}
	if g.IsEmpty(dns) {
		g.Log().Error(ctx, "gormDNS未配置", "请检查配置文件")
		return
	}
	gormDb, err := gorm.Open(mysql.New(mysql.Config{
		DSN:               dns.String(),
		DefaultStringSize: 255,
	}), &gorm.Config{})
	if err != nil {
		g.Log().Error(ctx, "gorm连接数据库失败", err)
		return
	}
	db = gormDb
}
func SetAutoMigrate(models ...interface{}) {
	NewAutoMigrate()
	if g.IsNil(db) {
		g.Log().Error(ctx, "数据库连接失败")
		return
	}
	db = db.Set("gorm:table_options", "ENGINE=InnoDB")
	err := db.AutoMigrate(models...)
	if err != nil {
		g.Log().Error(ctx, "数据库迁移失败", err)
	}
}
