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

var userSchema = struct {
	ID        NumericColumn[uint]
	CreatedAt TimeColumn
	UpdatedAt TimeColumn
	DeletedAt ValueColumn[gorm.DeletedAt]
	UserName  StringColumn[string]
	Email     StringColumn[string]
	Age       NumericColumn[int]
	Status    NumericColumn[int]
}{
	ID:        NewNumericColumn[uint]("", "id"),
	CreatedAt: NewTimeColumn("", "created_at"),
	UpdatedAt: NewTimeColumn("", "updated_at"),
	DeletedAt: NewValueColumn[gorm.DeletedAt]("", "deleted_at"),
	UserName:  NewStringColumn[string]("", "user_name"),
	Email:     NewStringColumn[string]("", "email"),
	Age:       NewNumericColumn[int]("", "age"),
	Status:    NewNumericColumn[int]("", "status"),
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

func applyFind(t *testing.T, db *gorm.DB, qb *Builder[user]) ([]user, error) {
	t.Helper()
	var out []user
	q := db.Model(&user{})
	if qb != nil {
		q = qb.Apply(q)
	}
	err := q.Find(&out).Error
	return out, err
}

func applyFirst(t *testing.T, db *gorm.DB, qb *Builder[user]) (*user, error) {
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

func applyDelete(t *testing.T, db *gorm.DB, qb *Builder[user]) (int64, error) {
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

	q := New[user]().Where(userSchema.Email.Eq("alice@example.com"))
	alice, err := applyFirst(t, db, q)
	require.NoError(t, err)
	require.NotNil(t, alice)
	assert.Equal(t, "Alice", alice.UserName)
}

func TestPagination_Page(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	q := New[user]().Order(userSchema.CreatedAt.Desc()).Page(1, 2)
	users, err := applyFind(t, db, q)
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "admin", users[0].UserName)
	assert.Equal(t, "David", users[1].UserName)
}

func TestQuery_StringHelpers(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.UserName.HasPrefix("Al")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)

	users, err = applyFind(t, db, New[user]().Where(userSchema.UserName.HasSuffix("lie")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Charlie", users[0].UserName)

	users, err = applyFind(t, db, New[user]().Where(userSchema.UserName.Contains("li")))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].UserName)
	assert.Equal(t, "Charlie", users[1].UserName)

	users, err = applyFind(t, db, New[user]().Where(userSchema.UserName.NotContains("a")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
}

func TestQuery_NotLike(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.UserName.NotLike("%a%")))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
}

func TestQuery_Select_Omit(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	qSelect := New[user]().Select(userSchema.UserName).Where(userSchema.UserName.Eq("Bob"))
	u, err := applyFirst(t, db, qSelect)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "Bob", u.UserName)
	assert.Empty(t, u.Email)

	qOmit := New[user]().Omit(userSchema.Email).Where(userSchema.UserName.Eq("Bob"))
	u, err = applyFirst(t, db, qOmit)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "Bob", u.UserName)
	assert.Empty(t, u.Email)
}

func TestQuery_Distinct(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	q := New[user]().Distinct(userSchema.UserName).Select(userSchema.UserName).Order(userSchema.UserName)
	users, err := applyFind(t, db, q)
	require.NoError(t, err)
	require.Len(t, users, 5)
	assert.Equal(t, "Alice", users[0].UserName)
}

func TestQuery_Between(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.Age.Between(20, 30)))
	require.NoError(t, err)
	require.Len(t, users, 3)
}

func TestQuery_ComparisonOps(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.Age.Gt(30)))
	require.NoError(t, err)
	require.Len(t, users, 2)

	users, err = applyFind(t, db, New[user]().Where(userSchema.Age.Gte(30)))
	require.NoError(t, err)
	require.Len(t, users, 3)

	users, err = applyFind(t, db, New[user]().Where(userSchema.Age.Lt(25)))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "David", users[0].UserName)

	users, err = applyFind(t, db, New[user]().Where(userSchema.Age.Lte(25)))
	require.NoError(t, err)
	require.Len(t, users, 2)
}

func TestQuery_In_NotIn(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	names := []string{"Alice", "Bob"}
	users, err := applyFind(t, db, New[user]().Where(userSchema.UserName.In(names)))
	require.NoError(t, err)
	require.Len(t, users, 2)

	users, err = applyFind(t, db, New[user]().Where(userSchema.UserName.NotIn(names)))
	require.NoError(t, err)
	require.Len(t, users, 3)
}

func TestQuery_Null_NotNull_WithSoftDelete(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.DeletedAt.IsNull()))
	require.NoError(t, err)
	require.Len(t, users, 5)

	alice, err := applyFirst(t, db, New[user]().Where(userSchema.UserName.Eq("Alice")))
	require.NoError(t, err)
	_, err = applyDelete(t, db, New[user]().Where(userSchema.ID.Eq(alice.ID)))
	require.NoError(t, err)

	users, err = applyFind(t, db, New[user]().Unscoped().Where(userSchema.DeletedAt.IsNotNull()))
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)
}

func TestQuery_Or_Not_Derived_EmptyNested(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	qOr := New[user]().Where(userSchema.UserName.Eq("Alice")).Or(userSchema.UserName.Eq("Bob"))
	users, err := applyFind(t, db, qOr)
	require.NoError(t, err)
	require.Len(t, users, 2)

	base := New[user]().Where(userSchema.Status.Eq(1))
	derived := base.Where(userSchema.UserName.Eq("Alice"))
	users, err = applyFind(t, db, base)
	require.NoError(t, err)
	require.Len(t, users, 5)
	users, err = applyFind(t, db, derived)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)

	qEmpty := New[user]().Where(userSchema.Status.Eq(1)).Or().Not()
	users, err = applyFind(t, db, qEmpty)
	require.NoError(t, err)
	require.Len(t, users, 5)

	qNot := New[user]().Not(userSchema.UserName.Eq("Alice"))
	users, err = applyFind(t, db, qNot)
	require.NoError(t, err)
	require.Len(t, users, 4)
}

func TestQuery_Group_Having(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	require.NoError(t, db.Create(&user{UserName: "Frank", Email: "frank@example.com", Age: 25, Status: 1}).Error)

	qGroup := New[user]().Select(userSchema.Age).Group(userSchema.Age).Having("count(*) > ?", 1)
	users, err := applyFind(t, db, qGroup)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, 25, users[0].Age)
}

func TestQuery_Limit_Offset_Unscoped_Order_Joins_Scope(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Limit(2).Order(userSchema.ID))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].UserName)
	assert.Equal(t, "Bob", users[1].UserName)

	users, err = applyFind(t, db, New[user]().Limit(2).Offset(2).Order(userSchema.ID))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "Charlie", users[0].UserName)
	assert.Equal(t, "David", users[1].UserName)

	users, err = applyFind(t, db, New[user]().Page(0, 0))
	require.NoError(t, err)
	require.Len(t, users, 5)

	activeScope := func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) }
	users, err = applyFind(t, db, New[user]().Scope(activeScope))
	require.NoError(t, err)
	require.Len(t, users, 5)

	users, err = applyFind(t, db, New[user]().Order(RawFragment("id DESC")))
	require.NoError(t, err)
	require.Len(t, users, 5)
	assert.Equal(t, "admin", users[0].UserName)

	users, err = applyFind(t, db, New[user]().Order(userSchema.ID))
	require.NoError(t, err)
	assert.Equal(t, "Alice", users[0].UserName)

	users, err = applyFind(t, db, New[user]().Order(userSchema.ID.Desc()))
	require.NoError(t, err)
	assert.Equal(t, "admin", users[0].UserName)
}

func TestColumn_Helpers_AndNeq(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.UserName.Neq("Alice")))
	require.NoError(t, err)
	require.Len(t, users, 4)

	// Qualified / aliased / aggregate SQL fragments.
	assert.Equal(t, "users.age", userSchema.Age.WithTable("users").SQL())
	assert.Equal(t, "DISTINCT age", userSchema.Age.Distinct().SQL())
	assert.Equal(t, "SUM(age)", userSchema.Age.Sum().SQL())
	assert.Equal(t, "COUNT(age)", userSchema.Age.Count().SQL())
	assert.Equal(t, "AVG(age)", userSchema.Age.Avg().SQL())
	assert.Equal(t, "MIN(age)", userSchema.Age.Min().SQL())
}

func TestQuery_NotBetween(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.Age.NotBetween(20, 30)))
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

	users, err := applyFind(t, db, New[user]().Order(userSchema.Age.Asc()))
	require.NoError(t, err)
	require.NotEmpty(t, users)
	assert.Equal(t, 20, users[0].Age)
	assert.Equal(t, 40, users[len(users)-1].Age)

	users, err = applyFind(t, db, New[user]().Order(userSchema.Age.Desc()))
	require.NoError(t, err)
	require.NotEmpty(t, users)
	assert.Equal(t, 40, users[0].Age)
	assert.Equal(t, 20, users[len(users)-1].Age)
}

func TestQuery_SelectHelpers_AsAndAgg(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	qAs := New[user]().Select(userSchema.Email.As("user_name")).Where(userSchema.UserName.Eq("Alice"))
	u, err := applyFirst(t, db, qAs)
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.Equal(t, "alice@example.com", u.UserName)

	qMax := db.Model(&user{})
	qMax = New[user]().Select(userSchema.Age.Max().As("age")).Apply(qMax)
	var out user
	require.NoError(t, qMax.Scan(&out).Error)
	assert.Equal(t, 40, out.Age)
}

func TestQuery_InEmptyDoesNotError(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.UserName.In([]string{})))
	require.NoError(t, err)
	assert.Empty(t, users)
}

// --- Association / Preload ---

type addr struct{}

type profile struct{}

func TestAssociation_Nested(t *testing.T) {
	userProfile := NewAssociation[user, profile]("Profile")
	profileAddr := NewAssociation[profile, addr]("Address")

	// user.Profile.Nested(profile.Address)  -> "Profile.Address"
	nested := userProfile.Nested(profileAddr)
	assert.Equal(t, "Profile.Address", nested.Path())
}

type preloadUser struct {
	gorm.Model
	Name   string
	Orders []preloadOrder
}

type preloadOrder struct {
	gorm.Model
	PreloadUserID uint
	Status        int
}

func openPreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:query_preload_%s?mode=memory&cache=shared", name)

	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&preloadUser{}, &preloadOrder{}))

	return gormDB
}

func TestQuery_Preload_WithConds(t *testing.T) {
	db := openPreloadTestDB(t)

	u1 := &preloadUser{Name: "U1"}
	u2 := &preloadUser{Name: "U2"}
	require.NoError(t, db.Create(u1).Error)
	require.NoError(t, db.Create(u2).Error)

	// u1 has both status 1 and 2 orders; u2 has only status 2.
	require.NoError(t, db.Create([]*preloadOrder{
		{PreloadUserID: u1.ID, Status: 1},
		{PreloadUserID: u1.ID, Status: 2},
		{PreloadUserID: u2.ID, Status: 2},
	}).Error)

	orders := NewAssociation[preloadUser, preloadOrder]("Orders")
	status := NewNumericColumn[int]("", "status")

	qb := New[preloadUser]().Preload(orders,
		status.Eq(1),
		func(tx *gorm.DB) *gorm.DB { return tx.Order("id asc") },
	)

	var out []preloadUser
	err := qb.Apply(db.Model(&preloadUser{})).Order("id asc").Find(&out).Error
	require.NoError(t, err)
	require.Len(t, out, 2)

	assert.Equal(t, "U1", out[0].Name)
	require.Len(t, out[0].Orders, 1)
	assert.Equal(t, 1, out[0].Orders[0].Status)

	assert.Equal(t, "U2", out[1].Name)
	assert.Empty(t, out[1].Orders)
}

// --- Joins ---

type joinUser struct {
	gorm.Model
	Name      string
	ProfileID uint
	Profile   joinProfile
}

type joinProfile struct {
	gorm.Model
	City string
}

func openJoinTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:query_join_%s?mode=memory&cache=shared", name)

	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&joinUser{}, &joinProfile{}))

	return gormDB
}

func seedJoinUsers(t *testing.T, db *gorm.DB) (sf, ny *joinProfile) {
	t.Helper()

	sf = &joinProfile{City: "SF"}
	ny = &joinProfile{City: "NY"}
	require.NoError(t, db.Create(sf).Error)
	require.NoError(t, db.Create(ny).Error)
	require.NoError(t, db.Create([]*joinUser{
		{Name: "Alice", ProfileID: sf.ID},
		{Name: "Bob", ProfileID: ny.ID},
		{Name: "Charlie"},
	}).Error)
	return sf, ny
}

func TestQuery_Joins_Association(t *testing.T) {
	db := openJoinTestDB(t)
	seedJoinUsers(t, db)

	profile := NewAssociation[joinUser, joinProfile]("Profile")

	// LEFT JOIN preserves users without a profile.
	var all []joinUser
	err := New[joinUser]().Joins(profile).Apply(db.Model(&joinUser{})).Order("`join_users`.id asc").Find(&all).Error
	require.NoError(t, err)
	require.Len(t, all, 3)
}

func TestQuery_Joins_AssociationWithConds(t *testing.T) {
	db := openJoinTestDB(t)
	seedJoinUsers(t, db)

	profile := NewAssociation[joinUser, joinProfile]("Profile")
	city := NewStringColumn[string]("", "city")

	var users []joinUser
	err := New[joinUser]().
		Joins(profile, city.Eq("SF")).
		Apply(db.Model(&joinUser{})).
		Order("`join_users`.id asc").
		Find(&users).Error
	require.NoError(t, err)
	// LEFT JOIN with ON city = 'SF': users without matched profile still appear.
	require.Len(t, users, 3)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, uint(1), users[0].Profile.ID)
	assert.Equal(t, uint(0), users[1].Profile.ID)
	assert.Equal(t, uint(0), users[2].Profile.ID)
}

func TestQuery_InnerJoins_Association(t *testing.T) {
	db := openJoinTestDB(t)
	seedJoinUsers(t, db)

	profile := NewAssociation[joinUser, joinProfile]("Profile")

	var users []joinUser
	err := New[joinUser]().InnerJoins(profile).Apply(db.Model(&joinUser{})).Order("`join_users`.id asc").Find(&users).Error
	require.NoError(t, err)
	// INNER JOIN drops Charlie who has no profile.
	require.Len(t, users, 2)
	assert.Equal(t, "Alice", users[0].Name)
	assert.Equal(t, "Bob", users[1].Name)
}

func TestQuery_InnerJoins_AssociationWithConds(t *testing.T) {
	db := openJoinTestDB(t)
	seedJoinUsers(t, db)

	profile := NewAssociation[joinUser, joinProfile]("Profile")
	city := NewStringColumn[string]("", "city")

	var users []joinUser
	err := New[joinUser]().
		InnerJoins(profile, city.Eq("SF")).
		Apply(db.Model(&joinUser{})).
		Find(&users).Error
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].Name)
}

func TestQuery_Preload_WithoutConds(t *testing.T) {
	db := openPreloadTestDB(t)

	u := &preloadUser{Name: "U"}
	require.NoError(t, db.Create(u).Error)
	require.NoError(t, db.Create([]*preloadOrder{
		{PreloadUserID: u.ID, Status: 1},
		{PreloadUserID: u.ID, Status: 2},
	}).Error)

	orders := NewAssociation[preloadUser, preloadOrder]("Orders")
	qb := New[preloadUser]().Preload(orders)

	var out []preloadUser
	err := qb.Apply(db.Model(&preloadUser{})).Find(&out).Error
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Len(t, out[0].Orders, 2)
}

func TestQuery_Select_Empty(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	// No columns: Select must return the receiver unchanged.
	qb := New[user]()
	assert.Same(t, qb, qb.Select())

	// And the resulting query still loads every row + every column.
	users, err := applyFind(t, db, qb)
	require.NoError(t, err)
	require.Len(t, users, 5)
}

func TestColumn_NameAndTable(t *testing.T) {
	bare := NewStringColumn[string]("", "user_name")
	assert.Equal(t, "user_name", bare.Name())
	assert.Equal(t, "", bare.Table())
	assert.Equal(t, "user_name", bare.SQL())

	qual := NewStringColumn[string]("users", "user_name")
	assert.Equal(t, "user_name", qual.Name())
	assert.Equal(t, "users", qual.Table())
	assert.Equal(t, "users.user_name", qual.SQL())
}

func TestStringColumn_LikeAndSet(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	users, err := applyFind(t, db, New[user]().Where(userSchema.UserName.Like("%li%")))
	require.NoError(t, err)
	require.Len(t, users, 2)

	// Set produces an Assignment whose Column matches the bare name.
	a := userSchema.UserName.Set("renamed")
	assert.Equal(t, "user_name", a.Column)
	assert.Equal(t, "renamed", a.Value)
}

func TestBoolColumn_Helpers(t *testing.T) {
	type boolRow struct {
		ID     uint
		Active bool
	}

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:query_bool_%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&boolRow{}))
	require.NoError(t, db.Create([]*boolRow{
		{Active: true}, {Active: true}, {Active: false},
	}).Error)

	active := NewBoolColumn("", "active")

	var trues []boolRow
	require.NoError(t, active.IsTrue()(db.Model(&boolRow{})).Find(&trues).Error)
	assert.Len(t, trues, 2)

	var falses []boolRow
	require.NoError(t, active.IsFalse()(db.Model(&boolRow{})).Find(&falses).Error)
	assert.Len(t, falses, 1)
}

func TestAssignments_ToMap(t *testing.T) {
	t.Run("empty returns nil", func(t *testing.T) {
		assert.Nil(t, Assignments(nil).ToMap())
		assert.Nil(t, Assignments{}.ToMap())
	})

	t.Run("later assignment wins for the same column", func(t *testing.T) {
		got := Assignments{
			userSchema.UserName.Set("first"),
			userSchema.Age.Set(10),
			userSchema.UserName.Set("second"),
		}.ToMap()
		assert.Equal(t, map[string]any{
			"user_name": "second",
			"age":       10,
		}, got)
	})
}

// TestAssociation_ParentMatching confirms that Association satisfies the
// nestable interface via its parentOf marker; this also exercises the marker
// so coverage tracks it.
func TestAssociation_ParentMatching(t *testing.T) {
	a := NewAssociation[preloadUser, preloadOrder]("Orders")
	var n nestable[preloadUser] = a
	assert.Equal(t, "Orders", n.Path())
}

// TestBuilder_Immutability verifies that chaining methods on a Builder never
// mutates the receiver, so the same base builder can be safely reused to
// derive several independent queries without calling Clone.
func TestBuilder_Immutability(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	base := New[user]().Where(userSchema.Status.Eq(1))

	// Two derivations from the same base, no Clone in between.
	adults := base.Where(userSchema.Age.Gte(30))
	minors := base.Where(userSchema.Age.Lt(30))

	// Each derived builder must own a fresh condition slice.
	assert.Len(t, base.conditions, 1, "base must remain unchanged")
	assert.Len(t, adults.conditions, 2)
	assert.Len(t, minors.conditions, 2)

	// Pointer identity: derivations are NEW Builders.
	assert.NotSame(t, base, adults)
	assert.NotSame(t, base, minors)
	assert.NotSame(t, adults, minors)

	// Behavioral check: the two derivations select different rows, and base
	// still selects everyone with Status == 1.
	adultUsers, err := applyFind(t, db, adults)
	require.NoError(t, err)
	assert.Len(t, adultUsers, 3)

	minorUsers, err := applyFind(t, db, minors)
	require.NoError(t, err)
	assert.Len(t, minorUsers, 2)

	allActive, err := applyFind(t, db, base)
	require.NoError(t, err)
	assert.Len(t, allActive, 5)
}

// TestBuilder_NoSharedBackingArray guards against the slice-aliasing bug:
// when bind reuses the parent's underlying array, two siblings derived from
// the same base can stomp on each other's conditions. With proper allocation
// the second sibling's append must not be visible to the first.
func TestBuilder_NoSharedBackingArray(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	base := New[user]().Where(userSchema.Status.Eq(1))

	first := base.Where(userSchema.UserName.Eq("Alice"))
	// If `bind` re-used base's slice, the next append below would overwrite
	// the third slot of `first` and turn its second condition into Bob.
	second := base.Where(userSchema.UserName.Eq("Bob"))

	users, err := applyFind(t, db, first)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)

	users, err = applyFind(t, db, second)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
}

// TestBuilder_ApplyDoesNotMutateBase guards against the side effect where Apply
// appends clauses onto the passed-in *gorm.DB. Two Apply calls sharing the same
// base session must stay isolated: the first derivation's conditions must not
// leak into the second.
func TestBuilder_ApplyDoesNotMutateBase(t *testing.T) {
	db := openTestDB(t)
	seedUsers(t, db)

	base := db.Model(&user{})

	alice := New[user]().Where(userSchema.UserName.Eq("Alice"))
	bob := New[user]().Where(userSchema.UserName.Eq("Bob"))

	var aliceUsers []user
	require.NoError(t, alice.Apply(base).Find(&aliceUsers).Error)
	require.Len(t, aliceUsers, 1)
	assert.Equal(t, "Alice", aliceUsers[0].UserName)

	// If Apply mutated base, this query would carry user_name = "Alice" too
	// and return zero rows.
	var bobUsers []user
	require.NoError(t, bob.Apply(base).Find(&bobUsers).Error)
	require.Len(t, bobUsers, 1)
	assert.Equal(t, "Bob", bobUsers[0].UserName)
}
