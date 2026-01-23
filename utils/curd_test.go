package utils

import (
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/stretchr/testify/assert"
)

func BuildWhere(req any, changeFiles map[string]any, caseSnake ...gstr.CaseType) map[string]any {
	if req == nil {
		return map[string]any{}
	}
	kType := gstr.Snake
	if len(caseSnake) > 0 {
		kType = caseSnake[0]
	}
	changMap := gmap.NewStrAnyMap()
	if changeFiles != nil {
		changMap.Sets(changeFiles)
	}
	reqM := gconv.Map(req)
	reqMap := gmap.NewStrAnyMapFrom(reqM)
	reqMap.Iterator(func(k string, v any) bool {
		if g.IsEmpty(v) {
			reqMap.Remove(k)
			return true
		}
		if gstr.InArray(pageInfo, k) {
			reqMap.Remove(k)
			return true
		}
		reqMap.Remove(k)
		newK := gstr.CaseConvert(k, kType)
		reqMap.Set(newK, v)
		if changMap != nil && changMap.Contains(k) {
			vMap := gmap.NewStrAnyMapFrom(gconv.Map(changMap.Get(k)))
			if vMap.Contains("op") {
				reqMap.Remove(k)
				reqMap.Set(fmt.Sprintf("%s %s", gstr.CaseConvert(k, kType), gstr.CaseConvert(gconv.String(vMap.Get("op")), kType)), vMap.Get("value"))
				return true
			}
		}
		return true
	})
	return reqMap.Map()
}

// 测试结构体
type TestStruct struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Sex     string `json:"sex"`
	AddRess string `json:"addRess"`
	Empty   string `json:"empty"` // 用于测试空值过滤
	Page    int    `json:"page"`
	PageNum int    `json:"pageNum"`
}

func TestBuildWhere(t *testing.T) {
	t.Run("基本功能测试 - 转换为snake_case", func(t *testing.T) {
		input := TestStruct{
			ID:      1,
			Name:    "张三",
			Age:     18,
			AddRess: "北京",
			Page:    1,
			PageNum: 10,
		}

		result := BuildWhere(input, nil)

		expected := map[string]any{
			"id":       1,
			"name":     "张三",
			"age":      18,
			"add_ress": "北京",
		}

		assert.Equal(t, expected, result)
	})

	t.Run("空值过滤测试", func(t *testing.T) {
		input := TestStruct{
			ID:    1,
			Name:  "",
			Age:   0, // 零值也会被过滤
			Sex:   "male",
			Empty: "", // 空字符串应被过滤
		}

		result := BuildWhere(input, nil)

		expected := map[string]any{
			"id":  1,
			"sex": "male",
		}

		assert.Equal(t, expected, result)
	})

	t.Run("带操作符的变更映射测试", func(t *testing.T) {
		input := TestStruct{
			ID:   3,
			Name: "李四",
			Age:  45,
		}

		changeMap := map[string]any{
			"name": map[string]any{
				"op":    "like",
				"value": "%李%",
			},
			"age": map[string]any{
				"op":    ">=",
				"value": 18,
			},
		}

		result := BuildWhere(input, changeMap)

		expected := map[string]any{
			"id":        3,
			"name like": "%李%",
			"age >=":    18,
		}

		assert.Equal(t, expected, result)
	})

	t.Run("自定义大小写转换测试", func(t *testing.T) {
		input := TestStruct{
			ID:   1,
			Name: "测试",
		}

		result := BuildWhere(input, nil, gstr.KebabScreaming)

		expected := map[string]any{
			"ID":   1,
			"NAME": "测试",
		}

		assert.Equal(t, expected, result)
	})

	t.Run(" pageInfo 字段过滤测试", func(t *testing.T) {
		// 模拟 pageInfo 包含的字段
		originalPageInfo := make([]string, len(pageInfo))
		copy(originalPageInfo, pageInfo)

		// 设置 pageInfo 的值进行测试
		pageInfo = []string{"page", "size"} // 假设这些是分页参数

		input := g.Map{
			"id":   7,
			"name": "测试",
			"page": 1,
			"size": 10,
		}

		result := BuildWhere(input, nil)

		// page 和 size 应该被过滤掉
		expected := map[string]any{
			"id":   7,
			"name": "测试",
		}

		assert.Equal(t, expected, result)

		// 恢复原始 pageInfo
		pageInfo = originalPageInfo
	})

	t.Run("混合测试 - 同时使用变更映射和大小写转换", func(t *testing.T) {
		input := TestStruct{
			ID:   1,
			Name: "王五",
			Age:  25,
		}

		changeMap := map[string]any{
			"name": map[string]any{
				"op":    "like",
				"value": "%王%",
			},
		}

		// 使用大写转换
		result := BuildWhere(input, changeMap, gstr.KebabScreaming)

		expected := map[string]any{
			"ID":        1,
			"NAME":      "王五",
			"AGE":       25,
			"NAME LIKE": "%王%",
		}

		assert.Equal(t, expected, result)
	})

	t.Run("空输入测试", func(t *testing.T) {
		result := BuildWhere(nil, nil)
		assert.Equal(t, map[string]any{}, result)
	})

	t.Run("nil变更映射测试", func(t *testing.T) {
		input := TestStruct{
			ID:   1,
			Name: "测试",
		}

		result := BuildWhere(input, nil)

		expected := map[string]any{
			"id":   1,
			"name": "测试",
		}

		assert.Equal(t, expected, result)
	})
}
