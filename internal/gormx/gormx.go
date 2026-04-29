package gormx

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func BuildNested[T ~func(*gorm.DB) *gorm.DB](db *gorm.DB, conds []T) *gorm.DB {
	nested := db.Session(&gorm.Session{NewDB: true})
	for _, cond := range conds {
		nested = cond(nested)
	}
	return nested
}

// RelationNames returns association (relationship) field names in a stable,
// field-declaration order, so generated files stay deterministic.
func RelationNames(sch *schema.Schema) []string {
	if sch.Relationships.Relations == nil {
		return nil
	}

	seen := make(map[string]struct{}, len(sch.Relationships.Relations))
	names := make([]string, 0, len(sch.Relationships.Relations))
	for _, field := range sch.Fields {
		if _, ok := sch.Relationships.Relations[field.Name]; !ok {
			continue
		}
		if _, ok := seen[field.Name]; ok {
			continue
		}
		seen[field.Name] = struct{}{}
		names = append(names, field.Name)
	}
	return names
}
