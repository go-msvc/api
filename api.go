package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Api interface {
	Run(addr string) error
}

func New(r R) Api {
	return api{
		r: r,
	}
}

type api struct {
	r R
}

func (api api) Handler(hdlr H) http.HandlerFunc {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		fmt.Printf("HTTP %s %s\n", httpReq.Method, httpReq.URL.Path)

		//call the handler
		hdlr(httpRes, httpReq)

		//http.Error(httpRes, "NYI", http.StatusNotFound)
	}
}

func (api api) Run(addr string) error {
	r := mux.NewRouter()
	for path, m := range api.r {
		for method, hdlr := range m {
			r.HandleFunc(path, api.Handler(hdlr)).Methods(method)
		}
	}
	http.Handle("/", r)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return err
	}
	return nil
}

func (api api) R(path string, methods M) Api {
	return api
}

type R map[string]M
type M map[string]H
type H http.HandlerFunc
