package gormx

import "gorm.io/gorm/schema"

// TableName returns the explicit table name declared on sch, falling back
// to the provided naming strategy when sch.Table is empty.
func TableName(sch *schema.Schema, ns schema.Namer) string {
	if sch.Table != "" {
		return sch.Table
	}
	return ns.TableName(sch.Name)
}
