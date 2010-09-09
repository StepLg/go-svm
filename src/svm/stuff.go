package svm

func scalarProduct(x, y []float) float {
	if len(x) != len(y) {
		panic("Size mismatch.")
	}

	res := 0.0
	for i, _ := range x {
		res += x[i] * y[i]
	}
	return res
}

func abs(x float) float {
	if x < 0.0 {
		return -x
	}
	return x
}

type Classificator func(x []float) int

func LincearClassificator(x, w []float, w0 float) float {
	return scalarProduct(x, w) - w0
}
