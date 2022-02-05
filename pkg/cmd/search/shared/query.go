package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cli/cli/v2/pkg/search"
)

var rangeRE = regexp.MustCompile(`(>|>=|<|<=|\*\.\.)?\d+(\.\.\*|\.\.\d+)?`)

func NewSearchQuery(kind string) search.Query {
	params := search.Query{
		Kind: kind,
		Parameters: search.Parameters{
			Order:   newParameter("order", "string", "desc", optsValidator([]string{"asc", "desc"})),
			Page:    newParameter("page", "int", "0", intValidator()),
			PerPage: newParameter("per_page", "int", "30", maxValidator(100)),
			Sort:    newParameter("sort", "string", "up", optsValidator([]string{"up", "down"})), // Variable based on kind
		},
		Qualifiers: search.Qualifiers{
			"Archived":         newQualifier("archived", "bool", "", boolValidator()),
			"Created":          newQualifier("created", "string", "", nil),
			"Followers":        newQualifier("followers", "string", "", rangeValidator()),
			"Fork":             newQualifier("fork", "string", "", optsValidator([]string{})),
			"Forks":            newQualifier("forks", "string", "", rangeValidator()),
			"GoodFirstIssues":  newQualifier("good-first-issues", "string", "", rangeValidator()),
			"HelpWantedIssues": newQualifier("help-wanted-issues", "string", "", rangeValidator()),
			"In":               newQualifier("in", "string", "name,descripton", optsValidator([]string{"name", "description", "readme"})), // Variable based on kind
			"Language":         newQualifier("language", "string", "", nil),
			"License":          newQualifier("license", "string", "", nil),
			"Mirror":           newQualifier("mirror", "bool", "", boolValidator()),
			"Org":              newQualifier("org", "string", "", nil),
			"Pushed":           newQualifier("pushed", "string", "", nil),
			"Repo":             newQualifier("repo", "string", "", nil),
			"Size":             newQualifier("size", "string", "", rangeValidator()),
			"Stars":            newQualifier("fork", "string", "", rangeValidator()),
			"Topic":            newQualifier("topic", "string", "", nil),
			"Topics":           newQualifier("topics", "string", "", rangeValidator()),
			"User":             newQualifier("user", "string", "", nil),
			"Visibility":       newQualifier("is", "string", "public", optsValidator([]string{"public", "private"})),
		},
	}

	return params
}

type validator func(string) error

type qualifier struct {
	key       string
	kind      string
	set       bool
	validator validator
	value     string
}

type parameter = qualifier

func newQualifier(key, kind, value string, validator func(string) error) *qualifier {
	return &qualifier{
		key:       key,
		kind:      kind,
		validator: validator,
		value:     value,
	}
}

func newParameter(key, kind, value string, validator func(string) error) *parameter {
	return &parameter{
		key:       key,
		kind:      kind,
		validator: validator,
		value:     value,
	}
}

func (q *qualifier) IsSet() bool {
	return q.set
}

func (q *qualifier) Key() string {
	return q.key
}

func (q *qualifier) Set(v string) error {
	if q.validator != nil {
		err := q.validator(v)
		if err != nil {
			return err
		}
	}
	q.set = true
	q.value = v
	return nil
}

func (q *qualifier) String() string {
	return q.value
}

func (q *qualifier) Type() string {
	return q.kind
}

// Validate that value is one of a list of options
func optsValidator(opts []string) validator {
	return func(v string) error {
		// TODO: Split v on comma and make sure all values are included
		// TODO: validating multiple values should be another validator...
		if !isIncluded(v, opts) {
			return fmt.Errorf("%s is not included in %s", v, strings.Join(opts, ", "))
		}
		return nil
	}
}

func isIncluded(v string, opts []string) bool {
	for _, opt := range opts {
		if v == opt {
			return true
		}
	}
	return false
}

// Validate that value is less than or equal to max
func maxValidator(max int) validator {
	return func(v string) error {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s is not an integer", v)
		}
		if i > max {
			return fmt.Errorf("%d is larger than the maximum %d", i, max)
		}
		return nil
	}
}

// Validate that value is an integer
func intValidator() validator {
	return func(v string) error {
		_, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s is not an integer value", v)
		}
		return nil
	}
}

// Validate that value is a boolean
func boolValidator() validator {
	return func(v string) error {
		_, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("%s is not a boolean value", v)
		}
		return nil
	}
}

// Validate that value is a correct range format
func rangeValidator() validator {
	return func(v string) error {
		if !rangeRE.MatchString(v) {
			return fmt.Errorf("%s is invalid format", v)
		}
		return nil
	}
}
