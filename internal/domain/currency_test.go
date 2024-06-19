package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloatToCurrency(t *testing.T) {
	for float, expectedCurrency := range map[float64]Currency{
		0.00:     0,
		0.55:     55,
		1.00:     100,
		-1.00:    -100,
		123.45:   12345,
		0.001:    0,
		-0.001:   0,
		0.123456: 12,
		5.555:    556,
		-5.555:   -555,
	} {
		currency := FloatCurrency(float)
		assert.Equal(t, expectedCurrency, currency)
	}
}

func TestCurrencyFLoat(t *testing.T) {
	for currency, expectedFloat := range map[Currency]float64{
		0:      0,
		1:      0.01,
		100:    1,
		1000:   10,
		10001:  100.01,
		-1520:  -15.2,
		493012: 4930.12,
	} {
		float := currency.Float()
		assert.Equal(t, expectedFloat, float)
	}
}

func TestCurrencyString(t *testing.T) {
	for currency, expectedString := range map[Currency]string{
		0:      "0.00",
		1:      "0.01",
		100:    "1.00",
		3211:   "32.11",
		590294: "5902.94",
		100000: "1000.00",
	} {
		str := currency.String()
		assert.Equal(t, expectedString, str)
	}
}
