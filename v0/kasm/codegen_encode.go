package kasm

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/kasm/ast"
)

// ---------------------------------------------------------------------------
// x86_64 register encoding table (FR-5.5)
// ---------------------------------------------------------------------------

// registerNumber maps upper-case x86_64 register names to their encoding
// numbers. Extended registers (R8–R15) have numbers >= 8 and require a REX
// prefix (FR-6).
var registerNumber = map[string]uint8{
	"RAX": 0, "RCX": 1, "RDX": 2, "RBX": 3,
	"RSP": 4, "RBP": 5, "RSI": 6, "RDI": 7,
	"R8": 8, "R9": 9, "R10": 10, "R11": 11,
	"R12": 12, "R13": 13, "R14": 14, "R15": 15,
}

// is64BitRegister returns true if the upper-cased register name refers to a
// 64-bit general-purpose register (RAX–R15).
func is64BitRegister(name string) bool {
	_, ok := registerNumber[strings.ToUpper(name)]
	return ok
}

// isExtendedRegister returns true if the register requires the REX.R or
// REX.B extension bit (R8–R15).
func isExtendedRegister(name string) bool {
	num, ok := registerNumber[strings.ToUpper(name)]
	return ok && num >= 8
}

// ---------------------------------------------------------------------------
// ast.Operand classification (FR-5.2)
// ---------------------------------------------------------------------------

// classifyOperand returns the operand-type string used for variant lookup.
func classifyOperand(op ast.Operand) string {
	switch op.(type) {
	case *ast.RegisterOperand:
		return "register"
	case *ast.ImmediateOperand:
		return "immediate"
	case *ast.MemoryOperand:
		return "memory"
	case *ast.IdentifierOperand:
		// Identifiers are label references; they resolve to relative offsets
		// at encoding time. Treated as "relative" for variant matching.
		return "relative"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// Instruction size computation (Pass 1)
// ---------------------------------------------------------------------------

// computeInstructionSize determines how many bytes an instruction will occupy
// without actually emitting bytes. This is used in Pass 1 to compute label
// offsets (FR-2.1).
func (g *Generator) computeInstructionSize(s *ast.InstructionStmt) int {
	mnemonic := strings.ToUpper(s.Mnemonic)
	instr, exists := g.instructions[mnemonic]
	if !exists {
		// Unknown instruction — error will be recorded in Pass 2.
		return 0
	}

	operandTypes := make([]string, len(s.Operands))
	for i, op := range s.Operands {
		operandTypes[i] = classifyOperand(op)
	}

	variant := instr.FindVariant(operandTypes...)
	if variant == nil {
		// No matching variant — error will be recorded in Pass 2.
		return 0
	}

	size := int(variant.Size)

	// FR-6: Account for REX prefix when 64-bit registers are used.
	if g.needsREX(s) {
		size++
	}

	return size
}

// ---------------------------------------------------------------------------
// Instruction encoding (Pass 2 — FR-5)
// ---------------------------------------------------------------------------

// encodeInstruction encodes a single ast.InstructionStmt into bytes and appends
// them to the current section buffer. Errors are recorded via addError
// (AR-4.3).
func (g *Generator) encodeInstruction(s *ast.InstructionStmt) {
	sec := g.currentSection()
	if sec == nil {
		return
	}

	mnemonic := strings.ToUpper(s.Mnemonic)

	// FR-5.1: Look up the instruction by upper-case mnemonic.
	instr, exists := g.instructions[mnemonic]
	if !exists {
		g.addError(
			fmt.Sprintf("unknown instruction '%s'", s.Mnemonic),
			s.Line, s.Column,
		)
		return
	}

	// FR-5.2: Build the operand-type signature.
	operandTypes := make([]string, len(s.Operands))
	for i, op := range s.Operands {
		operandTypes[i] = classifyOperand(op)
	}

	// FR-5.3: Find the matching variant.
	variant := instr.FindVariant(operandTypes...)
	if variant == nil {
		g.addError(
			fmt.Sprintf("no matching variant for '%s' with operands [%s]",
				s.Mnemonic, strings.Join(operandTypes, ", ")),
			s.Line, s.Column,
		)
		return
	}

	// FR-8.4: Trace the encoding.
	var encoded []byte

	// FR-6: Emit REX prefix if needed.
	if g.needsREX(s) {
		rex := g.buildREX(s)
		encoded = append(encoded, rex)
	}

	// FR-5.4: Emit the opcode byte.
	encoded = append(encoded, variant.Opcode)

	// Encode operands based on the variant encoding.
	operandBytes := g.encodeOperands(s, variant)
	encoded = append(encoded, operandBytes...)

	// FR-8.4: Verbose trace.
	if g.debugCtx != nil {
		g.debugCtx.Trace(
			g.debugCtx.Loc(s.Line, s.Column),
			fmt.Sprintf("encode %s [%s]: %X", s.Mnemonic, variant.Encoding, encoded),
		)
	}

	sec.data = append(sec.data, encoded...)
	sec.size += len(encoded)
}

// ---------------------------------------------------------------------------
// ast.Operand encoding
// ---------------------------------------------------------------------------

// encodeOperands encodes the operands of an instruction according to the
// variant's encoding scheme.
func (g *Generator) encodeOperands(s *ast.InstructionStmt, variant *InstructionVariant) []byte {
	switch variant.Encoding {
	case "RM":
		return g.encodeRM(s)
	case "MR":
		return g.encodeMR(s)
	case "RI":
		return g.encodeRI(s)
	case "R":
		return g.encodeRelative(s)
	case "F":
		return g.encodeFar(s)
	default:
		g.addError(
			fmt.Sprintf("unsupported encoding '%s' for '%s'", variant.Encoding, s.Mnemonic),
			s.Line, s.Column,
		)
		return nil
	}
}

// encodeRM encodes a register-to-register/memory instruction (e.g. MOV r/m64, r64).
// ModR/M byte: mod=11 (register-direct), reg=source, r/m=destination.
func (g *Generator) encodeRM(s *ast.InstructionStmt) []byte {
	if len(s.Operands) < 2 {
		return nil
	}
	dst := g.encodeRegOperand(s.Operands[0], s.Line, s.Column)
	src := g.encodeRegOperand(s.Operands[1], s.Line, s.Column)
	if dst < 0 || src < 0 {
		return nil
	}
	// ModR/M: mod=11, reg=src, r/m=dst
	modrm := byte(0xC0) | byte(src&0x07)<<3 | byte(dst&0x07)
	return []byte{modrm}
}

// encodeMR encodes a memory/register-to-register instruction.
// ModR/M byte: mod=11 (register-direct), reg=destination, r/m=source.
func (g *Generator) encodeMR(s *ast.InstructionStmt) []byte {
	if len(s.Operands) < 2 {
		return nil
	}
	dst := g.encodeRegOperand(s.Operands[0], s.Line, s.Column)
	src := g.encodeRegOperand(s.Operands[1], s.Line, s.Column)
	if dst < 0 || src < 0 {
		return nil
	}
	// ModR/M: mod=11, reg=dst, r/m=src
	modrm := byte(0xC0) | byte(dst&0x07)<<3 | byte(src&0x07)
	return []byte{modrm}
}

// encodeRI encodes a register-immediate instruction (e.g. MOV r64, imm32).
// The register is encoded in the low 3 bits of the opcode (already handled
// by variant selection); the immediate follows as a 4-byte little-endian
// value.
func (g *Generator) encodeRI(s *ast.InstructionStmt) []byte {
	if len(s.Operands) < 2 {
		return nil
	}

	// The register number is added to the opcode by the caller pattern in
	// x86_64. For RI encoding, the low 3 bits of the opcode are the register.
	// We encode the register offset and the immediate value.
	regNum := g.encodeRegOperand(s.Operands[0], s.Line, s.Column)
	if regNum < 0 {
		return nil
	}

	immVal, ok := g.parseImmediate(s.Operands[1], s.Line, s.Column)
	if !ok {
		return nil
	}

	// 4-byte little-endian immediate.
	imm := make([]byte, 4)
	binary.LittleEndian.PutUint32(imm, uint32(immVal))

	// For RI encoding, the register is encoded as an offset added to the base
	// opcode. We return the register offset byte followed by the immediate.
	return append([]byte{byte(regNum & 0x07)}, imm...)
}

// encodeRelative encodes a relative jump/call operand (e.g. JMP label).
// The operand is a 4-byte signed relative offset.
func (g *Generator) encodeRelative(s *ast.InstructionStmt) []byte {
	if len(s.Operands) < 1 {
		return nil
	}

	var targetOffset int
	switch op := s.Operands[0].(type) {
	case *ast.IdentifierOperand:
		resolved, ok := g.resolveLabel(op.Name, op.Line, op.Column)
		if !ok {
			return make([]byte, 4) // placeholder
		}
		// Relative offset: target - (current position + instruction size)
		sec := g.currentSection()
		currentPos := len(sec.data) + 4 // +4 for the 4-byte offset itself
		targetOffset = resolved - currentPos
	case *ast.ImmediateOperand:
		val, ok := g.parseImmediate(s.Operands[0], s.Line, s.Column)
		if !ok {
			return make([]byte, 4)
		}
		targetOffset = int(val)
	default:
		g.addError(
			fmt.Sprintf("unsupported operand type for relative encoding: %T", s.Operands[0]),
			s.Line, s.Column,
		)
		return make([]byte, 4)
	}

	rel := make([]byte, 4)
	binary.LittleEndian.PutUint32(rel, uint32(int32(targetOffset)))
	return rel
}

// encodeFar encodes a far jump/call operand. For now, treated the same as
// relative — a 4-byte offset.
func (g *Generator) encodeFar(s *ast.InstructionStmt) []byte {
	return g.encodeRelative(s)
}

// ---------------------------------------------------------------------------
// Register encoding helper
// ---------------------------------------------------------------------------

// encodeRegOperand extracts the register number from a ast.RegisterOperand.
// Returns -1 and records an error if the operand is not a register.
func (g *Generator) encodeRegOperand(op ast.Operand, line, column int) int {
	reg, ok := op.(*ast.RegisterOperand)
	if !ok {
		g.addError(
			fmt.Sprintf("expected register operand, got %T", op),
			line, column,
		)
		return -1
	}

	num, exists := registerNumber[strings.ToUpper(reg.Name)]
	if !exists {
		g.addError(
			fmt.Sprintf("unknown register '%s'", reg.Name),
			line, column,
		)
		return -1
	}
	return int(num)
}

// ---------------------------------------------------------------------------
// Immediate parsing (FR-5.6)
// ---------------------------------------------------------------------------

// parseImmediate extracts and parses an immediate value from an operand.
// Supports decimal, hexadecimal (0x), and binary (0b) formats.
func (g *Generator) parseImmediate(op ast.Operand, line, column int) (int64, bool) {
	imm, ok := op.(*ast.ImmediateOperand)
	if !ok {
		g.addError(
			fmt.Sprintf("expected immediate operand, got %T", op),
			line, column,
		)
		return 0, false
	}

	val := imm.Value

	// FR-5.6: Hexadecimal.
	if strings.HasPrefix(val, "0x") || strings.HasPrefix(val, "0X") {
		n, err := strconv.ParseInt(val[2:], 16, 64)
		if err != nil {
			g.addError(
				fmt.Sprintf("invalid hexadecimal immediate '%s': %v", val, err),
				line, column,
			)
			return 0, false
		}
		return n, true
	}

	// FR-5.6: Binary.
	if strings.HasPrefix(val, "0b") || strings.HasPrefix(val, "0B") {
		n, err := strconv.ParseInt(val[2:], 2, 64)
		if err != nil {
			g.addError(
				fmt.Sprintf("invalid binary immediate '%s': %v", val, err),
				line, column,
			)
			return 0, false
		}
		return n, true
	}

	// FR-5.6: Decimal (including negative).
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		g.addError(
			fmt.Sprintf("invalid immediate '%s': %v", val, err),
			line, column,
		)
		return 0, false
	}
	return n, true
}

// ---------------------------------------------------------------------------
// REX prefix (FR-6)
// ---------------------------------------------------------------------------

// needsREX returns true if any operand of the instruction references a
// 64-bit register, requiring a REX prefix (FR-6.1).
func (g *Generator) needsREX(s *ast.InstructionStmt) bool {
	for _, op := range s.Operands {
		if reg, ok := op.(*ast.RegisterOperand); ok {
			if is64BitRegister(reg.Name) {
				return true
			}
		}
	}
	return false
}

// buildREX constructs the REX prefix byte for the given instruction (FR-6.2–6.4).
func (g *Generator) buildREX(s *ast.InstructionStmt) byte {
	// Base REX prefix: 0100 WRXB
	rex := byte(0x40)

	// FR-6.2: REX.W — 64-bit operand size.
	rex |= 0x08

	if len(s.Operands) >= 2 {
		// FR-6.3: REX.R — extended register in ModR/M reg field (second operand for RM).
		if reg, ok := s.Operands[1].(*ast.RegisterOperand); ok {
			if isExtendedRegister(reg.Name) {
				rex |= 0x04
			}
		}
	}

	if len(s.Operands) >= 1 {
		// FR-6.4: REX.B — extended register in ModR/M r/m field (first operand).
		if reg, ok := s.Operands[0].(*ast.RegisterOperand); ok {
			if isExtendedRegister(reg.Name) {
				rex |= 0x01
			}
		}
	}

	return rex
}

// InstructionVariant is imported from the architecture package. This alias
// avoids qualifying every use with the full package path within this file.
type InstructionVariant = architecture.InstructionVariant
