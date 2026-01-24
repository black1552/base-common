package utils

import (
	"context"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
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

func (c Curd[R]) BuildWhere(req any, changeWhere any, subWhere any, removeFields []string, isSnake ...bool) map[string]any {
	// 默认使用小写下划线方式
	caseTypeValue := gstr.Snake
	if len(isSnake) > 0 && isSnake[0] == false {
		caseTypeValue = gstr.CamelLower
	}

	// 转换req为map
	reqMap := gconv.Map(req)

	// 清理空值和分页信息
	ctx := gctx.New()
	cleanedReq := make(map[string]any)
	for k, v := range reqMap {
		// 清理空值
		if g.IsEmpty(v) {
			glog.Debugf(ctx, "清理空值：%s", k)
			continue
		}
		// 清理分页信息
		if gstr.InArray(pageInfo, k) {
			glog.Debugf(ctx, "清理分页信息：%s", k)
			continue
		}
		if len(removeFields) > 0 && gstr.InArray(removeFields, k) {
			glog.Debugf(ctx, "清理字段：%s", k)
			continue
		}
		cleanedReq[gstr.CaseConvert(k, caseTypeValue)] = v
	}

	// 处理changeWhere
	if changeWhere != nil {
		changeMap := gconv.Map(changeWhere)
		for k, v := range changeMap {
			if _, hasKey := cleanedReq[k]; !hasKey {
				glog.Debugf(ctx, "处理changeWhere：%s", k)
				continue
			}
			if len(removeFields) > 0 && gstr.InArray(removeFields, k) {
				glog.Debugf(ctx, "清理应删除字段：%s", k)
				continue
			}
			// 转换v为map
			vMap := gconv.Map(v)
			value, hasValue := vMap["value"]
			op, hasOp := vMap["op"]

			if hasValue {
				glog.Debugf(ctx, "变更字段存在value：%s", k)
				// 构建新的键名
				newKey := k
				if hasOp && op != "" {
					glog.Debugf(ctx, "变更字段存在op：%s", k)
					newKey = k + " " + gconv.String(op)
					delete(cleanedReq, k)
				}
				cleanedReq[newKey] = value
			}
		}
	}

	// 变量名切换
	resultMap := make(map[string]any)
	for k, v := range cleanedReq {
		// 提取原始键名（去掉op部分）
		originalKey := k
		opStr := ""
		if opIndex := gstr.Pos(k, " "); opIndex > 0 {
			originalKey = k[:opIndex]
			opStr = k[opIndex+1:]
		}

		// 转换键名
		convertedKey := originalKey
		convertedKey = gstr.CaseConvert(convertedKey, caseTypeValue)

		// 如果有op，重新构建键名
		if opStr != "" {
			convertedKey = convertedKey + " " + opStr
		}

		resultMap[convertedKey] = v
	}
	// 合并subWhere
	if subWhere != nil {
		subMap := gconv.Map(subWhere)
		resultM := gmap.NewStrAnyMapFrom(resultMap)
		resultM.Merge(gmap.NewStrAnyMapFrom(subMap))
		resultMap = resultM.Map()
	}
	return resultMap
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
	err = db.Where(where).Order(order).Scan(&items)
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
	if order == nil {
		order = "create_time desc"
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
