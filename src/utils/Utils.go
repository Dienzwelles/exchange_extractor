package utils

func Sgn(a float64) float64 {
	switch {
	case a < 0:
		return -1
	case a > 0:
		return +1
	}
	return 0
}

func TernaryFloat64(test bool, a float64, b float64) float64{
	if(test){
		return a
	}

	return b
}

func Ternary(test bool, a string, b string) string{
	if(test){
		return a
	}

	return b
}