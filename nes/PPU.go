package nes

import "image"

// Ref: http://nesdev.com/2C02%20technical%20reference.TXT
//      And a lot of blogs, wikis, and so on.
//		Thanks to the authors of those documents.

type PPU struct {
	Memory // Memory interface
	NES    *NES

	Cycle    int    // 0-340
	ScanLine int    // 0-261
	Frame    uint64 // Frame counter

	palette       [32]byte
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

	// For Background
	nameTableByte      byte
	attributeTableByte byte
	lowTileByte        byte
	highTileByte       byte
	tileData           uint64

	// For Sprite
	spriteCount      int
	spritePatterns   [8]uint32
	spritePositions  [8]byte
	spritePriorities [8]byte
	spriteIndexes    [8]byte

	// PPUCTRL
	fNameTable       byte // 0: $2000; 1: $2400; 2: $2800; 3: $2C00
	fIncrement       byte // 0: add 1; 1: add 32
	fSpriteTable     byte // 0: $0000; 1: $1000; ignored in 8x16 mode
	fBackgroundTable byte // 0: $0000; 1: $1000
	fSpriteSize      byte // 0: 8x8; 1: 8x16
	fMasterSlave     byte // 0: read EXT; 1: write EXT

	// PPUMASK
	fGrayscale          byte // 0: color; 1: grayscale
	fShowLeftBackground byte // 0: hide; 1: show
	fShowLeftSprites    byte // 0: hide; 1: show
	fShowBackground     byte // 0: hide; 1: show
	fShowSprites        byte // 0: hide; 1: show
	fRedTint            byte // 0: normal; 1: emphasized
	fGreenTint          byte // 0: normal; 1: emphasized
	fBlueTint           byte // 0: normal; 1: emphasized

	// PPUSTATUS
	fSpriteZeroHit  byte
	fSpriteOverflow byte

	// OAMADDR
	oamAddress byte

	// PPUDATA
	bufferedData byte
}

func NewPPU(nes *NES) *PPU {
	ppu := PPU{Memory: NewPPUMemory(nes), NES: nes}
	ppu.front = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.back = image.NewRGBA(image.Rect(0, 0, 256, 240))
	ppu.Reset()
	return &ppu
}

func (p *PPU) Reset() {
	p.Cycle = 340
	p.ScanLine = 240
	p.Frame = 0
	p.wCtrl(0)
	p.wMask(0)
	p.wOAMAddr(0)
}

// Run runs a PPU cycle per execution

func (p *PPU) tik() {
	if p.nmiDelay > 0 {
		p.nmiDelay--
		if p.nmiDelay == 0 && p.nmiOutput && p.nmiOccurred {
			p.NES.CPU.tNMI()
		}
	}

	if p.fShowBackground != 0 || p.fShowSprites != 0 {
		if p.f == 1 && p.ScanLine == 261 && p.Cycle == 339 {
			p.Cycle = 0
			p.ScanLine = 0
			p.Frame++
			p.f ^= 1
			return
		}
	}
	p.Cycle++
	if p.Cycle > 340 {
		p.Cycle = 0
		p.ScanLine++
		if p.ScanLine > 261 {
			p.ScanLine = 0
			p.Frame++
			p.f ^= 1
		}
	}
}

func (p *PPU) Run() {
	p.tik()

	if p.fShowBackground != 0 || p.fShowSprites != 0 {
		if p.ScanLine < 240 && p.Cycle >= 1 && p.Cycle <= 256 {
			p.renderPixel()
		}
		if ((p.Cycle >= 321 && p.Cycle <= 336) || (p.Cycle >= 1 && p.Cycle <= 256)) && p.ScanLine < 240 {
			p.tileData <<= 4
			switch p.Cycle % 8 {
			case 1:
				p.getNtableByte()
			case 3:
				p.getATableByte()
			case 5:
				p.getLTileByte()
			case 7:
				p.getHTileByte()
			case 0:
				p.storeTData()
			}
		}
		if p.ScanLine == 261 && p.Cycle >= 280 && p.Cycle <= 304 {
			p.cpV()
		}
		if p.ScanLine < 240 || p.ScanLine == 261 {
			if ((p.Cycle >= 321 && p.Cycle <= 336) || (p.Cycle >= 1 && p.Cycle <= 256)) && p.Cycle%8 == 0 {
				p.iHori()
			}
			if p.Cycle == 256 {
				p.iVert()
			}
			if p.Cycle == 257 {
				p.cpH()
			}
		}
	}

	if p.fShowBackground != 0 || p.fShowSprites != 0 {
		if p.Cycle == 257 {
			if p.ScanLine < 240 {
				p.evalSprites()
			} else {
				p.spriteCount = 0
			}
		}
	}

	if p.ScanLine == 241 && p.Cycle == 1 {
		p.setVertBlank()
	}
	if p.ScanLine == 261 && p.Cycle == 1 {
		p.clrVertBlank()
		p.fSpriteZeroHit = 0
		p.fSpriteOverflow = 0
	}
}

// Palette

func (p *PPU) rPalette(address uint16) byte {
	if address >= 16 && address%4 == 0 {
		address = address - 16
	}
	return p.palette[address]
}

func (p *PPU) wPalette(address uint16, val byte) {
	if address >= 16 && address%4 == 0 {
		address = address - 16
	}
	p.palette[address] = val
}

// Registers
// Ref: http://wiki.nesdev.com/w/index.php/PPU_registers
// And a countless number of documents from the internet.

func (p *PPU) ReadRegister(address uint16) byte {
	switch address  {
	case 0x2002:
		return p.rStatus()
	case 0x2004:
		return p.rOAMData()
	case 0x2007:
		return p.rData()
	}
	return 0
}

func (p *PPU) WriteRegister(address uint16, val byte) {
	p.register = val
	switch address{
	case 0x2000:
		p.wCtrl(val)
	case 0x2001:
		p.wMask(val)
	case 0x2003:
		p.wOAMAddr(val)
	case 0x2004:
		p.wOAMData(val)
	case 0x2005:
		p.wScroll(val)
	case 0x2006:
		p.wAddr(val)
	case 0x2007:
		p.wData(val)
	case 0x4014:
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
	p.fNameTable = val & 3
	p.fIncrement = (val >> 2) & 1
	p.fSpriteTable = (val >> 3) & 1
	p.fBackgroundTable = (val >> 4) & 1
	p.fSpriteSize = (val >> 5) & 1
	p.fMasterSlave = (val >> 6) & 1
	p.nmiOutput = ((val >> 7) & 1) == 1
	p.nChange()
	p.t = (p.t & 0xF3FF) | ((uint16(val) & 0x03) << 10)
}

// $2001 - PPUMASK

func (p *PPU) wMask(val byte) {
	p.fGrayscale = (val >> 0) & 1
	p.fShowLeftBackground = (val >> 1) & 1
	p.fShowLeftSprites = (val >> 2) & 1
	p.fShowBackground = (val >> 3) & 1
	p.fShowSprites = (val >> 4) & 1
	p.fRedTint = (val >> 5) & 1
	p.fGreenTint = (val >> 6) & 1
	p.fBlueTint = (val >> 7) & 1
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
		p.t = (p.t & 0x8FFF) | ((uint16(val) & 0x7) << 12)
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
		p.v += 32
	}

	return out
}

func (p *PPU) wData(val byte) {
	p.Write(p.v, val)

	// Change address
	if p.fIncrement == 0 {
		p.v++
	} else {
		p.v += 32
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
	if p.NES.CPU.Cycles%2 == 1 {
		p.NES.CPU.stall++
	}
}

func (p *PPU) iHori() {
	if p.v&0x001F == 31 {
		p.v = p.v & 0xFFE0
		p.v = p.v ^ 0x400
	} else {
		p.v++
	}
}

func (p *PPU) iVert() {
	if p.v&0x7000 != 0x7000 {
		p.v += 0x1000
	} else {
		p.v = p.v & 0x8FFF
		y := (p.v & 0x03E0) >> 5
		if y == 29 {
			y = 0
			p.v = p.v ^ 0x0800
		} else if y == 31 {
			y = 0
		} else {
			y++
		}
		p.v = (p.v & 0xFC1F) | (y << 5)
	}
}

func (p *PPU) cpH() {
	p.v = (p.v & 0xFBE0) | (p.t & 0x041F)
}

func (p *PPU) cpV() {
	p.v = (p.v & 0x841F) | (p.t & 0x7BE0)
}

func (p *PPU) setVertBlank() {
	p.back, p.front = p.front, p.back
	p.nmiOccurred = true
	p.nChange()
}

func (p *PPU) clrVertBlank() {
	p.nmiOccurred = false
	p.nChange()
}

func (p *PPU) getNtableByte() {
	v := p.v
	address := 0x2000|(v & 0x0FFF)
	p.nameTableByte = p.Read(address)
}

func (p *PPU) getATableByte() {
	v := p.v
	address := 0x23C0 | (v & 0x0C00) | ((v >> 4) & 0x38) | ((v >> 2) & 0x7)
	i := ((v >> 4) & 4) | (v & 2)
	p.attributeTableByte = (((p.Read(address)) >> i) & 3) << 2
}

func (p *PPU) getLTileByte() {
	fineY := (p.v >> 12) & 7
	table := p.fBackgroundTable
	tile := p.nameTableByte
	address := 0x1000*uint16(table) + uint16(tile)*16 + fineY
	p.lowTileByte = p.Read(address)
}

func (p *PPU) getHTileByte() {
	fineY := (p.v >> 12) & 7
	table := p.fBackgroundTable
	tile := p.nameTableByte
	address := 0x1000*uint16(table) + uint16(tile)*16 + fineY
	p.lowTileByte = p.Read(address + 8)
}

func (p *PPU) getTData() uint32 {
	return uint32(p.tileData >> 32)
}

func (p *PPU) storeTData() {
	var data uint32
	for i := 0; i < 8; i++ {
		a := p.attributeTableByte
		p1 := (p.lowTileByte & 0x80) >> 7
		p2 := (p.highTileByte & 0x80) >> 6
		p.lowTileByte = p.lowTileByte << 1
		p.highTileByte = p.highTileByte << 1
		data = data << 4
		data |= uint32(a | p1 | p2)
	}
	p.tileData = p.tileData | uint64(data)
}

func (p *PPU) backPixel() byte {
	if p.fShowBackground == 0 {
		return 0
	}
	data := p.getTData() >> ((7 - p.x) * 4)
	return byte(data & 0x0F)
}

func (p *PPU) spritePixel() (byte, byte) {
	if p.fShowSprites == 0 {
		return 0, 0
	}
	for i := 0; i < p.spriteCount; i++ {
		temp := (p.Cycle - 1)- int(p.spritePositions[i])
		if temp < 0 || temp > 7 {
			continue
		}
		temp = 7 - temp
		color := byte((p.spritePatterns[i] >> byte(temp*4)) & 0xF)
		if color%4 == 0 {
			continue
		}
		return byte(i), color
	}
	return 0, 0
}

func (p *PPU) renderPixel() {
	x := p.Cycle - 1
	y := p.ScanLine

	back := p.backPixel()

	i, sprite := p.spritePixel()

	if x < 8 {
		if p.fShowLeftBackground == 0 {
			back = 0
		}
		if p.fShowLeftSprites == 0 {
			sprite = 0
		}
	}

	var color byte
	b := (back%4 + 3)/4
	s := (sprite%4 + 3)/4
	o := b << 1 | s
	switch o {
	case 0:
		color = 0
	case 1:
		color = sprite | 0x10
	case 2:
		color = back
	case 3:
		if p.spriteIndexes[i] == 0 && x < 255 {
			p.fSpriteZeroHit = 1
		}
		if p.spritePriorities[i] == 0 {
			color = sprite | 0x10
		} else {
			color = back
		}
	}

	realcolor := Palette[p.rPalette(uint16(color)%64)]

	p.back.SetRGBA(x, y, realcolor)
}

func (p *PPU) evalSprites() {
	var h int
	if p.fSpriteSize == 0 {
		h = 8
	} else {
		h = 16
	}
	count := 0
	for i := 0; i < 64; i++ {
		y := p.oamData[i*4]
		a := p.oamData[i*4+2]
		x := p.oamData[i*4+3]
		r := p.ScanLine - int(y)
		if r < 0 || r >= h {
			continue
		}
		if count < 8 {
			p.spritePatterns[count] = p.getSPattern(i, r)
			p.spritePositions[count] = x
			p.spritePriorities[count] = (a >> 5) & 1
			p.spriteIndexes[count] = byte(i)
		}
		count++
	}
	if count > 8 {
		count = 8
		p.fSpriteOverflow = 1
	}
	p.spriteCount = count
}

func (p *PPU) getSPattern(i, r int) uint32 {
	tile := p.oamData[i*4+1]
	attributes := p.oamData[i*4+2]
	var address uint16
	if p.fSpriteSize == 0 {
		if attributes&0x80 == 0x80 {
			r = 7 - r
		}
		table := p.fSpriteTable
		address = 0x1000*uint16(table) + uint16(tile)*16 + uint16(r)
	} else {
		if attributes&0x80 == 0x80 {
			r = 15 - r
		}
		table := tile & 1
		tile &= 0xFE
		if r > 7 {
			tile++
			r -= 8
		}
		address = 0x1000*uint16(table) + uint16(tile)*16 + uint16(r)
	}
	a := (attributes & 3) << 2
	lowTileByte := p.Read(address)
	highTileByte := p.Read(address + 8)
	var pattern uint32
	for i := 0; i < 8; i++ {
		var b, c byte
		if attributes&0x40 == 0x40 {
			b = lowTileByte & 1
			c = (highTileByte & 1) << 1
			lowTileByte = lowTileByte >> 1
			highTileByte = highTileByte >> 1
		} else {
			b = (lowTileByte & 0x80) >> 7
			c = (highTileByte & 0x80) >> 6
			lowTileByte = lowTileByte << 1
			highTileByte = highTileByte << 1
		}
		pattern = pattern << 4
		pattern = pattern | uint32(a|b|c)
	}
	return pattern
}
