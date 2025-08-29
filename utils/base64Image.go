package utils

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
)

// Base64ToImage 将Base64编码的数据转换为图片并保存到指定路径
func Base64ToImage(base64String, outputPath string) error {
	// 移除Base64数据URI前缀（如果有）
	if strings.Contains(base64String, ",") {
		base64String = strings.Split(base64String, ",")[1]
	}

	// 解码Base64字符串
	imageData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return fmt.Errorf("解码Base64数据时出错: %v", err)
	}

	// 将解码后的数据写入文件
	err = ioutil.WriteFile(outputPath, imageData, 0644)
	if err != nil {
		return fmt.Errorf("保存图片文件时出错: %v", err)
	}

	return nil
}

// Base64ToImageWithFormat 将Base64编码的数据转换为指定格式的图片并保存
func Base64ToImageWithFormat(base64String, outputPath, format string) error {
	// 移除Base64数据URI前缀（如果有）
	if strings.Contains(base64String, ",") {
		base64String = strings.Split(base64String, ",")[1]
	}

	// 解码Base64字符串
	imageData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return fmt.Errorf("解码Base64数据时出错: %v", err)
	}

	// 创建文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建图片文件时出错: %v", err)
	}
	defer file.Close()

	// 根据指定格式编码图片
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		// 解码图片
		img, _, err := image.Decode(strings.NewReader(string(imageData)))
		if err != nil {
			return fmt.Errorf("解码图片时出错: %v", err)
		}

		// 编码为JPEG格式
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return fmt.Errorf("编码JPEG图片时出错: %v", err)
		}
	case "png":
		// 解码图片
		img, _, err := image.Decode(strings.NewReader(string(imageData)))
		if err != nil {
			return fmt.Errorf("解码图片时出错: %v", err)
		}

		// 编码为PNG格式
		err = png.Encode(file, img)
		if err != nil {
			return fmt.Errorf("编码PNG图片时出错: %v", err)
		}
	default:
		// 直接写入原始数据
		_, err = file.Write(imageData)
		if err != nil {
			return fmt.Errorf("写入图片数据时出错: %v", err)
		}
	}

	return nil
}

// GetImageFormatFromBase64 从Base64数据中获取图片格式
func GetImageFormatFromBase64(base64String string) string {
	// 检查数据URI前缀
	if strings.HasPrefix(base64String, "data:image/") {
		// 提取格式信息
		format := strings.Split(strings.Split(base64String, ";")[0], "/")[1]
		return format
	}
	return "unknown"
}
