package test

import (
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/query"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestTypeSafeColumnUsage 测试类型安全列的使用
func TestTypeSafeColumnUsage(t *testing.T) {
	ctx, _, repo := setupTest(t)

	q := query.New().Where(model.UserProps.Email.Eq("alice@example.com"))

	alice, err := repo.First(ctx, q)
	require.NoError(t, err)
	require.NotNil(t, alice)
	assert.Equal(t, "Alice", alice.UserName)
}

// TestPagination 测试分页
func TestPagination(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 排序逻辑:
	// 创建顺序: Alice, Bob, Charlie, David, admin
	// CreatedAt 倒序: admin, David, Charlie, Bob, Alice
	// 查询: 按 CreatedAt 倒序, 第 1 页, 大小 2
	// 期望: admin, David

	q := query.New().
		Order(model.UserProps.CreatedAt.Desc()).
		Page(1, 2)

	pageUsers, err := repo.Find(ctx, q)

	require.NoError(t, err)
	require.Len(t, pageUsers, 2)
	assert.Equal(t, "admin", pageUsers[0].UserName)
	assert.Equal(t, "David", pageUsers[1].UserName)
}

// TestQuery_HasPrefix 测试查询功能 - HasPrefix
func TestQuery_HasPrefix(t *testing.T) {
	ctx, _, repo := setupTest(t)

	qPrefix := query.New().Where(model.UserProps.UserName.HasPrefix("Al"))
	usersPrefix, err := repo.Find(ctx, qPrefix)
	require.NoError(t, err)
	require.Len(t, usersPrefix, 1)
	assert.Equal(t, "Alice", usersPrefix[0].UserName)
}

// TestQuery_HasSuffix 测试查询功能 - HasSuffix
func TestQuery_HasSuffix(t *testing.T) {
	ctx, _, repo := setupTest(t)

	qSuffix := query.New().Where(model.UserProps.UserName.HasSuffix("lie"))
	usersSuffix, err := repo.Find(ctx, qSuffix)
	require.NoError(t, err)
	require.Len(t, usersSuffix, 1)
	assert.Equal(t, "Charlie", usersSuffix[0].UserName)
}

// TestQuery_Contains_NotContains 测试 Contains 和 NotContains 辅助方法
func TestQuery_Contains_NotContains(t *testing.T) {
	ctx, _, repo := setupTest(t)

	qContains := query.New().Where(model.UserProps.UserName.Contains("li"))
	usersContains, err := repo.Find(ctx, qContains)
	require.NoError(t, err)
	require.Len(t, usersContains, 2)
	assert.Equal(t, "Alice", usersContains[0].UserName)
	assert.Equal(t, "Charlie", usersContains[1].UserName)

	qNotContains := query.New().Where(model.UserProps.UserName.NotContains("a"))
	usersNotContains, err := repo.Find(ctx, qNotContains)
	require.NoError(t, err)
	require.Len(t, usersNotContains, 1)
	assert.Equal(t, "Bob", usersNotContains[0].UserName)
}

// TestQuery_NotLike 测试查询功能 - NotLike
func TestQuery_NotLike(t *testing.T) {
	ctx, _, repo := setupTest(t)

	qNotLike := query.New().Where(model.UserProps.UserName.NotLike("%a%"))
	usersNotLike, err := repo.Find(ctx, qNotLike)
	require.NoError(t, err)
	// Alice(a), Charlie(a), David(a), admin(a). 只有 Bob 没有 'a' (等等, "Bob" 确实没有 'a')
	require.Len(t, usersNotLike, 1)
	assert.Equal(t, "Bob", usersNotLike[0].UserName)
}

// TestQuery_Select_Omit 测试查询功能 - Select 和 Omit
func TestQuery_Select_Omit(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 仅选择 UserName
	qSelect := query.New().
		Select(model.UserProps.UserName).
		Where(model.UserProps.UserName.Eq("Bob"))
	userSelect, err := repo.First(ctx, qSelect)
	require.NoError(t, err)
	assert.Equal(t, "Bob", userSelect.UserName)
	assert.Empty(t, userSelect.Email) // Email 应该为空

	// 忽略 Email
	qOmit := query.New().
		Omit(model.UserProps.Email). // 直接传递 Column
		Where(model.UserProps.UserName.Eq("Bob"))
	userOmit, err := repo.First(ctx, qOmit)
	require.NoError(t, err)
	assert.Equal(t, "Bob", userOmit.UserName)
	assert.Empty(t, userOmit.Email) // Email 应该为空
}

// TestQuery_Distinct 测试查询功能 - Distinct
func TestQuery_Distinct(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 获取不重复的用户名
	qDistinct := query.New().
		Distinct(model.UserProps.UserName). // 直接传递 Column
		Order(model.UserProps.UserName).
		Select(model.UserProps.UserName) // 仅选择 UserName 以避免 ID 唯一性

	// 注意: Distinct 通常配合 Scan 到字符串切片或结构体切片使用。
	// BaseService.Find 扫描到 []*User。
	// 如果我们只选择 UserName，其他字段将为空。
	usersDistinct, err := repo.Find(ctx, qDistinct)
	require.NoError(t, err)
	// 我们插入了 5 个具有唯一名称的用户 (Alice, Bob, Charlie, David, admin)
	require.Len(t, usersDistinct, 5)
	assert.Equal(t, "Alice", usersDistinct[0].UserName)
}

// TestQuery_Between 测试查询功能 - Between
func TestQuery_Between(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 年龄在 20 到 30 之间 (Alice 25, Bob 30, David 20)
	// 应该包含 20 和 30。
	q := query.New().Where(model.UserProps.Age.Between(20, 30))
	users, err := repo.Find(ctx, q)
	require.NoError(t, err)
	// Alice(25), Bob(30), David(20) -> 3 个用户
	require.Len(t, users, 3)

	// 测试以 Column 作为参数
	// 例如 Age Between Age AND Age -> 应该返回所有 (Age = Age)
	// 这测试了 Between 中 Column 类型的处理
	qCol := query.New().Where(model.UserProps.Age.Between(model.UserProps.Age, model.UserProps.Age))
	usersCol, err := repo.Find(ctx, qCol)
	require.NoError(t, err)
	require.Len(t, usersCol, 5) // 所有用户

	// 测试 Like 配合 Column
	// 例如 UserName LIKE UserName -> 匹配所有
	qLikeCol := query.New().Where(model.UserProps.UserName.Like(model.UserProps.UserName))
	usersLikeCol, err := repo.Find(ctx, qLikeCol)
	require.NoError(t, err)
	require.Len(t, usersLikeCol, 5)
}

// TestQuery_Neq 测试查询功能 - Neq (Not Equal)
func TestQuery_Neq(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 获取不是 Alice 的用户
	qNeq := query.New().Where(model.UserProps.UserName.Neq("Alice"))
	usersNeq, err := repo.Find(ctx, qNeq)
	require.NoError(t, err)
	// Alice, Bob, Charlie, David, admin -> 5 users
	// Expect 4
	require.Len(t, usersNeq, 4)

	for _, u := range usersNeq {
		assert.NotEqual(t, "Alice", u.UserName)
	}

	// Test Column comparison: Age <> Age -> False for all (should return 0 results if Age is not null)
	// But Age is not null. So WHERE Age <> Age matches nothing.
	qCol := query.New().Where(model.UserProps.Age.Neq(model.UserProps.Age))
	usersCol, err := repo.Find(ctx, qCol)
	require.NoError(t, err)
	require.Empty(t, usersCol)
}

// TestQuery_Comparison 测试查询功能 - Gt, Gte, Lt, Lte
func TestQuery_Comparison(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Gt: Age > 30 (Charlie 35, admin 40)
	qGt := query.New().Where(model.UserProps.Age.Gt(30))
	usersGt, err := repo.Find(ctx, qGt)
	require.NoError(t, err)
	require.Len(t, usersGt, 2)

	// Gte: Age >= 30 (Bob 30, Charlie 35, admin 40)
	qGte := query.New().Where(model.UserProps.Age.Gte(30))
	usersGte, err := repo.Find(ctx, qGte)
	require.NoError(t, err)
	require.Len(t, usersGte, 3)

	// Lt: Age < 25 (David 20)
	qLt := query.New().Where(model.UserProps.Age.Lt(25))
	usersLt, err := repo.Find(ctx, qLt)
	require.NoError(t, err)
	require.Len(t, usersLt, 1)
	assert.Equal(t, "David", usersLt[0].UserName)

	// Lte: Age <= 25 (Alice 25, David 20)
	qLte := query.New().Where(model.UserProps.Age.Lte(25))
	usersLte, err := repo.Find(ctx, qLte)
	require.NoError(t, err)
	require.Len(t, usersLte, 2)

	// Column comparison
	// Age < Age -> False
	qLtCol := query.New().Where(model.UserProps.Age.Lt(model.UserProps.Age))
	usersLtCol, err := repo.Find(ctx, qLtCol)
	require.NoError(t, err)
	require.Empty(t, usersLtCol)

	// Age <= Age -> True (All)
	qLteCol := query.New().Where(model.UserProps.Age.Lte(model.UserProps.Age))
	usersLteCol, err := repo.Find(ctx, qLteCol)
	require.NoError(t, err)
	require.Len(t, usersLteCol, 5)

	// Age > Age -> False
	qGtCol := query.New().Where(model.UserProps.Age.Gt(model.UserProps.Age))
	usersGtCol, err := repo.Find(ctx, qGtCol)
	require.NoError(t, err)
	require.Empty(t, usersGtCol)

	// Age >= Age -> True (All)
	qGteCol := query.New().Where(model.UserProps.Age.Gte(model.UserProps.Age))
	usersGteCol, err := repo.Find(ctx, qGteCol)
	require.NoError(t, err)
	require.Len(t, usersGteCol, 5)
}

// TestQuery_In_NotIn 测试查询功能 - In, NotIn
func TestQuery_In_NotIn(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// In: Alice, Bob
	names := []string{"Alice", "Bob"}
	qIn := query.New().Where(model.UserProps.UserName.In(names))
	usersIn, err := repo.Find(ctx, qIn)
	require.NoError(t, err)
	require.Len(t, usersIn, 2)

	// NotIn: Alice, Bob -> Charlie, David, admin
	qNotIn := query.New().Where(model.UserProps.UserName.NotIn(names))
	usersNotIn, err := repo.Find(ctx, qNotIn)
	require.NoError(t, err)
	require.Len(t, usersNotIn, 3)
}

// TestQuery_Null_NotNull 测试查询功能 - IsNull, IsNotNull
func TestQuery_Null_NotNull(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Create a user with NULL Email (Assuming Email can be null, let's check model)
	// model.User struct has string for Email, usually empty string in Go is not NULL in DB unless pointer.
	// Looking at model/user.go: Email string `gorm:"uniqueIndex;size:128"`
	// It's a string, not *string, so it will be empty string "", not NULL.
	// But let's check if we can insert a NULL using map or raw SQL if we want to test IsNull.
	// Or we can test on DeletedAt which is gorm.DeletedAt (Time pointer wrapper).

	// For this test, let's use DeletedAt field which is nullable.

	// Default users are not deleted, so DeletedAt IS NULL
	qIsNull := query.New().Where(model.UserProps.DeletedAt.IsNull())
	usersIsNull, err := repo.Find(ctx, qIsNull)
	require.NoError(t, err)
	require.Len(t, usersIsNull, 5)

	// Soft delete Alice
	alice, _ := repo.First(ctx, query.New().Where(model.UserProps.UserName.Eq("Alice")))
	repo.Delete(ctx, query.New().Where(model.UserProps.ID.Eq(alice.ID)))

	// Now query Unscoped to find deleted user
	qIsNotNull := query.New().
		Unscoped().
		Where(model.UserProps.DeletedAt.IsNotNull())

	usersIsNotNull, err := repo.Find(ctx, qIsNotNull)
	require.NoError(t, err)
	require.Len(t, usersIsNotNull, 1)
	assert.Equal(t, "Alice", usersIsNotNull[0].UserName)
}

// TestQuery_Or 测试查询功能 - Or
func TestQuery_Or(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Name = Alice OR Name = Bob
	// 使用 Or 方法
	qOr := query.New().Where(
		model.UserProps.UserName.Eq("Alice"),
	).Or(
		model.UserProps.UserName.Eq("Bob"),
	)

	usersOr, err := repo.Find(ctx, qOr)
	require.NoError(t, err)
	// Alice, Bob
	require.Len(t, usersOr, 2)
}

// TestQuery_Clone 确保克隆后的 Builder 不会污染原始查询
func TestQuery_Clone(t *testing.T) {
	ctx, _, repo := setupTest(t)

	base := query.New().Where(model.UserProps.Status.Eq(1))
	derived := base.Clone().Where(model.UserProps.UserName.Eq("Alice"))

	baseUsers, err := repo.Find(ctx, base)
	require.NoError(t, err)
	require.Len(t, baseUsers, 5)

	derivedUsers, err := repo.Find(ctx, derived)
	require.NoError(t, err)
	require.Len(t, derivedUsers, 1)
	assert.Equal(t, "Alice", derivedUsers[0].UserName)
}

// TestQuery_EmptyNestedConditions 测试空 Or/Not 不会追加额外条件
func TestQuery_EmptyNestedConditions(t *testing.T) {
	ctx, _, repo := setupTest(t)

	q := query.New().Where(model.UserProps.Status.Eq(1)).Or().Not()
	users, err := repo.Find(ctx, q)
	require.NoError(t, err)
	require.Len(t, users, 5)
}

// TestQuery_Not 测试查询功能 - Not
func TestQuery_Not(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Not (Name = Alice)
	qNot := query.New().Not(
		model.UserProps.UserName.Eq("Alice"),
	)
	usersNot, err := repo.Find(ctx, qNot)
	require.NoError(t, err)
	// Bob, Charlie, David, admin -> 4 users
	require.Len(t, usersNot, 4)

	// Ensure Alice is not in the list
	for _, u := range usersNot {
		assert.NotEqual(t, "Alice", u.UserName)
	}
}

// TestQuery_Group_Having 测试查询功能 - Group & Having
func TestQuery_Group_Having(t *testing.T) {
	// 准备数据: 添加另一个 Age=25 的用户，以便分组测试
	ctx, _, repo := setupTest(t)
	repo.Create(ctx, &model.User{
		UserName: "Frank",
		Email:    "frank@example.com",
		Age:      25,
		Status:   1,
	})

	// Group by Age, Having Count(*) > 1
	// 应该找到 Age=25 (Alice, Frank)

	qGroup := query.New().
		Select(model.UserProps.Age).
		Group(model.UserProps.Age).
		Having("count(*) > ?", 1)

	// 注意: Find 的结果将只有 Age 字段被填充
	usersGroup, err := repo.Find(ctx, qGroup)
	require.NoError(t, err)

	require.Len(t, usersGroup, 1)
	assert.Equal(t, 25, usersGroup[0].Age)
}

// TestQuery_Preload 测试查询功能 - Preload
func TestQuery_Preload(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// User 模型目前没有定义关联关系。
	// 我们仅仅调用一下 Preload 看看是否 panic，或者 SQL 是否生成 (虽然没有任何效果)
	// 或者我们可以假装有一个关联 "Profile"

	// 这是一个简单的 smoke test，确保代码路径被覆盖
	qPreload := query.New().Preload("Orders") // 假设有 Orders
	// 这会报错: model.User' does not have relation 'Orders'
	// 所以我们应该期望错误

	_, err := repo.Find(ctx, qPreload)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Orders")
}

// TestQuery_Limit_Offset 测试查询功能 - Limit & Offset & Page
func TestQuery_Limit_Offset(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Limit 2
	qLimit := query.New().Limit(2).Order(model.UserProps.ID)
	usersLimit, err := repo.Find(ctx, qLimit)
	require.NoError(t, err)
	require.Len(t, usersLimit, 2)
	assert.Equal(t, "Alice", usersLimit[0].UserName)
	assert.Equal(t, "Bob", usersLimit[1].UserName)

	// Limit 2 Offset 2
	qOffset := query.New().Limit(2).Offset(2).Order(model.UserProps.ID)
	usersOffset, err := repo.Find(ctx, qOffset)
	require.NoError(t, err)
	require.Len(t, usersOffset, 2)
	assert.Equal(t, "Charlie", usersOffset[0].UserName)
	assert.Equal(t, "David", usersOffset[1].UserName)

	// Page edge cases
	qPage0 := query.New().Page(0, 0) // Should default to 1, 10
	usersPage0, err := repo.Find(ctx, qPage0)
	require.NoError(t, err)
	require.Len(t, usersPage0, 5) // Alice, Bob, Charlie, David, admin
}

// TestQuery_Scope 测试查询功能 - Scope
func TestQuery_Scope(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 定义一个 Scope
	activeScope := func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", 1)
	}

	qScope := query.New().Scope(activeScope)
	usersScope, err := repo.Find(ctx, qScope)
	require.NoError(t, err)

	// All seeded users have status 1
	require.Len(t, usersScope, 5)
}

// TestQuery_Unscoped 测试查询功能 - Unscoped
func TestQuery_Unscoped(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 删除 Alice
	alice, _ := repo.First(ctx, query.New().Where(model.UserProps.UserName.Eq("Alice")))
	repo.Delete(ctx, query.New().Where(model.UserProps.ID.Eq(alice.ID)))

	// Normal find - Alice missing
	users, _ := repo.Find(ctx, query.New())
	require.Len(t, users, 4)

	// Unscoped find - Alice present
	usersUnscoped, err := repo.Find(ctx, query.New().Unscoped())
	require.NoError(t, err)
	require.Len(t, usersUnscoped, 5)
}

// TestQuery_Order_Variants 测试 Order (String vs Column)
func TestQuery_Order_Variants(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// String "id DESC"
	qStr := query.New().Order("id DESC")
	usersStr, _ := repo.Find(ctx, qStr)
	assert.Equal(t, "admin", usersStr[0].UserName) // ID 5

	// Column Asc
	qColAsc := query.New().Order(model.UserProps.ID)
	usersColAsc, _ := repo.Find(ctx, qColAsc)
	assert.Equal(t, "Alice", usersColAsc[0].UserName) // ID 1

	// Column Desc
	qColDesc := query.New().Order(model.UserProps.ID.Desc())
	usersColDesc, _ := repo.Find(ctx, qColDesc)
	assert.Equal(t, "admin", usersColDesc[0].UserName) // ID 5
}

// TestJoins 测试 Joins
func TestJoins(t *testing.T) {
	ctx, _, repo := setupTest(t)
	// 虽然没有关联表，但我们可以测试生成的 SQL 不报错，或者测试 self-join 语法
	// SELECT users.* FROM users JOIN users as u2 ON users.id = u2.id

	// 使用 func(db *gorm.DB) *gorm.DB 作为 Where 参数
	q := query.New().Joins("JOIN users as u2 ON users.id = u2.id").Where(func(db *gorm.DB) *gorm.DB {
		return db.Where("u2.user_name = ?", "Alice")
	})
	users, err := repo.Find(ctx, q)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Alice", users[0].UserName)
}

// TestGroupString 测试 Group 字符串参数
func TestGroupString(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Group("age") instead of Column
	q := query.New().Select("age").Group("age").Having("count(*) > ?", 0)
	users, err := repo.Find(ctx, q)
	require.NoError(t, err)
	// Should have distinct ages: 20, 25, 30, 35, 40
	require.Len(t, users, 5)
}

// TestCoverageSupplement 测试 Coverage 补充用例
func TestCoverageSupplement(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 1. Select String
	qSelect := query.New().Select("user_name").Where(model.UserProps.UserName.Eq("Alice"))
	uSelect, err := repo.First(ctx, qSelect)
	require.NoError(t, err)
	assert.Equal(t, "Alice", uSelect.UserName)

	// 2. Omit String
	qOmit := query.New().Omit("email").Where(model.UserProps.UserName.Eq("Alice"))
	uOmit, err := repo.First(ctx, qOmit)
	require.NoError(t, err)
	assert.Empty(t, uOmit.Email)

	// 2.1 Omit Invalid Type (should gracefully degrade)
	qOmitInvalid := query.New().Omit(123).Where(model.UserProps.UserName.Eq("Alice"))
	uOmitInvalid, err := repo.First(ctx, qOmitInvalid)
	require.NoError(t, err)
	assert.Equal(t, "Alice", uOmitInvalid.UserName)
	assert.Equal(t, "alice@example.com", uOmitInvalid.Email)

	// 3. Order Invalid Type (should be ignored)
	qOrder := query.New().Order(123).Where(model.UserProps.UserName.Eq("Alice"))
	uOrder, err := repo.First(ctx, qOrder)
	require.NoError(t, err)
	assert.Equal(t, "Alice", uOrder.UserName)

	// 4. Eq(Column)
	// UserName = UserName -> All
	qEqCol := query.New().Where(model.UserProps.UserName.Eq(model.UserProps.UserName))
	usersEqCol, err := repo.Find(ctx, qEqCol)
	require.NoError(t, err)
	require.Len(t, usersEqCol, 5)

	// 5. NotLike(Column)
	// UserName NOT LIKE UserName -> None
	qNotLikeCol := query.New().Where(model.UserProps.UserName.NotLike(model.UserProps.UserName))
	usersNotLikeCol, err := repo.Find(ctx, qNotLikeCol)
	require.NoError(t, err)
	require.Len(t, usersNotLikeCol, 0)
}

// TestQuery_NotBetween 测试 NotBetween 方法
func TestQuery_NotBetween(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Alice(25), Bob(30), Charlie(35), David(20), admin(40)
	// NotBetween 20 and 30 -> Should be Charlie(35), admin(40)
	// 注意: BETWEEN 是包含边界的 (>= 20 AND <= 30)
	// NOT BETWEEN 是 (< 20 OR > 30)
	// 所以 20 和 30 应该被排除

	q := query.New().Where(model.UserProps.Age.NotBetween(20, 30))
	users, err := repo.Find(ctx, q)
	require.NoError(t, err)

	// Charlie(35), admin(40)
	require.Len(t, users, 2)
	for _, u := range users {
		if u.Age >= 20 && u.Age <= 30 {
			t.Errorf("User %s with age %d should be excluded", u.UserName, u.Age)
		}
	}
}

// TestQuery_Order_Helpers 测试 Desc 和 Asc 方法
func TestQuery_Order_Helpers(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// Asc
	qAsc := query.New().Order(model.UserProps.Age.Asc())
	usersAsc, err := repo.Find(ctx, qAsc)
	require.NoError(t, err)
	require.NotEmpty(t, usersAsc)
	assert.Equal(t, 20, usersAsc[0].Age)               // David
	assert.Equal(t, 40, usersAsc[len(usersAsc)-1].Age) // admin

	// Desc
	qDesc := query.New().Order(model.UserProps.Age.Desc())
	usersDesc, err := repo.Find(ctx, qDesc)
	require.NoError(t, err)
	require.NotEmpty(t, usersDesc)
	assert.Equal(t, 40, usersDesc[0].Age)                // admin
	assert.Equal(t, 20, usersDesc[len(usersDesc)-1].Age) // David
}

// TestQuery_Select_Helpers 测试 Select 相关的辅助方法
func TestQuery_Select_Helpers(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 测试 As (别名)
	// 将 user_name 别名为 name (但这在 GORM Model 映射中可能不会自动生效，除非 Result 结构体有匹配的 tag)
	// 这里我们反过来测试：将 user_name 别名为 user_name (无操作) 或者是 email 别名为 dummy
	// 为了验证效果，我们可以 Select("user_name AS name") 然后看日志，或者更简单的，验证方法本身返回字符串正确
	// 由于集成测试主要看 DB 行为，我们尝试用 As 别名到一个存在的字段，看是否能读取。
	// 例如：Select(Email.As("user_name")) -> 此时 User.UserName 应该填充的是 Email 的值

	qAs := query.New().
		Select(model.UserProps.Email.As("user_name")).
		Where(model.UserProps.UserName.Eq("Alice"))

	userAs, err := repo.First(ctx, qAs)
	require.NoError(t, err)
	// UserName 应该被填充为 Email 的值 "alice@example.com"
	assert.Equal(t, "alice@example.com", userAs.UserName)

	// 测试聚合函数 (Count, Max, Min, Avg, Sum)
	// 我们通常需要 scan 到特定的结构体，但 repo.Find 是 scan 到 []*User
	// 我们可以利用 Pluck 或者 Smart Select，但 BaseRepository 可能不支持任意 struct scan
	// 这里我们通过 Count() 的返回值作为 Select 参数，确保 SQL 执行无误即可

	// Max Age
	// Select MAX(age) as age
	qMax := query.New().
		Select(model.UserProps.Age.Max().As("age")).
		Where(model.UserProps.UserName.Neq("Unknown")) // Dummy where

	userMax, err := repo.First(ctx, qMax)
	require.NoError(t, err)
	assert.Equal(t, 40, userMax.Age) // admin(40) is max

	// Min Age
	qMin := query.New().
		Select(model.UserProps.Age.Min().As("age"))

	userMin, err := repo.First(ctx, qMin)
	require.NoError(t, err)
	assert.Equal(t, 20, userMin.Age) // David(20) is min

	// Sum Age (25+30+35+20+40 = 150)
	// sum(age) -> age
	qSum := query.New().
		Select(model.UserProps.Age.Sum().As("age"))

	userSum, err := repo.First(ctx, qSum)
	require.NoError(t, err)
	assert.Equal(t, 150, userSum.Age)

	// Count ID
	qCount := query.New().
		Select(model.UserProps.ID.Count().As("age"))

	userCount, err := repo.First(ctx, qCount)
	require.NoError(t, err)
	assert.Equal(t, 5, userCount.Age)

	// Avg Age
	qAvg := query.New().
		Select(model.UserProps.Age.Avg().As("age"))

	userAvg, err := repo.First(ctx, qAvg)
	require.NoError(t, err)
	assert.Equal(t, 30, userAvg.Age)

	// Distinct helper
	err = repo.Create(ctx, &model.User{
		UserName: "Frank",
		Email:    "frank@example.com",
		Age:      25,
		Status:   1,
	})
	require.NoError(t, err)

	qDistinct := query.New().
		Select(model.UserProps.Age.Distinct().As("age")).
		Order(model.UserProps.Age)

	usersDistinct, err := repo.Find(ctx, qDistinct)
	require.NoError(t, err)
	require.Len(t, usersDistinct, 5)
	assert.Equal(t, 20, usersDistinct[0].Age)
	assert.Equal(t, 25, usersDistinct[1].Age)
	assert.Equal(t, 30, usersDistinct[2].Age)
	assert.Equal(t, 35, usersDistinct[3].Age)
	assert.Equal(t, 40, usersDistinct[4].Age)

	// Table helper
	qTable := query.New().
		Select(model.UserProps.UserName.Table("users").As("user_name")).
		Where(model.UserProps.UserName.Eq("Alice"))

	userTable, err := repo.First(ctx, qTable)
	require.NoError(t, err)
	assert.Equal(t, "Alice", userTable.UserName)
}
