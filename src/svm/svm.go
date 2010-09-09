package svm

import (
	"sort"
)

// Logger function form SMO algorithm
//
// Function called after each iteration with these arguemnts:
// isAll - is during interations checks all alphas
// checkedAlpha - number of alphas, for which called takeStep function. Number of non-bound alphas
//                if !isAll and total number of alphas if isAll
// w, w0 - current solution
// changesMap - map of alphas, that changed value
type smoLoggerFunc func(isAll bool, checkedAlpha uint, w []float, w0 float, changesMap map[int]int)

type indexesSorter struct {
	weights []float
	indexes []int
}

func (s *indexesSorter) Len() int {
	return len(s.weights)
}

func (s *indexesSorter) Less(i, j int) bool {
	return s.weights[s.indexes[i]] < s.weights[s.indexes[j]]
}

func (s *indexesSorter) Swap(i, j int) {
	s.indexes[i], s.indexes[j] = s.indexes[j], s.indexes[i]
}

func newIndexesSorter(indexes []int, weights []float) *indexesSorter {
	if len(indexes) != len(weights) {
		panic("Size mismatch in sorter.")
	}

	return &indexesSorter{
		indexes: indexes,
		weights: weights,
	}
}

type sequentialMinimalOptimizationTask struct {
	points     [][]float
	target     []float
	C, eps     float
	w          []float
	w0         float
	size       int
	dimentions int

	alpha  []float
	errors []float
	// todo: add nonbound alphas map

	useHeurisitc bool
	log          smoLoggerFunc
}

func newSequentialMinimalOptimizationTask(points [][]float, target []float, C, eps float, useHeurisitc bool, log smoLoggerFunc) *sequentialMinimalOptimizationTask {
	// checking task
	size := len(points)
	if size <= 0 {
		panic("size < 0")
	}
	if size != len(target) {
		panic(" size != len(target)")
	}

	dimentions := len(points[0])
	if dimentions == 0 {
		panic("wrong dimentions cnt")
	}

	for i, p := range points {
		if len(p) != dimentions {
			panic("dimentions mismatch")
		}

		if target[i] != 1.0 && target[i] != -1.0 {
			panic("wrong class")
		}
	}

	if C <= 0 {
		panic("C<0")
	}

	if eps <= 0 {
		panic("eps<0")
	}
	result := &sequentialMinimalOptimizationTask{
		points:       points,
		target:       target,
		C:            C,
		eps:          eps,
		w:            make([]float, dimentions),
		w0:           0,
		size:         size,
		dimentions:   dimentions,
		alpha:        make([]float, size),
		errors:       make([]float, size),
		useHeurisitc: useHeurisitc,
		log:          log,
	}

	result.UpdateCache()
	return result
}

func (t *sequentialMinimalOptimizationTask) UpdateSeparatingHyperplane() {

	for i, p := range t.points {
		a := t.alpha[i] * t.target[i]
		if a != 0 {
			for k := 0; k < t.dimentions; k++ {
				t.w[k] += a * p[k]
			}
		}
	}

	thresholds := make([]float, t.size)
	for i, p := range t.points {
		thresholds[i] = scalarProduct(t.w, p) - t.target[i]
	}

	sort.SortFloats(thresholds)

	t.w0 = thresholds[len(thresholds)/2]
}

func (t *sequentialMinimalOptimizationTask) UpdateCache() {
	t.UpdateSeparatingHyperplane()
	for i, p := range t.points {
		t.errors[i] = scalarProduct(t.w, p) - t.w0 - t.target[i]
	}
}

func (t *sequentialMinimalOptimizationTask) TakeStep(i1, i2 int) bool {
	const eps = 1e-3
	if i1 == i2 {
		return false
	}

	alph1 := t.alpha[i1]
	alph2 := t.alpha[i2]
	y1 := t.target[i1]
	y2 := t.target[i2]
	E1 := t.errors[i1]
	E2 := t.errors[i2]
	s := y1 * y2

	var L float
	var H float
	if y1 != y2 {
		L = alph2 - alph1
		if L < 0 {
			L = 0
		}
		H = t.C + alph2 - alph1
		if H > t.C {
			H = t.C
		}
	} else {
		L = alph2 + alph1 - t.C
		if L < 0 {
			L = 0
		}
		H = alph2 + alph1
		if H > t.C {
			H = t.C
		}
	}

	if L == H {
		return false
	}

	k11 := scalarProduct(t.points[i1], t.points[i1])
	k12 := scalarProduct(t.points[i1], t.points[i2])
	k22 := scalarProduct(t.points[i2], t.points[i2])

	eta := k11 + k22 - 2*k12
	var a2 float
	if eta > 0 {
		a2 = alph2 + y2*(E1-E2)/eta
		if a2 < L {
			a2 = L
		} else if a2 > H {
			a2 = H
		}
	} else {
		f1 := y1*(E1+t.w0) - alph1*k11 - s*alph2*k12
		f2 := y2*(E2+t.w0) - s*alph1*k12 - alph2*k22
		L1 := alph1 + s*(alph2-L)
		H1 := alph1 + s*(alph2-H)
		Lobj := L1*f1 + L*f2 + 0.5*L1*L1*k11 + 0.5*L*L*k22 + s*L*L1*k12
		Hobj := H1*f1 + H*f2 + 0.5*H1*H1*k11 + 0.5*H*H*k22 + s*H*H1*k12
		if Lobj < Hobj-eps {
			a2 = L
		} else if Lobj > Hobj+eps {
			a2 = H
		} else {
			a2 = alph2
		}
	}
	if abs(a2-alph2) < t.eps*(a2+alph2+t.eps) {
		return false
	}

	a1 := alph1 + s*(alph2-a2)
	t.alpha[i1] = a1
	t.alpha[i2] = a2
	t.UpdateCache()
	return true
}

func (t *sequentialMinimalOptimizationTask) ExamineExample(i2 int) (int, bool) {
	y2 := t.target[i2]
	alph2 := t.alpha[i2]
	E2 := t.errors[i2]
	r2 := E2 * y2
	tol := 1e-3 // wtf?! i don't know, what't this! And what variable value should be :(
	if (r2 < -tol && alph2 < t.C) || (r2 > tol && alph2 > 0) {
		if t.useHeurisitc {
			// heuristic 2.2 choise
			indexes := make([]int, len(t.alpha))
			for i1, _ := range t.alpha {
				indexes[i1] = i1
			}

			sort.Sort(newIndexesSorter(indexes, t.errors))

			var i1 int
			if E2 > 0 {
				i1 = indexes[len(indexes)-1]
			} else {
				i1 = indexes[0]
			}

			if t.TakeStep(i1, i2) {
				return i1, true
			}
		}

		for i1, alph1 := range t.alpha {
			if alph1 > 0 || alph1 < t.C {
				if t.TakeStep(i1, i2) {
					return i1, true
				}
			}
		}

		for i1, alph1 := range t.alpha {
			if alph1 == 0 || alph1 == t.C {
				if t.TakeStep(i1, i2) {
					return i1, true
				}
			}
		}
	}

	return -1, false
}

func (t *sequentialMinimalOptimizationTask) Train() {
	numChanged := 0
	examineAll := true

	itersCnt := 0
	for (numChanged > 0) || examineAll {
		changesMap := make(map[int]int)
		itersCnt++
		scannedAlpha := uint(0)
		if examineAll {
			scannedAlpha = uint(len(t.points))
			for i, _ := range t.points {
				i1, isChanged := t.ExamineExample(i)
				if isChanged {
					changesMap[i] = i1
				}
			}
		} else {
			for i, a := range t.alpha {
				if a == 0.0 || a == t.C {
					continue
				}

				scannedAlpha++
				i1, isChanged := t.ExamineExample(i)
				if isChanged {
					changesMap[i] = i1
				}
			}
		}
		numChanged = len(changesMap)
		t.log(examineAll, scannedAlpha, t.w, t.w0, changesMap)

		if examineAll {
			examineAll = false
		} else if numChanged == 0 {
			examineAll = true
		}
	}
}

func SequentialMinimalOptimization(points [][]float, target []float, C, eps float) (w []float, w0 float) {
	t := newSequentialMinimalOptimizationTask(points, target, C, eps, true, nil)
	t.Train()
	return t.w, t.w0
}

func SequentialMinimalOptimization_logger(points [][]float, target []float, C, eps float, log smoLoggerFunc) (w []float, w0 float) {
	t := newSequentialMinimalOptimizationTask(points, target, C, eps, true, log)
	t.Train()
	return t.w, t.w0
}
