package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// IdPConnector holds the schema definition for the IdPConnector entity.
type IdPConnector struct {
	ent.Schema
}

// Fields of the IdPConnector.
func (IdPConnector) Fields() []ent.Field {
	return []ent.Field{
		field.String("issuer").
			NotEmpty(),
		field.String("client_id").
			NotEmpty(),
		field.String("client_secret").
			NotEmpty(),
	}
}
