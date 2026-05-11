package gormx

import (
	"reflect"

	"gorm.io/gorm/schema"
)

// Relation describes a single GORM association in stable declaration order.
type Relation struct {
	// Name is the field name on the parent model.
	Name string
	// ChildType is the model type referenced by the association.
	ChildType reflect.Type
}

// OrderedRelations returns the associations of sch sorted by the declaration
// order of the corresponding field on the owning model. Iterating
// sch.Relationships.Relations directly is non-deterministic because it is a
// map; this helper produces a stable order suitable for code generation and
// any other use case that benefits from determinism.
//
// It returns nil when sch is nil or has no associations.
func OrderedRelations(sch *schema.Schema) []Relation {
	if sch == nil || len(sch.Relationships.Relations) == 0 {
		return nil
	}
	out := make([]Relation, 0, len(sch.Relationships.Relations))
	seen := make(map[string]struct{}, len(sch.Relationships.Relations))
	for _, field := range sch.Fields {
		rel, ok := sch.Relationships.Relations[field.Name]
		if !ok {
			continue
		}
		if _, dup := seen[field.Name]; dup {
			continue
		}
		seen[field.Name] = struct{}{}
		out = append(out, Relation{Name: field.Name, ChildType: rel.FieldSchema.ModelType})
	}
	return out
}
