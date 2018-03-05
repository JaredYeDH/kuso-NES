package nes

import "log"

const (
	MirrorHorizontal = iota
	MirrorVertical
	MirrorQuad
)

type Cartridge struct {
	PRG     []byte
	CHR     []byte
	SRAM    []byte
	Mapper  int
	Mirror  int
	Battery bool
}

func (c *Cartridge) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		return c.CHR[address]
	case address >= 0x8000:
		idx := (int(address) - 0x8000) % len(c.PRG)
		return c.PRG[idx]
	case address >= 0x6000:
		idx := int(address) - 0x6000
		return c.SRAM[idx]
	default:
		log.Fatalf("Illegal cartridge read at address: $%04X", address)
	}
	return 0
}

func (c *Cartridge) Write(address uint16, val byte) {
	switch {
	case address < 0x2000:
		c.CHR[address] = val
	case address >= 0x8000:
		break
	case address >= 0x6000:
		index := int(address) - 0x6000
		c.SRAM[index] = val
	default:
		log.Fatalf("Illegal cartridge write at address: $%04X", address)
	}
}
