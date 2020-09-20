package mortgage_test

import (
	"github.com/keep94/mortgage"
	"github.com/keep94/toolbox/date_util"
	"reflect"
	"testing"
)

func TestTerm(t *testing.T) {
	term := &mortgage.Term{Payment: 62, Interest: 47}
	if term.Principal() != 15 {
		t.Error("Principal in Term wrong.")
	}
}

func TestRegular(t *testing.T) {
	l := mortgage.NewLoan(23800000, .04, 360)
	if amount := l.Payment(); amount != 113625 {
		t.Errorf("Expected payment 113625 but got %v", amount)
	}
	terms := l.Terms(2015, 1, 0)
	if len(terms) != 360 && len(terms) != 361 {
		t.Error("Term length wrong.")
		return
	}
	verifyTerms(t, terms, 23800000, 2015, 1)
	if !reflect.DeepEqual(l.Terms(2015, 1, 100), terms[:100]) {
		t.Error("Terms not properly truncated.")
	}
	if !reflect.DeepEqual(l.Terms(2015, 1, 1000), terms) {
		t.Error("Result should be the same as terms")
	}
}

func TestBigInterest(t *testing.T) {
	l := mortgage.NewLoan(23800000, .60, 360)
	if amount := l.Payment(); amount != 1190001 {
		t.Errorf("Expected payment 1190001 but got %v", amount)
	}
	verifyTerms(t, l.Terms(2015, 1, 0), 23800000, 2015, 1)
}

func TestNegativeInterest(t *testing.T) {
	l := mortgage.NewLoan(23800000, -0.60, 360)
	if amount := l.Payment(); amount != 1 {
		t.Errorf("Expected payment 1 but got %v", amount)
	}
	verifyTerms(t, l.Terms(2015, 1, 0), 23800000, 2015, 1)
}

func verifyTerms(
	t *testing.T, terms []*mortgage.Term, balanceForward int64,
	startYear, startMonth int) {
	date := date_util.YMD(startYear, startMonth, 1)
	prevBalance := balanceForward
	for _, term := range terms {
		date = date.AddDate(0, 1, 0)
		if prevBalance-term.Principal() != term.Balance {
			t.Error("Balance wrong")
		}
		if term.Date != date {
			t.Errorf("Date wrong: Expected %v, got %v", date, term.Date)
		}
		prevBalance = term.Balance
	}
	if prevBalance != 0 {
		t.Error("Final balance not zero.")
	}
}
