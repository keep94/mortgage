// Package static provides static content for the mortgage app.
package static

import (
	"github.com/keep94/toolbox/http_util"
	"net/http"
)

var (
	kThemeCss = `
.lineitem {background-color:#CCCCCC}
.error {color:#FF0000;font-weight:bold;}
`
)

func New() http.Handler {
	result := http.NewServeMux()
	http_util.AddStatic(result, "/theme.css", kThemeCss)
	return result
}
