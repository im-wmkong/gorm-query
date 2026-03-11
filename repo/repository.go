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

func New[T any](connector db.Connector) *BaseRepository[T] {
	return &BaseRepository[T]{connector: connector}
}

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

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB(ctx).Create(entity).Error
}

func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.DB(ctx).Save(entity).Error
}

func (r *BaseRepository[T]) Update(ctx context.Context, qb *query.Builder, column string, value interface{}) error {
	return r.buildQuery(ctx, qb).Update(column, value).Error
}

func (r *BaseRepository[T]) Updates(ctx context.Context, qb *query.Builder, values interface{}) error {
	return r.buildQuery(ctx, qb).Updates(values).Error
}

func (r *BaseRepository[T]) Delete(ctx context.Context, qb *query.Builder) error {
	var entity T
	return r.buildQuery(ctx, qb).Delete(&entity).Error
}

func (r *BaseRepository[T]) Find(ctx context.Context, qb *query.Builder) ([]*T, error) {
	var entities []*T
	err := r.buildQuery(ctx, qb).Find(&entities).Error
	return entities, err
}

func (r *BaseRepository[T]) First(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Count(ctx context.Context, qb *query.Builder) (int64, error) {
	var count int64
	err := r.buildQuery(ctx, qb).Count(&count).Error
	return count, err
}
