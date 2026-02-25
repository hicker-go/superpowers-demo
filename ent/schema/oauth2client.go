package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// OAuth2Client holds the schema definition for the OAuth2Client entity.
type OAuth2Client struct {
	ent.Schema
}

// Fields of the OAuth2Client.
func (OAuth2Client) Fields() []ent.Field {
	return []ent.Field{
		field.String("client_id").
			Unique().
			NotEmpty(),
		field.String("client_secret").
			NotEmpty(),
		field.JSON("redirect_uris", []string{}),
	}
}
