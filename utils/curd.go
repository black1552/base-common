package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"
)

type ctx = context.Context

type IDao interface {
	DB() gdb.DB
	Table() string
	Group() string
	Ctx(ctx context.Context) *gdb.Model
	Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error)
}

type Curd[R any] struct {
	Dao IDao
}

var pageInfo = []string{
	"page",
	"size",
	"num",
	"limit",
	"pagesize",
	"pageSize",
	"page_size",
	"pageNum",
	"pagenum",
	"page_num",
}

// 预编译正则：匹配key中的操作符（支持 >、>=、<、<=、=、!=、LIKE、IN、NOT IN 等）
var opRegex = regexp.MustCompile(`\s*(>=|<=|>|<|!=|=|LIKE|like|IN|in|NOT IN|not in)\s*`)

// BuildWhere 构建适配goframe orm的map格式查询条件
// req: 入参结构体/Map，存放查询条件（key可直接带操作符，如"age>="、"name like"）
// changeFiles: 字段映射+操作符配置，格式如 {"name": {"op": "like", "field": "user_name", "value": "张%"}}
// caseSnake: 字段命名格式转换类型（默认下划线命名）
// 返回值: 可直接给goframe orm使用的where条件map
func (c Curd[R]) BuildWhere(req any, changeFiles map[string]any, caseSnake ...gstr.CaseType) map[string]any {
	// 1. 空值快速返回
	if req == nil {
		req = map[string]any{} // 确保后续处理不报错
	}

	// 2. 初始化命名格式（默认下划线）
	kType := gstr.Snake
	if len(caseSnake) > 0 && caseSnake[0] != "" {
		kType = caseSnake[0]
	}

	// 3. 处理字段映射/操作符配置
	changeMap := gmap.NewStrAnyMap()
	if changeFiles != nil {
		changeMap.Sets(changeFiles)
	}

	// 4. 安全转换req为Map（支持结构体/Map/指针等任意类型）
	reqM := gconv.Map(req)
	if reqM == nil {
		reqM = map[string]any{}
	}
	reqMap := gmap.NewStrAnyMapFrom(reqM)

	// 5. 第一步：处理req中的原始key（原有逻辑不变）
	keys := reqMap.Keys()
	for _, originalKey := range keys {
		val := reqMap.Get(originalKey)
		// 跳过分页字段
		if gstr.InArray(pageInfo, originalKey) {
			reqMap.Remove(originalKey)
			continue
		}

		// 6. 精准空值判断：仅跳过真正的空值（保留0、false、0.0等合法零值）
		if gutil.IsEmpty(val) {
			reqMap.Remove(originalKey)
			continue
		}

		// 7. 解析req key中的「纯字段名」和「操作符」
		fieldName, keyOp := parseKeyWithOp(originalKey)
		// 字段名格式转换（如驼峰转下划线）
		convertedField := gstr.CaseConvert(fieldName, kType)

		// 8. 优先使用changeFiles配置
		finalOp := keyOp
		finalField := convertedField
		finalVal := val // 默认使用req的原始值
		if changeMap.Size() > 0 {
			// 分步判断：优先取转换后的字段名 → 原字段名 → 原始key
			var changeVal any
			if changeMap.Contains(convertedField) {
				changeVal = changeMap.Get(convertedField)
			} else if changeMap.Contains(fieldName) {
				changeVal = changeMap.Get(fieldName)
			} else if changeMap.Contains(originalKey) {
				changeVal = changeMap.Get(originalKey)
			}

			// 仅当获取到有效值时处理
			if changeVal != nil {
				opMap := gconv.Map(changeVal)
				if len(opMap) > 0 {
					// 提取changeFiles中的操作符和目标字段（value不覆盖req的val）
					confOp := gstr.ToLower(gconv.String(opMap["op"]))
					confField := gconv.String(opMap["field"])
					// 提取changeFiles中的value（关键修正点）
					confVal := opMap["value"]
					// 值优先级：req非空值 > changeFiles的value（关键修正点）
					if !gutil.IsEmpty(confVal) {
						finalVal = confVal // req值为空时，用changeFiles的value
					}
					if confField != "" {
						finalField = confField
					}
					if confOp != "" {
						// 转换为ORM标准操作符（如gt → >）
						finalOp = getORMOp(confOp)
					}
				}
			}
		}
		// 8. 最终空值判断：仅跳过真正的空值（保留0、false、0.0等合法零值）
		if gutil.IsEmpty(finalVal) {
			reqMap.Remove(originalKey)
			continue
		}
		// 9. 生成ORM合法的key（字段名 + 操作符）
		var ormKey string
		if finalOp != "" {
			ormKey = fmt.Sprintf("%s %s", finalField, finalOp)
		} else {
			ormKey = finalField
		}

		// 10. 移除原key，设置最终的ORM查询键值对
		reqMap.Remove(originalKey)
		reqMap.Set(ormKey, val)
	}

	// 11. 返回最终的orm查询map
	return reqMap.Map()
}

// parseKeyWithOp 解析key，拆分出「纯字段名」和「操作符」
// 支持的key格式：
// - "age>=" → 字段名"age"，操作符">="
// - "name like" → 字段名"name"，操作符"LIKE"
// - "id in" → 字段名"id"，操作符"IN"
// - "status" → 字段名"status"，操作符""
func parseKeyWithOp(key string) (fieldName string, op string) {
	// 去除首尾空格
	key = strings.TrimSpace(key)
	// 匹配操作符
	matches := opRegex.FindStringSubmatch(key)
	if len(matches) == 0 {
		// 无操作符，直接返回原key（去空格）
		return key, ""
	}

	// 提取操作符并标准化（如like → LIKE，in → IN）
	op = strings.ToUpper(matches[1])
	// 提取纯字段名（去除操作符部分，再去空格）
	fieldName = strings.TrimSpace(opRegex.ReplaceAllString(key, ""))
	return
}

// getORMOp 转换语义化操作符为goframe orm支持的标准操作符
func getORMOp(op string) string {
	opMap := map[string]string{
		"=":           "=",
		"!=":          "!=",
		">":           ">",
		"<":           "<",
		">=":          ">=",
		"<=":          "<=",
		"<>":          "<>",
		"between":     "BETWEEN",
		"is null":     "IS NULL",
		"is not null": "IS NOT NULL",
		"is":          "IS",
		"is not":      "IS NOT",
		"eq":          "=",
		"ne":          "!=",
		"gt":          ">",
		"gte":         ">=",
		"lt":          "<",
		"lte":         "<=",
		"like":        "LIKE",
		"in":          "IN",
		"not in":      "NOT IN",
		"":            "", // 空操作符
	}
	if val, ok := opMap[op]; ok {
		return val
	}
	return "=" // 默认等值匹配
}
func (c Curd[R]) BuildMap(op string, value any, field ...string) map[string]any {
	if len(field) > 0 {
		return map[string]any{
			"op":    op,
			"field": field[0],
			"value": value,
		}
	}
	return map[string]any{
		"op":    op,
		"field": "",
		"value": value,
	}
}
func (c Curd[R]) Builder(ctx context.Context) *gdb.WhereBuilder {
	return c.Dao.Ctx(ctx).Builder()
}
func (c Curd[R]) ClearField(req any, delField []string, subField ...map[string]any) map[string]any {
	m := gmap.NewStrAnyMapFrom(gconv.Map(req))
	if delField != nil && len(delField) > 0 {
		m.Iterator(func(k string, v any) bool {
			if g.IsEmpty(v) {
				m.Remove(k)
				return true
			}
			if gstr.InArray(delField, k) {
				m.Remove(k)
				return true
			}
			if gstr.InArray(pageInfo, k) {
				m.Remove(k)
				return true
			}
			return true
		})
	}
	if subField != nil && len(subField) > 0 {
		m.Merge(gmap.NewStrAnyMapFrom(subField[0]))
	}
	return m.Map()
}
func (c Curd[R]) ClearFieldPage(ctx ctx, req any, delField []string, where any, page *Paginate, order any, with bool) (items []*R, total int, err error) {
	db := c.Dao.Ctx(ctx)
	m := c.ClearField(req, delField)
	if with {
		db = db.WithAll()
	}
	db = db.Where(m)
	if !g.IsNil(where) {
		db = db.Where(where)
	}
	if order != nil {
		db = db.Order(order)
	}
	if !g.IsNil(page) {
		db = db.Page(page.Page, page.Limit)
	}
	err = db.ScanAndCount(&items, &total, false)
	return
}
func (c Curd[R]) ClearFieldList(ctx ctx, req any, delField []string, where any, order any, with bool) (items []*R, err error) {
	db := c.Dao.Ctx(ctx)
	m := c.ClearField(req, delField)
	db = db.Where(m)
	if !g.IsNil(where) {
		db = db.Where(where)
	}
	if with {
		db = db.WithAll()
	}
	if !g.IsNil(order) {
		db = db.Order(order)
	}
	err = db.Scan(&items)
	return
}
func (c Curd[R]) ClearFieldOne(ctx ctx, req any, delField []string, where any, order any, with bool) (items *R, err error) {
	db := c.Dao.Ctx(ctx)
	m := c.ClearField(req, delField)
	db = db.Where(m)
	if !g.IsNil(where) {
		db = db.Where(where)
	}
	if with {
		db = db.WithAll()
	}
	if !g.IsNil(order) {
		db = db.Order(order)
	}
	err = db.Scan(&items)
	return
}
func (c Curd[R]) Value(ctx ctx, where any, field any) (*gvar.Var, error) {
	return c.Dao.Ctx(ctx).Where(where).Fields(field).Value()
}
func (c Curd[R]) DeletePri(ctx ctx, primaryKey any) error {
	_, err := c.Dao.Ctx(ctx).WherePri(primaryKey).Delete()
	return err
}
func (c Curd[R]) DeleteWhere(ctx ctx, where any) error {
	_, err := c.Dao.Ctx(ctx).Where(where).Delete()
	return err
}

func (c Curd[R]) Sum(ctx ctx, where any, field string) (float64, error) {
	return c.Dao.Ctx(ctx).Where(where).Sum(field)
}

func (c Curd[R]) ArrayField(ctx ctx, where any, field any) ([]*gvar.Var, error) {
	if field == nil {
		field = "*"
	}
	return c.Dao.Ctx(ctx).Where(where).Fields(field).Array()
}

func (c Curd[R]) FindPri(ctx ctx, primaryKey any, with bool) (model *R, err error) {
	db := c.Dao.Ctx(ctx).WherePri(primaryKey)
	if with {
		db = db.WithAll()
	}
	err = db.Scan(&model)
	if err != nil {
		return
	}
	return
}

func (c Curd[R]) First(ctx ctx, where any, order any, with bool) (model *R, err error) {
	db := c.Dao.Ctx(ctx).Where(where)
	if with {
		db = db.WithAll()
	}
	if !g.IsNil(order) {
		db = db.Order(order)
	}
	err = db.Scan(&model)
	if err != nil {
		return
	}
	return
}

func (c Curd[R]) Exists(ctx ctx, where any) (exists bool, err error) {
	return c.Dao.Ctx(ctx).Where(where).Exist()
}

func (c Curd[R]) All(ctx ctx, where any, order any, with bool) (items []*R, err error) {
	db := c.Dao.Ctx(ctx)
	if with {
		db = db.WithAll()
	}
	if !g.IsNil(order) {
		db = db.Order(order)
	}
	err = db.Where(where).Scan(&items)
	if err != nil {
		return nil, err
	}
	return
}

func (c Curd[R]) Count(ctx ctx, where any) (count int, err error) {
	count, err = c.Dao.Ctx(ctx).Where(where).Count()
	return
}

func (c Curd[R]) Save(ctx ctx, data any) (id int64, err error) {
	result, err := c.Dao.Ctx(ctx).Save(data)
	if err != nil {
		return
	}
	id, err = result.LastInsertId()
	return
}

func (c Curd[R]) Update(ctx ctx, where any, data any) (count int64, err error) {
	result, err := c.Dao.Ctx(ctx).Where(where).Data(data).Update()
	if err != nil {
		return
	}
	count, err = result.RowsAffected()
	return
}

func (c Curd[R]) UpdatePri(ctx ctx, primaryKey any, data any) (count int64, err error) {
	result, err := c.Dao.Ctx(ctx).WherePri(primaryKey).Data(data).Update()
	if err != nil {
		return
	}
	count, err = result.RowsAffected()
	return
}

func (c Curd[R]) Paginate(ctx context.Context, where any, p Paginate, with bool, order any) (items []*R, total int, err error) {
	query := c.Dao.Ctx(ctx)
	if where != nil {
		query = query.Where(where)
	}
	query = query.Page(p.Page, p.Limit)
	if order != nil {
		query = query.Order(order)
	}
	if with == true {
		query = query.WithAll()
	}
	err = query.Order(order).ScanAndCount(&items, &total, false)
	if err != nil {
		return
	}
	return
}
