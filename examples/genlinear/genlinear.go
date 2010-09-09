package main

import (
	"fmt"
	"rand"
	"flag"
)

func main() {
	flag_help := flag.Bool("help", false, "display this help")
	flag_size := flag.Int("size", 1000, "number of points")
	flag_k := flag.Float("k", 4.0, "slope of linear function kx + b")
	flag_b := flag.Float("b", 3.0, "y-intercept of linear function kx + b")
	flag_width := flag.Float("width", 10.0, "width of each class")
	flag_d := flag.Float("distance", 5.0, `distance to linear separator. If negative then
there isn't separable without errors (two classes are intersept).`)
	flag_xstart := flag.Float("xstart", -25.0, "x lower bound")
	flag_xend := flag.Float("xend", 25.0, "x upper bound")

	flag.Parse()

	if *flag_help {
		flag.PrintDefaults()
		return
	}

	size := *flag_size
	k, b := *flag_k, *flag_b
	width, d := *flag_width, *flag_d
	xstart, xend := *flag_xstart, *flag_xend

	for i := 0; i < size/2; i++ {
		x := rand.Float()*(xend-xstart) + xstart
		y := x*k + b + rand.Float()*width + d
		t := -1.0
		fmt.Printf("%v\t%v\t%v\t\n", t, x, y)
	}

	for i := size / 2; i < size; i++ {
		x := rand.Float()*(xend-xstart) + xstart
		y := x*k + b - rand.Float()*width - d
		t := 1.0
		fmt.Printf("%v\t%v\t%v\t\n", t, x, y)
	}

	return
}
