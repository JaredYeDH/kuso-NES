package nes

type info struct {
	address uint16
	pc uint16
	mode byte
}

// 6502 CPU
// For more information, visit
//  	http://www.obelisk.me.uk/6502/

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
	O 		byte // Overflow Flag
	N 		byte // Negative Flag
	U 		byte // Ignored FLag
	inter   byte // Interrupt type
	stall   int  // Cycles to stall
	ins     [256]func(*info) // Function table
	Memory		 //Memory Interface
}


func NewCPU(memory Memory) *CPU {
	cpu := CPU{Memory:memory}
	cpu.Reset()
	return &cpu
}

func (cpu *CPU) Reset() {
	cpu.PC = cpu.Read16(0xFFFC)
}