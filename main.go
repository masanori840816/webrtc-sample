package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

type templateHandler struct {
	once      sync.Once
	filename  string
	templ     *template.Template
	serverUrl string
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, t.serverUrl)
}

func main() {
	serverUrl := "https://nginx.sample.jp/webrtc"
	target := getStrippingTargetPrefix(serverUrl)
	http.Handle(fmt.Sprintf("/%s/css/", target), http.StripPrefix(fmt.Sprintf("/%s", target), http.FileServer(http.Dir("templates"))))
	http.Handle(fmt.Sprintf("/%s/js/", target), http.StripPrefix(fmt.Sprintf("/%s", target), http.FileServer(http.Dir("templates"))))

	http.HandleFunc("/websocket", websocketHandler)
	http.Handle("/", &templateHandler{filename: "index.html", serverUrl: serverUrl})
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func getStrippingTargetPrefix(url string) string {
	sURL := strings.Split(url, "/")
	if len(sURL) <= 3 {
		return ""
	}
	for i := len(sURL) - 1; i >= 3; i-- {
		if sURL[i] != "" {
			return sURL[i]
		}
	}
	return ""
}
