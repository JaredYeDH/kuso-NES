package nes

import (
	"fmt"
	"log"
)

type Mapper interface {
	Read(address uint16) byte
	Write(address uint16, val byte)
	Run()
}

func NewMapper(nes *NES) (Mapper, error) {
	log.Printf("Mapper type: %d", nes.Cartridge.Mapper)
	switch nes.Cartridge.Mapper {
	case 0, 2:
		return NewMapper2(nes.Cartridge), nil
	case 1:
		return NewMapper1(nes.Cartridge), nil
	case 3:
		return NewMapper3(nes.Cartridge),nil
	case 4:
		return NewMapper4(nes,nes.Cartridge),nil
	case 7:
		return NewMapper7(nes.Cartridge),nil
	}
	return nil, fmt.Errorf("Unknown mapper unmber: %d", nes.Cartridge.Mapper)
}
