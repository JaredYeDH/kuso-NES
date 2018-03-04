package nes

import (
	"flag"
	"image"
	"image/color"
)

type PPU struct {
	Memory // Memory interface
	NES    *NES

	Cycle    int    // 0-340
	ScanLine int    // 0-261
	Frame    uint64 // Frame counter

	Palette       [32]byte
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
	spriteCount        int
	spritePatterns     [8]uint32
	spritePositions    [8]byte
	spritePriorities   [8]byte
	spriteIndexes      [8]byte

	// Flags
	fNameTable          byte // 0: $2000; 1: $2400; 2: $2800; 3: $2C00
	fIncrement          byte // 0: add 1; 1: add 32
	fSpriteTable        byte // 0: $0000; 1: $1000; ignored in 8x16 mode
	fBackgroundTable    byte // 0: $0000; 1: $1000
	fSpriteSize         byte // 0: 8x8; 1: 8x16
	fMasterSlave        byte // 0: read EXT; 1: write EXT
	fGrayscale          byte // 0: color; 1: grayscale
	fShowLeftBackground byte // 0: hide; 1: show
	fShowLeftSprites    byte // 0: hide; 1: show
	fShowBackground     byte // 0: hide; 1: show
	fShowSprites        byte // 0: hide; 1: show
	fRedTint            byte // 0: normal; 1: emphasized
	fGreenTint          byte // 0: normal; 1: emphasized
	fBlueTint           byte // 0: normal; 1: emphasized
	fSpriteZeroHit      byte
	fSpriteOverflow     byte

	oamAddress   byte
	bufferedData byte
}

func NewPPU(nes *NES) *PPU {
	ppu := PPU{Memory: nes.PPUMemory, NES: nes}
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

// Registers
// Ref: http://wiki.nesdev.com/w/index.php/PPU_registers
// And a countless number of documents from the internet.

func (p *PPU) ReadRegister(address uint16) byte {
	switch address % 0x2000 {
	case 2:
		return p.rStatus()
	case 4:
		return p.rOAMData()
	case 7:
		return p.rData()
	}
}

func (p *PPU) WriteRegister(address uint16, val byte) {
	p.register = val
	switch address % 0x2000 {
	case 0:
		p.wCtrl(val)
	case 1:
		p.wMask(val)
	case 3:
		p.wOAMAddr(val)
	case 4:
		p.wOAMData(val)
	case 5:
		p.wScroll(val)
	case 6:
		p.wAddr(val)
	case 7:
		p.wData(val)
	case 14:
		p.wOAMDMA(val)

	}
}

func (p *PPU) nChange() {
	nmi := p.nmiOutput && p.nmiOccurred
	if nmi && !p.nmiPrevious {
		p.nmiDelay = 20
	}
	p.nmiPrevious = nmi
}

// $2000 - PPUCTRL

func (p *PPU) wCtrl(val byte) {
	p.fNameTable = val&1 | (((val >> 1) & 1) << 1)
	p.fIncrement = (val >> 2) & 1
	p.fSpriteTable = (val >> 3) & 1
	p.fBackgroundTable = (val >> 4) & 1
	p.fSpriteTable = (val >> 5) & 1
	p.fMasterSlave = (val >> 6) & 1
	p.nmiOutput = (((val >> 7) & 1) == 1)
	p.nChange()
	p.t = (p.t & 0xF3FF) | uint16(val&1|(((val>>1)&1)<<1)<<10)
}

// $2001 - PPUMASK

func (p *PPU) wMask(value byte) {
	p.fGrayscale = (value >> 0) & 1
	p.fShowLeftBackground = (value >> 1) & 1
	p.fShowLeftSprites = (value >> 2) & 1
	p.fShowBackground = (value >> 3) & 1
	p.fShowSprites = (value >> 4) & 1
	p.fRedTint = (value >> 5) & 1
	p.fGreenTint = (value >> 6) & 1
	p.fBlueTint = (value >> 7) & 1
}

// $2002 - PPUSTATUS

func (p *PPU) rStatus() byte {
	out := p.register & 0x1F
	out |= p.fSpriteOverflow << 5
	out |= p.fSpriteZeroHit << 6
	if p.nmiOccurred {
		out = out | 1<<7
	}
	p.nmiOccurred = false
	p.nChange()
	p.w = 0
	return out
}

// $2003 - OAMADDR

func (p *PPU) wOAMAddr(val byte) {
	p.oamAddress = val
}

// $2004 - OAMDATA
func (p *PPU) rOAMData() byte {
	return p.oamData[p.oamAddress]
}

func (p *PPU) wOAMData(val byte) {
	p.oamData[p.oamAddress] = val
	p.oamAddress++
}

// $2005 - PPUSCROLL

func (p *PPU) wScroll(val byte) {
	if p.w == 0 {
		p.t = (p.t & 0xFFED) | (uint16(val) >> 3)
		p.x = val & 7
		p.w = 1
	} else {
		p.t = (p.t & 0x8FFF) | ((uint16(val) & 7) << 12)
		p.t = (p.t & 0xFC1F) | ((uint16(val) & 0xF8) << 2)
		p.w = 0
	}
}

// $2006 - PPUADDR

func (p *PPU) wAddr(val byte) {
	if p.w == 0 {
		p.t = (p.t & 0x80FF) | ((uint16(val) & 0x3F) << 8)
		p.w = 1
	} else {
		p.t = (p.t & 0xFF00) | (uint16(val))
		p.v = p.t
		p.w = 0
	}
}

// $2007 - PPUDATA

func (p *PPU) rData() byte {
	out := p.Read(p.v)
	if p.v%0x4000 < 0x3F00 {
		buf := p.bufferedData
		p.bufferedData = out
		out = buf
	} else {
		p.bufferedData = p.Read(p.v - 0x1000)
	}

	// Change address
	if p.fIncrement == 0 {
		p.v++
	} else {
		p.v += 1 << 5
	}

	return out
}

func (p *PPU) wData(val byte) {
	p.Write(p.v, val)

	// Change address
	if p.fIncrement == 0 {
		p.v++
	} else {
		p.v += 1 << 5
	}
}

// $4014 - OAMDMA

func (p *PPU) wOAMDMA(value byte) {
	address := uint16(value) << 8
	for i := 0; i < 256; i++ {
		p.oamData[p.oamAddress] = p.NES.CPU.Read(address)
		p.oamAddress++
		address++
	}
	p.NES.CPU.stall += 513
	if p.NES.CPU.Cycles&1 == 1 {
		p.NES.CPU.stall++
	}
}
