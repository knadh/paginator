package paginator

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginator(t *testing.T) {
	var (
		opt = Opt{
			DefaultPerPage: 10,
			MaxPerPage:     50,
			NumPageNums:    10,
			PageParam:      "page",
			PerPageParam:   "per_page",
			AllowAll:       false,
			AllowAllParam:  "all",
		}

		p = New(opt)
	)

	type testCase struct {
		Page    int
		PerPage int
		Total   int
	}
	type test struct {
		testCase
		result Set
	}
	cases := []test{
		{testCase{Page: 0, PerPage: 5, Total: 0}, Set{PerPage: 5, Page: 1, Offset: 0, Limit: 5}},
		{testCase{Page: -1, PerPage: 5, Total: 0}, Set{PerPage: 5, Page: 1, Offset: 0, Limit: 5}},
		{testCase{Page: 1, PerPage: 5, Total: 0}, Set{PerPage: 5, Page: 1, Offset: 0, Limit: 5}},
		{testCase{Page: 10, PerPage: 5, Total: 100}, Set{PerPage: 5, Page: 10, Offset: 45, Limit: 5}},
		{testCase{Page: 10, PerPage: 10, Total: 100}, Set{PerPage: 10, Page: 10, Offset: 90, Limit: 10}},
		{testCase{Page: -1, PerPage: 10, Total: 100}, Set{PerPage: 10, Page: 1, Offset: 0, Limit: 10}},
	}

	q := url.Values{}
	for _, c := range cases {
		s := p.New(c.testCase.Page, c.testCase.PerPage)
		assert.Equal(t, s.Page, c.result.Page)
		assert.Equal(t, s.PerPage, c.result.PerPage)
		assert.Equal(t, s.Offset, c.result.Offset)
		assert.Equal(t, s.Limit, c.result.Limit)

		q.Set("page", fmt.Sprintf("%v", s.Page))
		q.Set("per_page", fmt.Sprintf("%v", s.PerPage))

		s = p.NewFromURL(q)
		assert.Equal(t, s.Page, c.result.Page)
		assert.Equal(t, s.PerPage, c.result.PerPage)
		assert.Equal(t, s.Offset, c.result.Offset)
		assert.Equal(t, s.Limit, c.result.Limit)
	}

	// Exceed per page with AlloWall = false
	q.Set("page", "1")
	q.Set("per_page", "500")
	s := p.NewFromURL(q)
	assert.Equal(t, s.Page, 1)
	assert.Equal(t, s.PerPage, 50)

	// Exceed per page with AllowAll = true
	opt.AllowAll = true
	p = New(opt)
	s = p.NewFromURL(q)
	assert.Equal(t, s.Page, 1)
	// This is now allowed because AllowAll = true;
	assert.Equal(t, s.PerPage, 500)

	q.Set("per_page", "all")
	s = p.NewFromURL(q)
	assert.Equal(t, s.Page, 1)
	assert.Equal(t, s.PerPage, 0)
}
