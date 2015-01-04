// Package common provides routines common to all handlers in the mortgage
// web app.
package common

import (
  "github.com/keep94/finance/fin"
  "html/template"
  "time"
)

// NewTemplate returns a new template instance. name is the name
// of the template; templateStr is the template string.
func NewTemplate(name, templateStr string) *template.Template {
  return template.Must(template.New(name).Funcs(
      template.FuncMap{
          "FormatDate": formatDate,
          "FormatUSD": fin.FormatUSD}).Parse(templateStr))
}

func formatDate(t time.Time) string {
  return t.Format("01/02/2006")
}

