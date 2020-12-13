package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mikerowehl/asm/buf"
	"github.com/mikerowehl/asm/expr"
)

// val is an argument value, it can either be a direct immediate value or
// a symbol we need to resolve
type val struct {
	imm int
	sym string
}

type Register int

const (
	RegNone Register = 0
	RegA             = 1
	RegX             = 2
	RegY             = 3
)

// args is the argument set to a single machine instruction.
type args struct {
	reg  Register
	imm  *val
	addr *val
	ind  bool
}

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
	args     args
	chunk    binaryChunk
}

type program []*inst

var prg = program{}

type assembler struct {
	labels     []string
	prg        program
	sym        map[string]int
	exprParser expr.Parser
}

func parseArgs(a string) (ret args) {
	if strings.Compare(a, "A") == 0 {
		ret.reg = 1
	} else if a[0] == '(' {
		ret.ind = true
		if strings.HasSuffix(a, ",X)") {
			ret.reg = 2
			v, err := strconv.ParseInt(a[1:len(a)-3], 0, 16)
			if err != nil {
				log.Fatal("Error parsing int", a)
			}
			ret.addr = &val{imm: int(v)}
		} else if strings.HasSuffix(a, "),Y") {
			ret.reg = 3
			v, err := strconv.ParseInt(a[1:len(a)-3], 0, 16)
			if err != nil {
				log.Fatal("Error parsing int", a)
			}
			ret.addr = &val{imm: int(v)}
		}
	} else if a[0] == '#' {
		v, err := strconv.ParseInt(a[1:], 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.imm = &val{imm: int(v)}
	} else if strings.HasSuffix(a, ",X") {
		ret.reg = 2
		v, err := strconv.ParseInt(a[:len(a)-2], 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.addr = &val{imm: int(v)}
	} else if strings.HasSuffix(a, ",Y") {
		ret.reg = 3
		v, err := strconv.ParseInt(a[:len(a)-2], 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.addr = &val{imm: int(v)}
	} else {
		v, err := strconv.ParseInt(a, 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.addr = &val{imm: int(v)}
	}
	return
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

	op, remain := line.TakeWhile(buf.Word)
	/*
		if pseudo, found := pseudoOps[op.String()]; found {
			return nil
		}
	*/
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
	instruction := inst{
		op:       i,
		operands: operands,
	}
	fmt.Printf("instruction %+v", instruction)
	return nil
}

func (a *assembler) parseOperands(line buf.Buffer) (oper Operands, remain buf.Buffer, err error) {
	remain = line.Advance(line.Scan(buf.Whitespace))
	if remain.IsEmpty() {
		oper.mode = Immediate
		return
	}
	if remain.StartsWith(buf.Char('(')) {
		var e buf.Buffer
		oper.mode, e, remain, err = a.parseIndirect(remain)
		if err != nil {
			return
		}
		oper.expr, _, err = a.exprParser.Parse(e)
		if err != nil {
			return
		}
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

func assembleInstruction(i *inst, forms []OpcodeForm) (err error) {
	for _, f := range forms {
		switch f.mode {
		case Implied:
			if i.args.imm == nil && i.args.addr == nil {
				i.chunk.mem = []uint8{f.opcode}
				return
			}
		case Immediate:
			if i.args.imm != nil {
				i.chunk.mem = []uint8{f.opcode, uint8(i.args.imm.imm & 0xff)}
				return
			}
		case Accumulator:
			if i.args.reg == RegA {
				i.chunk.mem = []uint8{f.opcode}
				return
			}
		case Absolute:
			if i.args.addr != nil && !i.args.ind && i.args.reg == RegNone {
				i.chunk.mem = []uint8{
					f.opcode,
					uint8((i.args.addr.imm >> 8) & 0xff),
					uint8(i.args.addr.imm & 0xff),
				}
				return
			}
		case AbsoluteXIndex:
			if i.args.addr != nil && !i.args.ind && i.args.reg == RegX {
				i.chunk.mem = []uint8{
					f.opcode,
					uint8((i.args.addr.imm >> 8) & 0xff),
					uint8(i.args.addr.imm & 0xff),
				}
				return
			}
		case AbsoluteYIndex:
			if i.args.addr != nil && !i.args.ind && i.args.reg == RegY {
				i.chunk.mem = []uint8{
					f.opcode,
					uint8((i.args.addr.imm >> 8) & 0xff),
					uint8(i.args.addr.imm & 0xff),
				}
				return
			}
		case Relative:
			if i.args.addr != nil && i.args.addr.imm >= -128 && i.args.addr.imm <= 127 && !i.args.ind && i.args.reg == RegNone {
				i.chunk.mem = []uint8{f.opcode, uint8(i.args.addr.imm & 0xff)}
				return
			}
		case Indirect:
			if i.args.addr != nil && i.args.ind && i.args.reg == RegNone {
				i.chunk.mem = []uint8{
					f.opcode,
					uint8((i.args.addr.imm >> 8) & 0xff),
					uint8(i.args.addr.imm & 0xff),
				}
				return
			}
		case XIndexedIndirect:
			if i.args.addr != nil && i.args.ind && i.args.reg == RegX {
				i.chunk.mem = []uint8{
					f.opcode,
					uint8((i.args.addr.imm >> 8) & 0xff),
					uint8(i.args.addr.imm & 0xff),
				}
				return
			}
		case IndirectYIndexed:
			if i.args.addr != nil && i.args.ind && i.args.reg == RegY {
				i.chunk.mem = []uint8{
					f.opcode,
					uint8((i.args.addr.imm >> 8) & 0xff),
					uint8(i.args.addr.imm & 0xff),
				}
				return
			}
		case Zeropage:
			if i.args.addr != nil && i.args.addr.imm <= 255 && !i.args.ind && i.args.reg == RegNone {
				i.chunk.mem = []uint8{f.opcode, uint8(i.args.addr.imm & 0xff)}
				return
			}
		case ZeropageXIndexed:
			if i.args.addr != nil && i.args.addr.imm <= 255 && !i.args.ind && i.args.reg == RegX {
				i.chunk.mem = []uint8{f.opcode, uint8(i.args.addr.imm & 0xff)}
				return
			}
		case ZeropageYIndexed:
			if i.args.addr != nil && i.args.addr.imm <= 255 && !i.args.ind && i.args.reg == RegY {
				i.chunk.mem = []uint8{f.opcode, uint8(i.args.addr.imm & 0xff)}
				return
			}
		}
	}
	err = fmt.Errorf("Can't find matching form for instruction: %s", i.op)
	return
}

// firstPassAssemble walks through each entry in the program and creates the
// associated binary form. At this point we don't always have all the info we
// need for references (for instance forward references to labels, we need to
// figure out the address for the associated chunk). That means we might have
// to default to longer instruction forms. If we don't know the value of an
// expression and the instruction has both 8 and 16 bit addresses accepted, we
// just use the 16 bit version to be safe.
func firstPassAssemble(p program) error {
	for _, i := range p {
		fmt.Println("Instruction: ", i.op)
		fmt.Println("  Args: ", i.args)
		forms, ok := InstructionSet[i.op]
		if !ok {
			return fmt.Errorf("Invalid instruction: %v", i.op)
		}
		if err := assembleInstruction(i, forms); err != nil {
			log.Fatal("Error assembling: ", err)
		}
		if i.chunk.mem == nil {
			log.Fatalf("Can't find matching instruction for %s", i.op)
		}
		fmt.Printf("Assembled form %s\n", i.chunk)
	}
	return nil
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

func main() {
	a := assembler{}
	err := a.parseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if err := firstPassAssemble(prg); err != nil {
		log.Fatal(err)
	}
}
