package nes

import (
	"image"
	"image/color"
)

type PPU struct {
	Memory           // Memory interface
	NES *NES

	Cycle    int    // 0-340
	ScanLine int    // 0-261
	Frame    uint64 // Frame counter

	Palette		[64]color.RGBA
	nameTableData [2048]byte
	oamData       [256]byte
	front         *image.RGBA
	back          *image.RGBA

	// Registers
	v uint16 // Current vram address (15 bit)
	t uint16 // Temporary vram address (15 bit)
	x byte   // Fine x scroll (3 bit)
	w byte   // Write toggle (1 bit)
	f byte   // Even/odd frame flag (1 bit)

	register byte

	// NMI flag
	nmiOccurred bool
	nmiOutput   bool
	nmiPrevious bool
	nmiDelay    byte

	// Temporary
	nameTableByte      byte
	attributeTableByte byte
	lowTileByte        byte
	highTileByte       byte
	tileData           uint64
	spriteCount      int
	spritePatterns   [8]uint32
	spritePositions  [8]byte
	spritePriorities [8]byte
	spriteIndexes    [8]byte

	// Flags
	fNameTable       byte // 0: $2000; 1: $2400; 2: $2800; 3: $2C00
	fIncrement       byte // 0: add 1; 1: add 32
	fSpriteTable     byte // 0: $0000; 1: $1000; ignored in 8x16 mode
	fBackgroundTable byte // 0: $0000; 1: $1000
	fSpriteSize      byte // 0: 8x8; 1: 8x16
	fMasterSlave     byte // 0: read EXT; 1: write EXT
	fGrayscale          byte // 0: color; 1: grayscale
	fShowLeftBackground byte // 0: hide; 1: show
	fShowLeftSprites    byte // 0: hide; 1: show
	fShowBackground     byte // 0: hide; 1: show
	fShowSprites        byte // 0: hide; 1: show
	fRedTint            byte // 0: normal; 1: emphasized
	fGreenTint          byte // 0: normal; 1: emphasized
	fBlueTint           byte // 0: normal; 1: emphasized
	fSpriteZeroHit  byte
	fSpriteOverflow byte

	oamAddress byte
	bufferedData byte
}

func NewPPU(nes *NES) *PPU {
	ppu := PPU{Memory:nes.PPUMemory,NES:nes}
	ppu.Reset()
	ppu.front = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.back = image.NewRGBA(image.Rect(0, 0, 256, 240))
	return &ppu
}

func (p *PPU) Reset() {
	p.Cycle = 340
	p.ScanLine = 240
	p.Frame = 0
}

func (p *PPU) ReadRegister(address uint16) byte {
	return 0
}

func (ppu *PPU) WriteRegister(address uint16, val byte) {

}