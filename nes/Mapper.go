package nes

import "fmt"

type Mapper interface {
	Read(address uint16) byte
	Write(address uint16, val byte)
	Run()
}

func NewMapper(nes *NES) (Mapper, error) {
	switch nes.Cartridge.Mapper {
	case 0,2:
		return NewMapper2(nes.Cartridge),nil
	}
	return nil, fmt.Errorf("Unknown mapper unmber: %d", nes.Cartridge.Mapper)
}
