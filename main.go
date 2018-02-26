package main

import (
	"os"
	"github.com/kuso-kodo/kuso-NES/ui"
	"image/jpeg"
	"image/png"
	"strings"
	"log"
)

const (
	EXEC_SUCCESS = iota
	EXEC_FAILED
)

// At this time , the usage of main.go is to test whether we can read a png/jpeg file and show it properly.
// Usage: kuso-NES <file name>

func main() {

	argv := "assets/"+os.Args[1]
	if strings.HasSuffix(os.Args[1],"PNG") || strings.HasSuffix(os.Args[1],"png") {
		ui.Run(os.Args[1])
	} else {
		file,err := os.Open(argv)
		if err != nil {
			log.Printf("Readind file %s error:" + err.Error(), os.Args[1])
		}

		image,err := jpeg.Decode(file)

		if err != nil {
			log.Printf("Decoding jpeg file error:" + err.Error())
		}

		write,err := os.Create(argv+".png")
		if err != nil {
			log.Printf("Writing file %s error:" + err.Error(), os.Args[1])
		}

		png.Encode(write,image)
		ui.Run(os.Args[1]+ ".png")
	}
}
