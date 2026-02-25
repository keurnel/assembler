package kasm

// staticProfile is a concrete ArchitectureProfile backed by pre-built maps.
// Because the maps are assembled at construction time, lookups are O(1) and
// the profile is immediately ready for use — there is no separate
// initialisation step.
type staticProfile struct {
	registers    map[string]bool
	instructions map[string]bool
	keywords     map[string]bool
}

func (p *staticProfile) Registers() map[string]bool    { return p.registers }
func (p *staticProfile) Instructions() map[string]bool { return p.instructions }
func (p *staticProfile) Keywords() map[string]bool     { return p.keywords }

// NewX8664Profile returns an ArchitectureProfile populated with the x86_64
// register set, instruction set, and the default keyword set. Because all
// three sets are assembled at construction time, the returned profile is
// immediately ready for use — there is no separate initialisation step.
func NewX8664Profile() ArchitectureProfile {
	return &staticProfile{
		registers:    x8664Registers(),
		instructions: x8664Instructions(),
		keywords:     defaultKeywords(),
	}
}

// x8664Registers returns the full x86_64 register set (lower-case).
// Classification is case-insensitive because the lexer lower-cases the word
// before lookup.
func x8664Registers() map[string]bool {
	return map[string]bool{
		// 64-bit general-purpose
		"rax": true, "rbx": true, "rcx": true, "rdx": true,
		"rsi": true, "rdi": true, "rbp": true, "rsp": true,
		"r8": true, "r9": true, "r10": true, "r11": true,
		"r12": true, "r13": true, "r14": true, "r15": true,
		// 32-bit general-purpose
		"eax": true, "ebx": true, "ecx": true, "edx": true,
		"esi": true, "edi": true, "ebp": true, "esp": true,
		"r8d": true, "r9d": true, "r10d": true, "r11d": true,
		"r12d": true, "r13d": true, "r14d": true, "r15d": true,
		// 16-bit general-purpose
		"ax": true, "bx": true, "cx": true, "dx": true,
		"si": true, "di": true, "bp": true, "sp": true,
		// 8-bit
		"al": true, "bl": true, "cl": true, "dl": true,
		"ah": true, "bh": true, "ch": true, "dh": true,
		"sil": true, "dil": true, "bpl": true, "spl": true,
		// Segment registers
		"cs": true, "ds": true, "es": true, "fs": true, "gs": true, "ss": true,
		// Instruction pointer / flags
		"rip": true, "eip": true, "rflags": true, "eflags": true,
	}
}

// x8664Instructions returns the full x86_64 instruction mnemonic set
// (lower-case). These mnemonics must match those provided by
// v0/architecture/x86/_64 providers. Because the profile is the single source
// of truth, adding or removing a mnemonic from the architecture package has no
// effect on the lexer until the profile is updated.
func x8664Instructions() map[string]bool {
	return map[string]bool{
		// Data transfer
		"mov": true, "movzx": true, "movsx": true, "lea": true,
		"push": true, "pop": true, "xchg": true,
		// Arithmetic
		"add": true, "sub": true, "mul": true, "imul": true,
		"div": true, "idiv": true, "inc": true, "dec": true, "neg": true,
		// Bitwise / shift
		"and": true, "or": true, "xor": true, "not": true,
		"shl": true, "shr": true, "sal": true, "sar": true,
		"rol": true, "ror": true,
		// Comparison
		"cmp": true, "test": true,
		// Control flow
		"jmp": true, "je": true, "jne": true, "jz": true, "jnz": true,
		"jg": true, "jge": true, "jl": true, "jle": true,
		"ja": true, "jae": true, "jb": true, "jbe": true,
		"call": true, "ret": true, "syscall": true, "int": true,
		// System / misc
		"nop": true, "hlt": true, "cli": true, "sti": true,
		// Loop
		"loop": true, "loope": true, "loopne": true,
		// Conditional move
		"cmove": true, "cmovne": true, "cmovg": true, "cmovl": true,
		// Set byte
		"sete": true, "setne": true, "setg": true, "setl": true,
		// String / repeat
		"rep": true, "movsb": true, "stosb": true,
		// Sign extension
		"cbw": true, "cwd": true, "cdq": true, "cqo": true,
		// Custom
		"use": true,
	}
}
