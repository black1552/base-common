package excel

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"
)

func CreateExcel(sheetName string, headers, cols []string, drops []DropdownCol, fileName string, centerCol []string, creatCol func(f *excelize.File, cols []string) *excelize.File) string {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			panic(fmt.Sprintf("关闭Excel文件错误:%s", err))
		}
	}()
	index, err := f.NewSheet(sheetName)
	if err != nil {
		panic(fmt.Sprintf("创建工作表错误:%s", err))
	}
	f.SetActiveSheet(index)
	_ = f.SetRowHeight(sheetName, 0, 150)
	for i, v := range headers {
		_ = f.SetCellValue("sheet1", cols[i]+"1", v)
	}
	f = creatCol(f, cols)
	if !g.IsEmpty(drops) {
		ExcelDownCol(f, drops)
	}
	if !g.IsEmpty(centerCol) {
		for _, col := range centerCol {
			_ = f.SetColStyle("sheet1", col, CenterCol(f))
		}
	}
	basePath := filepath.Join("excel", fileName)
	fPath := filepath.Join(gfile.Pwd(), "resource", basePath)
	if err := f.SaveAs(fPath); err != nil {
		panic(fmt.Sprintf("导出房源Excel模板文件错误:%s", err))
	}
	return "/uploads/" + basePath
}

// ExcelDownCol 导出模板列下拉数据
// 添加下拉框和注释
// @param f *excelize.File
// @param dropdownCols []model.DropdownCol
// @return 返回文件
func ExcelDownCol(f *excelize.File, dropdownCols []DropdownCol) *excelize.File {
	// 为导入模板的列添加下拉数据
	for _, dropdown := range dropdownCols {
		if dropdown.EndRow < dropdown.StartRow {
			// 默认设置1000行下拉框
			dropdown.EndRow = dropdown.StartRow + 999
		}

		// 构建范围字符串 (如 "B2:B1000")
		dvRange := fmt.Sprintf(
			"%s%d:%s%d",
			dropdown.Column,
			dropdown.StartRow,
			dropdown.Column,
			dropdown.EndRow,
		)

		// 创建数据验证对象
		dv := excelize.NewDataValidation(true) // 允许空值
		dv.SetSqref(dvRange)                   // 设置应用范围

		// 设置下拉列表选项
		if err := dv.SetDropList(dropdown.Options); err != nil {
			fmt.Errorf("设置下拉选项失败: %w", err)
			continue
		}

		// 添加数据验证到工作表
		if err := f.AddDataValidation("sheet1", dv); err != nil {
			fmt.Errorf("添加下拉框失败: %w", err)
			continue
		}

		// 添加注释提示（使用新版API）
		commentCell := fmt.Sprintf("%s1", dropdown.Column)
		if err := f.AddComment("sheet1", excelize.Comment{
			Cell:   commentCell,
			Author: "系统提示",
			Paragraph: []excelize.RichTextRun{
				{Text: "请从下拉列表中选择:\n"},
				{Text: JoinOptions(dropdown.Options)},
			},
		}); err != nil {
			log.Printf("添加注释失败: %v", err)
			continue
		}
	}
	return f
}

// JoinOptions 将选项列表转换为逗号分隔的字符串（用于注释）
func JoinOptions(options []string) string {
	result := ""
	for i, opt := range options {
		if i > 0 {
			result += ", "
		}
		result += opt
	}
	return result
}

// TransMap 闯入MAP 并根据isCn查询key 返回value
func TransMap(maps map[string]string, key string, isCn bool) string {
	tagMap := gmap.NewStrStrMap()
	tagMap.Sets(maps)
	if !isCn {
		tagMap.Flip()
	}
	val, ok := tagMap.Search(key)
	if ok {
		return val
	} else {
		return ""
	}
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

func TimeTrans(value string) *gtime.Time {
	if value == "" {
		return nil
	}
	return gconv.GTime(value)
}

func CenterCol(f *excelize.File) int {
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		fmt.Println("创建样式失败:", err)
		return 0
	}
	return style
}
