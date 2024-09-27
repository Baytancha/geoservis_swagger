package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ReverseProxy struct {
	host string
	port string
}

func NewReverseProxy(host, port string) *ReverseProxy {
	return &ReverseProxy{
		host: host,
		port: port,
	}
}

// hugo:1313/static -> hugo
// hugo:1313/api -> proxy api

func (rp *ReverseProxy) ReverseProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
			return
		}
		link := fmt.Sprintf("http://%s:%s", rp.host, rp.port)
		uri, _ := url.Parse(link)

		if uri.Host == r.Host {
			next.ServeHTTP(w, r)
			return
		}
		r.Header.Set("Reverse-Proxy", "true")

		//proxy := httputil.NewSingleHostReverseProxy(uri)

		proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
			r.URL.Scheme = uri.Scheme
			r.URL.Host = uri.Host
			r.URL.Path = uri.Path + r.URL.Path
			r.Host = uri.Host
			fmt.Println("CONNECTING....", r.URL)
		}}
		proxy.ServeHTTP(w, r)
	})
}
