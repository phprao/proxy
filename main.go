package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type handle struct {
	host string
	port string
}

type Service struct {
	auth *handle
	user *handle
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var target *url.URL
	if strings.Contains(r.RequestURI, "/api/stats") {
		target, _ = url.Parse("http://" + s.auth.host + ":" + s.auth.port + "/show/stats")
	} else if strings.Contains(r.RequestURI, "/api/autoCommentList") {
		target, _ = url.Parse("http://" + s.user.host + ":" + s.user.port + "/show/autoCommentList")
	} else {
		fmt.Fprintf(w, "404 Not Found")
		return
	}

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	proxy := &httputil.ReverseProxy{Director: director}

	proxy.ModifyResponse = func(response *http.Response) error {
		cont, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(cont))

		response.Body = ioutil.NopCloser(bytes.NewReader(cont))
		return nil
	}
	proxy.ServeHTTP(w, r)
}

func startServer() {
	// 注册被代理的服务器 (host， port)
	service := &Service{
		auth: &handle{host: "192.168.2.157", port: "8007"},
		user: &handle{host: "192.168.2.157", port: "8007"},
	}
	err := http.ListenAndServe(":8888", service)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

func main() {
	startServer()
}
