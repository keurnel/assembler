package x86_64

import (
	"regexp"

	"github.com/keurnel/assembler/internal/asm"
)

type Assembler struct {
	asm.Architecture
	rawSource string
}

// AssemblerNew - returns a new instance of the x86_64 assembler
func AssemblerNew(rawSource string) *Assembler {
	return &Assembler{
		rawSource: rawSource,
	}
}

// ArchitectureName - returns the name of the architecture
func (a *Assembler) ArchitectureName() string {
	return "x86_64"
}

// Directives - returns the supported directives for the architecture
func (a *Assembler) Directives() map[string]bool {
	return map[string]bool{
		// Section directives
		".data":    true,
		".text":    true,
		".bss":     true,
		".rodata":  true,
		".section": true,

		// Symbol visibility directives
		".global":    true,
		".globl":     true,
		".extern":    true,
		".local":     true,
		".weak":      true,
		".hidden":    true,
		".protected": true,

		// Data definition directives
		".byte":   true,
		".word":   true,
		".short":  true,
		".long":   true,
		".int":    true,
		".quad":   true,
		".octa":   true,
		".float":  true,
		".double": true,
		".ascii":  true,
		".asciz":  true,
		".string": true,
		".zero":   true,
		".space":  true,
		".skip":   true,
		".fill":   true,

		// Alignment directives
		".align":   true,
		".balign":  true,
		".p2align": true,

		// Symbol definition directives
		".equ":   true,
		".set":   true,
		".equiv": true,

		// Type and size directives
		".type": true,
		".size": true,

		// Common symbol directives
		".comm":  true,
		".lcomm": true,

		// Syntax directives
		".intel_syntax": true,
		".att_syntax":   true,

		// Macro and conditional directives
		".macro":   true,
		".endm":    true,
		".if":      true,
		".ifdef":   true,
		".ifndef":  true,
		".else":    true,
		".endif":   true,
		".include": true,

		// Debug and metadata directives
		".file":    true,
		".line":    true,
		".ident":   true,
		".version": true,

		// CFI (Call Frame Information) directives
		".cfi_startproc":         true,
		".cfi_endproc":           true,
		".cfi_def_cfa":           true,
		".cfi_def_cfa_offset":    true,
		".cfi_def_cfa_register":  true,
		".cfi_offset":            true,
		".cfi_adjust_cfa_offset": true,
		".cfi_remember_state":    true,
		".cfi_restore_state":     true,
		".cfi_restore":           true,

		// Other common directives
		".org":        true,
		".rept":       true,
		".endr":       true,
		".irp":        true,
		".irpc":       true,
		".exitm":      true,
		".error":      true,
		".warning":    true,
		".print":      true,
		".purgem":     true,
		".altmacro":   true,
		".noaltmacro": true,
	}
}

// IsDirective - checks if a given line of assembly code is a valid directive for the architecture
func (a *Assembler) IsDirective(line string) bool {
	_, exists := a.Directives()[line]
	return exists
}

// Instructions - returns the parsed assembly instructions
func (a *Assembler) Instructions() map[string]asm.Instruction {
	return map[string]asm.Instruction{
		// Data Movement Instructions
		"MOV":   MOV,
		"MOVZX": MOVZX,
		"MOVSX": MOVSX,
		"LEA":   LEA,
		"PUSH":  PUSH,
		"POP":   POP,
		"XCHG":  XCHG,

		// Arithmetic Instructions
		"ADD":  ADD,
		"SUB":  SUB,
		"MUL":  MUL,
		"IMUL": IMUL,
		"DIV":  DIV,
		"IDIV": IDIV,
		"INC":  INC,
		"DEC":  DEC,
		"NEG":  NEG,
		"CMP":  CMP,

		// Logical Instructions
		"AND":  AND,
		"OR":   OR,
		"XOR":  XOR,
		"NOT":  NOT,
		"TEST": TEST,

		// Shift and Rotate Instructions
		"SHL": SHL,
		"SHR": SHR,
		"SAR": SAR,
		"ROL": ROL,
		"ROR": ROR,

		// Control Flow Instructions
		"JMP":  JMP,
		"JE":   JE,
		"JNE":  JNE,
		"JG":   JG,
		"JGE":  JGE,
		"JL":   JL,
		"JLE":  JLE,
		"JA":   JA,
		"JAE":  JAE,
		"JB":   JB,
		"JBE":  JBE,
		"CALL": CALL,
		"RET":  RET,

		// Miscellaneous Instructions
		"NOP":     NOP,
		"HLT":     HLT,
		"SYSCALL": SYSCALL,
		"SYSRET":  SYSRET,
		"INT":     INT,
		"IRET":    IRET,
		"CPUID":   CPUID,
		"RDTSC":   RDTSC,
	}
}

// IsInstruction - checks if a given line of assembly code is a valid instruction for the architecture
func (a *Assembler) IsInstruction(line string) bool {
	_, exists := a.Instructions()[line]
	if exists == true {
		return true
	}

	return false
}

// Operands - returns a map of supported operands for the architecture
func (a *Assembler) Operands() map[string]asm.OperandType {
	return map[string]asm.OperandType{
		// X86_8 - register operands
		"AL":   OperandReg8,
		"CL":   OperandReg8,
		"DL":   OperandReg8,
		"BL":   OperandReg8,
		"SPL":  OperandReg8,
		"BPL":  OperandReg8,
		"SIL":  OperandReg8,
		"DIL":  OperandReg8,
		"R8B":  OperandReg8,
		"R9B":  OperandReg8,
		"R10B": OperandReg8,
		"R11B": OperandReg8,
		"R12B": OperandReg8,
		"R13B": OperandReg8,
		"R14B": OperandReg8,
		"R15B": OperandReg8,

		// X86_16 - register operands
		"AX":   OperandReg16,
		"CX":   OperandReg16,
		"DX":   OperandReg16,
		"BX":   OperandReg16,
		"SP":   OperandReg16,
		"BP":   OperandReg16,
		"SI":   OperandReg16,
		"DI":   OperandReg16,
		"R8W":  OperandReg16,
		"R9W":  OperandReg16,
		"R10W": OperandReg16,
		"R11W": OperandReg16,
		"R12W": OperandReg16,
		"R13W": OperandReg16,
		"R14W": OperandReg16,
		"R15W": OperandReg16,

		// X86_32 - register operands
		"EAX":  OperandReg32,
		"ECX":  OperandReg32,
		"EDX":  OperandReg32,
		"EBX":  OperandReg32,
		"ESP":  OperandReg32,
		"EBP":  OperandReg32,
		"ESI":  OperandReg32,
		"EDI":  OperandReg32,
		"R8D":  OperandReg32,
		"R9D":  OperandReg32,
		"R10D": OperandReg32,
		"R11D": OperandReg32,
		"R12D": OperandReg32,
		"R13D": OperandReg32,
		"R14D": OperandReg32,
		"R15D": OperandReg32,

		// X86_64 - register operands
		"RAX": OperandReg64,
		"RBX": OperandReg64,
		"RCX": OperandReg64,
		"RDX": OperandReg64,
		"RSI": OperandReg64,
		"RDI": OperandReg64,
		"RBP": OperandReg64,
		"RSP": OperandReg64,
		"R8":  OperandReg64,
		"R9":  OperandReg64,
		"R10": OperandReg64,
		"R11": OperandReg64,
		"R12": OperandReg64,
		"R13": OperandReg64,
		"R14": OperandReg64,
		"R15": OperandReg64,

		// X86_8 - memory operands
		"[AL]":   OperandMem8,
		"[CL]":   OperandMem8,
		"[DL]":   OperandMem8,
		"[BL]":   OperandMem8,
		"[SPL]":  OperandMem8,
		"[BPL]":  OperandMem8,
		"[SIL]":  OperandMem8,
		"[DIL]":  OperandMem8,
		"[R8B]":  OperandMem8,
		"[R9B]":  OperandMem8,
		"[R10B]": OperandMem8,
		"[R11B]": OperandMem8,
		"[R12B]": OperandMem8,
		"[R13B]": OperandMem8,
		"[R14B]": OperandMem8,
		"[R15B]": OperandMem8,

		// X86_16 - memory operands
		"[AX]":   OperandMem16,
		"[CX]":   OperandMem16,
		"[DX]":   OperandMem16,
		"[BX]":   OperandMem16,
		"[SP]":   OperandMem16,
		"[BP]":   OperandMem16,
		"[SI]":   OperandMem16,
		"[DI]":   OperandMem16,
		"[R8W]":  OperandMem16,
		"[R9W]":  OperandMem16,
		"[R10W]": OperandMem16,
		"[R11W]": OperandMem16,
		"[R12W]": OperandMem16,
		"[R13W]": OperandMem16,
		"[R14W]": OperandMem16,
		"[R15W]": OperandMem16,

		// X86_32 - memory operands
		"[EAX]":  OperandMem32,
		"[ECX]":  OperandMem32,
		"[EDX]":  OperandMem32,
		"[EBX]":  OperandMem32,
		"[ESP]":  OperandMem32,
		"[EBP]":  OperandMem32,
		"[ESI]":  OperandMem32,
		"[EDI]":  OperandMem32,
		"[R8D]":  OperandMem32,
		"[R9D]":  OperandMem32,
		"[R10D]": OperandMem32,
		"[R11D]": OperandMem32,
		"[R12D]": OperandMem32,
		"[R13D]": OperandMem32,
		"[R14D]": OperandMem32,
		"[R15D]": OperandMem32,

		// X86_64 - memory operands
		"[RAX]": OperandMem64,
		"[RBX]": OperandMem64,
		"[RCX]": OperandMem64,
		"[RDX]": OperandMem64,
		"[RSI]": OperandMem64,
		"[RDI]": OperandMem64,
		"[RBP]": OperandMem64,
		"[RSP]": OperandMem64,
		"[R8]":  OperandMem64,
		"[R9]":  OperandMem64,
		"[R10]": OperandMem64,
		"[R11]": OperandMem64,
		"[R12]": OperandMem64,
		"[R13]": OperandMem64,
		"[R14]": OperandMem64,
		"[R15]": OperandMem64,

		// X86_64 - memory operands with displacement
		"[RBP-8]":  OperandMem64,
		"[RBP-16]": OperandMem64,
		"[RBP-32]": OperandMem64,
		"[RBP-64]": OperandMem64,

		// X86_8 - immediate operands
		"IMM8": OperandImm8,

		// X86_16 - immediate operands
		"IMM16": OperandImm16,

		// X86_32 - immediate operands
		"IMM32": OperandImm32,

		// X86_64 - immediate operands
		"IMM64": OperandImm64,
	}
}

// IsOperand - checks if a given string is a valid operand for the architecture
func (a *Assembler) IsOperand(operand string) bool {
	_, exists := a.Operands()[operand]
	if exists == true {
		return true
	}

	// Base displacement memory operand recognition (e.g., [RBP-8], [RAX], [RBX+RCX*4+16], [0x400000]).
	//
	pattern := `^\[` +
		`([A-Z][A-Z0-9]*)?` + // Optional base register
		`(\+|-)?` + // Optional operator
		`([A-Z][A-Z0-9]*)?` + // Optional index register
		`(\*[1248])?` + // Optional scale factor
		`(([+-])?(0x[0-9A-Fa-f]+|0o[0-7]+|0b[01]+|\d+))?` + // Optional displacement (with optional sign)
		`\]$`

	baseDisplacementMatches, err := regexp.MatchString(pattern, operand)
	if err != nil {
		return false
	}

	if baseDisplacementMatches == true {
		return true
	}

	// Immediate value (e.g., 123, 0x1A, -45, +0xFF) recognition.
	//
	matches, err := regexp.MatchString(`^[-+]?(0x[0-9a-fA-F]+|\d+)$`, operand)
	if err != nil {
		return false
	}

	return matches
}

// RegisterSet - returns a list of supported registers for the architecture
func (a *Assembler) Registers() []string {
	return []string{}
}

// IsRegister - checks if a given string is a valid register for the architecture
func (a *Assembler) IsRegister(name string) bool {
	return false
}

// OperandTypes - returns a list of supported operand types for the architecture
func (a *Assembler) OperandTypes() []asm.OperandType {
	return []asm.OperandType{
		OperandNone,
		OperandReg8,
		OperandReg16,
		OperandReg32,
		OperandReg64,
		OperandImm8,
		OperandImm16,
		OperandImm32,
		OperandImm64,
		OperandMem,
		OperandMem8,
		OperandMem16,
		OperandMem32,
		OperandMem64,
		OperandRel8,
		OperandRel32,
		OperandRegMem8,
		OperandRegMem16,
		OperandRegMem32,
		OperandRegMem64,
	}
}

// OperandCounts - returns a list of valid operand counts for the architecture
func (a *Assembler) OperandCounts() []int {
	return []int{OperandCountOne, OperandCountTwo, OperandCountThree}
}

// IsValidOperandCount - checks if a given operand count is valid for the architecture
func (a *Assembler) IsValidOperandCount(count int) bool {
	return count >= OperandCountOne && count <= OperandCountThree
}

// SourceOperandSupportsDestination - checks if a given source operand type can be used with a given destination operand type in an instruction
func (a *Assembler) SourceOperandSupportsDestination(sourceType, destType asm.OperandType) bool {
	// todo: implement this function based on the rules of operand compatibility for x86_64 instructions
	return false
}

// Is8BitInstruction - checks if a given instruction is an 8-bit instruction based on its operand types
func (a *Assembler) Is8BitInstruction(instr asm.Instruction) bool {
	// todo: implement this function based on the instruction's operand types
	return false
}

// RawSource - returns the raw assembly source code
func (a *Assembler) RawSource() string {
	return a.rawSource
}
