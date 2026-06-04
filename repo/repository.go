// Package repo provides a generic GORM repository implementation.
//
// BaseRepository[T] wraps the most common CRUD operations (Create, Update, Find, Count, etc.).
// With db.DBProvider it can automatically pick up transactions from context;
// with query.Builder[T] it can accept type-safe query conditions to avoid magic strings.
package repo

import (
	"context"

	"github.com/im-wmkong/gorm-query/db"
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
	Update(ctx context.Context, qb *query.Builder[T], assigns ...query.Assignment) (int64, error)
	Delete(ctx context.Context, qb *query.Builder[T]) (int64, error)
	Find(ctx context.Context, qb *query.Builder[T]) ([]*T, error)
	First(ctx context.Context, qb *query.Builder[T]) (*T, error)
	Take(ctx context.Context, qb *query.Builder[T]) (*T, error)
	Last(ctx context.Context, qb *query.Builder[T]) (*T, error)
	Count(ctx context.Context, qb *query.Builder[T]) (int64, error)
	Exists(ctx context.Context, qb *query.Builder[T]) (bool, error)
	Pluck(ctx context.Context, qb *query.Builder[T], column query.SQLFragment, dest any) error
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
//	tx := r.DB(ctx)
//	_ = tx
func (r *BaseRepository[T]) DB(ctx context.Context) *gorm.DB {
	return r.dbProvider.DB(ctx)
}

func (r *BaseRepository[T]) buildQuery(ctx context.Context, qb *query.Builder[T]) *gorm.DB {
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
//	u := &User{ID: 1, Name: "Alice"}
//	err := r.Save(ctx, u)
//	_ = err
func (r *BaseRepository[T]) Save(ctx context.Context, entity *T) error {
	return r.DB(ctx).Save(entity).Error
}

// Create inserts a new entity.
//
// Example:
//
//	u := &User{Name: "Alice"}
//	err := r.Create(ctx, u)
//	_ = err
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB(ctx).Create(entity).Error
}

// CreateInBatches inserts multiple entities in batches.
//
// Example:
//
//	users := []*User{{Name: "A"}, {Name: "B"}, {Name: "C"}}
//	rows, err := r.CreateInBatches(ctx, users, 100)
//	_, _ = rows, err
func (r *BaseRepository[T]) CreateInBatches(ctx context.Context, entities []*T, batchSize int) (int64, error) {
	result := r.DB(ctx).CreateInBatches(entities, batchSize)
	return result.RowsAffected, result.Error
}

// Update updates one or more columns for matched records.
// When no assignment is given, it is a no-op and returns (0, nil).
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.ID.Eq(1))
//	rows, err := r.Update(ctx, qb,
//	    schema.User.Status.Set(2),
//	    schema.User.Email.Set("a@b.com"),
//	)
//	_, _ = rows, err
func (r *BaseRepository[T]) Update(ctx context.Context, qb *query.Builder[T], assigns ...query.Assignment) (int64, error) {
	if len(assigns) == 0 {
		return 0, nil
	}
	values := query.Assignments(assigns).ToMap()
	result := r.buildQuery(ctx, qb).Updates(values)
	return result.RowsAffected, result.Error
}

// Delete deletes matched records.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.Eq(0))
//	rows, err := r.Delete(ctx, qb)
//	_, _ = rows, err
func (r *BaseRepository[T]) Delete(ctx context.Context, qb *query.Builder[T]) (int64, error) {
	var entity T
	result := r.buildQuery(ctx, qb).Delete(&entity)
	return result.RowsAffected, result.Error
}

// Find returns matched records.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1))
//	users, err := r.Find(ctx, qb)
//	_, _ = users, err
func (r *BaseRepository[T]) Find(ctx context.Context, qb *query.Builder[T]) ([]*T, error) {
	var entities []*T
	err := r.buildQuery(ctx, qb).Find(&entities).Error
	return entities, err
}

// First returns the first matched record (primary key ascending).
// When not found, it returns (nil, error).
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Email.Eq("a@b.com"))
//	u, err := r.First(ctx, qb)
//	_, _ = u, err
func (r *BaseRepository[T]) First(ctx context.Context, qb *query.Builder[T]) (*T, error) {
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
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1))
//	u, err := r.Take(ctx, qb)
//	_, _ = u, err
func (r *BaseRepository[T]) Take(ctx context.Context, qb *query.Builder[T]) (*T, error) {
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
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1))
//	u, err := r.Last(ctx, qb)
//	_, _ = u, err
func (r *BaseRepository[T]) Last(ctx context.Context, qb *query.Builder[T]) (*T, error) {
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
//	qb := schema.User.Query().Where(schema.User.Status.Eq(1))
//	n, err := r.Count(ctx, qb)
//	_, _ = n, err
func (r *BaseRepository[T]) Count(ctx context.Context, qb *query.Builder[T]) (int64, error) {
	var count int64
	err := r.buildQuery(ctx, qb).Count(&count).Error
	return count, err
}

// Exists reports whether any record matches the given conditions.
//
// Example:
//
//	qb := schema.User.Query().Where(schema.User.Email.Eq("a@b.com"))
//	ok, err := r.Exists(ctx, qb)
//	_, _ = ok, err
func (r *BaseRepository[T]) Exists(ctx context.Context, qb *query.Builder[T]) (bool, error) {
	var dst []int
	result := r.buildQuery(ctx, qb).Select("1").Limit(1).Scan(&dst)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// Pluck selects a single column into dest.
//
// Example:
//
//	var emails []string
//	err := r.Pluck(ctx, schema.User.Query().Where(schema.User.Status.Eq(1)), schema.User.Email, &emails)
//	_ = err
func (r *BaseRepository[T]) Pluck(ctx context.Context, qb *query.Builder[T], column query.SQLFragment, dest any) error {
	return r.buildQuery(ctx, qb).Pluck(column.SQL(), dest).Error
}
