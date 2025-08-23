package utils

import (
	"context"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gdb"
)

type ctx = context.Context

type ICurd[R any] interface {
	Save(ctx ctx, data any) (id int64, err error)
	Find(ctx ctx, primaryKey any) (model *R, err error)
	All(ctx ctx, where any, with []any, order any, limit ...int) (items []*R, err error)
	First(ctx ctx, where any, order ...any) (model *R, err error)
	Paginate(ctx ctx, where any, p Paginate, with []any, order any) (paginator *Paginator[*R], err error)
	Update(ctx ctx, where any, data any) (count int64, err error)
	UpdatePri(ctx ctx, primaryKey any, data any) (count int64, err error)
	Exists(ctx ctx, where any) (exists bool, err error)
	Delete(ctx ctx, primaryKey any) error
	Value(ctx ctx, where any, field any) (*gvar.Var, error)
	Sum(ctx ctx, where any, field string) (float64, error)
	ArrayField(ctx ctx, where any, field any) ([]*gvar.Var, error)
	Count(ctx ctx, where any) (count int, err error)
}

type IDao interface {
	DB() gdb.DB
	Table() string
	Group() string
	Ctx(ctx context.Context) *gdb.Model
	Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error)
}

type Curd[R any] struct {
	Dao    IDao
	Column R
}

func (c Curd[R]) Value(ctx ctx, where any, field any) (*gvar.Var, error) {
	return c.Dao.Ctx(ctx).Where(where).Fields(field).Value()
}
func (c Curd[R]) Array(ctx ctx, where any, field any) ([]*gvar.Var, error) {
	if field == nil {
		field = "*"
	}
	return c.Dao.Ctx(ctx).Where(where).Fields(field).Array()
}
func (c Curd[R]) Delete(ctx ctx, primaryKey any) error {
	_, err := c.Dao.Ctx(ctx).WherePri(primaryKey).Delete()
	return err
}

func (c Curd[R]) Sum(ctx ctx, where any, field string) (float64, error) {
	return c.Dao.Ctx(ctx).Where(where).Sum(field)
}

func (c Curd[R]) ArrayField(ctx ctx, where any, field any) ([]*gvar.Var, error) {
	return c.Dao.Ctx(ctx).Where(where).Fields(field).Array()
}

func (c Curd[R]) Find(ctx ctx, primaryKey any, with ...bool) (model *R, err error) {
	db := c.Dao.Ctx(ctx).WherePri(primaryKey)
	if len(with) > 0 && with[0] == true {
		db = db.WithAll()
	}
	err = db.Scan(&model)
	if err != nil {
		return
	}
	return
}

func (c Curd[R]) First(ctx ctx, where any, with bool, order ...any) (model *R, err error) {
	db := c.Dao.Ctx(ctx).Where(where)
	if with {
		db = db.WithAll()
	}
	if len(order) > 0 {
		db = db.Order(order...)
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

func (c Curd[R]) All(ctx ctx, where any, with bool, order any) (items []*R, err error) {
	db := c.Dao.Ctx(ctx)
	if with == true {
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
	query = query.Page(p.PageNum, p.PageSize)
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
