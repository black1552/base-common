package autoMigrate

import (
	"context"

	"github.com/black1552/base-common/log"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
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

type AutoMigrate struct {
	ctx context.Context
	db  *gorm.DB
	dns string
}

var am *AutoMigrate

func New() {
	am = &AutoMigrate{}
	am.ctx = gctx.New()
	err := g.DB().PingMaster()
	if err != nil {
		log.Errorf("数据库连接失败: %+v", err)
		return
	}
	dns, err := gcfg.Instance().Get(ctx, "dns", "")
	if err != nil {
		log.Errorf("获取配置失败: %+v", err)
		return
	}
	if g.IsEmpty(dns) {
		log.Error("gormDNS未配置", "请检查配置文件")
		return
	}
	am.dns = dns.String()
	am.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:               dns.String(),
		DefaultStringSize: 255,
	}), &gorm.Config{})
	if err != nil {
		log.Errorf("gorm连接数据库失败: %+v", err)
		return
	}
}
func SetAutoMigrate(models ...interface{}) {
	New()
	if g.IsNil(am.db) {
		log.Error("数据库连接失败")
		return
	}
	db = am.db.Set("gorm:table_options", "ENGINE=InnoDB")
	err := db.AutoMigrate(models...)
	if err != nil {
		log.Errorf("数据库迁移失败: %+v", err)
	}
}
func RenameColumn(dst interface{}, name, newName string) {
	if am.db.Migrator().HasColumn(dst, name) {
		err := am.db.Migrator().RenameColumn(dst, name, newName)
		if err != nil {
			log.Errorf("数据库修改字段失败: %+v", err)
		}
	} else {
		log.Infof("数据库字段不存在: %s", name)
	}
}

// DropColumn
// 删除字段
// 例：DropColumn(&User{}, "Sex")
func DropColumn(dst interface{}, name string) {
	if am.db.Migrator().HasColumn(dst, name) {
		err := am.db.Migrator().DropColumn(dst, name)
		if err != nil {
			log.Errorf("数据库删除字段失败: %+v", err)
		}
	} else {
		log.Infof("数据库字段不存在: %s", name)
	}
}
