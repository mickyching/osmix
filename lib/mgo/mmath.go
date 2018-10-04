package mgo

import (
	"math"
)

func IsMin(m float64, a ...float64) bool {
	if len(a) == 0 {
		return true
	}
	for _, v := range a {
		if v < m {
			return false
		}
	}
	return true
}
func IsMax(m float64, a ...float64) bool {
	if len(a) == 0 {
		return true
	}
	for _, v := range a {
		if v > m {
			return false
		}
	}
	return true
}
func Min(a ...float64) (int, float64) {
	if len(a) == 0 {
		return -1, 0.0
	}
	mii, min := 0, a[0]
	for i, v := range a {
		if v < min {
			mii, min = i, v
		}
	}
	return mii, min
}
func Max(a ...float64) (int, float64) {
	if len(a) == 0 {
		return -1, 0.0
	}
	mai, max := 0, a[0]
	for i, v := range a {
		if v > max {
			mai, max = i, v
		}
	}
	return mai, max
}
func Sum(a ...float64) float64 {
	sum := 0.0
	for _, v := range a {
		sum += v
	}
	return sum
}
func Mean(a ...float64) float64 {
	if len(a) == 0 {
		return 0
	}
	return Sum(a...) / float64(len(a))
}

// MSE return mean square error
// 方差：d2 = 1/n sum(xi-x)2
// 均方差=标准差：d = sqrt(d2)
func MSE(a ...float64) float64 {
	if len(a) == 0 {
		return 0
	}

	dd := 0.0
	m := Mean(a...)
	for _, x := range a {
		dd += (x - m) * (x - m)
	}

	return math.Sqrt(dd / float64(len(a)))
}

// LinearFit return (k, b) fit line y = kx + b
func LinearFit(sx []float64, sy []float64) (k float64, b float64) {
	if len(sx) != len(sy) {
		Fatalf("slice length not match x(%d) != y(%d)", len(sx), len(sy))
	}

	num := float64(len(sx))
	xy := 0.0
	xx := 0.0
	xs := 0.0
	ys := 0.0
	for i := 0; i < len(sx); i++ {
		x := sx[i]
		y := sy[i]
		xx += x * x
		xy += x * y
		xs += x
		ys += y
	}

	k = (num*xy - xs*ys) / (num*xx - xs*xs)
	b = (xx*ys - xs*xy) / (num*xx - xs*xs)

	return k, b
}

// CmpFloats compare slice
// a < b return -1
// a > b return 1
// else return 0
func CmpFloats(a []float64, b []float64) int {
	if len(a) != len(b) {
		Fatalf("length not same %d != %d", len(a), len(b))
	}

	aa := 0
	bb := 0
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			aa++
		} else if a[i] < b[i] {
			bb++
		} else {
			return 0
		}
		if aa != 0 && bb != 0 {
			return 0
		}
	}

	if aa == len(a) {
		return 1
	} else if bb == len(a) {
		return -1
	}
	Fatalf("unreachable code")
	return 0
}
