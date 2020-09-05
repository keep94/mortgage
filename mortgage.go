// Package mortgage calculates Amortization schedules
package mortgage

import (
	"github.com/keep94/appcommon/date_util"
	"math"
	"time"
)

// Term represents a single term within an amortization schedule.
type Term struct {
	Date     time.Time
	Payment  int64
	Interest int64
	Balance  int64
}

// Principal returns the principal paid during this term
func (t *Term) Principal() int64 {
	return t.Payment - t.Interest
}

// Loan represents a loan. Loan instances are immutable.
type Loan struct {
	amount  int64
	rate    float64
	length  int
	payment int64
}

// NewLoan returns a new loan. Payments on returned loan are monthly.
// amount is the amount borrowed;
// rate is the annual interest rate, 0.03 = 3%;
// durationInMonths is the number of months of the loan.
// amount and durationInMonths must be positive.
func NewLoan(amount int64, rate float64, durationInMonths int) *Loan {
	if amount <= 0 || durationInMonths <= 0 {
		panic("Amount and durationInMonths must be positive.")
	}
	payment := solveForPayment(amount, rate/12.0, durationInMonths)
	return &Loan{amount, rate, durationInMonths, payment}
}

// Amount returns the amount borrowed
func (l *Loan) Amount() int64 {
	return l.amount
}

// Rate returns the annual interest rate, 0.03 = 3%.
func (l *Loan) Rate() float64 {
	return l.rate
}

// DurationInMonths returns the number of months of the loan. Depending on the
// rounding of payment, this may be different than the actual number of months
// needed to pay off the loan.
func (l *Loan) DurationInMonths() int {
	return l.length
}

// Payment returns the payment due each term
func (l *Loan) Payment() int64 {
	return l.payment
}

// Terms returns all the terms needed to pay off this loan. year and
// month are the origination month of the loan.
// maxTerms is the maximum number of terms this method will return.
// The number of terms returned may differ from the duration of the loan
// depending on the rounding of the payment.
func (l *Loan) Terms(year, month, maxTerms int) []*Term {
	var result []*Term
	date := date_util.YMD(year, month, 1)
	balance := l.amount
	monthlyRate := l.rate / 12.0
	for balance > 0 {
		date = date.AddDate(0, 1, 0)
		interest := toInt64(float64(balance) * monthlyRate)
		balance += interest
		payment := l.payment
		if payment > balance {
			payment = balance
		}
		balance -= payment
		result = append(result, &Term{
			Date:     date,
			Payment:  payment,
			Interest: interest,
			Balance:  balance})
		if len(result) == maxTerms {
			break
		}
	}
	return result
}

func solveForPayment(
	amount int64, rate float64, length int) int64 {
	amountF := float64(amount)
	lengthF := float64(length)
	if rate == 0.0 {
		return toInt64(amountF / lengthF)
	}
	result := toInt64(amountF * rate * (1.0 + 1.0/(math.Pow((1.0+rate), lengthF)-1.0)))
	if result <= 0 {
		result = 1
	}
	interestOnly := toInt64(amountF * rate)
	if result <= interestOnly {
		result = interestOnly + 1
	}
	return result
}

func toInt64(x float64) int64 {
	return int64(x + 0.5)
}
