package home

import (
	"errors"
	"github.com/keep94/mortgage"
	"github.com/keep94/mortgage/apps/mortgage/common"
	"github.com/keep94/toolbox/http_util"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

const (
	kRowsAtTop = 1
)

var (
	kTemplateSpec = `
<html>
<head>
  <title>Mortgage</title>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
  <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
</head>
<body>
<h2>Mortgage</h2>
{{if .Error}}
  <span class="error">{{.Error}}</span>
{{end}}
<form>
  <table>
    <tr>
      <td>Start Date:</td>
      <td>
        <select name="month">
{{with .GetSelection .Months "month"}}
          <option value="{{.Value}}">{{.Name}}</option>
{{else}}
          <option value="">--Pick one--</option>
{{end}}
{{range .Months.Items}}
          <option value="{{.Value}}">{{.Name}}</option>
{{end}}
        </select>
        <input type="text" name="year" value="{{.Get "year"}}">
      </td>
    </tr>
    <tr>
     <td>Amount: </td>
     <td><input type="text" name="amount" value="{{.Get "amount"}}"></td>
    </tr>
    <tr>
     <td>Rate: </td>
     <td><input type="text" name="rate" value="{{.Get "rate"}}">%</td>
    </tr>
    <tr>
     <td>Term in months: </td>
     <td><input type="text" name="term" value="{{.Get "term"}}"></td>
    </tr>
  </table>
  <input type="submit" name="calculate" value="Calculate">
</form>
{{if .Loan}}
<table>
  <tr>
    <td><b>Monthly payment:</b></td>
    <td>{{FormatUSD .Loan.Payment}}</td>
  </tr>
  <tr>
    <td><b>&nbsp;</b></td>
    <td>&nbsp;</td>
  </tr>
  <tr>
    <td><b>Total Cost:</b></td>
    <td>{{FormatUSD .Totals.Payment}}</td>
  </tr>
  <tr>
    <td><b>Total Finance Charges:</b></td>
    <td>{{FormatUSD .Totals.Interest}}</td>
  </tr>
</table>
<br>
<table>
  <tr class="lineitem">
    <td><b>Date</b></td>
    <td><b>Payment</b></td>
    <td><b>Interest</b></td>
    <td><b>Principal</b></td>
    <td><b>Balance</b></td>
  </tr>
  {{range $year, $ytotals := .Totals.Years}}
    {{range $ytotals.Terms}}
  <tr>
    <td>{{FormatDate .Date}}</td>
    <td>{{FormatUSD .Payment}}</td>
    <td>{{FormatUSD .Interest}}</td>
    <td>{{FormatUSD .Principal}}</td>
    <td>{{FormatUSD .Balance}}</td>
  </tr>
    {{end}}
  <tr class="lineitem">
    <td><b>{{$year}} Totals</b></td>
    <td><b>{{FormatUSD $ytotals.Payment}}</b></td>
    <td><b>{{FormatUSD $ytotals.Interest}}</b></td>
    <td><b>{{FormatUSD $ytotals.Principal}}</b></td>
    <td>&nbsp</td>
  </tr>
  {{end}}
{{end}}
</table>

</body>
</html>`
)

var (
	kTemplate *template.Template
)

var (
	kMaxTerms = 1200
	kMonths   = http_util.ComboBox{
		{"January", 1},
		{"February", 2},
		{"March", 3},
		{"April", 4},
		{"May", 5},
		{"June", 6},
		{"July", 7},
		{"August", 8},
		{"September", 9},
		{"October", 10},
		{"November", 11},
		{"December", 12},
	}
)

type Handler struct {
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if hasNoneOf(r.Form, "month", "year", "amount", "rate", "term") {
		http_util.WriteTemplate(
			w,
			kTemplate,
			&view{
				Values: http_util.Values{r.Form},
				Months: kMonths,
			})
		return
	}
	month, _ := kMonths.ToValue(r.Form.Get("month")).(int)
	year, _ := strconv.Atoi(r.Form.Get("year"))
	amount, _ := mortgage.ParseUSD(r.Form.Get("amount"))
	rate, _ := strconv.ParseFloat(r.Form.Get("rate"), 64)
	rate /= 100.0
	length, _ := strconv.Atoi(r.Form.Get("term"))
	var message string
	if month == 0 {
		message = "Please choose a month."
	} else if year < 1900 || year > 2200 {
		message = "Please enter a year between 1900 and 2200"
	} else if amount <= 0 {
		message = "Please enter a positive amount."
	} else if length <= 0 {
		message = "Please enter a positive term in months."
	}
	if message != "" {
		http_util.WriteTemplate(
			w,
			kTemplate,
			&view{
				Values: http_util.Values{r.Form},
				Months: kMonths,
				Error:  errors.New(message),
			})
		return
	}
	loan := mortgage.NewLoan(amount, rate, length)
	terms := loan.Terms(year, month, kMaxTerms+1)
	if len(terms) == kMaxTerms+1 {
		http_util.WriteTemplate(
			w,
			kTemplate,
			&view{
				Values: http_util.Values{r.Form},
				Months: kMonths,
				Error:  errors.New("Too many terms to display"),
			})
		return
	}
	http_util.WriteTemplate(
		w,
		kTemplate,
		&view{
			Values: http_util.Values{r.Form},
			Loan:   loan,
			Totals: aggregate(terms),
			Months: kMonths,
		})
}

type yearTotals struct {
	Principal int64
	Interest  int64
	Payment   int64
	start     int
	end       int
	terms     []*mortgage.Term
}

func newYearTotals(terms []*mortgage.Term, idx int) *yearTotals {
	return &yearTotals{
		start: idx,
		end:   idx,
		terms: terms}
}

func (y *yearTotals) Terms() []*mortgage.Term {
	return y.terms[y.start:y.end]
}

func (y *yearTotals) add() {
	y.Principal += y.terms[y.end].Principal()
	y.Interest += y.terms[y.end].Interest
	y.Payment += y.terms[y.end].Payment
	y.end++
}

type loanTotals struct {
	Principal int64
	Interest  int64
	Payment   int64
	Years     map[int]*yearTotals
}

type view struct {
	http_util.Values
	Loan   *mortgage.Loan
	Totals *loanTotals
	Error  error
	Months http_util.ComboBox
}

func aggregate(terms []*mortgage.Term) *loanTotals {
	result := &loanTotals{Years: make(map[int]*yearTotals)}
	for i := range terms {
		result.Payment += terms[i].Payment
		result.Interest += terms[i].Interest
		result.Principal += terms[i].Principal()
		// This way payments due on Jan 1 are aggregated with previous year.
		// We assume payment is made one day before due date. While this
		// does the right thing for payments due on the 1st, generally it won't
		// work if we ever support payments due on a day other than the 1st
		// because then it depends on when payment is made.
		year := terms[i].Date.AddDate(0, 0, -1).Year()
		yTotal := result.Years[year]
		if yTotal == nil {
			yTotal = newYearTotals(terms, i)
			result.Years[year] = yTotal
		}
		yTotal.add()
	}
	return result
}

func hasNoneOf(values url.Values, paramNames ...string) bool {
	for _, name := range paramNames {
		if http_util.HasParam(values, name) {
			return false
		}
	}
	return true
}

func init() {
	kTemplate = common.NewTemplate("home", kTemplateSpec)
}
