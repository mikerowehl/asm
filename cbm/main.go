package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mikerowehl/asm/buf"
	"github.com/mikerowehl/asm/expr"
)

type Instruction int

const (
	ADC Instruction = iota
	AND
	ASL
	BCC
	BCS
	BEQ
	BIT
	BMI
	BNE
	BPL
	BRK
	BVC
	BVS
	CLC
	CLD
	CLI
	CLV
	CMP
	CPX
	CPY
	DEC
	DEX
	EOR
	INC
	INX
	INY
	JMP
	JSR
	LDA
	LDX
	LDY
	LSR
	NOP
	ORA
	PHA
	PHP
	PLA
	PLP
	ROL
	ROR
	RTI
	RTS
	SBC
	SEC
	SED
	SEI
	STA
	STX
	STY
	TAX
	TAY
	TSX
	TXA
	TXS
	TYA
)

var InstructionStrings = []string{
	"ADC",
	"AND",
	"ASL",
	"BCC",
	"BCS",
	"BEQ",
	"BIT",
	"BMI",
	"BNE",
	"BPL",
	"BRK",
	"BVC",
	"BVS",
	"CLC",
	"CLD",
	"CLI",
	"CLV",
	"CMP",
	"CPX",
	"CPY",
	"DEC",
	"DEX",
	"EOR",
	"INC",
	"INX",
	"INY",
	"JMP",
	"JSR",
	"LDA",
	"LDX",
	"LDY",
	"LSR",
	"NOP",
	"ORA",
	"PHA",
	"PHP",
	"PLA",
	"PLP",
	"ROL",
	"ROR",
	"RTI",
	"RTS",
	"SBC",
	"SEC",
	"SED",
	"SEI",
	"STA",
	"STX",
	"STY",
	"TAX",
	"TAY",
	"TSX",
	"TXA",
	"TXS",
	"TYA",
}

func (i Instruction) String() string {
	return InstructionStrings[i]
}

func ToInstruction(s string) (instruction Instruction, err error) {
	for i, v := range InstructionStrings {
		if v == s {
			instruction = Instruction(i)
			return
		}
	}
	err = fmt.Errorf("%s is not a valid instruction", s)
	return
}

type AddressingMode int

const (
	Accumulator AddressingMode = iota
	Absolute
	AbsoluteXIndex
	AbsoluteYIndex
	Immediate
	Implied
	Indirect
	XIndexedIndirect
	IndirectYIndexed
	Relative
	Zeropage
	ZeropageXIndexed
	ZeropageYIndexed
)

type OpcodeForm struct {
	mode   AddressingMode
	opcode uint8
}

var InstructionSet = map[Instruction][]OpcodeForm{
	ADC: {
		{mode: Immediate, opcode: 0x69},
		{mode: Zeropage, opcode: 0x65},
		{mode: ZeropageXIndexed, opcode: 0x75},
		{mode: Absolute, opcode: 0x6d},
		{mode: AbsoluteXIndex, opcode: 0x7d},
		{mode: AbsoluteYIndex, opcode: 0x79},
		{mode: XIndexedIndirect, opcode: 0x61},
		{mode: IndirectYIndexed, opcode: 0x71},
	},
	AND: {
		{mode: Immediate, opcode: 0x29},
		{mode: Zeropage, opcode: 0x25},
		{mode: ZeropageXIndexed, opcode: 0x35},
		{mode: Absolute, opcode: 0x2d},
		{mode: AbsoluteXIndex, opcode: 0x3d},
		{mode: AbsoluteYIndex, opcode: 0x39},
		{mode: XIndexedIndirect, opcode: 0x21},
		{mode: IndirectYIndexed, opcode: 0x31},
	},
	LDA: {
		{mode: Immediate, opcode: 0xa9},
		{mode: Zeropage, opcode: 0xa5},
		{mode: ZeropageXIndexed, opcode: 0xb5},
		{mode: Absolute, opcode: 0xad},
		{mode: AbsoluteXIndex, opcode: 0xbd},
		{mode: AbsoluteYIndex, opcode: 0xb9},
		{mode: XIndexedIndirect, opcode: 0xa1},
		{mode: IndirectYIndexed, opcode: 0xb1},
	},
}

func instructionEntry(i Instruction, m AddressingMode) (OpcodeForm, error) {
	forms, ok := InstructionSet[i]
	if !ok {
		return OpcodeForm{}, fmt.Errorf("Can't find instruction #%d in table", i)
	}

	for _, val := range forms {
		if val.mode == m {
			return val, nil
		}
	}
	return OpcodeForm{}, fmt.Errorf("Invalid addressing mode for %s", InstructionStrings[i])
}

type pseudoOpEntry struct {
	fn func(a *assembler, line buf.Buffer) error
}

var pseudoOps = map[string]pseudoOpEntry{
	".org": {fn: parseOrg},
}

type binaryChunk struct {
	addr int
	mem  []uint8
}

func (c binaryChunk) String() string {
	if c.mem == nil {
		return fmt.Sprintf("%x:", c.addr)
	}
	b := []string{}
	for _, m := range c.mem {
		b = append(b, fmt.Sprintf("0x%x", m))
	}
	return fmt.Sprintf("%x: [%s]", c.addr, strings.Join(b, ", "))
}

var mem binaryChunk

type Operands struct {
	mode AddressingMode
	e    *expr.Node
	imm  bool
	abs  bool
}

// inst is an instruction with arguments. Chunk holds the machine form of the
// instruction as we're assembling.
type inst struct {
	op       Instruction
	operands Operands
	chunk    binaryChunk
}

type program []*inst

var prg = program{}

type assembler struct {
	origin     int
	labels     []string
	prg        program
	sym        map[string]int
	exprParser expr.Parser
}

func (a *assembler) parseLine(line buf.Buffer) error {
	remain := line
	if !remain.StartsWith(buf.Whitespace) {
		remain = a.parseLabel(remain)
	}
	return a.parseOperation(remain)
}

func (a *assembler) parseLabel(line buf.Buffer) buf.Buffer {
	label, remain := line.TakeUntil(buf.Char(':'))
	a.labels = append(a.labels, label.String())
	return remain
}

func (a *assembler) parseOperation(line buf.Buffer) error {
	_, remain := line.TakeWhile(buf.Whitespace)
	if remain.IsEmpty() || remain.StartsWith(buf.Char(';')) {
		return nil
	}

	op, remain := remain.TakeWhile(buf.Word)
	if pseudo, found := pseudoOps[op.String()]; found {
		return pseudo.fn(a, remain)
	}
	return a.parseOpcode(op.String(), remain)
}

func (a *assembler) parseOpcode(opcode string, line buf.Buffer) error {
	i, err := ToInstruction(opcode)
	if err != nil {
		return err
	}
	operands, remain, err := a.parseOperands(line)
	if err != nil {
		return err
	}
	remain = remain.Advance(remain.Scan(buf.Whitespace))
	if !remain.IsEmpty() || remain.StartsWith(buf.Char(';')) {
		return fmt.Errorf("Unexpected text %v", remain.String())
	}
	form, err := instructionEntry(i, operands.mode)
	if err != nil {
		return err
	}
	instruction := inst{
		op:       i,
		operands: operands,
		chunk:    binaryChunk{addr: 0, mem: []uint8{form.opcode}},
	}
	a.prg = append(a.prg, &instruction)
	// fmt.Printf("instruction %+v\n", instruction)
	return nil
}

func (a *assembler) parseOperands(line buf.Buffer) (oper Operands, remain buf.Buffer, err error) {
	remain = line.Advance(line.Scan(buf.Whitespace))
	switch {
	case remain.IsEmpty():
		oper.mode = Immediate
	case remain.StartsWith(buf.Char('(')):
		var e buf.Buffer
		oper.mode, e, remain, err = a.parseIndirect(remain.Advance(1))
		if err != nil {
			return
		}
		oper.e, _, err = a.exprParser.Parse(e)
	case remain.StartsWith(buf.Char('#')):
		oper.mode = Immediate
		oper.imm = true
		oper.e, remain, err = a.exprParser.Parse(remain.Advance(1))
	default:
		var e buf.Buffer
		oper.mode, e, remain, err = a.parseAbsolute(remain)
		if err != nil {
			return
		}
		oper.e, _, err = a.exprParser.Parse(e)
	}
	return
}

func (a *assembler) parseIndirect(line buf.Buffer) (mode AddressingMode, expr buf.Buffer, remain buf.Buffer, err error) {
	expr, remain = line.TakeUntil(func(s string) bool { return s[0] == ',' || s[0] == ')' })

	if remain.StartsWith(buf.Str(",X)")) {
		mode = XIndexedIndirect
		remain = remain.Advance(3)
		return
	}
	if remain.StartsWith(buf.Str("),Y")) {
		mode = IndirectYIndexed
		remain = remain.Advance(3)
		return
	}
	if remain.StartsWith(buf.Char(')')) {
		mode = Indirect
		remain = remain.Advance(1)
		return
	}
	err = fmt.Errorf("Incorrect indirect format: %s", line.String())
	return
}

func (a *assembler) parseAbsolute(line buf.Buffer) (mode AddressingMode, expr buf.Buffer, remain buf.Buffer, err error) {
	expr, remain = line.TakeUntil(func(s string) bool { return s[0] == ',' || buf.Whitespace(s) })

	switch {
	case remain.StartsWith(buf.Str(",X")):
		mode = AbsoluteXIndex
		remain = remain.Advance(2)
	case remain.StartsWith(buf.Str(",Y")):
		mode = AbsoluteYIndex
		remain = remain.Advance(2)
	default:
		mode = Absolute
	}

	_, remain = remain.TakeWhile(buf.Whitespace)
	return
}

func parseOrg(a *assembler, line buf.Buffer) error {
	e, _, err := a.exprParser.Parse(line.Advance(line.Scan(buf.Whitespace)))
	if err != nil {
		return err
	}

	e.Eval(a.sym)
	a.origin, err = e.Value()
	return err
}

func (a *assembler) parseFile(fn string) (err error) {
	file, err := os.Open(os.Args[1])
	if err != nil {
		return
	}
	defer file.Close()
	return a.parseReader(file)
}

func (a *assembler) parseReader(r io.Reader) (err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		err = a.parseLine(buf.NewBuffer(scanner.Text()))
		if err != nil {
			return
		}
	}
	return nil
}

func (a *assembler) dumpAssembler(w io.Writer) {
	fmt.Fprintf(w, "%d segments in program\n", len(a.prg))
	fmt.Fprintf(w, "Starting address: %d\n", a.origin)
}

func main() {
	a := assembler{origin: 0xc00}
	err := a.parseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	a.dumpAssembler(os.Stdout)
}
