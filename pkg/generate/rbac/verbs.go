package rbac

import (
	"errors"
	"fmt"
)

type Verbs []string

var ErrInvalidVerb = errors.New("verb is invalid")

// Validate validates that the verbs are valid.
func (verbs *Verbs) Validate() error {

OUTER:
	for _, verb := range *verbs {
		var isValid bool

		for _, valid := range DefaultResourceVerbs() {
			if verb == valid {
				isValid = true

				continue OUTER
			}
		}

		if !isValid {
			return fmt.Errorf("%w : %s", ErrInvalidVerb, verb)
		}
	}

	return nil
}

// DefaultResourceVerbs is a helper function to define the default verbs that are allowed
// for resources that are managed by the scaffolded controller.
func DefaultResourceVerbs() []string {
	return []string{
		"get", "list", "watch", "create", "update", "patch", "delete",
	}
}

// defaultStatusVerbs is a helper function to define the default verbs which get placed
// onto resources that have a /status suffix.
func defaultStatusVerbs() []string {
	return []string{
		"get", "update", "patch",
	}
}
