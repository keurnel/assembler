package x86

const (

	// ========================================
	//
	// Data Transfer Instructions
	//
	// ========================================

	// General Purpose Instructions
	MOV  = "mov"
	PUSH = "push"
	POP  = "pop"
	XCHG = "xchg"
	XLAT = "xlat"

	// Input / Output Instructions
	IN  = "in"
	OUT = "out"

	// Address Object Instructions
	LEA = "lea"
	LDS = "lds"
	LES = "les"

	// Flag Transfer Instructions
	LAHF  = "lahf"
	SAHF  = "SAHF"
	PUSHF = "pushf"
	POPF  = "popf"

	// ========================================
	//
	// Arithmetic Instructions
	//
	// ========================================

	// Addition Instructions
	ADD = "ADD"
	ADC = "ADC"
	INC = "INC"
	AAA = "AAA"
	DAA = "DAA"

	// Subtraction Instructions
	SUB = "SUB"
	SBB = "SBB"
	DEC = "DEC"
	NEG = "NEG"
	CMP = "CMP"
	AAS = "AAS"
	DAS = "DAS"

	// Multiplication Instructions
	MUL  = "MUL"
	IMUL = "IMUL"
	AAM  = "AAM"

	// Division Instructions
	DIV  = "DIV"
	IDIV = "IDIV"
	AAD  = "AAD"
	CBW  = "CBW"
	CWD  = "CWD"

	// ========================================
	//
	// Bit manipulation Instructions
	//
	// ========================================

	// Logical Instructions
	NOT  = "NOT"
	AND  = "AND"
	OR   = "OR"
	XOR  = "XOR"
	TEST = "TEST"

	// Shifts Instructions
	SHL = "SHL"
	SAL = "SAL"
	SHR = "SHR"
	SAR = "SAR"

	// Rotates Instructions
	ROL = "ROL"
	ROR = "ROR"
	RCL = "RCL"
	RCR = "RCR"
)
