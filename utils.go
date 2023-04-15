package v2

import (
	"context"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/nfnt/resize"
	"image/jpeg"
	"os"
)

func Compress(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err.Error())
	}
	img, err := jpeg.Decode(file)
	if err != nil {
		panic(err.Error())
	}
	err = file.Close()
	if err != nil {
		panic(err.Error())
	}
	m := resize.Resize(960, 0, img, resize.Lanczos2)
	str := gstr.Split(filePath, "/")
	sta := gstr.Split(str[len(str)-1], ".")
	paths := gfile.Pwd() + "/resource/public/upload/" + sta[0] + "-cop." + sta[1]
	out, err := os.Create(paths)
	defer out.Close()
	err = jpeg.Encode(out, m, nil)
	if err != nil {
		panic(err.Error())
	}
	_ = gfile.Remove(filePath)
	return sta[0] + "-cop." + sta[1]
}

func InStrArray(ext string, code int) bool {
	if code == 1 {
		arr := garray.NewStrArrayFrom(g.SliceStr{".jpg", ".jpeg", ".png"})
		return arr.Contains(ext)
	} else {
		arr := garray.NewStrArrayFrom(g.SliceStr{".xlsx"})
		return arr.Contains(ext)
	}
}

var (
	log *glog.Logger
	ctx context.Context
)

func InitLog(path string) {
	ctx = gctx.New()
	log = glog.New()
	logPath := gfile.Pwd() + "/" + path
	if !gfile.IsDir(logPath) {
		err := gfile.Mkdir(logPath)
		if err != nil {
			panic(err.Error())
		}
	}
	log.SetStack(true)
	log.SetStdoutPrint(true)
	log.SetFile("{Y-m-d}.log")
	_ = log.SetLevelStr("ALL")
	_ = log.SetPath(logPath)
}

func LogInfo(text ...interface{}) {
	log.Info(ctx, text...)
}
func LogError(text ...interface{}) {
	log.Error(ctx, text...)
}
func LogDebug(text ...interface{}) {
	log.Debug(ctx, text...)
}
