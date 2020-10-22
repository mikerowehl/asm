package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// val is an argument value, it can either be a direct immediate value or
// a symbol we need to resolve
type val struct {
	imm int
	sym string
}

// args is the argument set to a single machine instruction.
type args struct {
	reg  int // 1 = a, 2 = x, 3 = y
	imm  val
	addr val
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

type InstructionForm struct {
	instruction Instruction
	modes       []AddressingMode
}

var InstructionForms = []InstructionForm{
	{
		instruction: ADC,
		modes:       []AddressingMode{Immediate, Zeropage},
	},
}

// inst is an instruction with arguments, addr is the address once we know it
type inst struct {
	addr int
	op   Instruction
	args args
}

var prg = []*inst{}

func parseArgs(a string) (ret args) {
	if strings.Compare(a, "A") == 0 {
		ret.reg = 1
	} else if a[0] == '(' {
		ret.ind = true
		if strings.HasSuffix(a, ",X)") {
			ret.reg = 2
			v, err := strconv.ParseInt(a[:len(a)-3], 0, 16)
			if err != nil {
				log.Fatal("Error parsing int", a)
			}
			ret.addr.imm = int(v)
		} else if strings.HasSuffix(a, "),Y") {
			ret.reg = 3
			v, err := strconv.ParseInt(a[:len(a)-3], 0, 16)
			if err != nil {
				log.Fatal("Error parsing int", a)
			}
			ret.addr.imm = int(v)
		}
	} else if a[0] == '#' {
		v, err := strconv.ParseInt(a[1:], 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.imm.imm = int(v)
	} else if strings.HasSuffix(a, ",X") {
		ret.reg = 2
		v, err := strconv.ParseInt(a[:len(a)-2], 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.addr.imm = int(v)
	} else if strings.HasSuffix(a, ",Y") {
		ret.reg = 3
		v, err := strconv.ParseInt(a[:len(a)-2], 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.addr.imm = int(v)
	} else {
		v, err := strconv.ParseInt(a, 0, 16)
		if err != nil {
			log.Fatal("Error parsing int", a)
		}
		ret.addr.imm = int(v)
	}
	return
}

func parseLine(l string) error {
	parts := strings.Fields(l)
	i, err := ToInstruction(parts[0])
	if err != nil {
		return err
	}
	args := args{}
	if len(parts) > 1 {
		args = parseArgs(parts[1])
	}
	prg = append(prg, &inst{op: i, args: args})
	return nil
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err = parseLine(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for _, o := range prg {
		fmt.Println("Instruction: ", o.op)
		fmt.Println("  Args: ", o.args)
	}
}
