package main

import (
	"fmt"
	"log"
	"os"
	"github.com/kuso-kodo/kuso-NES/nes"
)

const (
	EXEC_SUCCESS = iota
	EXEC_FAILED
)
// At this time , the usage of main.go is to test weather we can read a iNES file currectly.
// Usage: kuso-NES <file name>

func main() {
	cartridge , err := nes.LoadNES(os.Args[1])

	if err != nil {
		log.Fatalln(err)
		os.Exit(EXEC_FAILED)
	}

	fmt.Println(cartridge)
}
