package errx

import (
	"fmt"
	"strings"
)

type NamedError struct {
	Field string
	Err   error
}

func (ve NamedError) Error() string {
	return fmt.Sprintf("{%s} - %s", ve.Field, ve.Err.Error())
}

type NamedErrors []NamedError

func (vs NamedErrors) Error() string {
	if len(vs) == 0 {
		return ""
	}
	err := strings.Builder{}
	for i, ve := range vs {
		err.WriteString(ve.Error())
		if i > 0 {
			err.WriteString("; ")
		}
	}
	return err.String()
}

func (vs *NamedErrors) Add(ve NamedError) {
	*vs = append(*vs, ve)
}

func (vs NamedErrors) Empty() bool {
	return len(vs) == 0
}
