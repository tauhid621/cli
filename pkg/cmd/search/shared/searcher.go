package shared

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/pkg/search"
)

type searcher struct {
	host   string
	client *http.Client
}

func NewSearcher(host string, client *http.Client) *searcher {
	return &searcher{
		host:   host,
		client: client,
	}
}

//TODO: Handle pagination
func (s *searcher) Search(query search.Query) (string, error) {
	path := fmt.Sprintf("https://api.%s/search/%s", s.host, query.Kind)

	qs := url.Values{}
	for k, v := range query.Parameters.ListSet() {
		qs.Add(k, v)
	}

	q := strings.Builder{}
	q.WriteString(strings.Join(query.Keywords, " "))
	if !query.Raw {
		for k, v := range query.Qualifiers.ListSet() {
			q.WriteString(fmt.Sprintf(" %s:%s", k, v))
		}
	}
	qs.Add("q", q.String())

	url := fmt.Sprintf("%s?%s", path, qs.Encode())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		//TODO: Have specialized error handling code
		//TODO: Handle search failures due to too long of query
		//TODO: Handle too many results
		//TODO: Handle validation failure
		return "", api.HandleHTTPError(resp)
	}

	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

func (s *searcher) URL(query search.Query) string {
	path := fmt.Sprintf("https://%s/search", s.host)
	qs := url.Values{}
	qs.Add("type", query.Kind)
	for k, v := range query.Parameters.ListSet() {
		qs.Add(k, v)
	}
	q := strings.Builder{}
	q.WriteString(strings.Join(query.Keywords, " "))
	if !query.Raw {
		for k, v := range query.Qualifiers.ListSet() {
			q.WriteString(fmt.Sprintf(" %s:%s", k, v))
		}
	}
	qs.Add("q", q.String())
	url := fmt.Sprintf("%s?%s", path, qs.Encode())
	return url
}
