package nes

// https://github.com/asfdfdfd/fceux/blob/master/src/boards/mmc3.cpp
import (
	"log"
)

type Mapper4 struct {
	*Cartridge
	nes        *NES
	register   byte
	registers  [8]byte
	prgMode    byte
	chrMode    byte
	prgOffsets [4]int
	chrOffsets [8]int
	reload     byte
	counter    byte
	irqEnable  bool
}

func NewMapper4(nes *NES, cartridge *Cartridge) Mapper {
	m := Mapper4{Cartridge: cartridge, nes: nes}
	m.prgOffsets[0] = m.prgBankOffset(0)
	m.prgOffsets[1] = m.prgBankOffset(1)
	m.prgOffsets[2] = m.prgBankOffset(-2)
	m.prgOffsets[3] = m.prgBankOffset(-1)
	return &m
}

func (m *Mapper4) HandleScanLine() {
	if m.counter == 0 {
		m.counter = m.reload
	} else {
		m.counter--
		if m.counter == 0 && m.irqEnable {
			m.nes.CPU.tIRQ()
		}
	}
}

func (m *Mapper4) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		bank := address / 0x0400
		offset := address % 0x0400
		return m.CHR[m.chrOffsets[bank]+int(offset)]
	case address >= 0x8000:
		address = address - 0x8000
		bank := address / 0x2000
		offset := address % 0x2000
		return m.PRG[m.prgOffsets[bank]+int(offset)]
	case address >= 0x6000:
		return m.SRAM[int(address)-0x6000]
	default:
		log.Fatalf("Illegal mapper4 read at address: $%04X", address)
	}
	return 0
}

func (m *Mapper4) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		bank := address / 0x0400
		offset := address % 0x0400
		m.CHR[m.chrOffsets[bank]+int(offset)] = val
	case address >= 0x8000:
		m.wRegister(address, val)
	case address >= 0x6000:
		m.SRAM[int(address)-0x6000] = val
	default:
		log.Fatalf("Illegal mapper4 read at address: $%04X", address)
	}
}

func (m *Mapper4) wRegister(address uint16, val byte) {
	switch {
	case address <= 0x9FFF && address%2 == 0:
		m.wBankSelect(val)
	case address <= 0x9FFF && address%2 == 1:
		m.wBankData(val)
	case address <= 0xBFFF && address%2 == 0:
		m.wMirror(val)
	case address <= 0xBFFF && address%2 == 1:
		m.wProtect(val)
	case address <= 0xDFFF && address%2 == 0:
		m.wIRQLatch(val)
	case address <= 0xDFFF && address%2 == 1:
		m.wIRQReload(val)
	case address <= 0xFFFF && address%2 == 0:
		m.wIRQDisable(val)
	case address <= 0xFFFF && address%2 == 1:
		m.wIRQEnable(val)
	}
}

func (m *Mapper4) wBankSelect(val byte) {
	m.prgMode = (val >> 6) & 1
	m.chrMode = (val >> 7) & 1
	m.register = val & 7
	m.updateOffsets()
}

func (m *Mapper4) wBankData(val byte) {
	m.registers[m.register] = val
	m.updateOffsets()
}

func (m *Mapper4) updateOffsets() {
	switch m.prgMode {
	case 0:
		m.prgOffsets[0] = m.prgBankOffset(int(m.registers[6]))
		m.prgOffsets[1] = m.prgBankOffset(int(m.registers[7]))
		m.prgOffsets[2] = m.prgBankOffset(-2)
		m.prgOffsets[3] = m.prgBankOffset(-1)
	case 1:
		m.prgOffsets[0] = m.prgBankOffset(-2)
		m.prgOffsets[1] = m.prgBankOffset(int(m.registers[7]))
		m.prgOffsets[2] = m.prgBankOffset(int(m.registers[6]))
		m.prgOffsets[3] = m.prgBankOffset(-1)
	}
	switch m.chrMode {
	case 0:
		m.chrOffsets[0] = m.chrBankOffset(int(m.registers[0] & 0xFE))
		m.chrOffsets[1] = m.chrBankOffset(int(m.registers[0] | 0x01))
		m.chrOffsets[2] = m.chrBankOffset(int(m.registers[1] & 0xFE))
		m.chrOffsets[3] = m.chrBankOffset(int(m.registers[1] | 0x01))
		m.chrOffsets[4] = m.chrBankOffset(int(m.registers[2]))
		m.chrOffsets[5] = m.chrBankOffset(int(m.registers[3]))
		m.chrOffsets[6] = m.chrBankOffset(int(m.registers[4]))
		m.chrOffsets[7] = m.chrBankOffset(int(m.registers[5]))
	case 1:
		m.chrOffsets[0] = m.chrBankOffset(int(m.registers[2]))
		m.chrOffsets[1] = m.chrBankOffset(int(m.registers[3]))
		m.chrOffsets[2] = m.chrBankOffset(int(m.registers[4]))
		m.chrOffsets[3] = m.chrBankOffset(int(m.registers[5]))
		m.chrOffsets[4] = m.chrBankOffset(int(m.registers[0] & 0xFE))
		m.chrOffsets[5] = m.chrBankOffset(int(m.registers[0] | 0x01))
		m.chrOffsets[6] = m.chrBankOffset(int(m.registers[1] & 0xFE))
		m.chrOffsets[7] = m.chrBankOffset(int(m.registers[1] | 0x01))
	}
}

func (m *Mapper4) wMirror(val byte) {
	switch val & 1 {
	case 0:
		m.Cartridge.Mirror = MirrorVertical
	case 1:
		m.Cartridge.Mirror = MirrorHorizontal
	}
}

func (m *Mapper4) wProtect(val byte) {
	return
}

func (m *Mapper4) wIRQLatch(val byte) {
	m.reload = val
}

func (m *Mapper4) wIRQReload(val byte) {
	m.counter = 0
}

func (m *Mapper4) wIRQDisable(val byte) {
	m.irqEnable = false
}

func (m *Mapper4) wIRQEnable(val byte) {
	m.irqEnable = true
}

func (m *Mapper4) prgBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.PRG) / 0x2000
	offset := index * 0x2000
	if offset < 0 {
		offset += len(m.PRG)
	}
	return offset
}

func (m *Mapper4) chrBankOffset(index int) int {
	if index >= 0x80 {
		index -= 0x100
	}
	index %= len(m.CHR) / 0x0400
	offset := index * 0x0400
	if offset < 0 {
		offset += len(m.CHR)
	}
	return offset
}

func (m *Mapper4) Run() {
	ppu := m.nes.PPU
	if ppu.Cycle != 280 {
		return
	}
	if ppu.ScanLine > 239 && ppu.ScanLine < 261 {
		return
	}
	if ppu.fShowBackground == 0 && ppu.fShowSprites == 0 {
		return
	}
	m.HandleScanLine()
}
