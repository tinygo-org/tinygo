//go:build ignore
// +build ignore

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"os/exec"
)

//go:embed pjw.jpg
var b []byte

func main() {
	img, err := jpeg.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}

	c := exec.Command("goimports")
	w, err := c.StdinPipe()
	if err != nil {
		log.Fatalf("stdinpipe:%v", err)
	}

	if c.Stdout, err = os.Create("img.go"); err != nil {
		log.Fatalf("creating output:%v", err)
	}

	c.Stderr = os.Stderr

	if err := c.Start(); err != nil {
		log.Fatalf("Starting %v:%v", c, err)
	}

	fmt.Fprintf(w, "package main\ntype pixel struct{\nx int16\ny int16\nval uint8\n}\nvar pixels = []pixel{\n")

	const scale = 5
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += scale {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x += scale {
			r, _, _, _ := img.At(x, y).RGBA()
			fmt.Fprintf(w, "pixel{x:%d,y:%d,val:%d,},\n", x/scale, y/scale, uint8(r&0xff))
		}
	}
	fmt.Fprintf(w, "\n}")
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
	if err := c.Wait(); err != nil {
		log.Fatalf("%v:%v", c, err)
	}

}
