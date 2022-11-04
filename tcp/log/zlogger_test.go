package log_test

import (
	"github.com/black1552/base-common/tcp/log"
	"testing"
)

func TestStdZLog(t *testing.T) {

	//测试 默认debug输出
	log.Debug("zinx debug content1")
	log.Debug("zinx debug content2")

	log.Debugf(" zinx debug a = %d\n", 10)

	//设置log标记位，加上长文件名称 和 微秒 标记
	log.ResetFlags(log.BitDate | log.BitLongFile | log.BitLevel)
	log.Info("zinx info content")

	//设置日志前缀，主要标记当前日志模块
	log.SetPrefix("MODULE")
	log.Error("zinx error content")

	//添加标记位
	log.AddFlag(log.BitShortFile | log.BitTime)
	log.Stack(" Zinx Stack! ")

	//设置日志写入文件
	log.SetLogFile("./log", "testfile.log")
	log.Debug("===> zinx debug content ~~666")
	log.Debug("===> zinx debug content ~~888")
	log.Error("===> zinx Error!!!! ~~~555~~~")

	//关闭debug调试
	log.CloseDebug()
	log.Debug("===> 我不应该出现~！")
	log.Debug("===> 我不应该出现~！")
	log.Error("===> zinx Error  after debug close !!!!")
}
