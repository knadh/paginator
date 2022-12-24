// Package paginator provides a simple abstracting for handling pagination
// requests and offset/limit generation for HTTP requests. The most common
// usecase is arbitrary queries that need to be paginated with query params
// coming in from a UI, for instance, /things/all?page=2&per_page=5. paginator
// can parse and sanitize these values and provide offset and limit values that
// can be passed to the database query there by avoiding boilerplate code for
// basic pagination. In addition, it can also generate HTML-ready page number
// series (Google search style).
package paginator

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"strconv"
)

// Opt represents paginator options.
type Opt struct {
	// DefaultPerPage is the default number of items per page.
	DefaultPerPage int

	// MaxPerPage is the number number of items per page. Usually, this has the
	// the same value as PerPage. In some cases, it may be desirable to have
	// a small default value but still allow users to request a larger number.
	MaxPerPage int

	// NumPageNums is the of number of page numbers to generate when
	// generating page numbers to be printed (eg: 1, 2 ... 10 ..)
	NumPageNums int

	// PerPageParam is the name of the query param (in url.Values) from which
	// NewFromURL() will pick up the the per_page value in case it is coming
	// from the frontend.
	PerPageParam string

	// PageParam is the name of the query param (in url.Values) from which
	// NewFromURL() will pick up the current page number.
	PageParam string

	// If this is set to true, `per_page=all` is allowed and LIMIT is set as 0,
	// allowing queries to fetch all records in the database (by typically issuing
	// LIMIT NULL in an SQL query)
	AllowAll bool

	// Query param value for the `page` query to use in NewFromURL() if AllowAll
	// is set to true. Default value is `all`.
	AllowAllParam string
}

// Paginator represents a Paginator instance.
type Paginator struct {
	o Opt
}

// Set represents pagination values for a query.
type Set struct {
	// These values are json tagged in case they need to be embedded
	// in a struct that's sent to the outside world.
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Total      int `json:"total"`

	// Computed values for queries.
	Offset int `json:"-"`
	Limit  int `json:"-"`

	// Fields for rendering page numbers.
	PinFirstPage bool  `json:"-"`
	PinLastPage  bool  `json:"-"`
	Pages        []int `json:"-"`
	pg           *Paginator
}

// Default returns a paginator.Opt with default values set.
func Default() Opt {
	return Opt{
		DefaultPerPage: 10,
		MaxPerPage:     50,
		NumPageNums:    10,
		PageParam:      "page",
		PerPageParam:   "per_page",
		AllowAll:       false,
		AllowAllParam:  "all",
	}
}

// New returns a new Paginator instance.
func New(o Opt) *Paginator {
	if o.AllowAllParam == "" {
		o.AllowAllParam = "all"
	}

	return &Paginator{
		o: o,
	}
}

// NewFromURL returns a new pagination Set by .
func (p *Paginator) NewFromURL(q url.Values) Set {
	var (
		perPage, _ = strconv.Atoi(q.Get("per_page"))
		page, _    = strconv.Atoi(q.Get("page"))
	)

	if q.Get("per_page") == p.o.AllowAllParam {
		perPage = -1
	}

	return p.New(page, perPage)
}

// New returns a page Set.
func (p *Paginator) New(page, perPage int) Set {
	if perPage < 0 && p.o.AllowAll {
		perPage = 0
	} else if perPage < 1 {
		perPage = p.o.DefaultPerPage
	} else if perPage > p.o.MaxPerPage {
		perPage = p.o.MaxPerPage
	}
	if page < 1 {
		page = 1
	}

	return Set{
		Page:    page,
		PerPage: perPage,
		Offset:  (page - 1) * perPage,
		Limit:   perPage,
		pg:      p,
	}
}

// SetTotal sets the total count of results after a Set has been used to fetch
// results. This is necessary to generate page numbers.
func (s *Set) SetTotal(t int) {
	s.Total = t
	s.generateNumbers()
}

// generateNumbers generates page numbers on a Set and fills the .PageFirst,
// .Pages[], and .PageLast values.
func (s *Set) generateNumbers() {
	if s.Total <= s.PerPage {
		return
	}
	numPages := int(math.Ceil(float64(s.Total) / float64(s.PerPage)))
	s.TotalPages = numPages
	half := (s.pg.o.NumPageNums / 2)

	// First and last page numbers to print, half towards the back
	// and half towards the front.
	var (
		first = s.Page - half
		last  = s.Page + half
	)
	if first < 1 {
		first = 1
	}
	if last > numPages {
		last = numPages
	}
	if numPages > s.pg.o.NumPageNums {
		if last < numPages && s.Page <= half {
			last = first + s.pg.o.NumPageNums - 1
		}
		if s.Page > numPages-half {
			first = last - s.pg.o.NumPageNums
		}
	}

	// If first in the page number series isn't 1, pin it.
	if first != 1 {
		s.PinFirstPage = true
	}

	// If last page in the page number series is not the actual last page,
	// pin it.
	if last != numPages {
		s.PinLastPage = true
	}

	s.Pages = make([]int, 0, last-first+1)
	for i := first; i <= last; i++ {
		s.Pages = append(s.Pages, i)
	}
}

// HTML prints pagination as HTML.
func (s *Set) HTML(uri string) string {
	var b bytes.Buffer
	if s.PinFirstPage {
		b.WriteString(`<a class="pg-page-first" href="` + fmt.Sprintf(uri, 1) + `">`)
		b.WriteString("1")
		b.WriteString(`</a> `)
		b.WriteString(`<span class="pg-page-ellipsis-first">...</span> `)
	}
	for _, p := range s.Pages {
		c := ""
		if s.Page == p {
			c = " pg-selected"
		}
		b.WriteString(`<a class="pg-page` + c + `" href="` + fmt.Sprintf(uri, p) + `">`)
		b.WriteString(fmt.Sprintf("%d", p))
		b.WriteString(`</a> `)
	}
	if s.PinLastPage {
		b.WriteString(`<span class="pg-page-ellipsis-last">...</span> `)
		b.WriteString(`<a class="pg-page-last" href="` + fmt.Sprintf(uri, s.TotalPages) + `">`)
		b.WriteString(fmt.Sprintf("%d", s.TotalPages))
		b.WriteString(`</a> `)
	}
	return b.String()
}
