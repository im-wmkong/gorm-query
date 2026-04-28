// Package repo provides a generic GORM repository implementation.
//
// BaseRepository[T] wraps the most common CRUD operations (Create, Update, Find, Count, etc.).
// With db.DBProvider it can automatically pick up transactions from context;
// with query.Builder it can accept type-safe query conditions to avoid magic strings.
package repo

import (
	"context"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/internal/column"
	"github.com/im-wmkong/gorm-query/query"

	"gorm.io/gorm"
)

// Repository defines a generic repository interface.
// T is the model type.
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

// New creates a new BaseRepository.
//
// Example:
//
//	r := repo.New[User](dbClient) // dbClient implements db.DBProvider
//	_ = r
func New[T any](dbProvider db.DBProvider) *BaseRepository[T] {
	return &BaseRepository[T]{dbProvider: dbProvider}
}

// DB returns the GORM DB instance for the given context.
//
// Example:
//
//	session := r.DB(ctx)
//	_ = session
func (r *BaseRepository[T]) DB(ctx context.Context) *gorm.DB {
	return r.dbProvider.DB(ctx)
}

func (r *BaseRepository[T]) buildQuery(ctx context.Context, qb *query.Builder) *gorm.DB {
	var entity T
	session := r.DB(ctx).Model(&entity)
	if qb != nil {
		session = qb.Apply(session)
	}
	return session
}

// Save persists the entity (insert or update).
//
// Example:
//
//	err := r.Save(ctx, &User{ID: 1, UserName: "Alice"})
//	_ = err
func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.DB(ctx).Save(entity).Error
}

// Create inserts a new entity.
//
// Example:
//
//	err := r.Create(ctx, &User{UserName: "Alice"})
//	_ = err
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB(ctx).Create(entity).Error
}

// CreateInBatches inserts multiple entities in batches.
//
// Example:
//
//	rows, err := r.CreateInBatches(ctx, []*User{{UserName: "A"}, {UserName: "B"}}, 100)
//	_, _ = rows, err
func (r *BaseRepository[T]) CreateInBatches(ctx context.Context, entities []*T, batchSize int) (int64, error) {
	result := r.DB(ctx).CreateInBatches(entities, batchSize)
	return result.RowsAffected, result.Error
}

// Update updates a single column for matched records.
//
// Example:
//
//	qb := query.New().Where(columns.User.ID.Eq(1))
//	rows, err := r.Update(ctx, qb, columns.User.Status, 2)
//	_, _ = rows, err
func (r *BaseRepository[T]) Update(ctx context.Context, qb *query.Builder, column query.Column, value any) (int64, error) {
	result := r.buildQuery(ctx, qb).Update(column.String(), value)
	return result.RowsAffected, result.Error
}

// Updates updates multiple columns for matched records.
//
// Example:
//
//	qb := query.New().Where(columns.User.ID.Eq(1))
//	rows, err := r.Updates(ctx, qb, map[query.Column]any{columns.User.Status: 2, columns.User.Email: "a@b.com"})
//	_, _ = rows, err
func (r *BaseRepository[T]) Updates(ctx context.Context, qb *query.Builder, values any) (int64, error) {
	if mapVals, ok := values.(map[query.Column]any); ok {
		values = column.ToStringMap(mapVals)
	}
	result := r.buildQuery(ctx, qb).Updates(values)
	return result.RowsAffected, result.Error
}

// Delete deletes matched records.
//
// Example:
//
//	qb := query.New().Where(columns.User.ID.Eq(1))
//	rows, err := r.Delete(ctx, qb)
//	_, _ = rows, err
func (r *BaseRepository[T]) Delete(ctx context.Context, qb *query.Builder) (int64, error) {
	var entity T
	result := r.buildQuery(ctx, qb).Delete(&entity)
	return result.RowsAffected, result.Error
}

// Find returns matched records.
//
// Example:
//
//	qb := query.New().Where(columns.User.Status.Eq(1)).Order(columns.User.CreatedAt.Desc())
//	users, err := r.Find(ctx, qb)
//	_, _ = users, err
func (r *BaseRepository[T]) Find(ctx context.Context, qb *query.Builder) ([]*T, error) {
	var entities []*T
	err := r.buildQuery(ctx, qb).Find(&entities).Error
	return entities, err
}

// First returns the first matched record (primary key ascending).
// When not found, it returns (nil, error).
//
// Example:
//
//	qb := query.New().Where(columns.User.Email.Eq("alice@example.com"))
//	user, err := r.First(ctx, qb)
//	_, _ = user, err
func (r *BaseRepository[T]) First(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Take returns one matched record without specifying order.
// When not found, it returns (nil, error).
//
// Example:
//
//	user, err := r.Take(ctx, query.New().Where(columns.User.ID.Eq(1)))
//	_, _ = user, err
func (r *BaseRepository[T]) Take(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).Take(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Last returns the last matched record (primary key descending).
// When not found, it returns (nil, error).
//
// Example:
//
//	user, err := r.Last(ctx, query.New())
//	_, _ = user, err
func (r *BaseRepository[T]) Last(ctx context.Context, qb *query.Builder) (*T, error) {
	var entity T
	err := r.buildQuery(ctx, qb).Last(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Count returns the count of matched records.
//
// Example:
//
//	n, err := r.Count(ctx, query.New().Where(columns.User.Status.Eq(1)))
//	_, _ = n, err
func (r *BaseRepository[T]) Count(ctx context.Context, qb *query.Builder) (int64, error) {
	var count int64
	err := r.buildQuery(ctx, qb).Count(&count).Error
	return count, err
}

// Pluck selects a single column into dest.
//
// Example:
//
//	var emails []string
//	err := r.Pluck(ctx, query.New().Where(columns.User.Status.Eq(1)), columns.User.Email, &emails)
//	_ = err
func (r *BaseRepository[T]) Pluck(ctx context.Context, qb *query.Builder, column query.Column, dest any) error {
	return r.buildQuery(ctx, qb).Pluck(column.String(), dest).Error
}
