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
	Cycles	uint16
	PC		uint16 // program counter
	SP 		uint16 // stack pointer
	A 		byte // Accumulator
	X		byte // Index Register X
	Y 		byte // Index Register Y
	C 		byte // Carry FLag
	Z 		byte // Zero Flag
	I 		byte // Interrupt Disable
	D 		byte // Decimal Mode
	B 		byte // Break Command
	U 		byte // Ignored FLag
	O 		byte // Overflow Flag
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

func (c *CPU) ReadFlags() byte {
	var flags byte
	flags |= c.C << 0
	flags |= c.Z << 1
	flags |= c.I << 2
	flags |= c.D << 3
	flags |= c.B << 4
	flags |= c.U << 5
	flags |= c.O << 6
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
	c.O = (flags >> 6) & 1
	c.N = (flags >> 7) & 1
}

// For Debug
func (c *CPU) DebugPrint() {
	opcode := c.Read(c.PC)
	bytes := insSizes[opcode]
	name := insName[opcode]
	bytep0 := fmt.Sprintf("%02X",c.Read(c.PC + 0))
	bytep1 := fmt.Sprintf("%02X",c.Read(c.PC + 1))
	bytep2 := fmt.Sprintf("%02X",c.Read(c.PC + 2))

	if bytes < 3 {
		bytep2 = "	"
	}
	if bytes < 2 {
		bytep1 = "	"
	}
	fmt.Println()
	fmt.Printf(
		"PC: %4X  %s %s %s  %s %28s\n"+
			"A: %02X\nX: %02X\nY:%02X\nP: %02X\nSP: %02X\nCYC:%3d\n",
		c.PC, bytep0, bytep1, bytep2, name, "",
		c.A, c.X, c.Y, c.ReadFlags(), c.SP, (c.Cycles*3)%341)
	fmt.Println()
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

var insSizes = [256]byte {
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

var insCycles = [256]byte{
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

var insPCycles = [256]byte{
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
// Instructions by Name:
// ADC	....	add with carry
// AND	....	and (with accumulator)
// ASL	....	arithmetic shift left
// BCC	....	branch on carry clear
// BCS	....	branch on carry set
// BEQ	....	branch on equal (zero set)
// BIT	....	bit test
// BMI	....	branch on minus (negative set)
// BNE	....	branch on not equal (zero clear)
// BPL	....	branch on plus (negative clear)
// BRK	....	interrupt
// BVC	....	branch on overflow clear
// BVS	....	branch on overflow set
// CLC	....	clear carry
// CLD	....	clear decimal
// CLI	....	clear interrupt disable
// CLV	....	clear overflow
// CMP	....	compare (with accumulator)
// CPX	....	compare with X
// CPY	....	compare with Y
// DEC	....	decrement
// DEX	....	decrement X
// DEY	....	decrement Y
// EOR	....	exclusive or (with accumulator)
// INC	....	increment
// INX	....	increment X
// INY	....	increment Y
// JMP	....	jump
// JSR	....	jump subroutine
// LDA	....	load accumulator
// LDY	....	load X
// LDY	....	load Y
// LSR	....	logical shift right
// NOP	....	no operation
// ORA	....	or with accumulator
// PHA	....	push accumulator
// PHP	....	push processor status (SR)
// PLA	....	pull accumulator
// PLP	....	pull processor status (SR)
// ROL	....	rotate left
// ROR	....	rotate right
// RTI	....	return from interrupt
// RTS	....	return from subroutine
// SBC	....	subtract with carry
// SEC	....	set carry
// SED	....	set decimal
// SEI	....	set interrupt disable
// STA	....	store accumulator
// STX	....	store X
// STY	....	store Y
// TAX	....	transfer accumulator to X
// TAY	....	transfer accumulator to Y
// TSX	....	transfer stack pointer to X
// TXA	....	transfer X to accumulator
// TXS	....	transfer X to stack pointer
// TYA	....	transfer Y to accumulator
// EOR  ....    functions should not appear here
// For more detials ,
// See: http://e-tradition.net/bytes/6502/6502_instruction_set.html
// TODO: Implement those functions

func (c CPU) adc(info *info) {

}

func (c CPU) and(info *info) {

}

func (c CPU) asl(info *info) {

}

func (c CPU) bcc(info *info) {

}

func (c CPU) bcs(info *info) {

}

func (c CPU) bpl(info *info) {

}

func (c CPU) beq(info *info) {

}

func (c CPU) bit(info *info) {

}

func (c CPU) bmi(info *info) {

}

func (c CPU) bne(info *info) {

}

func (c CPU) brl(info *info) {

}

func (c CPU) brk(info *info) {

}

func (c CPU) bvc(info *info) {

}

func (c CPU) bvs(info *info) {

}

func (c CPU) clc(info *info) {

}

func (c CPU) cld(info *info) {

}

func (c CPU) cli(info *info) {

}

func (c CPU) clv(info *info) {

}

func (c CPU) cmp(info *info) {

}

func (c CPU) cpx(info *info) {

}

func (c CPU) cpy(info *info) {

}

func (c CPU) dec(info *info) {

}

func (c CPU) dex(info *info) {

}

func (c CPU) dey(info *info) {

}

func (c CPU) eor(info *info) {

}

func (c CPU) inc(info *info) {

}

func (c CPU) inx(info *info) {

}

func (c CPU) iny(info *info) {

}

func (c CPU) jmp(info *info) {

}

func (c CPU) jsr(info *info) {

}

func (c CPU) lda(info *info) {

}

func (c CPU) ldx(info *info) {

}

func (c CPU) ldy(info *info) {

}

func (c CPU) lsr(info *info) {

}

func (c CPU) nop(info *info) {

}

func (c CPU) ora(info *info) {

}

func (c CPU) pha(info *info) {

}

func (c CPU) php(info *info) {

}

func (c CPU) pla(info *info) {

}

func (c CPU) plp(info *info) {

}

func (c CPU) rol(info *info) {

}

func (c CPU) ror(info *info) {

}

func (c CPU) rti(info *info) {

}

func (c CPU) rts(info *info) {

}

func (c CPU) sbc(info *info) {

}

func (c CPU) sec(info *info) {

}

func (c CPU) sed(info *info) {

}

func (c CPU) sei(info *info) {

}

func (c CPU) sta(info *info) {

}

func (c CPU) stx(info *info) {

}

func (c CPU) sty(info *info) {

}

func (c CPU) tax(info *info) {

}

func (c CPU) tay(info *info) {

}

func (c CPU) tsx(info *info) {

}

func (c CPU) txa(info *info) {

}

func (c CPU) txs(info *info) {

}

func (c CPU) tya(info *info) {

}

func (c CPU) err(info *info) {

}