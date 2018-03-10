package nes

import "log"

type Cartridge struct {
	PRG      []byte
	CHR      []byte
	SRAM     []byte
	Mapper   byte
	Mirror   byte
	Battery  byte
	prgBank  int
	chrBank  int
	prgBank1 int
	prgBank2 int
	chrBank1 int
}

func NewCartridge(prg, chr []byte, mapper, mirror, battery byte) *Cartridge {
	prgBank := len(prg) / 0x4000
	chrBank := len(chr) / 0x2000
	prgBank2 := prgBank - 1
	sram := make([]byte, 0x2000)
	cartridge := Cartridge{
		prg, chr, sram, mapper, mirror, battery,
		prgBank, chrBank, 0, prgBank2, 0}
	return &cartridge
}

func (c *Cartridge) Read(address uint16) byte {
	switch {
	case address < 0x2000:
		idx := c.chrBank1*0x2000 + int(address)
		//log.Printf("%d %x %x %x",c.chrBank,address,idx,len(c.CHR))
		return c.CHR[idx]
	case address >= 0xC000:
		idx := c.prgBank2*0x4000 + int(address-0xC000)
		return c.PRG[idx]
	case address >= 0x8000:
		idx := c.prgBank1*0x4000 + int(address-0x8000)
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
		idx := c.chrBank1*0x2000 + int(address)
		c.CHR[idx] = val
	case address >= 0x8000:
		c.prgBank1 = int(val)
	case address >= 0x6000:
		idx := int(address) - 0x6000
		c.SRAM[idx] = val
	default:
		log.Fatalf("Illegal cartridge write at address: $%04X", address)
	}
}
