package nes

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
	case address > 0x8000:
		idx := (int(address) - 0x8000) % len(c.PRG)
		return c.PRG[idx]
	case address >= 0x6000:
		idx := int(address) - 0x6000
		return c.SRAM[idx]
	default:
		return 0
	}
}

func (c *Cartridge) Write(address uint16, val byte) {
	if address >= 0x6000 && address < 0x8000 {
		idx := int(address) - 0x6000
		c.SRAM[idx] = val
	}
}
