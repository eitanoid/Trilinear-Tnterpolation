package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

//TODO: user input colors

// used as RGBA
type Vec []float64 // vector type
var Supported_Formats = map[string]bool{"rgba": true, "oklab": true}

func Print_Input(verts [2][2][2]Vec, depth int, format string) {
	_depth := strconv.Itoa(depth - 1)

	front_face := map[string]string{
		"top":    fmt.Sprintf("%s%s", ParseFormat(verts[0][0][0], format).To_Ansi("000"), ParseFormat(verts[0][0][1], format).To_Ansi("00"+_depth)),
		"bottom": fmt.Sprintf("%s%s", ParseFormat(verts[0][1][0], format).To_Ansi("010"), ParseFormat(verts[0][1][1], format).To_Ansi("0"+_depth+_depth)),
	}
	back_face := map[string]string{
		"top":    fmt.Sprintf("%s%s", ParseFormat(verts[1][0][0], format).To_Ansi(_depth+"00"), ParseFormat(verts[1][0][1], format).To_Ansi(_depth+"0"+_depth)),
		"bottom": fmt.Sprintf("%s%s", ParseFormat(verts[1][1][0], format).To_Ansi(_depth+_depth+"0"), ParseFormat(verts[1][1][1], format).To_Ansi(_depth+_depth+_depth)),
	}

	fmt.Printf("Interpolating %d times, in %s format, between: \n%s   %s\n%s   %s\n", depth, format, front_face["top"], back_face["top"], front_face["bottom"], back_face["bottom"])
}

func Parse_Input(input string) []RGBA {
	var verts = make([]RGBA, 8)
	entries := strings.Split(input, ",")
	if len(entries) != 8 {
		panic("must contain exactly 8 codes")
	}
	dimension := len(entries[0]) - 1
	if (dimension != 6) && (dimension != 8) && dimension != 0 {
		panic("all hex codes must contain 3 or 4 channels")
	}

	for i, entry := range entries {
		entry = entry[1:]
		if len(entry) != dimension {
			panic("all hex codes must be of same length")
		}

		switch dimension {
		case 6:
			r, _ := strconv.ParseInt(entry[:2], 16, 64)
			g, _ := strconv.ParseInt(entry[2:4], 16, 64)
			b, _ := strconv.ParseInt(entry[4:6], 16, 64)
			verts[i] = RGBA{float64(r), float64(g), float64(b)}
		case 8:
			r, _ := strconv.ParseInt(entry[:2], 16, 64)
			g, _ := strconv.ParseInt(entry[2:4], 16, 64)
			b, _ := strconv.ParseInt(entry[4:6], 16, 64)
			a, _ := strconv.ParseInt(entry[6:8], 16, 64)
			verts[i] = RGBA{float64(r), float64(g), float64(b), float64(a)}
		case 0: // create random verts
			for i := 0; i < 8; i++ {
				verts[i] = RGBA{
					float64(rand.Intn((i + 1) * 30)),
					float64(rand.Intn((i + 1) * 30)),
					float64(rand.Intn((i + 1) * 30)),
					255}
			}
		}

	}
	return verts
}

func main() {
	//Handle user input
	_format := flag.String("format", "rgba", "In which format to interpolate.")
	_depth := flag.Int("depth", 6, "Interpolate to 'depth' points.")
	_verbose := flag.Bool("v", false, "Verbose")
	_debug := flag.Bool("d", false, "Set vertecies to debug mode")
	_hex := flag.Bool("H", false, "If not returning image, return terminal print as hex codes rather than indicies")
	_generate_images := flag.Bool("i", false, "Generate 'depth' images of the interpolation")
	_input_verts := flag.String("verts", "", "Enter 8 verticies as RGBA HEX codes seperated by commas: '#000000,#FFFFFF...' indecies will read as Front Top: 01, Front Bottom: 23, Back Top: 45, Back Bottom: 67")

	flag.Parse()

	format := *_format
	depth := *_depth
	verbose := *_verbose
	generate_images := *_generate_images
	hex := *_hex
	debug := *_debug
	input_verts := Parse_Input(*_input_verts)
	// fmt.Printf("%v,%v,%v,%v\n", format, depth, verbose, generate_images)
	// fmt.Println(input_verts)
	if debug {
		input_verts = []RGBA{ // constant rather than random values for debugging
			{0, 0, 0, 255},       //#000000
			{0, 0, 255, 255},     //#0000ff
			{0, 255, 0, 255},     //#00ff00
			{0, 255, 255, 255},   //#00ffff
			{255, 0, 0, 255},     //#ff0000
			{255, 0, 255, 255},   //#ff00ff
			{255, 255, 0, 255},   //#ffff00
			{255, 255, 255, 255}, //#ffffff
		}
	}

	corners := make([]Vec, 8)
	switch format { // which format to interpolate as:
	case "oklab":
		for i := range corners {
			corners[i] = input_verts[i].ToLAB().ToRaw()
		}
	default:
	case "rgba":
		for i := range corners {
			corners[i] = input_verts[i].ToRaw()
		}
	}

	// verts is indexed as [forward 0 / backward 1][top 0 / bottom 1][left 0/ right 1]
	verts := [2][2][2]Vec{ // corners[n].ToLAB().ToRaw() to lerpas as OKLAB.
		{
			{corners[0], corners[1]},
			{corners[2], corners[3]}},

		{
			{corners[4], corners[5]},
			{corners[6], corners[7]}},
	}

	if verbose {
		Print_Input(verts, depth, format)
	}

	// Run Trilerp
	now := time.Now()
	cube := Trilinear_interp(verts, depth)
	if verbose {
		fmt.Printf("Trilinear interp took: %v \n", time.Since(now))
	}

	// Return images
	now = time.Now()
	if generate_images {

		images := Export_Cube(cube, format)
		for i, image := range images {
			Save_PNG(image, "./images/"+strconv.Itoa(i)+".png")
		}
	} else { // print colors to terminal in groups of 3 planes per row

		_cspace := 1          // how much space between each color in the planes
		_hspace := 2          // how much space between planes horizontally
		_vspace := 1          // how much space between planes vertically
		show_hex_codes := hex // option to show hex or index

		hspace := strings.Repeat(" ", _hspace)
		vspace := strings.Repeat("\n", _vspace)

		ansi_cube := Export_Cube_Ansi(cube, format, _cspace, show_hex_codes)

		for dep, res := 0, len(cube); dep < res; dep += 3 {

			for row := 0; row < res; row++ {
				switch {
				case dep+2 < res: // print 3 points
					fmt.Printf("%s%s%s%s%s\n", ansi_cube[dep][row], hspace, ansi_cube[dep+1][row], hspace, ansi_cube[dep+2][row]) // col+space+col+space
				case dep+2 == res: // print 2 planes
					fmt.Printf("%s%s%s\n", ansi_cube[dep][row], hspace, ansi_cube[dep+1][row]) // col+space+col+space
				case dep+1 == res: // print 1 plane
					fmt.Printf("%s\n", ansi_cube[dep][row]) // col
				}
			}
			fmt.Print(vspace) // print vspace at the end of each series of planes
		}
	}

	if verbose {
		fmt.Printf("Output took: %v \n", time.Since(now))
	}
}
