package nes

import "fmt"

// 6502 CPU
// For more information, visit
//  	http://www.obelisk.me.uk/6502/

const CPUFrequency = 1789773 // From http://wiki.nesdev.com/w/index.php/CPU


// Interrupt type

const (
	_ = iota
	interNone
	interNMI
	interIRQ
)

// Addressing Modes
const (
	_ = iota
	mAbsolute
	mAbsoluteX
	mAbsoluteY
	mAccumulator
	mImmediate
	mImplied
	mIndexedIndirect
	mIndirect
	mIndirectIndexed
	mRelative
	mZeroPage
	mZeroPageX
	mZeroPageY
)

// CPU structure

type CPU struct {
	Cycles	uint64 // Should be big enough
	PC		uint16 // Program counter
	SP 		byte // Stack pointer
	A 		byte // Accumulator
	X		byte // Index Register X
	Y 		byte // Index Register Y
	C 		byte // Carry FLag
	Z 		byte // Zero Flag
	I 		byte // Interrupt Disable
	D 		byte // Decimal Mode
	B 		byte // Break Command
	U 		byte // Ignored FLag
	V 		byte // Overflow Flag
	N 		byte // Negative Flag
	inter   byte // Interrupt type
	stall   int  // Cycles to stall
	ins     [256]func(*info) // Function table
	Memory		 //Memory Interface
}

// CPU operations

type info struct {
	address uint16
	pc uint16
	mode byte
}

func NewCPU(memory Memory) *CPU {
	cpu := CPU{Memory:memory}
	cpu.createTable()
	cpu.Reset()
	return &cpu
}

func (cpu *CPU) Reset() {
	cpu.PC = cpu.Read16(0xFFFC)
	cpu.SP = 0xFD
	cpu.SetFlags(0x24)
}

// For Debug
func (c *CPU) DebugPrint() {
	opcode := c.Read(c.PC)
	bytes := insSizes[opcode]
	name := insName[opcode]
	bytep0 := fmt.Sprintf("PC: %02X",c.Read(c.PC + 0))
	bytep1 := fmt.Sprintf("PC + 1: %02X",c.Read(c.PC + 1))
	bytep2 := fmt.Sprintf("PC + 2: %02X",c.Read(c.PC + 2))

	if bytes < 3 {
		bytep2 = "PC + 2: --"
	}
	if bytes < 2 {
		bytep1 = "PC + 1: --"
	}
	fmt.Println()
	fmt.Printf(
		"PC: %04X\n%s\n%s\n%s\nOption: %s %28s\n"+
			"A: %02X\nX: %02X\nY: %02X\nP: %02X\nSP: %02X\nCYC:%3d\n",
		c.PC, bytep0, bytep1, bytep2, name, "",
		c.A, c.X, c.Y, c.ReadFlags(), c.SP, (c.Cycles*3)%341)
	fmt.Println()
}

// Some basic functions

func (c *CPU) Read16(address uint16) uint16 {
	l := uint16(c.Read(address))
	h := uint16(c.Read(address +1))

	return h << 8 | l
}

// Emulate a bug which is used by those fucking game makers
func (c *CPU) readbug(address uint16) uint16 {
	l := uint16(c.Read(address))
	h := uint16(c.Read((address & 0xFF00) | uint16((byte(address)) + 1)))

	return h << 8 | l
}

// Common pull/push instrustion

func (c *CPU) push(val byte) {
	c.Write(0x100 | uint16(c.SP),val)
	c.SP--
}

func (c *CPU) pull() byte {
	c.SP++
	return c.Read(0x100 | uint16(c.SP))
}

func (c *CPU) push16(val uint16) {
	c.push(byte(val >> 8)) // High 8 bit
	c.push(byte(val & 0xFF)) // Low 8 bit
}

func (c *CPU) pull16() uint16 {
	l := uint16(c.pull())
	h := uint16(c.pull())

	return h << 8 | l
}

// Functions about flag

func (c *CPU) ReadFlags() byte {
	var flags byte
	flags |= c.C << 0
	flags |= c.Z << 1
	flags |= c.I << 2
	flags |= c.D << 3
	flags |= c.B << 4
	flags |= c.U << 5
	flags |= c.V << 6
	flags |= c.N << 7
	return flags
}

func (c *CPU) SetFlags(flags byte) {
	c.C = (flags >> 0) & 1
	c.Z = (flags >> 1) & 1
	c.I = (flags >> 2) & 1
	c.D = (flags >> 3) & 1
	c.B = (flags >> 4) & 1
	c.U = (flags >> 5) & 1
	c.V = (flags >> 6) & 1
	c.N = (flags >> 7) & 1
}

// setZ sets Z flag if val is equal to 0
func (c *CPU) setZ(val byte) {
	if val == 0 {
		c.Z = 1
	} else {
		c.Z = 0
	}
}

// setN sets Nflag if val is negative, which means the highest bit is set to 1
func (c *CPU) setN(val byte) {
	c.N = val&0x80
}

// setNZ set flag Z and flag N at one "operation"
func (c *CPU) setNZ(val byte) {
	c.setZ(val)
	c.setN(val)
}


// Some other functions to make the implementation of the following instructions more easier

// pageDiff tests if a & b reference different pages
func (c *CPU) pageDiff(a,b uint16) bool {
	return a & 0xFF00 != b & 0xFF00
}

// addBCycles adds a cycle for taking branch
func (c *CPU) addBCycles(info *info) {
	c.Cycles++
	if c.pageDiff(info.pc, info.address) {
		c.Cycles ++
	}
}


// Now let's run the CPU

// Run runs a instruction each time

func (c *CPU) Run() int {
	if c.stall >0 {
		c.stall --
		return 1
	}

	cycles := c.Cycles

	// Detect interrupts

	switch c.inter {
	case interIRQ:
		c.irq()
	case interNMI:
		c.nmi()
	}

	// Clear interrupt
	c.inter = interNone

	// Read instruction
	opcode := c.Read(c.PC)
	mode := insModes[opcode]

	// Build address for different addressing modes
	var address uint16
	var crossed bool

	switch mode {
	case mAbsolute:
		address = c.Read16(c.PC + 1)
	case mAbsoluteX:
		address = c.Read16(c.PC+1) + uint16(c.X)
		crossed = c.pageDiff(address-uint16(c.X), address)
	case mAbsoluteY:
		address = c.Read16(c.PC+1) + uint16(c.Y)
		crossed = c.pageDiff(address-uint16(c.Y), address)
	case mAccumulator:
		address = 0
	case mImmediate:
		address = c.PC + 1
	case mImplied:
		address = 0
	case mIndexedIndirect:
		address = c.readbug(uint16(c.Read(c.PC+1) + c.X))
	case mIndirect:
		address = c.readbug(c.Read16(c.PC + 1))
	case mIndirectIndexed:
		address = c.readbug(uint16(c.Read(c.PC+1))) + uint16(c.Y)
		crossed = c.pageDiff(address-uint16(c.Y), address)
	case mRelative:
		offset := uint16(c.Read(c.PC + 1))
		if offset < 0x80 {
			address = c.PC + 2 + offset
		} else {
			address = c.PC + 2 + offset - 0x100
		}
	case mZeroPage:
		address = uint16(c.Read(c.PC + 1))
	case mZeroPageX:
		address = uint16(c.Read(c.PC+1)+c.X) & 0xff
	case mZeroPageY:
		address = uint16(c.Read(c.PC+1)+c.Y) & 0xff
	}

	c.PC += insSizes[opcode]

	c.Cycles += insCycles[opcode]
	if crossed { // Page crossed is found
		c.Cycles += insPCycles[opcode]
	}

	// Build info
	info := &info{address,c.PC,mode}

	c.ins[opcode](info)

	return int(c.Cycles - cycles) // Should be int...
}

// Interrupts

// NMI Handler
func (c *CPU) nmi() {
	c.push16(c.PC)
	c.php(nil)
	c.PC = c.Read16(0xFFFA)
	c.I = 1
	c.Cycles += 7
}

// IRQ handler
func (c *CPU) irq() {
	c.push16(c.PC)
	c.php(nil)
	c.PC = c.Read16(0xFFFE)
	c.I = 1
	c.Cycles += 7
}


// Instruction set
// Ref: http://e-tradition.net/bytes/6502/6502_instruction_set.html
// As arrays.
// AUTO GENERATED.

var insModes = [256]byte {
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	1, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	6, 7, 6, 7, 11, 11, 11, 11, 6, 5, 4, 5, 8, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 13, 13, 6, 3, 6, 3, 2, 2, 3, 3,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 13, 13, 6, 3, 6, 3, 2, 2, 3, 3,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
	5, 7, 5, 7, 11, 11, 11, 11, 6, 5, 6, 5, 1, 1, 1, 1,
	10, 9, 6, 9, 12, 12, 12, 12, 6, 3, 6, 3, 2, 2, 2, 2,
}

var insSizes = [256]uint16 {
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	3, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	1, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 0, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 0, 3, 0, 0,
	2, 2, 2, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 2, 1, 0, 3, 3, 3, 0,
	2, 2, 0, 0, 2, 2, 2, 0, 1, 3, 1, 0, 3, 3, 3, 0,
}

var insCycles = [256]uint64{
	7, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	6, 6, 2, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	2, 6, 2, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
	2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
	2, 5, 2, 5, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4,
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
	2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
	2, 5, 2, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
}

var insPCycles = [256]uint64{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 0, 0,
}

var insName = [256]string{
	"BRK", "ORA", "ERR", "ERR", "NOP", "ORA", "ASL", "ERR",
	"PHP", "ORA", "ASL", "ERR", "NOP", "ORA", "ASL", "ERR",
	"BPL", "ORA", "ERR", "ERR", "NOP", "ORA", "ASL", "ERR",
	"CLC", "ORA", "NOP", "ERR", "NOP", "ORA", "ASL", "ERR",
	"JSR", "AND", "ERR", "ERR", "BIT", "AND", "ROL", "ERR",
	"PLP", "AND", "ROL", "ERR", "BIT", "AND", "ROL", "ERR",
	"BMI", "AND", "ERR", "ERR", "NOP", "AND", "ROL", "ERR",
	"SEC", "AND", "NOP", "ERR", "NOP", "AND", "ROL", "ERR",
	"RTI", "EOR", "ERR", "ERR", "NOP", "EOR", "LSR", "ERR",
	"PHA", "EOR", "LSR", "ERR", "JMP", "EOR", "LSR", "ERR",
	"BVC", "EOR", "ERR", "ERR", "NOP", "EOR", "LSR", "ERR",
	"CLI", "EOR", "NOP", "ERR", "NOP", "EOR", "LSR", "ERR",
	"RTS", "ADC", "ERR", "ERR", "NOP", "ADC", "ROR", "ERR",
	"PLA", "ADC", "ROR", "ERR", "JMP", "ADC", "ROR", "ERR",
	"BVS", "ADC", "ERR", "ERR", "NOP", "ADC", "ROR", "ERR",
	"SEI", "ADC", "NOP", "ERR", "NOP", "ADC", "ROR", "ERR",
	"NOP", "STA", "NOP", "ERR", "STY", "STA", "STX", "ERR",
	"DEY", "NOP", "TXA", "ERR", "STY", "STA", "STX", "ERR",
	"BCC", "STA", "ERR", "ERR", "STY", "STA", "STX", "ERR",
	"TYA", "STA", "TXS", "ERR", "ERR", "STA", "ERR", "ERR",
	"LDY", "LDA", "LDX", "ERR", "LDY", "LDA", "LDX", "ERR",
	"TAY", "LDA", "TAX", "ERR", "LDY", "LDA", "LDX", "ERR",
	"BCS", "LDA", "ERR", "ERR", "LDY", "LDA", "LDX", "ERR",
	"CLV", "LDA", "TSX", "ERR", "LDY", "LDA", "LDX", "ERR",
	"CPY", "CMP", "NOP", "ERR", "CPY", "CMP", "DEC", "ERR",
	"INY", "CMP", "DEX", "ERR", "CPY", "CMP", "DEC", "ERR",
	"BNE", "CMP", "ERR", "ERR", "NOP", "CMP", "DEC", "ERR",
	"CLD", "CMP", "NOP", "ERR", "NOP", "CMP", "DEC", "ERR",
	"CPX", "SBC", "NOP", "ERR", "CPX", "SBC", "INC", "ERR",
	"INX", "SBC", "NOP", "SBC", "CPX", "SBC", "INC", "ERR",
	"BEQ", "SBC", "ERR", "ERR", "NOP", "SBC", "INC", "ERR",
	"SED", "SBC", "NOP", "ERR", "NOP", "SBC", "INC", "ERR",
}


func (c *CPU) createTable() {
	c.ins = [256]func(*info){
		c.brk, c.ora, c.err, c.err, c.nop, c.ora, c.asl, c.err,
		c.php, c.ora, c.asl, c.err, c.nop, c.ora, c.asl, c.err,
		c.bpl, c.ora, c.err, c.err, c.nop, c.ora, c.asl, c.err,
		c.clc, c.ora, c.nop, c.err, c.nop, c.ora, c.asl, c.err,
		c.jsr, c.and, c.err, c.err, c.bit, c.and, c.rol, c.err,
		c.plp, c.and, c.rol, c.err, c.bit, c.and, c.rol, c.err,
		c.bmi, c.and, c.err, c.err, c.nop, c.and, c.rol, c.err,
		c.sec, c.and, c.nop, c.err, c.nop, c.and, c.rol, c.err,
		c.rti, c.eor, c.err, c.err, c.nop, c.eor, c.lsr, c.err,
		c.pha, c.eor, c.lsr, c.err, c.jmp, c.eor, c.lsr, c.err,
		c.bvc, c.eor, c.err, c.err, c.nop, c.eor, c.lsr, c.err,
		c.cli, c.eor, c.nop, c.err, c.nop, c.eor, c.lsr, c.err,
		c.rts, c.adc, c.err, c.err, c.nop, c.adc, c.ror, c.err,
		c.pla, c.adc, c.ror, c.err, c.jmp, c.adc, c.ror, c.err,
		c.bvs, c.adc, c.err, c.err, c.nop, c.adc, c.ror, c.err,
		c.sei, c.adc, c.nop, c.err, c.nop, c.adc, c.ror, c.err,
		c.nop, c.sta, c.nop, c.err, c.sty, c.sta, c.stx, c.err,
		c.dey, c.nop, c.txa, c.err, c.sty, c.sta, c.stx, c.err,
		c.bcc, c.sta, c.err, c.err, c.sty, c.sta, c.stx, c.err,
		c.tya, c.sta, c.txs, c.err, c.err, c.sta, c.err, c.err,
		c.ldy, c.lda, c.ldx, c.err, c.ldy, c.lda, c.ldx, c.err,
		c.tay, c.lda, c.tax, c.err, c.ldy, c.lda, c.ldx, c.err,
		c.bcs, c.lda, c.err, c.err, c.ldy, c.lda, c.ldx, c.err,
		c.clv, c.lda, c.tsx, c.err, c.ldy, c.lda, c.ldx, c.err,
		c.cpy, c.cmp, c.nop, c.err, c.cpy, c.cmp, c.dec, c.err,
		c.iny, c.cmp, c.dex, c.err, c.cpy, c.cmp, c.dec, c.err,
		c.bne, c.cmp, c.err, c.err, c.nop, c.cmp, c.dec, c.err,
		c.cld, c.cmp, c.nop, c.err, c.nop, c.cmp, c.dec, c.err,
		c.cpx, c.sbc, c.nop, c.err, c.cpx, c.sbc, c.inc, c.err,
		c.inx, c.sbc, c.nop, c.sbc, c.cpx, c.sbc, c.inc, c.err,
		c.beq, c.sbc, c.err, c.err, c.nop, c.sbc, c.inc, c.err,
		c.sed, c.sbc, c.nop, c.err, c.nop, c.sbc, c.inc, c.err,
	}
}



// Instructions
// For more detials ,
// See: http://e-tradition.net/bytes/6502/6502_instruction_set.html



// ADC - Add Memory to Accumulator with Carry
// A + M + C -> A, C
// N Z C I D V
// + + + - - +
func (c CPU) adc(info *info) {
	a := c.A
	m := c.Read(info.address)
	cf := c.C

	c.A = a + m + cf

	c.setNZ(c.A)

	// Set C flag
	if uint16(a) + uint16(m) + uint16(cf) > 0xFF {
		c.C = 1
	} else {
		c.C = 0
	}

	// Set V flag
	if ((a^m) >> 7) & 1  == 0 && ((a^c.A) >> 7) & 1 != 0 {
		c.V = 1
	} else {
		c.V = 0
	}
}

// AND - AND Memory with Accumulator
// A AND M -> A
// N Z C I D V
// + + - - - -
func (c CPU) and(info *info) {
	c.A = c.A | c.Read(info.address)
	c.setNZ(c.A)
}

// ASL - Shift Left One Bit (Memory or Accumulator)
// C <- [76543210] <- 0
// N Z C I D V
// + + + - - -
func (c CPU) asl(info *info) {
	if info.mode == mAccumulator { // Handle Accumulator Mode
		c.C = (c.A >> 7) & 1
		c.A = c.A << 1
		c.setNZ(c.A)
	} else { // Other Mode
		val := c.Read(info.address)
		c.C = (val >> 7) & 1
		val = val << 1
		c.Write(info.address, val)
		c.setNZ(val)
	}
}

// BCC - Branch on Carry Clear
// branch on C = 0
// N Z C I D V
// - - - - - -
func (c CPU) bcc(info *info) {
	if c.C == 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// BCC - Branch on Carry Clear
// branch on C = 1
// N Z C I D V
// - - - - - -
func (c CPU) bcs(info *info) {
	if c.C != 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// BEQ - Branch on Result Zero
// branch on Z = 1
// N Z C I D V
// - - - - - -
func (c CPU) beq(info *info) {
	if c.Z != 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// BIT - Test Bits in Memory with Accumulator
// bits 7 and 6 of operand are transfered to bit 7 and 6 of SR (N,V); the zeroflag is set to the result of operand AND accumulator.
// A AND M, M7 -> N, M6 -> V
// N Z C I D V
// M7 + - - - M6
func (c CPU) bit(info *info) {
	val := c.Read(info.address)

	c.setZ(val & c.A)

	c.V = (val >> 6) & 1
	c.N = (val >> 7) & 1
}

// BMI - Branch on Result Minus
// branch on N = 1
// N Z C I D V
// - - - - - -
func (c CPU) bmi(info *info) {
	if c.N != 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// BNE - Branch on Result not Zero
// branch on Z = 0
// N Z C I D V
// - - - - - -
func (c CPU) bne(info *info) {
	if c.Z == 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// BPL - Branch if Positive
// branch on N = 0
// N Z C I D V
// - - - - - -
func (c *CPU) bpl(info *info) {
	if c.N == 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// BRK - Force Break
// interrupt, 			N Z C I D V
// push PC+2, push SR 	- - - 1 - -
func (c CPU) brk(info *info) {
	c.push16(c.PC)
	c.php(info)
	c.sei(info)
	c.PC = c.Read16(0xFFFE)
}

// BVC - Branch on Overflow Clear
// branch on V = 0
// N Z C I D V
// - - - - - -
func (c CPU) bvc(info *info) {
	if c.V == 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

//BVS - Branch on Overflow Set
// branch on V = 1
// N Z C I D V
// - - - - - -
func (c CPU) bvs(info *info) {
	if c.V != 0 {
		c.PC = info.address
		c.addBCycles(info)
	}
}

// CLC - Clear Carry Flag
// 0 -> C
// N Z C I D V
// - - 0 - - -
func (c CPU) clc(info *info) {
	c.C = 0
}

// CLD - Clear Decimal Mode
// 0 -> D N Z C I D V
// - - - - 0 -
func (c CPU) cld(info *info) {
	c.D = 0
}

// CLI - Clear Interrupt Disable Bit
// 0 -> I
// N Z C I D V
// - - - 0 - -
func (c CPU) cli(info *info) {
	c.I = 0
}

// CLV - Clear Overflow Flag
// 0 -> V N Z C I D V
// - - - - - 0
func (c CPU) clv(info *info) {
	c.V = 0
}

// compare function for the following 3 functions
func (c *CPU) compare(a, b byte) {
	c.setNZ(a - b)
	if a >= b {
		c.C = 1
	} else {
		c.C = 0
	}
}

// CMP - Compare Memory with Accumulator
// A - M
// N Z C I D V
// + + + - - -
func (c CPU) cmp(info *info) {
	val := c.Read(info.address)
	c.compare(c.A,val)
}

// CPX - Compare Memory with IndexX
// X - M
// N Z C I D V
// + + + - - -
func (c CPU) cpx(info *info) {
	val := c.Read(info.address)
	c.compare(c.X,val)
}

// CPY - Compare Memory with IndexY
// Y - M
// N Z C I D V
// + + + - - -
func (c CPU) cpy(info *info) {
	val := c.Read(info.address)
	c.compare(c.Y,val)
}

// DEC - Decrement Memory by One
// M - 1 -> M
// N Z C I D V
// + + - - - -
func (c CPU) dec(info *info) {
	val := c.Read(info.address) - 1
	c.Write(info.address,val)
	c.setNZ(val)
}

// DEX - Decrement Index X by One
// X - 1 -> X
// N Z C I D V
// + + - - - -
func (c CPU) dex(info *info) {
	c.X --
	c.setNZ(c.X)
}

// DEY - Decrement Index Y by One
// Y - 1 -> Y
// N Z C I D V
// + + - - - -
func (c CPU) dey(info *info) {
	c.Y --
	c.setNZ(c.Y)
}

// EOR - Exclusive-OR Memory with Accumulator
// A EOR M -> A
// N Z C I D V
// + + - - - -
func (c CPU) eor(info *info) {
	val := c.Read(info.address)
	c.A = c.A ^ val
	c.setNZ(c.A)
}

// INC - Increment Memory by One
// M + 1 -> M
// N Z C I D V
// + + - - - -
func (c CPU) inc(info *info) {
	val := c.Read(info.address) + 1
	c.Write(info.address,val)
	c.setNZ(val)
}

// INX - Increment Index X by One
// X + 1 -> X
// N Z C I D V
// + + - - - -
func (c CPU) inx(info *info) {
	c.X ++
	c.setNZ(c.X)
}

// INY - Increment Index Y by One
// Y + 1 -> Y
// N Z C I D V
// + + - - - -
func (c CPU) iny(info *info) {
	c.Y ++
	c.setNZ(c.Y)
}

// JMP - Jump to New Location
// N Z C I D V
// - - - - - -
func (c CPU) jmp(info *info) {
	c.PC = info.address
}

// JSR - Jump to New Location Saving Return Address
// N Z C I D V
// - - - - - -
func (c CPU) jsr(info *info) {
	// Saving address...
	c.push16(c.PC - 1)
	// And then jump !!
	c.PC = info.address
}

// LDA - Load Accumulator with Memory
// M -> A
// N Z C I D V
// + + - - - -
func (c CPU) lda(info *info) {
	c.A = c.Read(info.address)
	c.setNZ(c.A)
}

// LDX Load Index X with Memory
// M -> X
// N Z C I D V
// + + - - - -
func (c CPU) ldx(info *info) {
	c.X = c.Read(info.address)
	c.setNZ(c.X)
}

// LDY Load Index Y with Memory
// M -> Y
// N Z C I D V
// + + - - - -
func (c CPU) ldy(info *info) {
	c.Y = c.Read(info.address)
	c.setNZ(c.Y)
}

// LSR - Logical Shift Right
// 0 -> [76543210] -> C
// N Z C I D V
// + + + - - -
func (c CPU) lsr(info *info) {
		if info.mode == mAccumulator {
			c.C = c.A & 1
			c.A >>= 1
			c.setNZ(c.A)
		} else {
			value := c.Read(info.address)
			c.C = value & 1
			value >>= 1
			c.Write(info.address, value)
			c.setNZ(value)
		}
}

// NOP - No Operation
// ------
// N Z C I D V
// - - - - - -
func (c CPU) nop(info *info) {
	// Indeed .. no operation
}

// ORA - OR Memory with Accumulator
// A OR M -> A
// N Z C I D V
// + + - - - -
func (c CPU) ora(info *info) {
	val := c.Read(info.address)
	c.A = c.A | val
	c.setNZ(c.A)
}

// PHA - Push Accumulator on Stack
// push A N Z C I D V
// - - - - - -
func (c CPU) pha(info *info) {
	c.push(c.A)
}

// PHP - Push Processor Status on Stack
// push SR
// N Z C I D V
// - - - - - -
func (c CPU) php(info *info) {
	c.push(c.ReadFlags() | 0x10)
}

// PLA - Pull Accumulator from Stack
// pull A
// N Z C I D V
// + + - - - -
func (c CPU) pla(info *info) {
	c.A = c.pull()
	c.setNZ(c.A)
}

// PLP - Pull Processor Status from Stack
// pull SR
// N Z C I D V
// from stack
func (c CPU) plp(info *info) {
	c.SetFlags((c.pull() & 0xEF | 0x20))
}

// ROL Rotate One Bit Left (Memory or Accumulator)
// C <- [76543210] <- C
// N Z C I D V
// + + + - - -
func (c CPU) rol(info *info) {
	if info.mode == mAccumulator {
		cf := c.C
		c.C = (c.A >> 7) & 1
		c.A = (c.A << 1) | cf
		c.setNZ(c.A)
	} else {
		cf := c.C
		val := c.Read(info.address)
		c.C = (val >> 7) & 1
		val = (val << 1) | cf
		c.Write(info.address, val)
		c.setNZ(val)
	}
}

// ROR - Rotate One Bit Right (Memory or Accumulator)
// C -> [76543210] -> C
// N Z C I D V
// + + + - - -
func (c CPU) ror(info *info) {
	if info.mode == mAccumulator {
		cf := c.C
		c.C = c.A & 1
		c.A = (c.A >> 1) | (cf << 7)
		c.setNZ(c.A)
	} else {
		cf := c.C
		value := c.Read(info.address)
		c.C = value & 1
		value = (value >> 1) | (cf << 7)
		c.Write(info.address, value)
		c.setNZ(value)
	}
}

// RTI - Return from Interrupt
// pull SR, pull PC
// N Z C I D V
// from stack
func (c CPU) rti(info *info) {
	c.plp(info)
	c.PC = c.pull16()
}

// RTS - Return from Subroutine
// pull PC, PC+1 -> PC
// N Z C I D V
// - - - - - -
func (c CPU) rts(info *info) {
	c.PC = c.pull16() + 1
}

// SBC - Subtract Memory from Accumulator with Borrow
// A - M - C -> A
// N Z C I D V
// + + + - - +
func (c CPU) sbc(info *info) {
	a := c.A
	m := c.Read(info.address)
	cf := c.C

	c.A = a - m - (1 - cf)

	c.setNZ(c.A)

	// Set C flag
	if int16(a) + int16(m) + int16(1 - cf) > 0 {
		c.C = 1
	} else {
		c.C = 0
	}

	// Set V flag
	if ((a^m) >> 7) & 1  != 0 && ((a^c.A) >> 7) & 1 != 0 {
		c.V = 1
	} else {
		c.V = 0
	}
}

// SEC - Set Carry Flag
// 1 -> C
// N Z C I D V
// - - 1 - - -
func (c CPU) sec(info *info) {
	c.C = 1
}

// SED - Set Decimal Flag
// 1 -> D N Z C I D V
// - - - - 1 -
func (c CPU) sed(info *info) {
	c.D = 1
}

// SEI - Set Interrupt Disable Status
// 1 -> I N Z C I D V
// - - - 1 - -
func (c CPU) sei(info *info) {
	c.I = 1
}

// STA Store Accumulator in Memory
// A -> M
// N Z C I D V
// - - - - - -
func (c CPU) sta(info *info) {
	c.Write(info.address,c.A)
}

// STX - Store Index X in Memory
// X -> M
// N Z C I D V
// - - - - - -
func (c CPU) stx(info *info) {
	c.Write(info.address,c.X)
}

// STY - Store Index Y in Memory
// Y -> M
// N Z C I D V
// - - - - - -
func (c CPU) sty(info *info) {
	c.Write(info.address,c.Y)
}

// TAX - Transfer Accumulator to Index X
// A -> X
// N Z C I D V
// + + - - - -
func (c CPU) tax(info *info) {
	c.X = c.A
	c.setNZ(c.X)
}

// TAY - Transfer Accumulator to Index Y
// A -> Y
// N Z C I D V
// + + - - - -
func (c CPU) tay(info *info) {
	c.Y = c.A
	c.setNZ(c.Y)
}

// TSX - Transfer Accumulator to Index X
// SP -> X
// N Z C I D V
// + + - - - -
func (c CPU) tsx(info *info) {
	c.X = c.SP
	c.setNZ(c.X)
}

// TXA - Transfer Index X to Accumulator
// X -> A
// N Z C I D V
// + + - - - -
func (c CPU) txa(info *info) {
	c.A = c.X
	c.setNZ(c.A)
}

// TXS - Transfer Index X to Stack Register
// X -> SP
// N Z C I D V
// + + - - - -
func (c CPU) txs(info *info) {
	c.SP = c.X
	c.setNZ(c.SP)
}

//TYA - Transfer Index Y to Accumulator
// Y -> A
// N Z C I D V
// + + - - - -
func (c CPU) tya(info *info) {
	c.A = c.Y
	c.setNZ(c.A)
}

// ERR - Function for those opcodes which should not appear in ANY iNES roms
// ------
// N Z C I D V
// - - - - - -
func (c CPU) err(info *info) {
	// What can I do ????
	// I can just forgive those ancient masters.
	// Thanks to those NES video game makers, we can have a happy childhood.
	// QAQ
}