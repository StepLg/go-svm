package svm

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

func ReadPoints(f io.Reader) (points [][]float, target []float) {
	points = make([][]float, 1, 10)
	target = make([]float, 1, 10)

	reader := bufio.NewReader(f)

	// reading first line
	dimentions := 0
	var err os.Error
	var line string
	line, err = reader.ReadString('\n');
	line = strings.Trim(line, " \t\n")
	chunks := strings.Split(line, "\t", -1)
	if len(chunks)<3 {
		panic("Too small numbers in a first line.")
	}
	
	target[0], err = strconv.Atof(chunks[0])
	if err!=nil {
		panic(err)
	}
	dimentions = len(chunks)-1
	points[0] = make([]float, dimentions)
	for i:=1; i<len(chunks); i++ {
		points[0][i-1], err = strconv.Atof(chunks[i])
		if err!=nil {
			panic(err)
		}
	}
	
	// reading other points
	line, err = reader.ReadString('\n');
	for err==nil || err==os.EOF {
		line = strings.Trim(line, " \t\n")
		if line=="" {
			if err==os.EOF {
				break
			}
			line, err = reader.ReadString('\n');
			continue
		}
		chunks := strings.Split(line, "\t", -1)
		if len(chunks)-1 != dimentions {
			panic("Dimentions mismatch.")
		}
		
		if len(points)+1>=cap(points) {
			// resize arrays
			{
				tmp := make([][]float, len(points), 2*len(points))
				copy(tmp, points)
				points = tmp
			}
			
			{
				tmp := make([]float, len(target), 2*len(target))
				copy(tmp, target)
				target = tmp
			}
		}
		
		pos := len(points)
		points = points[0:len(points)+1]
		target = target[0:len(target)+1]
		
		target[pos], err = strconv.Atof(chunks[0])
		if err!=nil {
			panic(err)
		}
		points[pos] = make([]float, dimentions)
		for i:=1; i<len(chunks); i++ {
			points[pos][i-1], err = strconv.Atof(chunks[i])
			if err!=nil {
				panic(err)
			}
		}
		if err==os.EOF {
			break
		}
		line, err = reader.ReadString('\n');
	}
	if err!=nil && err!=os.EOF {
		panic(err)
	}
	
	return
}
