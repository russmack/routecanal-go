package routecanal

import (
	"bytes"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
)

type regexRouter struct {
	routes []*route
}

func New() *regexRouter {
	return &regexRouter{}
}

type RegexHandler struct{}

type route struct {
	pattern *regexp.Regexp
	handler Handler
}

type Handler func(http.ResponseWriter, *http.Request, map[string]string) error

func NewRoute() *route {
	return &route{}
}

func (r *route) SetPattern(s string) *route {
	r.pattern = regexp.MustCompile(s)
	return r
}

func (r *route) SetHandler(h Handler) *route {
	r.handler = h
	return r
}

func (h *regexRouter) AddRoute(r *route) {
	h.routes = append(h.routes, r)
}

func (h *regexRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Sort the routes so that they are matched longest first to shortest.
	sort.Slice(h.routes, func(i, j int) bool {
		return h.routes[i].pattern.String() > h.routes[j].pattern.String()
	})

	path := r.URL.Path

	for _, route := range h.routes {
		log.Println("Routing: ", path,
			", checking route:", route.pattern.String())

		if route.pattern.MatchString(path) {
			log.Println("Found route.")

			params := h.ParsePath(path)

			err := route.handler(w, r, params)
			if err != nil {
				log.Println("error:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}
	}

	http.NotFound(w, r)
}

func (h *regexRouter) ParsePath(path string) map[string]string {
	m := make(map[string]string)

	idCounter := 0

	var buf bytes.Buffer

	for _, j := range path {
		if j == '/' {
			if buf.Len() != 0 {
				m[strconv.Itoa(idCounter)] = buf.String()
				idCounter++
				buf.Reset()
			}
		} else {
			buf.WriteString(string(j))
		}
	}

	return m
}
