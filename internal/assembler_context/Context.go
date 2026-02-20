package assembler_context

import "github.com/keurnel/assembler/internal/asm"

type AssemblerContext struct {
	// Architecture - the assembly architecture being used (e.g., 64, ...). This field allows the assembler
	// to perform architecture-specific operations, such as validating instructions, registers, addressing modes,
	// and generating machine code according to the rules of the specified architecture.
	Architecture asm.Architecture
}
