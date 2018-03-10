package nes

import "image"

// Ref: http://nesdev.com/2C02%20technical%20reference.TXT
//      And a lot of blogs, wikis, and so on.
//		Thanks to the authors of those documents.
// Refactor on 2018.3.10.

type PPU struct {
	Memory // memory interface
	NES    *NES

	Cycle    int // 0-340
	ScanLine int // 0-261
	Frame    uint64

	palette       [32]byte
	nameTableData [2048]byte
	oamData       [256]byte
	front         *image.RGBA
	back          *image.RGBA

	// Registers
	v uint16
	t uint16
	x byte
	w byte
	f byte

	register  byte
	nOccurred bool
	nOutput   bool
	nPrevious bool
	nDelay    byte

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

	// PPUCTRL
	fNameTable       byte
	fIncrement       byte
	fSpriteTable     byte
	fBackgroundTable byte
	fSpriteSize      byte
	fMasterSlave     byte

	// PPUMASK
	fGrayscale          byte
	fShowLeftBackground byte
	fShowLeftSprites    byte
	fShowBackground     byte
	fShowSprites        byte
	fRedTint            byte
	fGreenTint          byte
	fBlueTint           byte

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

func (p *PPU) tik() {
	if p.nDelay > 0 {
		p.nDelay--
		if p.nDelay == 0 && p.nOutput && p.nOccurred {
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

// Run runs the ppu.
func (p *PPU) Run() {
	p.tik()

	a := p.fShowBackground != 0 || p.fShowSprites != 0
	b := p.ScanLine == 261
	c := p.ScanLine < 241
	d := p.Cycle >= 321 && p.Cycle <= 336
	e := p.Cycle >= 1 && p.Cycle <= 256
	f := b || c
	g := d || e

	if a {
		if c && e {
			p.renderPixel()
		}
		if f && g {
			p.tileData <<= 4
			switch p.Cycle % 8 {
			case 1:
				p.getNameTableByte()
			case 3:
				p.getAttributeTableByte()
			case 5:
				p.getLowTileByte()
			case 7:
				p.getHighTileByte()
			case 0:
				p.storeTileData()
			}
		}
		if b && p.Cycle >= 280 && p.Cycle <= 304 {
			p.cpY()
		}
		if f {
			if g && p.Cycle%8 == 0 {
				p.iX()
			}
			if p.Cycle == 256 {
				p.iY()
			}
			if p.Cycle == 257 {
				p.cpX()
			}
		}
	}
	if a {
		if p.Cycle == 257 {
			if c {
				p.evaluateSprites()
			} else {
				p.spriteCount = 0
			}
		}
	}
	if p.ScanLine == 241 && p.Cycle == 1 {
		p.setVerticalBlank()
	}
	if b && p.Cycle == 1 {
		p.clearVerticalBlank()
		p.fSpriteZeroHit = 0
		p.fSpriteOverflow = 0
	}
}

func (p *PPU) Reset() {
	p.Cycle = 340
	p.ScanLine = 240
	p.Frame = 0
	p.wControl(0)
	p.wMask(0)
	p.wOAMAddress(0)
}

func (p *PPU) ReadRegister(address uint16) byte {
	switch address {
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
	switch address {
	case 0x2000:
		p.wControl(val)
	case 0x2001:
		p.wMask(val)
	case 0x2003:
		p.wOAMAddress(val)
	case 0x2004:
		p.wOAMData(val)
	case 0x2005:
		p.wScroll(val)
	case 0x2006:
		p.wAddress(val)
	case 0x2007:
		p.wData(val)
	case 0x4014:
		p.wDMA(val)
	}
}

// $2000: PPUCTRL
func (p *PPU) wControl(val byte) {
	p.fNameTable = (val >> 0) & 3
	p.fIncrement = (val >> 2) & 1
	p.fSpriteTable = (val >> 3) & 1
	p.fBackgroundTable = (val >> 4) & 1
	p.fSpriteSize = (val >> 5) & 1
	p.fMasterSlave = (val >> 6) & 1
	p.nOutput = (val>>7)&1 == 1
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
	result := p.register & 0x1F
	result |= p.fSpriteOverflow << 5
	result |= p.fSpriteZeroHit << 6
	if p.nOccurred {
		result |= 1 << 7
	}
	p.nOccurred = false
	p.nChange()
	p.w = 0
	return result
}

// $2003 -  OAMADDR
func (p *PPU) wOAMAddress(val byte) {
	p.oamAddress = val
}

// $2004 - OAMDATA (r)
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
		p.t = (p.t & 0xFFE0) | (uint16(val) >> 3)
		p.x = val & 0x07
		p.w = 1
	} else {
		p.t = (p.t & 0x8FFF) | ((uint16(val) & 0x07) << 12)
		p.t = (p.t & 0xFC1F) | ((uint16(val) & 0xF8) << 2)
		p.w = 0
	}
}

// $2006 - PPUADDR
func (p *PPU) wAddress(val byte) {
	if p.w == 0 {
		p.t = (p.t & 0x80FF) | ((uint16(val) & 0x3F) << 8)
		p.w = 1
	} else {
		p.t = (p.t & 0xFF00) | uint16(val)
		p.v = p.t
		p.w = 0
	}
}

// $2007 - PPUDATA
func (p *PPU) rData() byte {
	val := p.Read(p.v)
	if p.v%0x4000 < 0x3F00 {
		buffered := p.bufferedData
		p.bufferedData = val
		val = buffered
	} else {
		p.bufferedData = p.Read(p.v - 0x1000)
	}
	if p.fIncrement == 0 {
		p.v += 1
	} else {
		p.v += 32
	}
	return val
}

func (p *PPU) wData(val byte) {
	p.Write(p.v, val)
	if p.fIncrement == 0 {
		p.v += 1
	} else {
		p.v += 32
	}
}

// $4014 - OAMDMA
func (p *PPU) wDMA(val byte) {
	cpu := p.NES.CPU
	address := uint16(val) << 8
	for i := 0; i < 256; i++ {
		p.oamData[p.oamAddress] = cpu.Read(address)
		p.oamAddress++
		address++
	}
	cpu.stall += 513
	if cpu.Cycles%2 == 1 {
		cpu.stall++
	}
}

func (p *PPU) iX() {
	if p.v&0x001F == 31 {
		p.v &= 0xFFE0
		p.v ^= 0x0400
	} else {
		p.v++
	}
}

func (p *PPU) iY() {
	if p.v&0x7000 != 0x7000 {
		p.v += 0x1000
	} else {
		p.v &= 0x8FFF
		y := (p.v & 0x03E0) >> 5
		if y == 29 {
			y = 0
			p.v ^= 0x0800
		} else if y == 31 {
			y = 0
		} else {
			y++
		}
		p.v = (p.v & 0xFC1F) | (y << 5)
	}
}

func (p *PPU) cpX() {
	p.v = (p.v & 0xFBE0) | (p.t & 0x041F)
}

func (p *PPU) cpY() {
	p.v = (p.v & 0x841F) | (p.t & 0x7BE0)
}

func (p *PPU) nChange() {
	n := p.nOutput && p.nOccurred
	if n && !p.nPrevious {
		p.nDelay = 20
	}
	p.nPrevious = n
}

func (p *PPU) setVerticalBlank() {
	p.front, p.back = p.back, p.front
	p.nOccurred = true
	p.nChange()
}

func (p *PPU) clearVerticalBlank() {
	p.nOccurred = false
	p.nChange()
}

func (p *PPU) getNameTableByte() {
	v := p.v
	address := 0x2000 | (v & 0x0FFF)
	p.nameTableByte = p.Read(address)
}

func (p *PPU) getAttributeTableByte() {
	v := p.v
	address := 0x23C0 | (v & 0x0C00) | ((v >> 4) & 0x38) | ((v >> 2) & 0x07)
	shift := ((v >> 4) & 4) | (v & 2)
	p.attributeTableByte = ((p.Read(address) >> shift) & 3) << 2
}

func (p *PPU) getLowTileByte() {
	fineY := (p.v >> 12) & 7
	table := p.fBackgroundTable
	tile := p.nameTableByte
	address := 0x1000*uint16(table) + uint16(tile)*16 + fineY
	p.lowTileByte = p.Read(address)
}

func (p *PPU) getHighTileByte() {
	fineY := (p.v >> 12) & 7
	table := p.fBackgroundTable
	tile := p.nameTableByte
	address := 0x1000*uint16(table) + uint16(tile)*16 + fineY
	p.highTileByte = p.Read(address + 8)
}

func (p *PPU) storeTileData() {
	var data uint32
	for i := 0; i < 8; i++ {
		a := p.attributeTableByte
		p1 := (p.lowTileByte & 0x80) >> 7
		p2 := (p.highTileByte & 0x80) >> 6
		p.lowTileByte <<= 1
		p.highTileByte <<= 1
		data <<= 4
		data |= uint32(a | p1 | p2)
	}
	p.tileData |= uint64(data)
}

func (p *PPU) getTileData() uint32 {
	return uint32(p.tileData >> 32)
}

// Palette

func (p *PPU) rPalette(address uint16) byte {
	if address >= 16 && address%4 == 0 {
		address -= 16
	}
	return p.palette[address]
}

func (p *PPU) wPalette(address uint16, val byte) {
	if address >= 16 && address%4 == 0 {
		address -= 16
	}
	p.palette[address] = val
}

func (p *PPU) backgroundPixel() byte {
	if p.fShowBackground == 0 {
		return 0
	}
	data := p.getTileData() >> ((7 - p.x) * 4)
	return byte(data & 0x0F)
}

func (p *PPU) spritePixel() (byte, byte) {
	if p.fShowSprites == 0 {
		return 0, 0
	}
	for i := 0; i < p.spriteCount; i++ {
		offset := (p.Cycle - 1) - int(p.spritePositions[i])
		if offset < 0 || offset > 7 {
			continue
		}
		offset = 7 - offset
		color := byte((p.spritePatterns[i] >> byte(offset*4)) & 0x0F)
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
	background := p.backgroundPixel()
	i, sprite := p.spritePixel()
	if x < 8 && p.fShowLeftBackground == 0 {
		background = 0
	}
	if x < 8 && p.fShowLeftSprites == 0 {
		sprite = 0
	}
	b := background%4 != 0
	s := sprite%4 != 0
	var color byte
	if !b && !s {
		color = 0
	} else if !b && s {
		color = sprite | 0x10
	} else if b && !s {
		color = background
	} else {
		if p.spriteIndexes[i] == 0 && x < 255 {
			p.fSpriteZeroHit = 1
		}
		if p.spritePriorities[i] == 0 {
			color = sprite | 0x10
		} else {
			color = background
		}
	}
	c := Palette[p.rPalette(uint16(color))%64]
	p.back.SetRGBA(x, y, c)
}

func (p *PPU) getSpritePattern(i, row int) uint32 {
	tile := p.oamData[i*4+1]
	attributes := p.oamData[i*4+2]
	var address uint16
	if p.fSpriteSize == 0 {
		if attributes&0x80 == 0x80 {
			row = 7 - row
		}
		table := p.fSpriteTable
		address = 0x1000*uint16(table) + uint16(tile)*16 + uint16(row)
	} else {
		if attributes&0x80 == 0x80 {
			row = 15 - row
		}
		table := tile & 1
		tile &= 0xFE
		if row > 7 {
			tile++
			row -= 8
		}
		address = 0x1000*uint16(table) + uint16(tile)*16 + uint16(row)
	}
	a := (attributes & 3) << 2
	lowTileByte := p.Read(address)
	highTileByte := p.Read(address + 8)
	var data uint32
	for i := 0; i < 8; i++ {
		var p1, p2 byte
		if attributes&0x40 == 0x40 {
			p1 = (lowTileByte & 1) << 0
			p2 = (highTileByte & 1) << 1
			lowTileByte >>= 1
			highTileByte >>= 1
		} else {
			p1 = (lowTileByte & 0x80) >> 7
			p2 = (highTileByte & 0x80) >> 6
			lowTileByte <<= 1
			highTileByte <<= 1
		}
		data <<= 4
		data |= uint32(a | p1 | p2)
	}
	return data
}

func (p *PPU) evaluateSprites() {
	var h int
	if p.fSpriteSize == 0 {
		h = 8
	} else {
		h = 16
	}
	count := 0
	for i := 0; i < 64; i++ {
		y := p.oamData[i*4+0]
		a := p.oamData[i*4+2]
		x := p.oamData[i*4+3]
		row := p.ScanLine - int(y)
		if row < 0 || row >= h {
			continue
		}
		if count < 8 {
			p.spritePatterns[count] = p.getSpritePattern(i, row)
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
