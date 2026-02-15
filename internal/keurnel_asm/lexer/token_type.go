package lexer

import (
	"errors"
	"regexp"
)

const (
	ILLEGAL          TokenType = "ILLEGAL"
	ILLEGAL_RESERVED TokenType = "ILLEGAL_RESERVED"
	ILLEGAL_PATTERN  TokenType = "ILLEGAL_PATTERN"
	EOF              TokenType = "EOF"

	// Identifiers and literals
	IDENT  TokenType = "IDENT"  // variable/label names
	INT    TokenType = "INT"    // integer literals: 42, 0x2A, 0b101010
	FLOAT  TokenType = "FLOAT"  // floating point: 3.14
	STRING TokenType = "STRING" // string literals: "hello"
	CHAR   TokenType = "CHAR"   // character literals: 'a'

	// Architecture-specific tokens
	INSTRUCTION TokenType = "INSTRUCTION" // MOV, ADD, SUB, etc.
	REGISTER    TokenType = "REGISTER"    // rax, rbx, eax, etc.

	// Memory and addressing
	MEMORY_REF TokenType = "MEMORY_REF" // [rbx], [rsp+8]
	IMMEDIATE  TokenType = "IMMEDIATE"  // immediate values with prefix

	// Directives
	DIRECTIVE TokenType = "DIRECTIVE" // .data, .text, .section, etc.
	LABEL     TokenType = "LABEL"     // main:, loop:

	// Keurnel-specific
	NAMESPACE TokenType = "NAMESPACE"
	USE       TokenType = "USE"

	// Operators and delimiters
	COMMA    TokenType = "COMMA"    // ,
	COLON    TokenType = "COLON"    // :
	LBRACKET TokenType = "LBRACKET" // [
	RBRACKET TokenType = "RBRACKET" // ]
	LBRACE   TokenType = "LBRACE"   // {
	RBRACE   TokenType = "RBRACE"   // }
	PLUS     TokenType = "PLUS"     // +
	MINUS    TokenType = "MINUS"    // -
	ASTERISK TokenType = "ASTERISK" // *

	// Comments
	COMMENT TokenType = "COMMENT" // ; comment or // comment

	// Special
	NEWLINE TokenType = "NEWLINE" // line breaks (if significant)
	MACRO   TokenType = "MACRO"   // %define, %macro, etc.

	// Errors
	ErrIllegalTokenPrefix  = "illegal-token-prefix"
	ErrIllegalTokenPattern = "illegal-token-pattern"
)

type TokenType string

type InvalidTokenTypeError struct {
	Value   TokenType
	Message string
}

// Valid - verifies if the value of the TokenType is valid. Returns nil if
// valid, otherwise returns an error.
func (t *TokenType) Valid() *InvalidTokenTypeError {
	switch *t {
	default:
		return &InvalidTokenTypeError{
			Value:   *t,
			Message: "Illegal token type received.",
		}
	case ILLEGAL, EOF, IDENT, INT, FLOAT, STRING, CHAR, INSTRUCTION, REGISTER,
		MEMORY_REF, IMMEDIATE, DIRECTIVE, LABEL, NAMESPACE, USE,
		COMMA, COLON, LBRACKET, RBRACKET, LBRACE, RBRACE,
		PLUS, MINUS, ASTERISK,
		COMMENT, NEWLINE, MACRO:
		return nil
	}
}

// TokenTypeDetermine - determines the token type of given literal string. The
// literal should already be trimmed of whitespace and comments before being passed
// to this function.
func TokenTypeDetermine(literal string) TokenType {
	// =========================================================
	//
	// Handling of directives
	//
	// =========================================================
	isDirective, err := isDirective(literal)
	if err != nil {
		switch err.Error() {
		default:
			return ILLEGAL
		case ErrIllegalTokenPattern:
			return ILLEGAL_PATTERN
		case ErrIllegalTokenPrefix:
			return ILLEGAL_RESERVED
		}
	}

	if isDirective {
		return DIRECTIVE
	}

	// =========================================================
	//
	// Handling of labels (e.g., main:, loop:, etc.)
	//
	// =========================================================
	isLabel, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*:$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isLabel {
		return LABEL
	}

	// =========================================================
	//
	// Handling of identifiers (e.g., variable names, label names without colon, etc.)
	//
	// =========================================================
	isIdent, err := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isIdent {

		// Cannot be reserved keywords such as CPU opcodes or machine instructions.
		//
		if isCPUOpcode(literal) {
			return INSTRUCTION
		}

		if isMachineInstruction(literal) {
			return REGISTER
		}

		// When the identifier matches any known blacklisted keyword
		// (e.g., `data`), it cannot be an identifier as this would
		// cause ambiguity with directives. In such case we continue
		// checking other rules.
		identifierBlacklist := map[string]bool{
			"data": true,
		}
		if _, exists := identifierBlacklist[literal]; !exists {
			return IDENT
		}
	}

	// =========================================================
	//
	// Handle int literals (e.g., 42, 0x2A, 0b101010)
	//
	// =========================================================
	isInt, err := regexp.MatchString(`^(0x[0-9a-fA-F]+|0b[01]+|0o[0-7]+|0O[0-7]+|\d+)$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isInt {
		return INT
	}

	// =========================================================
	//
	// Handle float literals (e.g., 3.14, 1.23e-4, .5, 2.)
	//
	// =========================================================
	isFloat, err := regexp.MatchString(`^(\d+\.\d*|\.\d+)([eE][+-]?\d+)?$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isFloat {
		return FLOAT
	}

	// =========================================================
	//
	// Handle char literals (e.g., 'a', '\n')
	//
	// =========================================================
	isChar, err := regexp.MatchString(`^'(\\.|[^\\'])'$`, literal)
	if err != nil {
		return ILLEGAL
	}
	if isChar {
		return CHAR
	}

	// =========================================================
	//
	// Handle string literals (e.g., "hello", 'world', "Line 1\nLine 2")
	//
	// =========================================================
	isString, err := regexp.MatchString(`^("([^"\\]|\\.)*"|'([^'\\]|\\.)*')$`, literal)
	if err != nil {
		return ILLEGAL
	}

	if isString {
		return STRING
	}

	// =========================================================
	//
	// Instruction mnemonics
	//
	// =========================================================
	if isCPUOpcode(literal) {
		return INSTRUCTION
	}

	if isMachineInstruction(literal) {
		return REGISTER
	}

	// No match found, return ILLEGAL token type to indicate
	// an unrecognized token was encountered.
	//
	return ILLEGAL
}

// isDirective - checks if the given literal matches a known directive pattern.
func isDirective(literal string) (bool, error) {

	// 1. Check if literal corresponds to any CPU opcode or machine instruction. If
	// it does, then it cannot be a directive.
	//
	if isCPUOpcode(literal) || isMachineInstruction(literal) {
		return false, nil
	}

	// 2. Check if the literal matches known directive patterns. If it exists
	// in the known directives map, then it is a directive. Otherwise, continue
	// checking other rules.
	//
	knownDirectives := map[string]bool{
		".data":         true,
		".data:":        true,
		".text":         true,
		".text:":        true,
		".section":      true,
		".section:":     true,
		".global":       true,
		".global:":      true,
		".globl":        true, // alternative spelling
		".globl:":       true,
		".bss":          true, // uninitialized data section
		".bss:":         true,
		".rodata":       true, // read-only data section
		".rodata:":      true,
		".extern":       true, // external symbol declaration
		".extern:":      true,
		".byte":         true, // define byte
		".word":         true, // define word (2 bytes)
		".long":         true, // define long (4 bytes)
		".quad":         true, // define quad (8 bytes)
		".ascii":        true, // ASCII string
		".asciz":        true, // null-terminated ASCII string
		".string":       true, // string data
		".align":        true, // alignment directive
		".balign":       true, // byte alignment
		".p2align":      true, // power-of-2 alignment
		".comm":         true, // common symbol
		".local":        true, // local symbol
		".type":         true, // symbol type
		".size":         true, // symbol size
		".set":          true, // set symbol value
		".equ":          true, // equate symbol
		".equiv":        true, // equivalent symbol
		".intel_syntax": true, // Intel syntax mode
		".att_syntax":   true, // AT&T syntax mode
		".file":         true, // source file info
		".ident":        true, // identification string
	}
	if _, exists := knownDirectives[literal]; exists {
		return true, nil
	}

	// 3. Check if the literal starts with a single dot followed by characters (e.g., my_directive) and does
	// end with a colon (`:`). If it does, then it is a directive.
	//
	matched, err := regexp.MatchString(`^\.[a-zA-Z_][a-zA-Z0-9_]*:$`, literal)
	if err != nil {
		return false, errors.New(ErrIllegalTokenPattern)
	}

	if matched {

		// Directives cannot start with `.kasm` as this is reserved for Keurnel-specific directives.
		//
		if regexp.MustCompile(`^\.kasm`).MatchString(literal) {
			return false, errors.New(ErrIllegalTokenPrefix)
		}

		return true, nil
	}

	return matched, nil
}

// isCPUOpcode - checks if the given literal matches a known CPU opcode.
func isCPUOpcode(literal string) bool {
	println("CPU OPCODE CHECK: " + literal)

	opcodes := map[string]bool{
		// Data Transfer Instructions
		"MOV":    true,
		"MOVSX":  true,
		"MOVZX":  true,
		"MOVSXD": true,
		"MOVABS": true,
		"PUSH":   true,
		"POP":    true,
		"XCHG":   true,
		"LEA":    true,
		"CMOVE":  true,
		"CMOVNE": true,
		"CMOVG":  true,
		"CMOVGE": true,
		"CMOVL":  true,
		"CMOVLE": true,
		"CMOVA":  true,
		"CMOVAE": true,
		"CMOVB":  true,
		"CMOVBE": true,
		"CMOVS":  true,
		"CMOVNS": true,
		"CMOVO":  true,
		"CMOVNO": true,
		"CMOVP":  true,
		"CMOVNP": true,

		// Arithmetic Instructions
		"ADD":  true,
		"ADC":  true,
		"SUB":  true,
		"SBB":  true,
		"MUL":  true,
		"IMUL": true,
		"DIV":  true,
		"IDIV": true,
		"INC":  true,
		"DEC":  true,
		"NEG":  true,
		"CMP":  true,

		// Logical Instructions
		"AND":  true,
		"OR":   true,
		"XOR":  true,
		"NOT":  true,
		"TEST": true,

		// Shift and Rotate Instructions
		"SHL":  true,
		"SHR":  true,
		"SAL":  true,
		"SAR":  true,
		"ROL":  true,
		"ROR":  true,
		"RCL":  true,
		"RCR":  true,
		"SHLD": true,
		"SHRD": true,

		// Control Flow Instructions
		"JMP":    true,
		"JE":     true,
		"JZ":     true,
		"JNE":    true,
		"JNZ":    true,
		"JG":     true,
		"JNLE":   true,
		"JGE":    true,
		"JNL":    true,
		"JL":     true,
		"JNGE":   true,
		"JLE":    true,
		"JNG":    true,
		"JA":     true,
		"JNBE":   true,
		"JAE":    true,
		"JNB":    true,
		"JB":     true,
		"JNAE":   true,
		"JBE":    true,
		"JNA":    true,
		"JS":     true,
		"JNS":    true,
		"JO":     true,
		"JNO":    true,
		"JP":     true,
		"JPE":    true,
		"JNP":    true,
		"JPO":    true,
		"JCXZ":   true,
		"JECXZ":  true,
		"JRCXZ":  true,
		"LOOP":   true,
		"LOOPE":  true,
		"LOOPZ":  true,
		"LOOPNE": true,
		"LOOPNZ": true,
		"CALL":   true,
		"RET":    true,
		"RETN":   true,
		"RETF":   true,

		// String Instructions
		"MOVS":  true,
		"MOVSB": true,
		"MOVSW": true,
		"MOVSQ": true,
		"CMPS":  true,
		"CMPSB": true,
		"CMPSW": true,
		"CMPSQ": true,
		"SCAS":  true,
		"SCASB": true,
		"SCASW": true,
		"SCASD": true,
		"SCASQ": true,
		"LODS":  true,
		"LODSB": true,
		"LODSW": true,
		"LODSD": true,
		"LODSQ": true,
		"STOS":  true,
		"STOSB": true,
		"STOSW": true,
		"STOSD": true,
		"STOSQ": true,
		"REP":   true,
		"REPE":  true,
		"REPZ":  true,
		"REPNE": true,
		"REPNZ": true,

		// Flag Instructions
		"CLC":    true,
		"STC":    true,
		"CMC":    true,
		"CLD":    true,
		"STD":    true,
		"CLI":    true,
		"STI":    true,
		"LAHF":   true,
		"SAHF":   true,
		"PUSHF":  true,
		"PUSHFD": true,
		"PUSHFQ": true,
		"POPF":   true,
		"POPFD":  true,
		"POPFQ":  true,

		// Bit Manipulation Instructions
		"BT":     true,
		"BTS":    true,
		"BTR":    true,
		"BTC":    true,
		"BSF":    true,
		"BSR":    true,
		"BSWAP":  true,
		"POPCNT": true,
		"LZCNT":  true,
		"TZCNT":  true,

		// Set Byte on Condition Instructions
		"SETE":   true,
		"SETZ":   true,
		"SETNE":  true,
		"SETNZ":  true,
		"SETG":   true,
		"SETNLE": true,
		"SETGE":  true,
		"SETNL":  true,
		"SETL":   true,
		"SETNGE": true,
		"SETLE":  true,
		"SETNG":  true,
		"SETA":   true,
		"SETNBE": true,
		"SETAE":  true,
		"SETNB":  true,
		"SETB":   true,
		"SETNAE": true,
		"SETBE":  true,
		"SETNA":  true,
		"SETS":   true,
		"SETNS":  true,
		"SETO":   true,
		"SETNO":  true,
		"SETP":   true,
		"SETPE":  true,
		"SETNP":  true,
		"SETPO":  true,

		// System Instructions
		"NOP":      true,
		"HLT":      true,
		"INT":      true,
		"INTO":     true,
		"IRET":     true,
		"IRETD":    true,
		"IRETQ":    true,
		"SYSCALL":  true,
		"SYSRET":   true,
		"SYSENTER": true,
		"SYSEXIT":  true,
		"UD2":      true,
		"CPUID":    true,
		"RDTSC":    true,
		"RDTSCP":   true,
		"RDMSR":    true,
		"WRMSR":    true,
		"IN":       true,
		"OUT":      true,
		"INS":      true,
		"INSB":     true,
		"INSW":     true,
		"INSD":     true,
		"OUTS":     true,
		"OUTSB":    true,
		"OUTSW":    true,
		"OUTSD":    true,

		// Stack Frame Instructions
		"ENTER": true,
		"LEAVE": true,

		// Special Purpose Instructions
		"CBW":   true,
		"CWDE":  true,
		"CDQE":  true,
		"CWD":   true,
		"CDQ":   true,
		"CQO":   true,
		"XLATB": true,
		"XLAT":  true,

		// x87 FPU Instructions
		"FLD":     true,
		"FST":     true,
		"FSTP":    true,
		"FILD":    true,
		"FIST":    true,
		"FISTP":   true,
		"FADD":    true,
		"FADDP":   true,
		"FIADD":   true,
		"FSUB":    true,
		"FSUBP":   true,
		"FSUBR":   true,
		"FSUBRP":  true,
		"FISUB":   true,
		"FISUBR":  true,
		"FMUL":    true,
		"FMULP":   true,
		"FIMUL":   true,
		"FDIV":    true,
		"FDIVP":   true,
		"FDIVR":   true,
		"FDIVRP":  true,
		"FIDIV":   true,
		"FIDIVR":  true,
		"FCHS":    true,
		"FABS":    true,
		"FSQRT":   true,
		"FCOM":    true,
		"FCOMP":   true,
		"FCOMPP":  true,
		"FICOM":   true,
		"FICOMP":  true,
		"FTST":    true,
		"FXAM":    true,
		"FSIN":    true,
		"FCOS":    true,
		"FSINCOS": true,
		"FPTAN":   true,
		"FPATAN":  true,
		"F2XM1":   true,
		"FYL2X":   true,
		"FYL2XP1": true,
		"FLDZ":    true,
		"FLD1":    true,
		"FLDPI":   true,
		"FLDL2E":  true,
		"FLDL2T":  true,
		"FLDLG2":  true,
		"FLDLN2":  true,

		// SSE/SSE2 Instructions
		"MOVSS":     true,
		"MOVSD":     true, // Also string instruction
		"MOVAPS":    true,
		"MOVAPD":    true,
		"MOVUPS":    true,
		"MOVUPD":    true,
		"MOVLPS":    true,
		"MOVLPD":    true,
		"MOVHPS":    true,
		"MOVHPD":    true,
		"MOVDQA":    true,
		"MOVDQU":    true,
		"ADDSS":     true,
		"ADDSD":     true,
		"ADDPS":     true,
		"ADDPD":     true,
		"SUBSS":     true,
		"SUBSD":     true,
		"SUBPS":     true,
		"SUBPD":     true,
		"MULSS":     true,
		"MULSD":     true,
		"MULPS":     true,
		"MULPD":     true,
		"DIVSS":     true,
		"DIVSD":     true,
		"DIVPS":     true,
		"DIVPD":     true,
		"SQRTSS":    true,
		"SQRTSD":    true,
		"SQRTPS":    true,
		"SQRTPD":    true,
		"MAXSS":     true,
		"MAXSD":     true,
		"MAXPS":     true,
		"MAXPD":     true,
		"MINSS":     true,
		"MINSD":     true,
		"MINPS":     true,
		"MINPD":     true,
		"CMPSS":     true,
		"CMPSD":     true, // Also string instruction
		"CMPPS":     true,
		"CMPPD":     true,
		"COMISS":    true,
		"COMISD":    true,
		"UCOMISS":   true,
		"UCOMISD":   true,
		"ANDPS":     true,
		"ANDPD":     true,
		"ANDNPS":    true,
		"ANDNPD":    true,
		"ORPS":      true,
		"ORPD":      true,
		"XORPS":     true,
		"XORPD":     true,
		"CVTSS2SD":  true,
		"CVTSD2SS":  true,
		"CVTSI2SS":  true,
		"CVTSI2SD":  true,
		"CVTTSS2SI": true,
		"CVTTSD2SI": true,
		"CVTSS2SI":  true,
		"CVTSD2SI":  true,

		// AVX Instructions
		"VADDSS":  true,
		"VADDSD":  true,
		"VADDPS":  true,
		"VADDPD":  true,
		"VSUBSS":  true,
		"VSUBSD":  true,
		"VSUBPS":  true,
		"VSUBPD":  true,
		"VMULSS":  true,
		"VMULSD":  true,
		"VMULPS":  true,
		"VMULPD":  true,
		"VDIVSS":  true,
		"VDIVSD":  true,
		"VDIVPS":  true,
		"VDIVPD":  true,
		"VMOVSS":  true,
		"VMOVSD":  true,
		"VMOVAPS": true,
		"VMOVAPD": true,
		"VMOVUPS": true,
		"VMOVUPD": true,
		"VMOVDQA": true,
		"VMOVDQU": true,

		// Miscellaneous
		"PAUSE":      true,
		"LFENCE":     true,
		"SFENCE":     true,
		"MFENCE":     true,
		"XADD":       true,
		"CMPXCHG":    true,
		"CMPXCHG8B":  true,
		"CMPXCHG16B": true,
		"LOCK":       true,
	}

	if _, exists := opcodes[literal]; exists {
		return true
	}

	return false
}

// isMachineInstruction - checks if the given literal matches a known machine instruction.
func isMachineInstruction(literal string) bool {
	// This is a simplified check. In a real implementation, you would have a comprehensive list of machine instructions.
	instructions := []string{"RAX", "RBX", "EAX", "EBX"}
	for _, instr := range instructions {
		if literal == instr {
			return true
		}
	}
	return false
}
