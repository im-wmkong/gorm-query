// Package repo 提供了一个基于泛型的企业级 GORM 通用仓储（Repository）实现。
//
// BaseRepository[T] 封装了最常用的 CRUD 操作（如 Create, Update, Find, Count 等）。
// 结合 db 包的 DBProvider，它能够自动感知上下文中的事务；
// 结合 query 包的 Builder，它能够接收强类型的查询条件，彻底消除魔法字符串。
package repo

import (
	"context"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/internal/cast"
	"github.com/im-wmkong/gorm-query/query"

	"gorm.io/gorm"
)

// Repository 定义了通用的仓储接口
// T 是 Model 类型
type Repository[T any] interface {
	DB(ctx context.Context) *gorm.DB
	Save(ctx context.Context, entity *T) error
	Create(ctx context.Context, entity *T) error
	CreateInBatches(ctx context.Context, entities []*T, batchSize int) (int64, error)
	Update(ctx context.Context, qb *query.Builder, column query.Column, value any) (int64, error)
	Updates(ctx context.Context, qb *query.Builder, values any) (int64, error)
	Delete(ctx context.Context, qb *query.Builder) (int64, error)
	Find(ctx context.Context, qb *query.Builder) ([]*T, error)
	First(ctx context.Context, qb *query.Builder) (*T, error)
	Take(ctx context.Context, qb *query.Builder) (*T, error)
	Last(ctx context.Context, qb *query.Builder) (*T, error)
	Count(ctx context.Context, qb *query.Builder) (int64, error)
	Pluck(ctx context.Context, qb *query.Builder, column query.Column, dest any) error
}

var _ Repository[any] = (*BaseRepository[any])(nil)

type BaseRepository[T any] struct {
	dbProvider db.DBProvider
}

// New 创建一个新的 BaseRepository 实例
func New[T any](dbProvider db.DBProvider) *BaseRepository[T] {
	return &BaseRepository[T]{dbProvider: dbProvider}
}

// DB 返回当前上下文的 GORM DB 实例
func (r *BaseRepository[T]) DB(ctx context.Context) *gorm.DB {
	return r.dbProvider.DB(ctx)
}

func (r *BaseRepository[T]) buildQuery(ctx context.Context, qb *query.Builder) *gorm.DB {
	var entity T
	db := r.DB(ctx).Model(&entity)
	if qb != nil {
		db = qb.Apply(db)
	}
	return db
}

// Save 保存记录
func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.DB(ctx).Save(entity).Error
}

// Create 创建一个新记录
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB(ctx).Create(entity).Error
}

// CreateInBatches 创建多条记录
func (r *BaseRepository[T]) CreateInBatches(ctx context.Context, entities []*T, batchSize int) (int64, error) {
	result := r.DB(ctx).CreateInBatches(entities, batchSize)
	return result.RowsAffected, result.Error
}

// Update 更新记录的指定字段
func (r *BaseRepository[T]) Update(ctx context.Context, qb *query.Builder, column query.Column, value any) (int64, error) {
	result := r.buildQuery(ctx, qb).Update(column.String(), value)
	return result.RowsAffected, result.Error
}

// Updates 更新记录的多个字段
func (r *BaseRepository[T]) Updates(ctx context.Context, qb *query.Builder, values any) (int64, error) {
	if mapVals, ok := values.(map[query.Column]any); ok {
		values = cast.ToStringMap(mapVals)
	}
	result := r.buildQuery(ctx, qb).Updates(values)
	return result.RowsAffected, result.Error
}

// Delete 删除记录
func (r *BaseRepository[T]) Delete(ctx context.Context, qb *query.Builder) (int64, error) {
	var entity T
	result := r.buildQuery(ctx, qb).Delete(&entity)
	return result.RowsAffected, result.Error
}

// Find 查询多条记录
func (r *BaseRepository[T]) Find(ctx context.Context, qb *query.Builder) ([]*T, error) {
	var entities []*T
	err := r.buildQuery(ctx, qb).Find(&entities).Error
	return entities, err
}

// First 获取第一条记录（主键升序）
func (r *BaseRepository[T]) First(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).First(&entity).Error
	return &entity, err
}

// Take 获取一条记录，没有指定排序字段
func (r *BaseRepository[T]) Take(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).Take(&entity).Error
	return &entity, err
}

// Last 获取最后一条记录（主键降序）
func (r *BaseRepository[T]) Last(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).Last(&entity).Error
	return &entity, err
}

// Count 查询记录数
func (r *BaseRepository[T]) Count(ctx context.Context, qb *query.Builder) (int64, error) {
	var count int64
	err := r.buildQuery(ctx, qb).Count(&count).Error
	return count, err
}

// Pluck 获取记录的指定字段值
func (r *BaseRepository[T]) Pluck(ctx context.Context, qb *query.Builder, column query.Column, dest any) error {
	return r.buildQuery(ctx, qb).Pluck(column.String(), dest).Error
}
