// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mmcloughlin/cb/app/brand"
)

func main() {
	os.Exit(main1())
}

func main1() int {
	if err := mainerr(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

var (
	output    = flag.String("output", "", "path to output file (default stdout)")
	numshades = flag.Int("shades", 9, "number of lighter shades")
)

func mainerr() error {
	flag.Parse()

	b, err := Palette(*numshades)
	if err != nil {
		return err
	}

	if *output != "" {
		err = ioutil.WriteFile(*output, b, 0644)
	} else {
		_, err = os.Stdout.Write(b)
	}

	return err
}

func Palette(shades int) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// Header.
	_, self, _, _ := runtime.Caller(0)
	fmt.Fprintf(buf, "/* Generated by %s. DO NOT EDIT. */\n\n", filepath.Base(self))

	fmt.Fprintf(buf, ":root {\n")
	for _, name := range brand.ColorNames {
		fmt.Fprintf(buf, "\t--%s: %s;\n", name, brand.Colors[name])
		for s := 1; s <= shades; s++ {
			shade, err := brand.Lighten(name, float64(s)/float64(shades+1))
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(buf, "\t--%s-%d: %s;\n", name, s, shade)
		}
	}
	fmt.Fprint(buf, "}\n")

	return buf.Bytes(), nil
}
