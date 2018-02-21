package nes

const (
	MirrorHorizontal = 0
	MirrorVertical   = 1
	MirrorQuad       = 2
)

type Cartridge struct {
	PRG     []byte
	CHR     []byte
	Mapper  int
	Mirror  int
	Battery bool
}

func (c *Cartridge) Read(address uint16) byte {
	idx := int(address - 0x8000) % len(c.PRG)
	return c.PRG[idx]
}