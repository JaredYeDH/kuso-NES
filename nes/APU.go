package nes

import (
	"math"
)

const fCounterRate = CPUFrequency / 240.0

var (
	lTable = []byte{
		10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
		12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
	}

	dTable = [][]byte{
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 0, 0, 0},
		{1, 0, 0, 1, 1, 1, 1, 1},
	}

	triangleTable = []byte{
		15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}

	noiseTable = []uint16{
		4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068,
	}

	dmcTable = []byte{
		214, 190, 170, 160, 143, 127, 113, 107, 95, 80, 71, 64, 53, 42, 36, 27,
	}
)
var pulseTable [31]float32
var tndTable [203]float32

func init() {
	for i := 0; i < 31; i++ {
		pulseTable[i] = 95.52 / (8128.0/float32(i) + 100)
	}
	for i := 0; i < 203; i++ {
		tndTable[i] = 163.67 / (24329.0/float32(i) + 100)
	}
}

// Square

type Square struct {
	enabled  bool
	channel  byte
	lEnabled bool
	lValue   byte
	tPeriod  uint16
	tValue   uint16
	dMode    byte
	dValue   byte
	sReload  bool
	sEnabled bool
	sNegate  bool
	sShift   byte
	sPeriod  byte
	sValue   byte
	eEnabled bool
	eLoop    bool
	eStart   bool
	ePeriod  byte
	eValue   byte
	eVolume  byte
	cVolume  byte
}

func (s *Square) wCtrl(val byte) {
	s.dMode = (val >> 6) & 3
	s.lEnabled = (val>>5)&1 == 0
	s.eLoop = (val>>5)&1 == 1
	s.eEnabled = (val>>4)&1 == 0
	s.ePeriod = val & 15
	s.cVolume = val & 15
	s.eStart = true
}

func (s *Square) wSweep(val byte) {
	s.sEnabled = (val>>7)&1 == 1
	s.sPeriod = (val>>4)&7 + 1
	s.sNegate = (val>>3)&1 == 1
	s.sShift = val & 7
	s.sReload = true
}

func (s *Square) wTimerLow(val byte) {
	s.tPeriod = (s.tPeriod & 0xFF00) | uint16(val)
}

func (s *Square) wTimerHigh(val byte) {
	s.lValue = lTable[val>>3]
	s.tPeriod = (s.tPeriod & 0x00FF) | (uint16(val&7) << 8)
	s.eStart = true
	s.dValue = 0
}

func (s *Square) rTimer() {
	if s.tValue == 0 {
		s.tValue = s.tPeriod
		s.dValue = (s.dValue + 1) % 8
	} else {
		s.tValue--
	}
}

func (s *Square) rEnvelope() {
	if s.eStart {
		s.eVolume = 15
		s.eValue = s.ePeriod
		s.eStart = false
	} else if s.eValue > 0 {
		s.eValue--
	} else {
		if s.eVolume > 0 {
			s.eVolume--
		} else if s.eLoop {
			s.eVolume = 15
		}
		s.eValue = s.ePeriod
	}
}

func (s *Square) rSweep() {
	if s.sReload {
		if s.sEnabled && s.sValue == 0 {
			s.s()
		}
		s.sValue = s.sPeriod
		s.sReload = false
	} else if s.sValue > 0 {
		s.sValue--
	} else {
		if s.sEnabled {
			s.s()
		}
		s.sValue = s.sPeriod
	}
}

func (s *Square) rLength() {
	if s.lEnabled && s.lValue > 0 {
		s.lValue--
	}
}

func (s *Square) s() {
	delta := s.tPeriod >> s.sShift
	if s.sNegate {
		s.tPeriod -= delta
		if s.channel == 1 {
			s.tPeriod--
		}
	} else {
		s.tPeriod += delta
	}
}

func (s *Square) output() byte {
	if !s.enabled {
		return 0
	}
	if s.lValue == 0 {
		return 0
	}
	if dTable[s.dMode][s.dValue] == 0 {
		return 0
	}
	if s.tPeriod < 8 || s.tPeriod > 0x7FF {
		return 0
	}
	if s.eEnabled {
		return s.eVolume
	} else {
		return s.cVolume
	}
}

// Triangle

type Triangle struct {
	enabled  bool
	lEnabled bool
	lValue   byte
	tPeriod  uint16
	tValue   uint16
	dValue   byte
	cPeriod  byte
	cValue   byte
	cReload  bool
}

func (t *Triangle) wCtrl(val byte) {
	t.lEnabled = (val>>7)&1 == 0
	t.cPeriod = val & 0x7F
}

func (t *Triangle) wTimerLow(val byte) {
	t.tPeriod = (t.tPeriod & 0xFF00) | uint16(val)
}

func (t *Triangle) wTimerHigh(val byte) {
	t.lValue = lTable[val>>3]
	t.tPeriod = (t.tPeriod & 0x00FF) | (uint16(val&7) << 8)
	t.tValue = t.tPeriod
	t.cReload = true
}

func (t *Triangle) rTimer() {
	if t.tValue == 0 {
		t.tValue = t.tPeriod
		if t.lValue > 0 && t.cValue > 0 {
			t.dValue = (t.dValue + 1) % 32
		}
	} else {
		t.tValue--
	}
}

func (t *Triangle) rLength() {
	if t.lEnabled && t.lValue > 0 {
		t.lValue--
	}
}

func (t *Triangle) rCounter() {
	if t.cReload {
		t.cValue = t.cPeriod
	} else if t.cValue > 0 {
		t.cValue--
	}
	if t.lEnabled {
		t.cReload = false
	}
}

func (t *Triangle) output() byte {
	if !t.enabled {
		return 0
	}
	if t.lValue == 0 {
		return 0
	}
	if t.cValue == 0 {
		return 0
	}
	return triangleTable[t.dValue]
}

// Noise

type Noise struct {
	enabled   bool
	mode      bool
	sRegister uint16
	lEnabled  bool
	lValue    byte
	tPeriod   uint16
	tValue    uint16
	eEnabled  bool
	eLoop     bool
	eStart    bool
	ePeriod   byte
	eValue    byte
	eVolume   byte
	cVolume   byte
}

func (n *Noise) wCtrl(val byte) {
	n.lEnabled = (val>>5)&1 == 0
	n.eLoop = (val>>5)&1 == 1
	n.eEnabled = (val>>4)&1 == 0
	n.ePeriod = val & 15
	n.cVolume = val & 15
	n.eStart = true
}

func (n *Noise) wPeriod(val byte) {
	n.mode = val&0x80 == 0x80
	n.tPeriod = noiseTable[val&0x0F]
}

func (n *Noise) wLength(val byte) {
	n.lValue = lTable[val>>3]
	n.eStart = true
}

func (n *Noise) rTimer() {
	if n.tValue == 0 {
		n.tValue = n.tPeriod
		var s byte
		if n.mode {
			s = 6
		} else {
			s = 1
		}
		b1 := n.sRegister & 1
		b2 := (n.sRegister >> s) & 1
		n.sRegister >>= 1
		n.sRegister |= (b1 ^ b2) << 14
	} else {
		n.tValue--
	}
}

func (n *Noise) rEnvelope() {
	if n.eStart {
		n.eVolume = 15
		n.eValue = n.ePeriod
		n.eStart = false
	} else if n.eValue > 0 {
		n.eValue--
	} else {
		if n.eVolume > 0 {
			n.eVolume--
		} else if n.eLoop {
			n.eVolume = 15
		}
		n.eValue = n.ePeriod
	}
}

func (n *Noise) rLength() {
	if n.lEnabled && n.lValue > 0 {
		n.lValue--
	}
}

func (n *Noise) output() byte {
	if !n.enabled {
		return 0
	}
	if n.lValue == 0 {
		return 0
	}
	if n.sRegister&1 == 1 {
		return 0
	}
	if n.eEnabled {
		return n.eVolume
	} else {
		return n.cVolume
	}
}

// DMC

type DMC struct {
	cpu            *CPU
	enabled        bool
	val            byte
	sampleAddress  uint16
	sampleLength   uint16
	currentAddress uint16
	currentLength  uint16
	sRegister      byte
	bitCount       byte
	tickPeriod     byte
	tickValue      byte
	loop           bool
	irq            bool
}

func (d *DMC) wCtrl(val byte) {
	d.irq = val&0x80 == 0x80
	d.loop = val&0x40 == 0x40
	d.tickPeriod = dmcTable[val&0x0F]
}

func (d *DMC) wValue(val byte) {
	d.val = val & 0x7F
}

func (d *DMC) wAddress(val byte) {
	d.sampleAddress = 0xC000 | (uint16(val) << 6)
}

func (d *DMC) wLength(val byte) {
	d.sampleLength = (uint16(val) << 4) | 1
}

func (d *DMC) restart() {
	d.currentAddress = d.sampleAddress
	d.currentLength = d.sampleLength
}

func (d *DMC) rTimer() {
	if !d.enabled {
		return
	}
	d.rReader()
	if d.tickValue == 0 {
		d.tickValue = d.tickPeriod
		d.rShifter()
	} else {
		d.tickValue--
	}
}

func (d *DMC) rReader() {
	if d.currentLength > 0 && d.bitCount == 0 {
		d.cpu.stall += 4
		d.sRegister = d.cpu.Read(d.currentAddress)
		d.bitCount = 8
		d.currentAddress++
		if d.currentAddress == 0 {
			d.currentAddress = 0x8000
		}
		d.currentLength--
		if d.currentLength == 0 && d.loop {
			d.restart()
		}
	}
}

func (d *DMC) rShifter() {
	if d.bitCount == 0 {
		return
	}
	if d.sRegister&1 == 1 {
		if d.val <= 125 {
			d.val += 2
		}
	} else {
		if d.val >= 2 {
			d.val -= 2
		}
	}
	d.sRegister >>= 1
	d.bitCount--
}

func (d *DMC) output() byte {
	return d.val
}

type Filter interface {
	Run(x float32) float32
}

type FirstOrderFilter struct {
	B0    float32
	B1    float32
	A1    float32
	prevX float32
	prevY float32
}

func (f *FirstOrderFilter) Run(x float32) float32 {
	y := f.B0*x + f.B1*f.prevX - f.A1*f.prevY
	f.prevY = y
	f.prevX = x
	return y
}

func LPassFilter(sampleRate float32, cutoffFreq float32) Filter {
	c := sampleRate / math.Pi / cutoffFreq
	a0i := 1 / (1 + c)
	return &FirstOrderFilter{
		B0: a0i,
		B1: a0i,
		A1: (1 - c) * a0i,
	}
}

func HPassFilter(sampleRate float32, cutoffFreq float32) Filter {
	c := sampleRate / math.Pi / cutoffFreq
	a0i := 1 / (1 + c)
	return &FirstOrderFilter{
		B0: c * a0i,
		B1: -c * a0i,
		A1: (1 - c) * a0i,
	}
}

type FilterChain []Filter

func (fc FilterChain) Run(x float32) float32 {
	if fc != nil {
		for i := range fc {
			x = fc[i].Run(x)
		}
	}
	return x
}

type APU struct {
	nes        *NES
	channel    chan float32
	sampleRate float64
	square1    Square
	square2    Square
	triangle   Triangle
	noise      Noise
	dmc        DMC
	cycle      uint64
	fPeriod    byte
	fValue     byte
	fIRQ       bool
	fChain     FilterChain
}

func NewAPU(nes *NES) *APU {
	apu := APU{}
	apu.nes = nes
	apu.noise.sRegister = 1
	apu.square1.channel = 1
	apu.square2.channel = 2
	apu.dmc.cpu = nes.CPU
	return &apu
}

func (a *APU) Run() {
	cycle1 := a.cycle
	a.cycle++
	cycle2 := a.cycle
	a.rTimer()
	f1 := int(float64(cycle1) / fCounterRate)
	f2 := int(float64(cycle2) / fCounterRate)
	if f1 != f2 {
		a.rFrameCounter()
	}
	s1 := int(float64(cycle1) / a.sampleRate)
	s2 := int(float64(cycle2) / a.sampleRate)
	if s1 != s2 {
		a.sendSample()
	}
}

func (a *APU) sendSample() {
	output := a.fChain.Run(a.output())
	select {
	case a.channel <- output:
	default:
	}
}

func (a *APU) output() float32 {
	p1 := a.square1.output()
	p2 := a.square2.output()
	t := a.triangle.output()
	n := a.noise.output()
	d := a.dmc.output()
	pulseOut := pulseTable[p1+p2]
	tndOut := tndTable[3*t+2*n+d]
	return pulseOut + tndOut
}

func (a *APU) rFrameCounter() {
	switch a.fPeriod {
	case 4:
		a.fValue = (a.fValue + 1) % 4
		switch a.fValue {
		case 0, 2:
			a.rEnvelope()
		case 1:
			a.rEnvelope()
			a.rSweep()
			a.rLength()
		case 3:
			a.rEnvelope()
			a.rSweep()
			a.rLength()
			a.fireIRQ()
		}
	case 5:
		a.fValue = (a.fValue + 1) % 5
		switch a.fValue {
		case 1, 3:
			a.rEnvelope()
		case 0, 2:
			a.rEnvelope()
			a.rSweep()
			a.rLength()
		}
	}
}

func (a *APU) rTimer() {
	if a.cycle%2 == 0 {
		a.square1.rTimer()
		a.square2.rTimer()
		a.noise.rTimer()
		a.dmc.rTimer()
	}
	a.triangle.rTimer()
}

func (a *APU) rEnvelope() {
	a.square1.rEnvelope()
	a.square2.rEnvelope()
	a.triangle.rCounter()
	a.noise.rEnvelope()
}

func (a *APU) rSweep() {
	a.square1.rSweep()
	a.square2.rSweep()
}

func (a *APU) rLength() {
	a.square1.rLength()
	a.square2.rLength()
	a.triangle.rLength()
	a.noise.rLength()
}

func (a *APU) fireIRQ() {
	if a.fIRQ {
		a.nes.CPU.tIRQ()
	}
}

func (a *APU) ReadRegister(address uint16) byte {
	switch address {
	case 0x4015:
		return a.rStatus()
	}
	return 0
}

func (a *APU) WriteRegister(address uint16, val byte) {
	switch address {
	case 0x4000:
		a.square1.wCtrl(val)
	case 0x4001:
		a.square1.wSweep(val)
	case 0x4002:
		a.square1.wTimerLow(val)
	case 0x4003:
		a.square1.wTimerHigh(val)
	case 0x4004:
		a.square2.wCtrl(val)
	case 0x4005:
		a.square2.wSweep(val)
	case 0x4006:
		a.square2.wTimerLow(val)
	case 0x4007:
		a.square2.wTimerHigh(val)
	case 0x4008:
		a.triangle.wCtrl(val)
	case 0x4009:
	case 0x4010:
		a.dmc.wCtrl(val)
	case 0x4011:
		a.dmc.wValue(val)
	case 0x4012:
		a.dmc.wAddress(val)
	case 0x4013:
		a.dmc.wLength(val)
	case 0x400A:
		a.triangle.wTimerLow(val)
	case 0x400B:
		a.triangle.wTimerHigh(val)
	case 0x400C:
		a.noise.wCtrl(val)
	case 0x400D:
	case 0x400E:
		a.noise.wPeriod(val)
	case 0x400F:
		a.noise.wLength(val)
	case 0x4015:
		a.wCtrl(val)
	case 0x4017:
		a.wFrameCounter(val)
	}
}

func (a *APU) rStatus() byte {
	var result byte
	if a.square1.lValue > 0 {
		result |= 1
	}
	if a.square2.lValue > 0 {
		result |= 2
	}
	if a.triangle.lValue > 0 {
		result |= 4
	}
	if a.noise.lValue > 0 {
		result |= 8
	}
	if a.dmc.currentLength > 0 {
		result |= 16
	}
	return result
}

func (a *APU) wCtrl(val byte) {
	a.square1.enabled = val&1 == 1
	a.square2.enabled = val&2 == 2
	a.triangle.enabled = val&4 == 4
	a.noise.enabled = val&8 == 8
	a.dmc.enabled = val&16 == 16
	if !a.square1.enabled {
		a.square1.lValue = 0
	}
	if !a.square2.enabled {
		a.square2.lValue = 0
	}
	if !a.triangle.enabled {
		a.triangle.lValue = 0
	}
	if !a.noise.enabled {
		a.noise.lValue = 0
	}
	if !a.dmc.enabled {
		a.dmc.currentLength = 0
	} else {
		if a.dmc.currentLength == 0 {
			a.dmc.restart()
		}
	}
}

func (a *APU) wFrameCounter(val byte) {
	a.fPeriod = 4 + (val>>7)&1
	a.fIRQ = (val>>6)&1 == 0
	if a.fPeriod == 5 {
		a.rEnvelope()
		a.rSweep()
		a.rLength()
	}
}
