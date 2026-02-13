package x86_64

import "github.com/keurnel/assembler/internal/asm"

const (
	PrefixNone        asm.Prefix = 0x00
	PrefixLock        asm.Prefix = 0xF0 // LOCK prefix
	PrefixRepNE       asm.Prefix = 0xF2 // REPNE/REPNZ prefix
	PrefixRep         asm.Prefix = 0xF3 // REP/REPE/REPZ prefix
	PrefixCS          asm.Prefix = 0x2E // CS segment override
	PrefixSS          asm.Prefix = 0x36 // SS segment override
	PrefixDS          asm.Prefix = 0x3E // DS segment override
	PrefixES          asm.Prefix = 0x26 // ES segment override
	PrefixFS          asm.Prefix = 0x64 // FS segment override
	PrefixGS          asm.Prefix = 0x65 // GS segment override
	PrefixOperandSize asm.Prefix = 0x66 // Operand-size override
	PrefixAddressSize asm.Prefix = 0x67 // Address-size override
	PrefixREX         asm.Prefix = 0x40 // REX prefix base (REX.W = 0x48)
)
