package domain

import (
	"fmt"
	"math"
)

// naive implementation of decimal, using float64 as currency is a terrible idea
// with this approach the whole system should be aware that the "Currency" is an int with decimals included, which is not perfect as well
// probably a matured library like github.com/shopspring/decimal would be much better
type Currency int64

func (c Currency) Float() float64 {
	return float64(c) / 100.
}

func FloatCurrency(v float64) Currency {
	return Currency(math.Floor(v*100. + .5))
}

func (c Currency) String() string {
	return fmt.Sprintf("%.2f", c.Float())
}
