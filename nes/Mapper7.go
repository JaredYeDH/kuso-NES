package nes

import (
	"log"
)

type Mapper7 struct {
	*Cartridge
	prgBank int
}

func NewMapper7(cartridge *Cartridge) Mapper {
	return &Mapper7{cartridge, 0}
}

func (m *Mapper7) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return m.CHR[address]
	case address >= 0x8000:
		index := m.prgBank*0x8000 + int(address-0x8000)
		return m.PRG[index]
	case address >= 0x6000:
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		log.Fatalf("Illegal mapper7 read at address: $%04X", address)
	}
	return 0
}

func (m *Mapper7) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		m.CHR[address] = val
	case address >= 0x8000:
		m.prgBank = int(val & 7)
		switch val & 0x10 {
		case 0x00:
			m.Cartridge.Mirror = MirrorSingle0
		case 0x10:
			m.Cartridge.Mirror = MirrorSingle1
		}
	case address >= 0x6000:
		index := int(address) - 0x6000
		m.SRAM[index] = val
	default:
		log.Fatalf("Illegal mapper7 write at address: $%04X", address)
	}
}

func (m *Mapper7) Run() {
	return
}