package nes

import "log"

type Mapper2 struct {
	*Cartridge
	prgBank  int
	prgBank1 int
	prgBank2 int
}

func NewMapper2(c *Cartridge) Mapper {
	prgBank := len(c.PRG) / 0x4000
	return &Mapper2{c, prgBank, 0, prgBank - 1}
}

func (m *Mapper2) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return m.CHR[address]
	case address >= 0xC000:
		idx := m.prgBank2*0x4000 + int(address-0xC000)
		return m.PRG[idx]
	case address >= 0x8000:
		idx := m.prgBank1*0x4000 + int(address-0x8000)
		return m.PRG[idx]
	case address >= 0x6000:
		idx := int(address) - 0x6000
		return m.SRAM[idx]
	default:
		log.Fatalf("Illegal mapper2 read at address: $%04X", address)
	}
	return 0
}

func (m *Mapper2) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		m.CHR[address] = val
	case address >= 0x8000:
		m.prgBank1 = int(val) % m.prgBank
	case address >= 0x6000:
		idx := int(address) - 0x6000
		m.SRAM[idx] = val
	default:
		log.Fatalf("Illegal mapper2 write at address: 0x%04X", address)
	}
}

func (m *Mapper2) Run() {
	return
}
