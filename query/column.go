package query

import (
	"fmt"

	"gorm.io/gorm"
)

// Column 代表数据库中的列名
type Column string

var _ fmt.Stringer = (*Column)(nil)

// String 转换为字符串 (Field)
func (c Column) String() string {
	return string(c)
}

// Eq 等于 (Field = Value)
func (c Column) Eq(val any) Condition {
	return c.compare("=", val)
}

// Neq 不等于 (Field <> Value)
func (c Column) Neq(val any) Condition {
	return c.compare("<>", val)
}

// Gt 大于 (Field > Value)
func (c Column) Gt(val any) Condition {
	return c.compare(">", val)
}

// Gte 大于等于 (Field >= Value)
func (c Column) Gte(val any) Condition {
	return c.compare(">=", val)
}

// Lt 小于 (Field < Value)
func (c Column) Lt(val any) Condition {
	return c.compare("<", val)
}

// Lte 小于等于 (Field <= Value)
func (c Column) Lte(val any) Condition {
	return c.compare("<=", val)
}

// Like 模糊查询 (Field LIKE Value)
func (c Column) Like(val any) Condition {
	return c.compare("LIKE", val)
}

// NotLike 模糊查询否定 (Field NOT LIKE Value)
func (c Column) NotLike(val any) Condition {
	return c.compare("NOT LIKE", val)
}

// Contains 包含匹配 (Field LIKE %Value%)
func (c Column) Contains(val string) Condition {
	return c.compare("LIKE", "%"+val+"%")
}

// NotContains 不包含匹配 (Field NOT LIKE %Value%)
func (c Column) NotContains(val string) Condition {
	return c.compare("NOT LIKE", "%"+val+"%")
}

// HasPrefix 前缀匹配 (Field LIKE Value%)
func (c Column) HasPrefix(val string) Condition {
	return c.compare("LIKE", val+"%")
}

// HasSuffix 后缀匹配 (Field LIKE %Value)
func (c Column) HasSuffix(val string) Condition {
	return c.compare("LIKE", "%"+val)
}

// In 包含 (Field IN Values)
func (c Column) In(vals any) Condition {
	return c.compare("IN", vals)
}

// NotIn 不包含 (Field NOT IN Values)
func (c Column) NotIn(vals any) Condition {
	return c.compare("NOT IN", vals)
}

// Between 范围匹配 (Field BETWEEN Start AND End)
func (c Column) Between(start, end any) Condition {
	return c.between("BETWEEN", start, end)
}

// NotBetween 范围匹配否定 (Field NOT BETWEEN Start AND End)
func (c Column) NotBetween(start, end any) Condition {
	return c.between("NOT BETWEEN", start, end)
}

// IsNull 为空 (Field IS NULL)
func (c Column) IsNull() Condition {
	return c.clause("IS NULL")
}

// IsNotNull 不为空 (Field IS NOT NULL)
func (c Column) IsNotNull() Condition {
	return c.clause("IS NOT NULL")
}

// Desc 降序 (Field DESC) (用于 Order By)
func (c Column) Desc() string {
	return c.String() + " DESC"
}

// Asc 升序 (Field ASC) (用于 Order By)
func (c Column) Asc() string {
	return c.String() + " ASC"
}

// Table 表名 (Table.Field) (用于 Select)
func (c Column) Table(name string) Column {
	return Column(name + "." + c.String())
}

// As 别名 (Field AS Alias) (用于 Select)
func (c Column) As(alias string) Column {
	return Column(c.String() + " AS " + alias)
}

// Sum 求和 (SUM(Field)) (用于 Select)
func (c Column) Sum() Column {
	return Column("SUM(" + c.String() + ")")
}

// Count 计数 (COUNT(Field)) (用于 Select)
func (c Column) Count() Column {
	return Column("COUNT(" + c.String() + ")")
}

// Avg 平均值 (AVG(Field)) (用于 Select)
func (c Column) Avg() Column {
	return Column("AVG(" + c.String() + ")")
}

// Max 最大值 (MAX(Field)) (用于 Select)
func (c Column) Max() Column {
	return Column("MAX(" + c.String() + ")")
}

// Min 最小值 (MIN(Field)) (用于 Select)
func (c Column) Min() Column {
	return Column("MIN(" + c.String() + ")")
}

func (c Column) compare(op string, val any) Condition {
	return c.clause(op+" ?", val)
}

func (c Column) between(op string, start, end any) Condition {
	return c.clause(op+" ? AND ?", start, end)
}

func (c Column) clause(suffix string, args ...any) Condition {
	return func(db *gorm.DB) *gorm.DB {
		for i, arg := range args {
			if col, ok := arg.(Column); ok {
				args[i] = gorm.Expr(col.String())
			}
		}
		return db.Where(c.String()+" "+suffix, args...)
	}
}
