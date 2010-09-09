package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func prod(x, y []float) float {
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

func makeClassificateVector(points [][]float, target []float, alpha []float) ([]float, float) {
	w := make([]float, len(points[0]))
	for i := 0; i < len(points); i++ {
		if alpha[i] > 0 {
			for k := 0; k < len(w); k++ {
				w[k] += target[i] * alpha[i] * points[i][k]
			}
		}
	}

	// todo: median
	thresholds := make([]float, len(alpha))
	for i, p := range points {
		thresholds[i] = prod(w, p) - target[i]
	}

	sort.SortFloats(thresholds)

	w0 := thresholds[len(thresholds)/2]
	return w, w0
}

func takeStep(points [][]float, target []float, C float, alpha []float, i1, i2 int, w []float, w0 float) bool {
	const eps = 1e-3
	if i1 == i2 {
		return false
	}

	alph1 := alpha[i1]
	alph2 := alpha[i2]
	y1 := target[i1]
	y2 := target[i2]
	E1 := prod(w, points[i1]) - w0 - y1
	E2 := prod(w, points[i2]) - w0 - y2
	s := y1 * y2

	var L float
	var H float
	if y1 != y2 {
		L = alph2 - alph1
		if L < 0 {
			L = 0
		}
		H = C + alph2 - alph1
		if H > C {
			H = C
		}
	} else {
		L = alph2 + alph1 - C
		if L < 0 {
			L = 0
		}
		H = alph2 + alph1
		if H > C {
			H = C
		}
	}

	if L == H {
		return false
	}

	k11 := prod(points[i1], points[i1])
	k12 := prod(points[i1], points[i2])
	k22 := prod(points[i2], points[i2])

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
		f1 := y1*(E1+w0) - alph1*k11 - s*alph2*k12
		f2 := y2*(E2+w0) - s*alph1*k12 - alph2*k22
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
	if abs(a2-alph2) < eps*(a2+alph2+eps) {
		return false
	}

	a1 := alph1 + s*(alph2-a2)
	alpha[i1] = a1
	alpha[i2] = a2
	return true
}

type indexesSorter struct {
	weights []float
	indexes []int
}

func (s *indexesSorter) Len() int {
	return len(s.weights)
}

func (s *indexesSorter) Less(i, j int) bool {
	return s.weights[i] < s.weights[j]
}

func (s *indexesSorter) Swap(i, j int) {
	s.weights[i], s.weights[j] = s.weights[j], s.weights[i]
	// s.indexes[i], s.indexes[j] = s.indexes[j], s.indexes[i]
}

func newIndexesSorter(indexes []int, weights []float) *indexesSorter {
	if len(indexes) != len(weights) {
		panic("Size mismatch.")
	}

	return &indexesSorter{
		indexes: indexes,
		weights: weights,
	}
}

func examineExample(points [][]float, target []float, C float, alpha []float, i2 int) int {
	y2 := target[i2]
	alph2 := alpha[i2]
	w, w0 := makeClassificateVector(points, target, alpha)
	E2 := prod(w, points[i2]) - w0 - y2
	r2 := E2 * y2
	tol := 1e-3 // wtf?! i don't know, what't this! And what variable value should be :(
	if (r2 < -tol && alph2 < C) || (r2 > tol && alph2 > 0) {
		// heuristic 2.2 choise
		/*
			{
				errors := make([]float, len(alpha))
				indexes := make([]int, len(alpha))
				for i1, _ := range alpha {
					indexes[i1] = i1
					errors[i1] = prod(w, points[i1]) - w0 - target[i1]
				}

				sort.Sort(newIndexesSorter(indexes, errors))

				var i1 int
				if E2>0 {
					i1 = indexes[len(indexes)-1]
				} else {
					i1 = indexes[0]
				}

				if takeStep(points, target, C, alpha, i1, i2, w, w0) {
					return 1
				}
			}
		*/

		for i1, alph1 := range alpha {
			if alph1 > 0 || alph1 < C {
				if takeStep(points, target, C, alpha, i1, i2, w, w0) {
					return 1
				}
			}
		}

		for i1, alph1 := range alpha {
			if alph1 == 0 || alph1 == C {
				if takeStep(points, target, C, alpha, i1, i2, w, w0) {
					return 1
				}
			}
		}
	}

	return 0
}

func main() {
	flag_help := flag.Bool("help", false, "display this help")
	flag_infile := flag.String("in", "",
		`input file with data. Each row is a data point. First number in
every line must be 1 or -1 and means point class. Next float numbers
are point coordinates.`)

	flag.Parse()

	if *flag_help {
		flag.PrintDefaults()
		return
	}

	// configuring input file
	infile := os.Stdin
	var err os.Error
	if *flag_infile != "" {
		infile, err = os.Open(*flag_infile, os.O_RDONLY, 0000)
		if err != nil {
			panic(err)
		}
	}

	// reading data
	points := make([][]float, 1, 10)
	target := make([]float, 1, 10)
	{
		reader := bufio.NewReader(infile)

		// reading first line
		dimentions := 0
		var err os.Error
		var line string
		line, err = reader.ReadString('\n')
		line = strings.Trim(line, " \t\n")
		chunks := strings.Split(line, "\t", -1)
		if len(chunks) < 3 {
			panic("Too small numbers in a line")
		}

		target[0], err = strconv.Atof(chunks[0])
		if err != nil {
			panic(err)
		}
		dimentions = len(chunks) - 1
		points[0] = make([]float, dimentions)
		for i := 1; i < len(chunks); i++ {
			points[0][i-1], err = strconv.Atof(chunks[i])
			if err != nil {
				panic(err)
			}
		}

		// reading other points
		line, err = reader.ReadString('\n')
		for err == nil || err == os.EOF {
			line = strings.Trim(line, " \t\n")
			if line == "" {
				if err == os.EOF {
					break
				}
				line, err = reader.ReadString('\n')
				continue
			}
			chunks := strings.Split(line, "\t", -1)
			if len(chunks)-1 != dimentions {
				panic("Dimentions mismatch.")
			}

			if len(points)+1 >= cap(points) {
				// resize arrays
				{
					tmp := make([][]float, len(points), 2*len(points))
					copy(tmp, points)
					points = tmp
				}

				{
					tmp := make([]float, len(points), 2*len(points))
					copy(tmp, target)
					target = tmp
				}
			}

			pos := len(points)
			points = points[0 : len(points)+1]
			target = target[0 : len(points)+1]

			target[pos], err = strconv.Atof(chunks[0])
			if err != nil {
				panic(err)
			}
			points[pos] = make([]float, dimentions)
			for i := 1; i < len(chunks); i++ {
				points[pos][i-1], err = strconv.Atof(chunks[i])
				if err != nil {
					panic(err)
				}
			}
			if err == os.EOF {
				break
			}
			line, err = reader.ReadString('\n')
		}
		if err != nil && err != os.EOF {
			panic(err)
		}
	}

	size := len(points)
	C := 5.0

	fmt.Println("Classification started for size", size)

	numChanged := 0
	examineAll := true
	alpha := make([]float, size)

	itersCnt := 0
	for (numChanged > 0) || examineAll {
		itersCnt++
		numChanged = 0
		if examineAll {
			for i, _ := range points {
				numChanged += examineExample(points, target, C, alpha, i)
			}
		} else {
			for i, a := range alpha {
				if a == 0.0 || a == C {
					continue
				}

				numChanged += examineExample(points, target, C, alpha, i)
			}
		}

		if examineAll {
			examineAll = false
		} else if numChanged == 0 {
			examineAll = true
		}
	}

	w, w0 := makeClassificateVector(points, target, alpha)
	errors := 0
	for i := 0; i < len(points)/2; i++ {
		if prod(w, points[i])-w0 >= 0 {
			errors++
		}
	}
	for i := len(points) / 2; i < len(points); i++ {
		if prod(w, points[i])-w0 <= 0 {
			errors++
		}
	}
	if errors > 0 {
		for i, p := range points {
			fmt.Println(i, prod(w, p)-w0)
		}
	}
	fmt.Println("w  =", w)
	fmt.Println("w0 =", w0)
	fmt.Println("iters =", itersCnt)
	fmt.Println("errors =", errors, " /", len(points), " =", float(errors)/float(len(points))*100.0, "%")
}
