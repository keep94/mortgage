// Package common provides routines common to all handlers in the mortgage
// web app.
package common

import (
	"html/template"
	"time"

	"github.com/keep94/mortgage"
)

// NewTemplate returns a new template instance. name is the name
// of the template; templateStr is the template string.
func NewTemplate(name, templateStr string) *template.Template {
	return template.Must(template.New(name).Funcs(
		template.FuncMap{
			"FormatDate": formatDate,
			"FormatUSD":  mortgage.FormatUSD}).Parse(templateStr))
}

func formatDate(t time.Time) string {
	return t.Format("01/02/2006")
}
