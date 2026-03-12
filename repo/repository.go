// Package repo 提供了一个基于泛型的企业级 GORM 通用仓储（Repository）实现。
//
// BaseRepository[T] 封装了最常用的 CRUD 操作（如 Create, Update, Find, Count 等）。
// 结合 db 包的 Connector，它能够自动感知上下文中的事务；
// 结合 query 包的 Builder，它能够接收强类型的查询条件，彻底消除魔法字符串。
package repo

import (
	"context"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/query"

	"gorm.io/gorm"
)

// Repository 定义了通用的仓储接口
// T 是 Model 类型
type Repository[T any] interface {
	DB(ctx context.Context) *gorm.DB
	Create(ctx context.Context, entity *T) error
	Save(ctx context.Context, entity *T) error
	Update(ctx context.Context, qb *query.Builder, column string, value interface{}) error
	Updates(ctx context.Context, qb *query.Builder, values interface{}) error
	Delete(ctx context.Context, qb *query.Builder) error
	Find(ctx context.Context, qb *query.Builder) ([]*T, error)
	First(ctx context.Context, qb *query.Builder) (*T, error)
	Count(ctx context.Context, qb *query.Builder) (int64, error)
}

var _ Repository[any] = (*BaseRepository[any])(nil)

type BaseRepository[T any] struct {
	connector db.Connector
}

// New 创建一个新的 BaseRepository 实例
func New[T any](connector db.Connector) *BaseRepository[T] {
	return &BaseRepository[T]{connector: connector}
}

// DB 返回当前上下文的 GORM DB 实例
func (r *BaseRepository[T]) DB(ctx context.Context) *gorm.DB {
	return r.connector.DB(ctx)
}

func (r *BaseRepository[T]) buildQuery(ctx context.Context, qb *query.Builder) *gorm.DB {
	var entity T
	db := r.DB(ctx).Model(&entity)
	if qb != nil {
		db = qb.Apply(db)
	}
	return db
}

// Create 创建一个新记录
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB(ctx).Create(entity).Error
}

// Save 保存记录
func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.DB(ctx).Save(entity).Error
}

// Update 更新记录的指定字段
func (r *BaseRepository[T]) Update(ctx context.Context, qb *query.Builder, column string, value interface{}) error {
	return r.buildQuery(ctx, qb).Update(column, value).Error
}

// Updates 更新记录的多个字段
func (r *BaseRepository[T]) Updates(ctx context.Context, qb *query.Builder, values interface{}) error {
	return r.buildQuery(ctx, qb).Updates(values).Error
}

// Delete 删除记录
func (r *BaseRepository[T]) Delete(ctx context.Context, qb *query.Builder) error {
	var entity T
	return r.buildQuery(ctx, qb).Delete(&entity).Error
}

// Find 查询多条记录
func (r *BaseRepository[T]) Find(ctx context.Context, qb *query.Builder) ([]*T, error) {
	var entities []*T
	err := r.buildQuery(ctx, qb).Find(&entities).Error
	return entities, err
}

// First 查询第一条记录
func (r *BaseRepository[T]) First(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Count 查询记录数
func (r *BaseRepository[T]) Count(ctx context.Context, qb *query.Builder) (int64, error) {
	var count int64
	err := r.buildQuery(ctx, qb).Count(&count).Error
	return count, err
}
