package utils

import "math"

type Money int64

// ToMoney converts a float64 amount in dollars to Money (cents)
func ToMoney(dollars float64) Money {
	return Money(math.Round(dollars * 1000))
}

// ToDollars converts Money (cents) to float64 dollars
func (m Money) ToDollars() float64 {
	return float64(m) / 1000
}

// Add adds two Money values
func (m Money) Add(other Money) Money {
	return m + other
}

// Subtract subtracts another Money value from the current Money value
func (m Money) Subtract(other Money) Money {
	return m - other
}

// Multiply multiplies Money by a factor, keeping two decimal places
func (m Money) Multiply(factor float64) Money {
	result := float64(m) * factor
	return ToMoney(result / 1000) // Divide by 1000 to get the correct result in cents
}

// Divide divides Money by a divisor, keeping two decimal places
func (m Money) Divide(divisor float64) Money {
	result := float64(m) / divisor
	return ToMoney(result / 1000) // Divide by 1000 to get the correct result in cents
}
