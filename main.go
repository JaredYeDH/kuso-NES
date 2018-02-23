package main

import (
	"log"
	"os"
	"github.com/kuso-kodo/kuso-NES/nes"
)

const (
	EXEC_SUCCESS = iota
	EXEC_FAILED
)
// At this time , the usage of main.go is to test wether we can read instructions from aan iNES file properly.
// Usage: kuso-NES <file name>

func main() {

	nes,err := nes.NewNES(os.Args[1])

	if err != nil {
		log.Fatalln(err)
		os.Exit(EXEC_FAILED)
	}

	log.Printf("Read iNES file : %s\n",os.Args[1])

	for i := 0 ; i < 0xFF ; i ++ {
		nes.CPU.DebugPrint()
		nes.CPU.Run()
	}
}
