package validator

import (
	"github.com/meta-quick/opa/internal/gqlparser/ast"

	//nolint:revive // Validator rules each use dot imports for convenience.
	. "github.com/meta-quick/opa/internal/gqlparser/validator"
)

func init() {
	AddRule("UniqueFragmentNames", func(observers *Events, addError AddErrFunc) {
		seenFragments := map[string]bool{}

		observers.OnFragment(func(walker *Walker, fragment *ast.FragmentDefinition) {
			if seenFragments[fragment.Name] {
				addError(
					Message(`There can be only one fragment named "%s".`, fragment.Name),
					At(fragment.Position),
				)
			}
			seenFragments[fragment.Name] = true
		})
	})
}
