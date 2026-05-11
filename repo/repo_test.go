package repo

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type user struct {
	gorm.Model
	UserName string `gorm:"column:user_name;size:255;not null"`
	Email    string `gorm:"column:email;size:255;unique"`
	Age      int    `gorm:"column:age"`
	Status   int    `gorm:"column:status;default:1"`
}

var userSchema = struct {
	ID       query.NumericColumn[uint]
	UserName query.StringColumn[string]
	Email    query.StringColumn[string]
	Age      query.NumericColumn[int]
	Status   query.NumericColumn[int]
}{
	ID:       query.NewNumericColumn[uint]("users", "id"),
	UserName: query.NewStringColumn[string]("users", "user_name"),
	Email:    query.NewStringColumn[string]("users", "email"),
	Age:      query.NewNumericColumn[int]("users", "age"),
	Status:   query.NewNumericColumn[int]("users", "status"),
}

func openRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:repo_%s?mode=memory&cache=shared", name)

	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&user{}))
	return gormDB
}

func seedUsers(t *testing.T, r *BaseRepository[user], ctx context.Context) {
	t.Helper()
	require.NoError(t, r.Create(ctx, &user{UserName: "Alice", Email: "alice@example.com", Age: 25, Status: 1}))
	require.NoError(t, r.Create(ctx, &user{UserName: "Bob", Email: "bob@example.com", Age: 30, Status: 1}))
	require.NoError(t, r.Create(ctx, &user{UserName: "Charlie", Email: "charlie@example.com", Age: 35, Status: 1}))
}

func TestBaseRepository_CRUD(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	seedUsers(t, r, ctx)

	count, err := r.Count(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Find
	users, err := r.Find(ctx, query.New[user]().Where(userSchema.Age.Gte(30)).Order(userSchema.Age.Asc()))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Bob", users[0].UserName)
	assert.Equal(t, "Charlie", users[1].UserName)

	// First
	alice, err := r.First(ctx, query.New[user]().Where(userSchema.UserName.Eq("Alice")))
	require.NoError(t, err)
	require.NotNil(t, alice)
	assert.Equal(t, "alice@example.com", alice.Email)

	// Last (primary key descending)
	last, err := r.Last(ctx, query.New[user]())
	require.NoError(t, err)
	require.NotNil(t, last)
	assert.Equal(t, "Charlie", last.UserName)
}

func TestBaseRepository_NotFoundReturnsNilEntity(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	got, err := r.Take(ctx, query.New[user]().Where(userSchema.UserName.Eq("nobody")))
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, got)
}

func TestBaseRepository_UpdateAndUpdates(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	seedUsers(t, r, ctx)

	rows, err := r.Update(ctx, query.New[user]().Where(userSchema.UserName.Eq("Bob")), userSchema.Age.Set(31))
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	rows, err = r.Updates(ctx, query.New[user]().Where(userSchema.UserName.Eq("Bob")),
		userSchema.Status.Set(2),
		userSchema.Email.Set("bob2@example.com"),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	bob, err := r.First(ctx, query.New[user]().Where(userSchema.UserName.Eq("Bob")))
	require.NoError(t, err)
	require.NotNil(t, bob)
	assert.Equal(t, 31, bob.Age)
	assert.Equal(t, 2, bob.Status)
	assert.Equal(t, "bob2@example.com", bob.Email)
}

func TestBaseRepository_CreateInBatches(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	rows, err := r.CreateInBatches(ctx, []*user{
		{UserName: "A", Email: "a@ex.com", Age: 1, Status: 1},
		{UserName: "B", Email: "b@ex.com", Age: 2, Status: 1},
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), rows)

	count, err := r.Count(ctx, query.New[user]().Where(userSchema.UserName.In([]string{"A", "B"})))
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestBaseRepository_Delete(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	seedUsers(t, r, ctx)

	rows, err := r.Delete(ctx, query.New[user]().Where(userSchema.UserName.Eq("Alice")))
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	count, err := r.Count(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestBaseRepository_DeleteWithoutWhereIsRejectedByGorm(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	_, err := r.Delete(ctx, query.New[user]())
	require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
}

func TestBaseRepository_Pluck(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	seedUsers(t, r, ctx)

	var names []string
	err := r.Pluck(ctx, query.New[user]().Order(userSchema.ID.Asc()), userSchema.UserName, &names)
	require.NoError(t, err)
	require.Equal(t, []string{"Alice", "Bob", "Charlie"}, names)
}

func TestBaseRepository_TransactionCommitAndRollback(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	seedUsers(t, r, ctx)

	// commit
	require.NoError(t, client.Transaction(ctx, func(txCtx context.Context) error {
		return r.Create(txCtx, &user{UserName: "Commit", Email: "c@ex.com", Age: 1, Status: 1})
	}))

	count, err := r.Count(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)

	// rollback
	rollbackErr := fmt.Errorf("rollback")
	err = client.Transaction(ctx, func(txCtx context.Context) error {
		if err := r.Create(txCtx, &user{UserName: "Rollback", Email: "r@ex.com", Age: 1, Status: 1}); err != nil {
			return err
		}
		return rollbackErr
	})
	require.ErrorIs(t, err, rollbackErr)

	count, err = r.Count(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)
}

func TestBaseRepository_Save(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	u := &user{UserName: "Alice", Email: "alice@example.com", Age: 25, Status: 1}
	require.NoError(t, r.Create(ctx, u))

	u.Age = 26
	require.NoError(t, r.Save(ctx, u))

	got, err := r.First(ctx, query.New[user]().Where(userSchema.Email.Eq("alice@example.com")))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 26, got.Age)
}

func TestBaseRepository_FirstTakeLast_SuccessAndNotFound(t *testing.T) {
	gormDB := openRepoTestDB(t)
	client := db.NewClient(gormDB)
	r := New[user](client)
	ctx := context.Background()

	seedUsers(t, r, ctx)

	// Take success
	got, err := r.Take(ctx, query.New[user]().Where(userSchema.UserName.Eq("Bob")))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Bob", got.UserName)

	// First not found
	gotFirst, err := r.First(ctx, query.New[user]().Where(userSchema.UserName.Eq("nobody")))
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, gotFirst)

	// Last not found
	gotLast, err := r.Last(ctx, query.New[user]().Where(userSchema.UserName.Eq("nobody")))
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Nil(t, gotLast)
}
