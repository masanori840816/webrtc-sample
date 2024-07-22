package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
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
	//serverUrl := "https://goapp.sample.jp/webrtc"
	//target := "/webrtc"
	serverUrl := "http://localhost:8083"
	target := ""
	conn := NewWebRTCConnection()
	http.Handle(fmt.Sprintf("%s/css/", target), http.StripPrefix(target, http.FileServer(http.Dir("templates"))))
	http.Handle(fmt.Sprintf("%s/js/", target), http.StripPrefix(target, http.FileServer(http.Dir("templates"))))

	http.HandleFunc(fmt.Sprintf("%s/websocket", target), conn.websocketHandler)
	http.Handle("/", &templateHandler{filename: "index.html", serverUrl: serverUrl})

	go func() {
		for range time.NewTicker(time.Second * 3).C {
			conn.dispatchKeyFrame()
		}
	}()
	log.Fatal(http.ListenAndServe("localhost:8083", nil))
}
