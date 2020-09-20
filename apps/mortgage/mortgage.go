package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/context"
	"github.com/keep94/mortgage/apps/mortgage/home"
	"github.com/keep94/mortgage/apps/mortgage/static"
	"github.com/keep94/toolbox/http_util"
	"github.com/keep94/toolbox/logging"
	"github.com/keep94/weblogs"
	"net/http"
)

var (
	fPort string
	fIcon string
)

func main() {
	flag.Parse()
	http.HandleFunc("/", rootRedirect)
	http.Handle("/static/", http.StripPrefix("/static", static.New()))
	if fIcon != "" {
		err := http_util.AddStaticFromFile(
			http.DefaultServeMux, "/images/favicon.ico", fIcon)
		if err != nil {
			fmt.Printf("Icon file not found - %s\n", fIcon)
		}
	}
	http.Handle("/home", &home.Handler{})
	defaultHandler := context.ClearHandler(
		weblogs.HandlerWithOptions(
			http.DefaultServeMux,
			&weblogs.Options{Logger: logging.ApacheCommonLoggerWithLatency()}))
	if err := http.ListenAndServe(fPort, defaultHandler); err != nil {
		fmt.Println(err)
	}
}

func rootRedirect(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http_util.Redirect(w, r, "/home")
	} else {
		http_util.Error(w, http.StatusNotFound)
	}
}

func init() {
	flag.StringVar(&fPort, "http", ":8080", "Port to bind")
	flag.StringVar(&fIcon, "icon", "", "Path to icon file")
}
