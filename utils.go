package v2

import (
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
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
	Path     = gfile.Pwd() + "/device-log/"
	Name     = "info_"
	Exp      = ".log"
	FilePath = ""
)

func DeBug(content string) {
	time := gtime.Datetime()
	timeHour := gstr.Split(time, ":")[0]
	name := Name + timeHour + Exp
	FilePath = Path + name
	if !gfile.IsFile(FilePath) {
		_, err := gfile.Create(FilePath)
		if err != nil {
			panic(err)
		}
	}
	file, err := gfile.OpenFile(FilePath, os.O_APPEND|os.O_WRONLY, gfile.DefaultPermCopy)

	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	_, err = file.Write(gconv.Bytes("\n" + gtime.Datetime() + ">>>>" + content))
	if err != nil {
		panic(err)
	}
}
