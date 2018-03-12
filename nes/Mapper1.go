package nes

import (
	"log"
)

type Mapper1 struct {
	*Cartridge
	shiftRegister byte
	control       byte
	prgMode       byte
	chrMode       byte
	prgBank       byte
	chrBank0      byte
	chrBank1      byte
	prgOffset     [2]int
	chrOffset     [2]int
}

func NewMapper1(c *Cartridge) Mapper {
	m := Mapper1{}
	m.Cartridge = c
	m.shiftRegister = 0x10
	m.prgOffset[1] = m.prgBankOffset(-1)
	return &m
}

func (m *Mapper1) Run() {
}

func (m *Mapper1) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		bank := address / 0x1000
		offset := address % 0x1000
		return m.CHR[m.chrOffset[bank]+int(offset)]
	case address >= 0x8000:
		address = address - 0x8000
		bank := address / 0x4000
		offset := address % 0x4000
		return m.PRG[m.prgOffset[bank]+int(offset)]
	case address >= 0x6000:
		return m.SRAM[int(address)-0x6000]
	default:
		log.Fatalf("Illegal mapper1 read at address: $%04X", address)
	}
	return 0
}

func (m *Mapper1) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		bank := address / 0x1000
		offset := address % 0x1000
		m.CHR[m.chrOffset[bank]+int(offset)] = val
	case address >= 0x8000:
		m.loadRegister(address, val)
	case address >= 0x6000:
		m.SRAM[int(address)-0x6000] = val
	default:
		log.Fatalf("Illegal mapper1 write at address: $%04X", address)
	}
}

func (m *Mapper1) loadRegister(address uint16, val byte) {
	if val&0x80 == 0x80 {
		m.shiftRegister = 0x10
		m.wCtrl(m.control | 0x0C)
	} else {
		complete := m.shiftRegister&1 == 1
		m.shiftRegister >>= 1
		m.shiftRegister |= (val & 1) << 4
		if complete {
			m.wRegister(address, m.shiftRegister)
			m.shiftRegister = 0x10
		}
	}
}

func (m *Mapper1) wRegister(address uint16, val byte) {
	switch {
	case address <= 0x9FFF:
		m.wCtrl(val)
	case address <= 0xBFFF:
		m.wCHRBank0(val)
	case address <= 0xDFFF:
		m.wCHRBank1(val)
	case address <= 0xFFFF:
		m.wPRGBank(val)
	}
}

// Control - $8000-$9FFF
func (m *Mapper1) wCtrl(val byte) {
	m.control = val
	m.chrMode = (val >> 4) & 1
	m.prgMode = (val >> 2) & 3
	mirror := val & 3
	switch mirror {
	case 0:
		m.Cartridge.Mirror = MirrorSingle0
	case 1:
		m.Cartridge.Mirror = MirrorSingle1
	case 2:
		m.Cartridge.Mirror = MirrorVertical
	case 3:
		m.Cartridge.Mirror = MirrorHorizontal
	}
	m.updateOffset()
}

// CHR bank 0 - $A000-$BFFF
func (m *Mapper1) wCHRBank0(val byte) {
	m.chrBank0 = val
	m.updateOffset()
}

// CHR bank 1 - $C000-$DFFF
func (m *Mapper1) wCHRBank1(val byte) {
	m.chrBank1 = val
	m.updateOffset()
}

// PRG - $E000-$FFFF
func (m *Mapper1) wPRGBank(val byte) {
	m.prgBank = val & 0x0F
	m.updateOffset()
}

func (m *Mapper1) prgBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.PRG) / 0x4000
	offset := index * 0x4000
	if offset < 0 {
		offset += len(m.PRG)
	}
	return offset
}

func (m *Mapper1) chrBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.CHR) / 0x1000
	offset := index * 0x1000
	if offset < 0 {
		offset += len(m.CHR)
	}
	return offset
}

func (m *Mapper1) updateOffset() {
	switch m.prgMode {
	case 0, 1:
		m.prgOffset[0] = m.prgBankOffset(int(m.prgBank & 0xFE))
		m.prgOffset[1] = m.prgBankOffset(int(m.prgBank | 0x01))
	case 2:
		m.prgOffset[0] = 0
		m.prgOffset[1] = m.prgBankOffset(int(m.prgBank))
	case 3:
		m.prgOffset[0] = m.prgBankOffset(int(m.prgBank))
		m.prgOffset[1] = m.prgBankOffset(-1)
	}
	switch m.chrMode {
	case 0:
		m.chrOffset[0] = m.chrBankOffset(int(m.chrBank0 & 0xFE))
		m.chrOffset[1] = m.chrBankOffset(int(m.chrBank0 | 0x01))
	case 1:
		m.chrOffset[0] = m.chrBankOffset(int(m.chrBank0))
		m.chrOffset[1] = m.chrBankOffset(int(m.chrBank1))
	}
}
