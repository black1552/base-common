package v2

import (
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

func GetPort(agent string) int {
	path := "D:\\go-object\\port.txt"
	port := 8114
	agent = gstr.Replace(agent, " ", "_")
	if !gfile.IsFile(path) {
		_, _ = gfile.Create(path)
		_ = gfile.PutContents(path, "8081 "+agent)
		return port
	}
	oldStr := gfile.GetContents(path)
	oldArr := gstr.Split(oldStr, " ")
	old := gconv.Int(oldArr[0])
	if oldArr[1] != agent {
		newPort := old + 1
		_ = gfile.PutContents(path, gconv.String(newPort)+" "+agent)
		return newPort
	}
	return old
}
