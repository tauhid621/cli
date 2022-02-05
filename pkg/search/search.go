package search

import "github.com/spf13/pflag"

type Qualifiers map[string]Qualifier

type Parameters struct {
	Order   Parameter
	Page    Parameter
	PerPage Parameter
	Sort    Parameter
}

type Query struct {
	Keywords   []string
	Kind       string
	Paginate   bool
	Parameters Parameters
	Qualifiers Qualifiers
	Raw        bool
}

type Searcher interface {
	Search(Query) (string, error)
	URL(Query) string
}

type Qualifier interface {
	IsSet() bool
	Key() string
	pflag.Value
}

type Parameter = Qualifier

func (q *Qualifiers) ListSet() map[string]string {
	m := map[string]string{}
	for _, v := range *q {
		if v.IsSet() {
			m[v.Key()] = v.String()
		}
	}
	return m
}

func (p *Parameters) ListSet() map[string]string {
	m := map[string]string{}
	if p.Order.IsSet() {
		m[p.Order.Key()] = p.Order.String()
	}
	if p.Page.IsSet() {
		m[p.Page.Key()] = p.Page.String()
	}
	if p.PerPage.IsSet() {
		m[p.PerPage.Key()] = p.PerPage.String()
	}
	if p.Sort.IsSet() {
		m[p.Sort.Key()] = p.Sort.String()
	}
	return m
}
