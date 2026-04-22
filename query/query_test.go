package query

import (
	"fmt"
	"strings"
	"testing"
	"time"

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

var userProps = struct {
	ID        Column
	CreatedAt Column
	UpdatedAt Column
	DeletedAt Column
	UserName  Column
	Email     Column
	Age       Column
	Status    Column
}{
	ID:        "id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
	UserName:  "user_name",
	Email:     "email",
	Age:       "age",
	Status:    "status",
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:query_%s?mode=memory&cache=shared", name)

	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&user{}))

	return gormDB
}

func seedUsers(t *testing.T, db *gorm.DB) {
	t.Helper()

	users := []*user{
		{UserName: "Alice", Email: "alice@example.com", Age: 25, Status: 1},
		{UserName: "Bob", Email: "bob@example.com", Age: 30, Status: 1},
		{UserName: "Charlie", Email: "charlie@example.com", Age: 35, Status: 1},
		{UserName: "David", Email: "david@example.com", Age: 20, Status: 1},
		{UserName: "admin", Email: "admin@example.com", Age: 40, Status: 1},
	}
	for _, u := range users {
		require.NoError(t, db.Create(u).Error)
		// Ensure created_at has sortable differences.
		time.Sleep(1 * time.Millisecond)
	}
}

func applyFind(t *testing.T, db *gorm.DB, qb *Builder) ([]user, error) {
	t.Helper()
	var out []user
	q := db.Model(&user{})
	if qb != nil {
		q = qb.Apply(q)
	}
	err := q.Find(&out).Error
	return out, err
}

func applyFirst(t *testing.T, db *gorm.DB, qb *Builder) (*user, error) {
	t.Helper()
	var out user
	q := db.Model(&user{})
	if qb != nil {
		q = qb.Apply(q)
	}
	err := q.First(&out).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func applyDelete(t *testing.T, db *gorm.DB, qb *Builder) (int64, error) {
	t.Helper()
	q := db.Model(&user{})
	if qb != nil {
		q = qb.Apply(q)
	}
	res := q.Delete(&user{})
	return res.RowsAffected, res.Error
}

func TestTypeSafeColumnUsage(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	q := New().Where(userProps.Email.Eq("alice@example.com"))
	alice, err := applyFirst(t, db, q)
	require.NoError(t, err)
	require.NotNil(t, alice)
	assert.Equal(t, "Alice", alice.UserName)
}

func TestPagination_Page(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	q := New().Order(userProps.CreatedAt.Desc()).Page(1, 2)
	users, err := applyFind(t, db, q)
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "admin", users[0].UserName)
	assert.Equal(t, "David", users[1].UserName)
}

func TestQuery_StringHelpers(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.UserName.HasPrefix("Al")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)

	users, err = applyFind(t, db, New().Where(userProps.UserName.HasSuffix("lie")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Charlie", users[0].UserName)

	users, err = applyFind(t, db, New().Where(userProps.UserName.Contains("li")))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].UserName)
	assert.Equal(t, "Charlie", users[1].UserName)

	users, err = applyFind(t, db, New().Where(userProps.UserName.NotContains("a")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
}

func TestQuery_NotLike(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.UserName.NotLike("%a%")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
}

func TestQuery_Select_Omit(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	qSelect := New().Select(userProps.UserName).Where(userProps.UserName.Eq("Bob"))
	u, err := applyFirst(t, db, qSelect)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "Bob", u.UserName)
	assert.Empty(t, u.Email)

	qOmit := New().Omit(userProps.Email).Where(userProps.UserName.Eq("Bob"))
	u, err = applyFirst(t, db, qOmit)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "Bob", u.UserName)
	assert.Empty(t, u.Email)
}

func TestQuery_Distinct(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	q := New().Distinct(userProps.UserName).Select(userProps.UserName).Order(userProps.UserName)
	users, err := applyFind(t, db, q)
	require.NoError(t, err)
	require.Len(t, users, 5)
	assert.Equal(t, "Alice", users[0].UserName)
}

func TestQuery_Between_AndColumnComparison(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.Age.Between(20, 30)))
	require.NoError(t, err)
	require.Len(t, users, 3)

	users, err = applyFind(t, db, New().Where(userProps.Age.Between(userProps.Age, userProps.Age)))
	require.NoError(t, err)
	require.Len(t, users, 5)

	users, err = applyFind(t, db, New().Where(userProps.UserName.Like(userProps.UserName)))
	require.NoError(t, err)
	require.Len(t, users, 5)
}

func TestQuery_ComparisonOps(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.Age.Gt(30)))
	require.NoError(t, err)
	require.Len(t, users, 2)

	users, err = applyFind(t, db, New().Where(userProps.Age.Gte(30)))
	require.NoError(t, err)
	require.Len(t, users, 3)

	users, err = applyFind(t, db, New().Where(userProps.Age.Lt(25)))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "David", users[0].UserName)

	users, err = applyFind(t, db, New().Where(userProps.Age.Lte(25)))
	require.NoError(t, err)
	require.Len(t, users, 2)

	users, err = applyFind(t, db, New().Where(userProps.Age.Lt(userProps.Age)))
	require.NoError(t, err)
	require.Empty(t, users)

	users, err = applyFind(t, db, New().Where(userProps.Age.Lte(userProps.Age)))
	require.NoError(t, err)
	require.Len(t, users, 5)

	users, err = applyFind(t, db, New().Where(userProps.Age.Gt(userProps.Age)))
	require.NoError(t, err)
	require.Empty(t, users)

	users, err = applyFind(t, db, New().Where(userProps.Age.Gte(userProps.Age)))
	require.NoError(t, err)
	require.Len(t, users, 5)
}

func TestQuery_In_NotIn(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	names := []string{"Alice", "Bob"}
	users, err := applyFind(t, db, New().Where(userProps.UserName.In(names)))
	require.NoError(t, err)
	require.Len(t, users, 2)

	users, err = applyFind(t, db, New().Where(userProps.UserName.NotIn(names)))
	require.NoError(t, err)
	require.Len(t, users, 3)
}

func TestQuery_Null_NotNull_WithSoftDelete(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.DeletedAt.IsNull()))
	require.NoError(t, err)
	require.Len(t, users, 5)

	alice, err := applyFirst(t, db, New().Where(userProps.UserName.Eq("Alice")))
	require.NoError(t, err)
	_, err = applyDelete(t, db, New().Where(userProps.ID.Eq(alice.ID)))
	require.NoError(t, err)

	users, err = applyFind(t, db, New().Unscoped().Where(userProps.DeletedAt.IsNotNull()))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)
}

func TestQuery_Or_Not_Clone_EmptyNested(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	qOr := New().Where(userProps.UserName.Eq("Alice")).Or(userProps.UserName.Eq("Bob"))
	users, err := applyFind(t, db, qOr)
	require.NoError(t, err)
	require.Len(t, users, 2)

	base := New().Where(userProps.Status.Eq(1))
	derived := base.Clone().Where(userProps.UserName.Eq("Alice"))
	users, err = applyFind(t, db, base)
	require.NoError(t, err)
	require.Len(t, users, 5)
	users, err = applyFind(t, db, derived)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)

	qEmpty := New().Where(userProps.Status.Eq(1)).Or().Not()
	users, err = applyFind(t, db, qEmpty)
	require.NoError(t, err)
	require.Len(t, users, 5)

	qNot := New().Not(userProps.UserName.Eq("Alice"))
	users, err = applyFind(t, db, qNot)
	require.NoError(t, err)
	require.Len(t, users, 4)
}

func TestQuery_Group_Having(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	require.NoError(t, db.Create(&user{UserName: "Frank", Email: "frank@example.com", Age: 25, Status: 1}).Error)

	qGroup := New().Select(userProps.Age).Group(userProps.Age).Having("count(*) > ?", 1)
	users, err := applyFind(t, db, qGroup)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, 25, users[0].Age)
}

func TestQuery_Preload_NoRelationShouldError(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	q := New().Preload("Orders")
	_, err := applyFind(t, db, q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Orders")
}

func TestQuery_Limit_Offset_Unscoped_Order_Joins_Scope(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Limit(2).Order(userProps.ID))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].UserName)
	assert.Equal(t, "Bob", users[1].UserName)

	users, err = applyFind(t, db, New().Limit(2).Offset(2).Order(userProps.ID))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Charlie", users[0].UserName)
	assert.Equal(t, "David", users[1].UserName)

	users, err = applyFind(t, db, New().Page(0, 0))
	require.NoError(t, err)
	require.Len(t, users, 5)

	activeScope := func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) }
	users, err = applyFind(t, db, New().Scope(activeScope))
	require.NoError(t, err)
	require.Len(t, users, 5)

	// order variants
	users, err = applyFind(t, db, New().Order("id DESC"))
	require.NoError(t, err)
	require.Len(t, users, 5)
	assert.Equal(t, "admin", users[0].UserName)

	users, err = applyFind(t, db, New().Order(userProps.ID))
	require.NoError(t, err)
	assert.Equal(t, "Alice", users[0].UserName)

	users, err = applyFind(t, db, New().Order(userProps.ID.Desc()))
	require.NoError(t, err)
	assert.Equal(t, "admin", users[0].UserName)

	// joins
	qJoin := New().Joins("JOIN users as u2 ON users.id = u2.id").Where(func(db *gorm.DB) *gorm.DB {
		return db.Where("u2.user_name = ?", "Alice")
	})
	users, err = applyFind(t, db, qJoin)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)
}

func TestColumn_Helpers_AndNeq(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.UserName.Neq("Alice")))
	require.NoError(t, err)
	require.Len(t, users, 4)

	assert.Equal(t, Column("users.age"), userProps.Age.Table("users"))
	assert.Equal(t, Column("DISTINCT age"), userProps.Age.Distinct())
	assert.Equal(t, Column("SUM(age)"), userProps.Age.Sum())
	assert.Equal(t, Column("COUNT(age)"), userProps.Age.Count())
	assert.Equal(t, Column("AVG(age)"), userProps.Age.Avg())
	assert.Equal(t, Column("MIN(age)"), userProps.Age.Min())
}

func TestQuery_NotBetween(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	// NOT BETWEEN 20 AND 30 -> Charlie(35), admin(40)
	users, err := applyFind(t, db, New().Where(userProps.Age.NotBetween(20, 30)))
	require.NoError(t, err)
	require.Len(t, users, 2)
	for _, u := range users {
		if u.Age >= 20 && u.Age <= 30 {
			t.Fatalf("unexpected age %d for user %s", u.Age, u.UserName)
		}
	}
}

func TestQuery_OrderHelpers_AscDesc(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Order(userProps.Age.Asc()))
	require.NoError(t, err)
	require.NotEmpty(t, users)
	assert.Equal(t, 20, users[0].Age)
	assert.Equal(t, 40, users[len(users)-1].Age)

	users, err = applyFind(t, db, New().Order(userProps.Age.Desc()))
	require.NoError(t, err)
	require.NotEmpty(t, users)
	assert.Equal(t, 40, users[0].Age)
	assert.Equal(t, 20, users[len(users)-1].Age)
}

func TestQuery_SelectHelpers_AsAndAgg(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	// Email.As("user_name"): map the email column to the UserName field.
	qAs := New().Select(userProps.Email.As("user_name")).Where(userProps.UserName.Eq("Alice"))
	u, err := applyFirst(t, db, qAs)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "alice@example.com", u.UserName)

	// Max(Age)
	qMax := db.Model(&user{})
	qMax = New().Select(userProps.Age.Max().As("age")).Apply(qMax)
	var out user
	require.NoError(t, qMax.Scan(&out).Error)
	assert.Equal(t, 40, out.Age)
}

func TestQuery_OmitAndOrder_InvalidTypeAreSafe(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	// Omit(123) only produces an invalid column name and should not break normal queries.
	qOmitInvalid := New().Omit(123).Where(userProps.UserName.Eq("Alice"))
	u, err := applyFirst(t, db, qOmitInvalid)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "alice@example.com", u.Email)

	// Order(123) is equivalent to ORDER BY 123 in GORM (a valid constant expression) and should run.
	qOrderInvalid := New().Order(123).Where(userProps.UserName.Eq("Alice"))
	u, err = applyFirst(t, db, qOrderInvalid)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "Alice", u.UserName)
}

func TestQuery_InEmptyDoesNotError(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New().Where(userProps.UserName.In([]string{})))
	require.NoError(t, err)
	assert.Empty(t, users)
}
