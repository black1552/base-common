package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/jpeg"
	"math/big"
	"os"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gres"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nfnt/resize"
)

// Compress 图片压缩
/*
 * @param filePath string 图片路径
 * @return string
 * 压缩图片
 */
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
	_ = gfile.RemoveFile(filePath)
	return sta[0] + "-cop." + sta[1]
}

// Sha256 sha256加密
/*
 * @param src string 被加密字符串
 * @return string
 * 将字符串进行hash加密
 */
func Sha256(src string) string {
	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

// InStrArray 判断是否在数组中
/*
 * @param ext string 要判断的字符串
 * @param code int
 * 判断是否在字符串数组中
 */
func InStrArray(ext string, code int) bool {
	if code == 1 {
		arr := garray.NewStrArrayFrom(g.SliceStr{".jpg", ".jpeg", ".png"})
		return arr.Contains(ext)
	} else {
		arr := garray.NewStrArrayFrom(g.SliceStr{".xlsx"})
		return arr.Contains(ext)
	}
}

// ResAddFile 添加文件到资源包
/*
 * @param onePath string
 * 例：gf pack resource/dist internal/boot/boot_resource.go -n boot
 * 需要在boot中 打包引入 并在main.go中引入boot
 */
func ResAddFile(onePath string) {
	if gfile.Exists(gfile.Pwd() + gfile.Separator + onePath) {
		err := gfile.RemoveFile(onePath)
		if err != nil {
			panic(err)
		}
	}
	g.Log().Debug(gctx.GetInitCtx(), onePath)
	gres.Dump()
	if gres.IsEmpty() {
		return
	}
	if gstr.Contains(onePath, "/") {
		strs := gstr.Split(onePath, "/")
		err := gres.Export(strs[1], strs[0])
		if err != nil {
			panic(err)
		}

	} else {
		err := gres.Export(onePath, onePath)
		if err != nil {
			panic(err)
		}
	}
}

const (
	CharsetLower   = "abcdefghijklmnopqrstuvwxyz"
	CharsetUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetNumber  = "0123456789"
	CharsetDefault = CharsetLower + CharsetUpper + CharsetNumber
)

// GenerateString 生成安全随机字符串
func GenerateString(length int) (str string) {
	bytes := make([]byte, length)
	charsetLen := big.NewInt(int64(len(CharsetDefault)))

	for i := range bytes {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			panic(err)
		}
		bytes[i] = CharsetDefault[n.Int64()]
	}
	str = string(bytes)
	return
}

// GetFileList 获取文件列表
/*
 * 根据传入的路径返回该路径下的所有可访问的列表
 * 要求必须将静态文件访问路径定义为static，且静态文件访问目录为resource
 */
func GetFileList(path string) []string {
	if path == "" {
		path = "/"
	}
	filePath := fmt.Sprintf("%s%sresource", gfile.Pwd(), gfile.Separator)
	if path != "/" {
		filePath += gfile.Separator + path + gfile.Separator
	}
	paths, _ := gfile.DirNames(filePath)
	pathArr := garray.NewStrArray()
	for _, v := range paths {
		if gstr.Contains(v, ".") {
			if path != "/" {
				pathArr.Append("/static/" + path + "/" + v)
			} else {
				pathArr.Append("/static/" + v)
			}
		} else {
			pathArr.Append(v)
		}
	}
	return pathArr.Slice()
}
func Float64Trans(value string) *float64 {
	if value == "" {
		return nil
	}
	f := gconv.Float64(value)
	return &f
}

func IntTrans(value string) *int {
	if value == "" {
		return nil
	}
	i := gconv.Int(value)
	return &i
}
func StrTrans(value string) *string {
	if value == "" {
		return nil
	}
	i := gconv.String(value)
	return &i
}

func TimeTrans(value string) *gtime.Time {
	if value == "" {
		return nil
	}
	return gconv.GTime(value)
}
