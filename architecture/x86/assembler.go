package x86

import "github.com/keurnel/assembler/internal/asm"

type Assembler struct {
	asm.Architecture
}

// AssemblerNew - Create a new x86 assembler
func AssemblerNew() *Assembler {
	return &Assembler{}
}

// ArchitectureName - returns the name of the architecture.
func (a *Assembler) ArchitectureName() string {
	return "x86"
}

// Directives - returns a map of directives supported by x86 architecture.
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

// IsDirective - is the given line a valid directive for x86 architecture?
func (a *Assembler) IsDirective(line string) bool {
	return asm.IsDirectiveLine(line)
}

// Instructions - returns a map of instructions supported by x86 architecture.
func (a *Assembler) Instructions() map[string]asm.Instruction {
	return map[string]asm.Instruction{

		// ========================================
		//
		// Data Transfer Instructions
		//
		// ========================================

		// General Purpose Instructions
		MOV:  asm.Instruction{Mnemonic: MOV},
		PUSH: asm.Instruction{Mnemonic: PUSH},
		POP:  asm.Instruction{Mnemonic: POP},
		XCHG: asm.Instruction{Mnemonic: XCHG},
		XLAT: asm.Instruction{Mnemonic: XLAT},

		// Input / Output Instructions
		IN:  asm.Instruction{Mnemonic: IN},
		OUT: asm.Instruction{Mnemonic: OUT},

		// Address Object Instructions
		LEA: asm.Instruction{Mnemonic: LEA},
		LDS: asm.Instruction{Mnemonic: LDS},
		LES: asm.Instruction{Mnemonic: LES},

		// Flag Transfer Instructions
		LAHF:  asm.Instruction{Mnemonic: LAHF},
		SAHF:  asm.Instruction{Mnemonic: SAHF},
		PUSHF: asm.Instruction{Mnemonic: PUSHF},
		POPF:  asm.Instruction{Mnemonic: POPF},

		// ========================================
		//
		// Arithmetic Instructions
		//
		// ========================================

		// Addition Instructions
		ADD: asm.Instruction{Mnemonic: ADD},
		ADC: asm.Instruction{Mnemonic: ADC},
		INC: asm.Instruction{Mnemonic: INC},
		AAA: asm.Instruction{Mnemonic: AAA},
		DAA: asm.Instruction{Mnemonic: DAA},

		// Subtraction Instructions
		SUB: asm.Instruction{Mnemonic: SUB},
		SBB: asm.Instruction{Mnemonic: SBB},
		DEC: asm.Instruction{Mnemonic: DEC},
		NEG: asm.Instruction{Mnemonic: NEG},
		CMP: asm.Instruction{Mnemonic: CMP},
		AAS: asm.Instruction{Mnemonic: AAS},
		DAS: asm.Instruction{Mnemonic: DAS},

		// Multiplication Instructions
		MUL:  asm.Instruction{Mnemonic: MUL},
		IMUL: asm.Instruction{Mnemonic: IMUL},
		AAM:  asm.Instruction{Mnemonic: AAM},

		// Division Instructions
		DIV:  asm.Instruction{Mnemonic: DIV},
		IDIV: asm.Instruction{Mnemonic: IDIV},
		AAD:  asm.Instruction{Mnemonic: AAD},
		CBW:  asm.Instruction{Mnemonic: CBW},
		CWD:  asm.Instruction{Mnemonic: CWD},

		// ========================================
		//
		// Bit manipulation Instructions
		//
		// ========================================

		// Logical Instructions
		NOT:  asm.Instruction{Mnemonic: NOT},
		AND:  asm.Instruction{Mnemonic: AND},
		OR:   asm.Instruction{Mnemonic: OR},
		XOR:  asm.Instruction{Mnemonic: XOR},
		TEST: asm.Instruction{Mnemonic: TEST},

		// Shifts Instructions
		SHL: asm.Instruction{Mnemonic: SHL},
		SAL: asm.Instruction{Mnemonic: SAL},
		SHR: asm.Instruction{Mnemonic: SHR},
		SAR: asm.Instruction{Mnemonic: SAR},

		// Rotates Instructions
		ROL: asm.Instruction{Mnemonic: ROL},
		ROR: asm.Instruction{Mnemonic: ROR},
		RCL: asm.Instruction{Mnemonic: RCL},
		RCR: asm.Instruction{Mnemonic: RCR},
	}
}
