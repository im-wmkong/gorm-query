// Package query 提供了强类型的 GORM 查询构建器核心能力。
//
// 它通过 Condition 函数式选项模式动态构建查询，支持深拷贝 (Clone) 以防止
// 基础查询条件被衍生查询污染。它通常与 cmd/gen-props 生成的属性字典配合使用，
// 为开发者提供极致顺滑、类型安全的链式 SQL 构建体验。
package query

import (
	"github.com/im-wmkong/gorm-query/internal/column"
	"github.com/im-wmkong/gorm-query/internal/gormx"
	"gorm.io/gorm"
)

// Condition 定义一个修改 *gorm.DB 对象的函数
// 这是核心抽象，所有的 WHERE 条件最终都转换为此函数
type Condition func(db *gorm.DB) *gorm.DB

// Builder 查询构建器
type Builder struct {
	conditions []Condition
}

// New 创建一个新的查询构建器
func New() *Builder {
	return &Builder{}
}

// Apply 将所有累积的查询条件应用到 gorm.DB 上
func (b *Builder) Apply(db *gorm.DB) *gorm.DB {
	for _, cond := range b.conditions {
		db = cond(db)
	}
	return db
}

// Clone 深度拷贝当前的 Builder，方便复用公共查询条件而不互相污染
func (b *Builder) Clone() *Builder {
	conditions := make([]Condition, len(b.conditions))
	copy(conditions, b.conditions)
	return &Builder{conditions: conditions}
}

// Where 接受一个或多个 Condition
func (b *Builder) Where(conds ...Condition) *Builder {
	return b.bind(conds...)
}

// Or 添加 OR 条件
func (b *Builder) Or(conds ...Condition) *Builder {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Or(nested)
	})
}

// Not 添加 NOT 条件
func (b *Builder) Not(conds ...Condition) *Builder {
	return b.nested(conds, func(db, nested *gorm.DB) *gorm.DB {
		return db.Not(nested)
	})
}

// Select 指定查询字段
func (b *Builder) Select(query any, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Select(column.Value(query), column.Values(args)...)
	})
}

// Omit 忽略字段
func (b *Builder) Omit(columns ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Omit(column.ValuesTo[string](columns)...)
	})
}

// Joins 连接查询
func (b *Builder) Joins(query string, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Joins(query, column.Values(args)...)
	})
}

// Preload 预加载关联
func (b *Builder) Preload(query string, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Preload(query, column.Values(args)...)
	})
}

// Group 分组
func (b *Builder) Group(name any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Group(column.ValueTo[string](name))
	})
}

// Having 分组后过滤
func (b *Builder) Having(query any, args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Having(column.Value(query), column.Values(args)...)
	})
}

// Order 排序
func (b *Builder) Order(col any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Order(column.Value(col))
	})
}

// Page 分页
func (b *Builder) Page(page, pageSize int) *Builder {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Limit(pageSize).Offset(offset)
	})
}

// Limit 限制返回结果数量
func (b *Builder) Limit(limit int) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	})
}

// Offset 偏移量
func (b *Builder) Offset(offset int) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	})
}

// Distinct 去重
func (b *Builder) Distinct(args ...any) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Distinct(column.Values(args)...)
	})
}

// Unscoped 忽略软删除等 Scope
func (b *Builder) Unscoped() *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	})
}

// Scope 支持 GORM Scopes
func (b *Builder) Scope(funcs ...func(*gorm.DB) *gorm.DB) *Builder {
	return b.bind(func(db *gorm.DB) *gorm.DB {
		return db.Scopes(funcs...)
	})
}

func (b *Builder) nested(conds []Condition, applier func(db, nested *gorm.DB) *gorm.DB) *Builder {
	if len(conds) == 0 {
		return b
	}

	return b.bind(func(db *gorm.DB) *gorm.DB {
		return applier(db, gormx.BuildNested(db, conds))
	})
}

func (b *Builder) bind(conds ...Condition) *Builder {
	b.conditions = append(b.conditions, conds...)
	return b
}
