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
	bytes  uint8 // Length of this form of the instruction
}

// Built out from http://www.6502.org/users/obelisk/6502/reference.html
var InstructionSet = map[Instruction][]OpcodeForm{
	ADC: {
		{mode: Immediate, opcode: 0x69, bytes: 2},
		{mode: Zeropage, opcode: 0x65, bytes: 2},
		{mode: ZeropageXIndexed, opcode: 0x75, bytes: 2},
		{mode: Absolute, opcode: 0x6d, bytes: 3},
		{mode: AbsoluteXIndex, opcode: 0x7d, bytes: 3},
		{mode: AbsoluteYIndex, opcode: 0x79, bytes: 3},
		{mode: XIndexedIndirect, opcode: 0x61, bytes: 2},
		{mode: IndirectYIndexed, opcode: 0x71, bytes: 2},
	},
	AND: {
		{mode: Immediate, opcode: 0x29, bytes: 2},
		{mode: Zeropage, opcode: 0x25, bytes: 2},
		{mode: ZeropageXIndexed, opcode: 0x35, bytes: 2},
		{mode: Absolute, opcode: 0x2d, bytes: 3},
		{mode: AbsoluteXIndex, opcode: 0x3d, bytes: 3},
		{mode: AbsoluteYIndex, opcode: 0x39, bytes: 3},
		{mode: XIndexedIndirect, opcode: 0x21, bytes: 2},
		{mode: IndirectYIndexed, opcode: 0x31, bytes: 2},
	},
	LDA: {
		{mode: Immediate, opcode: 0xa9, bytes: 2},
		{mode: Zeropage, opcode: 0xa5, bytes: 2},
		{mode: ZeropageXIndexed, opcode: 0xb5, bytes: 2},
		{mode: Absolute, opcode: 0xad, bytes: 3},
		{mode: AbsoluteXIndex, opcode: 0xbd, bytes: 3},
		{mode: AbsoluteYIndex, opcode: 0xb9, bytes: 3},
		{mode: XIndexedIndirect, opcode: 0xa1, bytes: 2},
		{mode: IndirectYIndexed, opcode: 0xb1, bytes: 2},
	},
	STA: {
		{mode: Zeropage, opcode: 0x85, bytes: 2},
		{mode: ZeropageXIndexed, opcode: 0x95, bytes: 2},
		{mode: Absolute, opcode: 0x8d, bytes: 3},
		{mode: AbsoluteXIndex, opcode: 0x9d, bytes: 3},
		{mode: AbsoluteYIndex, opcode: 0x99, bytes: 3},
		{mode: XIndexedIndirect, opcode: 0x81, bytes: 2},
		{mode: IndirectYIndexed, opcode: 0x91, bytes: 2},
	},
	RTS: {
		{mode: Implied, opcode: 0x60, bytes: 1},
	},
}

func instructionEntry(i Instruction, m AddressingMode) (OpcodeForm, error) {
	forms, ok := InstructionSet[i]
	if !ok {
		return OpcodeForm{}, fmt.Errorf("can't find instruction #%d in table", i)
	}

	for _, val := range forms {
		if val.mode == m {
			return val, nil
		}
	}
	return OpcodeForm{}, fmt.Errorf("invalid addressing mode for %s", InstructionStrings[i])
}

func maxSize(forms []OpcodeForm) uint8 {
	max := uint8(0)
	for _, form := range forms {
		if form.bytes > max {
			max = form.bytes
		}
	}
	return max
}

func instructionMax(i Instruction) (uint8, error) {
	forms, ok := InstructionSet[i]
	if !ok {
		return 0, fmt.Errorf("can't find instruction #%d in table", i)
	}
	return maxSize(forms), nil
}

type pseudoOpEntry struct {
	fn func(a *assembler, line buf.Buffer) error
}

var pseudoOps = map[string]pseudoOpEntry{
	".ORG":  {fn: parseOrg},
	".BYTE": {fn: parseOrg},
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

type Operands struct {
	mode AddressingMode
	e    *expr.Node
	imm  bool
	abs  bool
}

// ----- New Node style
type Node interface {
	Pos() int
	node()
}

type LabelNode struct {
	Name     string
	position int
}

func (n *LabelNode) Pos() int { return n.position }
func (n *LabelNode) node()    {}

type InstructionNode struct {
	inst     *inst
	position int
}

func (n *InstructionNode) Pos() int { return n.position }
func (n *InstructionNode) node()    {}

type PseudoOpKind int

const (
	PseudoOrg PseudoOpKind = iota
	PseudoByte
	PseudoEqu
)

var PseudoOpMap = map[string]PseudoOpKind{
	".ORG":  PseudoOrg,
	".BYTE": PseudoByte,
	".EQU":  PseudoEqu,
}

type PseudoOp struct {
	Kind PseudoOpKind
	Args []*expr.Node
}

type PseudoNode struct {
	Pseudo   *PseudoOp
	position int
}

func (n *PseudoNode) Pos() int { return n.position }
func (n *PseudoNode) node()    {}

// ----- End Node style

// inst is an instruction with arguments. Chunk holds the machine form of the
// instruction as we're assembling.
// This is going to need a refactor, there should probably be another layer of
// abstraction here. This isn't just instructions. It could also be a chunk of
// memory the user has requested, or a string constant
type inst struct {
	labels   []string
	op       Instruction
	operands Operands
	size     uint8
	chunk    binaryChunk
}

type program []Node

type assembler struct {
	origin     int
	currLabel  []string // Needs to be copied when assigned to inst.labels
	prg        program
	sym        map[string]int
	constants  map[string]int
	exprParser expr.Parser
	line       int
}

func (a *assembler) parseLine(line buf.Buffer) error {
	remain := line
	if !remain.StartsWith(buf.Whitespace) {
		remain = a.parseLabel(remain)
		if remain.StartsWith(buf.Char('=')) {
			return a.parseConst(remain.Advance(1))
		}
	}
	return a.parseOperation(remain)
}

func (a *assembler) parseLabel(line buf.Buffer) buf.Buffer {
	label, remain := line.TakeWhile(buf.Letter)
	labelNode := LabelNode{Name: label.String(), position: a.line}
	a.prg = append(a.prg, &labelNode)
	if remain.StartsWith(buf.Char(':')) {
		remain = remain.Advance(1)
	}
	return remain
}

func (a *assembler) parseOperation(line buf.Buffer) error {
	_, remain := line.TakeWhile(buf.Whitespace)
	if remain.IsEmpty() || remain.StartsWith(buf.Char(';')) {
		return nil
	}

	op, remain := remain.TakeWhile(buf.Word)
	if pseudoKind, found := PseudoOpMap[strings.ToUpper(op.String())]; found {
		return a.parsePseudo(pseudoKind, remain)
	}
	return a.parseOpcode(op.String(), remain)
}

func (a *assembler) parsePseudo(pseudo PseudoOpKind, line buf.Buffer) error {
	pseudoOp := PseudoOp{Kind: pseudo}
	remain := line.Advance(line.Scan(buf.Whitespace))
	for !remain.IsEmpty() {
		expr, newRemain, err := a.exprParser.Parse(remain)
		if err != nil {
			return err
		}
		pseudoOp.Args = append(pseudoOp.Args, expr)
		remain = newRemain
		if remain.StartsWith(buf.Char(',')) {
			remain = remain.Advance(1)
		}
	}
	pseudoNode := PseudoNode{Pseudo: &pseudoOp, position: a.line}
	a.prg = append(a.prg, &pseudoNode)
	return nil
}

func (a *assembler) parseOpcode(opcode string, line buf.Buffer) error {
	i, err := ToInstruction(opcode)
	if err != nil {
		return err
	}
	maxBytes, err := instructionMax(i)
	if err != nil {
		return err
	}
	operands, remain, err := a.parseOperands(line)
	if err != nil {
		return err
	}
	remain = remain.Advance(remain.Scan(buf.Whitespace))
	if !remain.IsEmpty() || remain.StartsWith(buf.Char(';')) {
		return fmt.Errorf("unexpected text %v", remain.String())
	}
	instruction := inst{
		labels:   append([]string{}, a.currLabel...),
		op:       i,
		operands: operands,
		size:     maxBytes,
		chunk:    binaryChunk{addr: 0},
	}
	instructionNode := InstructionNode{inst: &instruction, position: a.line}
	a.prg = append(a.prg, &instructionNode)
	return nil
}

func (a *assembler) parseOperands(line buf.Buffer) (oper Operands, remain buf.Buffer, err error) {
	remain = line.Advance(line.Scan(buf.Whitespace))
	switch {
	case remain.IsEmpty():
		oper.mode = Implied
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
	err = fmt.Errorf("incorrect indirect format: %s", line.String())
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

func (a *assembler) parseConst(line buf.Buffer) error {
	e, _, err := a.exprParser.Parse(line)
	if err != nil {
		return fmt.Errorf("parseConst failed to parse expression %w", err)
	}
	ok, err := e.Eval(map[string]int{})
	if !ok || err != nil {
		return fmt.Errorf("parseConst failed to evaluate expression %w", err)
	}
	val, err := e.Value()
	if err != nil {
		return fmt.Errorf("parseConst failed getting Value() %w", err)
	}
	for _, v := range a.currLabel {
		a.constants[v] = val
	}
	a.currLabel = []string{}
	return nil
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

func (a *assembler) parseFile(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	return a.parseReader(file)
}

func (a *assembler) parseReader(r io.Reader) (err error) {
	a.line = 1
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		err = a.parseLine(buf.NewBuffer(scanner.Text()))
		if err != nil {
			return
		}
		a.line++
	}
	return nil
}

func (a *assembler) dumpAssembler(w io.Writer) {
	fmt.Fprintf(w, "%d segments in program\n", len(a.prg))
	fmt.Fprintf(w, "Starting address: %d\n", a.origin)
}

func (a *assembler) binaryImage() []uint8 {
	bytes := []uint8{}
	for _, node := range a.prg {
		// bytes = append(bytes, val.chunk.mem...)
		switch n := node.(type) {
		case *InstructionNode:
			fmt.Printf("Process instruction %v\n", n.inst.op)
		}
	}
	return bytes
}

func writeProgram(startAddr int, bytes []uint8, filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file %v", err)
	}
	defer file.Close()

	startBytes := []uint8{uint8(startAddr & 0xff), uint8((startAddr >> 8) & 0xff)}

	_, err = file.Write(startBytes)
	if err != nil {
		return fmt.Errorf("error writing bytes %v", err)
	}
	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing bytes %v", err)
	}
	return nil
}

func main() {
	a := assembler{origin: 0xc000}
	err := a.parseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// a.dumpAssembler(os.Stdout)
	bytes := a.binaryImage()
	for i, val := range bytes {
		fmt.Printf("%d = %X\n", i, val)
	}
	err = writeProgram(a.origin, bytes, "out.prg")
	if err != nil {
		log.Fatal(err)
	}
}
