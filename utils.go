package v2

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gres"
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

func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
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
	ctx context.Context
)

func ResAddFile(onePath, twoPath string) {
	if twoPath == "" {
		err := gres.Load(fmt.Sprintf("%s", onePath))
		if err != nil {
			g.Log().Debug(gctx.GetInitCtx(), err)
		}
	} else {
		err := gres.Load(fmt.Sprintf("%s/%s", onePath, twoPath))
		if err != nil {
			g.Log().Debug(gctx.GetInitCtx(), err)
		}
	}
	if !gres.IsEmpty() {
		g.Log().Debug(gctx.GetInitCtx(), fmt.Sprintf("%s is not empty", onePath))
		if twoPath == "" {
			err := gres.Export(onePath, onePath)
			if err != nil {
				panic(err)
			}
		} else {
			err := gres.Export(twoPath, onePath)
			if err != nil {
				panic(err)
			}
		}
	}
}
