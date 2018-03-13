package nes

import "log"

type Mapper3 struct {
	*Cartridge
	chrBank  int
	prgBank1 int
	prgBank2 int
}

func NewMapper3(cartridge *Cartridge) Mapper {
	prgBanks := len(cartridge.PRG) / 0x4000
	return &Mapper3{cartridge, 0, 0, prgBanks - 1}
}

func (m *Mapper3) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		index := m.chrBank*0x2000 + int(address)
		return m.CHR[index]
	case address >= 0xC000:
		index := m.prgBank2*0x4000 + int(address-0xC000)
		return m.PRG[index]
	case address >= 0x8000:
		index := m.prgBank1*0x4000 + int(address-0x8000)
		return m.PRG[index]
	case address >= 0x6000:
		index := int(address) - 0x6000
		return m.SRAM[index]
	default:
		log.Fatalf("Illegal mapper3 read at address: $%04X", address)
	}
	return 0
}

func (m *Mapper3) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		index := m.chrBank*0x2000 + int(address)
		m.CHR[index] = val
	case address >= 0x8000:
		m.chrBank = int(val & 3)
	case address >= 0x6000:
		index := int(address) - 0x6000
		m.SRAM[index] = val
	default:
		log.Fatalf("Illegal mapper2 write at address: $%04X", address)
	}
}

func (m *Mapper3) Run() {
	return
}
