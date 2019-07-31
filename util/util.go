package util

import (
	"math"
	"math/rand"
	"time"
)

//Max : integer max
func Max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

//Min : integer min
func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

//Abs : integer absolute value
func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

//LinVal : linearly scaled quantity
func LinVal(C float64, v int) int {
	return int(math.Max(1, C*float64(v)))
}

//SqrtVal : a quantity scaled to the power of 0.5
func SqrtVal(C float64, v int) int {
	return int(math.Max(1, C*math.Sqrt(float64(v))))
}

//Seed : renew random seed
func Seed() {
	rand.Seed(time.Now().UnixNano())
}

//RandInt : Sample from {0,1,...,n-1}
func RandInt(n int) int {
	return rand.Intn(n)
}

//RandBool : Sample from {false,true}
func RandBool() bool {
	if RandInt(2) == 0 {
		return false
	}
	return true
}

//RandNorm : Sample from standard normal distribution
func RandNorm() float64 {
	return rand.NormFloat64()
}

//Poisson : sample from Poisson distribution of mean lambda
func Poisson(lambda int) int {
	sum := 0
	for i := 0; i < lambda; i++ {
		sum += unitPoisson()
	}
	return sum
}

func unitPoisson() int {
	p := rand.Float64()
	k := -1
	q := 0.0
	for q < p {
		k++
		q += math.Exp(-1) / float64(factorial(k))
	}
	return k
}

func factorial(n int) int {
	m := 1
	for i := 1; i <= n; i++ {
		m *= i
	}
	return m
}
