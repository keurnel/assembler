package x86_64

import "github.com/keurnel/assembler/internal/asm"

const (
	// EncodingLegacy - represents the legacy encoding of 64 instructions (no prefix)
	EncodingLegacy asm.InstructionEncoding = iota
	// EncodingVEX - represents the VEX prefix encoding used for AVX instructions
	EncodingVEX asm.InstructionEncoding = 1
	// EncodingEVEX - represents the EVEX prefix encoding used for AVX-512 instructions
	EncodingEVEX asm.InstructionEncoding = 2
	// EncodingXOP - represents the XOP prefix encoding used for AMD-specific instructions
	EncodingXOP asm.InstructionEncoding = 3
)
