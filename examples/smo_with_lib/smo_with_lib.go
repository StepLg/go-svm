package main

import (
	"../../src/svm/_obj/svm"
	
	"flag"
	"fmt"
	"os"
)

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
	if *flag_infile!="" {
		infile, err = os.Open(*flag_infile, os.O_RDONLY, 0000)
		if err!=nil {
			panic(err)
		}
	}
	
	points, target := svm.ReadPoints(infile)
	
	C := 5.0
	
	w, w0 := svm.SequentialMinimalOptimization(points, target, C, 1e-3, true)
	errors := 0
	for i:=0; i<len(points)/2; i++ {
		if svm.LincearClassificator(points[i], w, w0) >= 0 {
			errors++
		}
	}
	for i:=len(points)/2; i<len(points); i++ {
		if svm.LincearClassificator(points[i], w, w0) <= 0 {
			errors++
		}
	}
	/*
	if errors>0 {
		for i, p := range points {
			fmt.Println(i, svm.LincearClassificator(p, w, w0))
		}
	}
	*/
	fmt.Println(w, w0)
	fmt.Println(errors)
}
