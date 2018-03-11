package main

import (
	"fmt"
	"github.com/kuso-kodo/kuso-NES/nes"
	"github.com/kuso-kodo/kuso-NES/ui"
	"log"
	"os"
)

const (
	EXEC_SUCCESS = iota
	EXEC_FAILED
)

// Trying to connect UI with the f***ing PPU.
func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: kuso-NES <NES Rom Path>")
		os.Exit(EXEC_FAILED)
	}
	path,hastmp := nes.ReadFile(os.Args[1])
	log.Print(path)
	NES, err := nes.NewNES(path)
	if err != nil {
		log.Fatalln(err)
	}
	if hastmp {
		err := nes.RemoveDir(nes.Tmpdir)
		if err != nil {
			log.Printf("Remove tmp dir %v failed: %v",nes.Tmpdir,err)
		}
	}
	ui.Run(NES)
}
