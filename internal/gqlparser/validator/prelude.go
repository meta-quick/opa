package validator

import (
	_ "embed"

	"github.com/meta-quick/opa/internal/gqlparser/ast"
)

//go:embed prelude.graphql
var preludeGraphql string

var Prelude = &ast.Source{
	Name:    "prelude.graphql",
	Input:   preludeGraphql,
	BuiltIn: true,
}
