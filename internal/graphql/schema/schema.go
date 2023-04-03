package schema

//go:generate sh ./tools/make_bundle.sh

// Schema provides textual representation of the GraphQL schema content.
func Schema() string {
	return schema
}
