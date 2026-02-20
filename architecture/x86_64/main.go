package x86_64

import (
	"strconv"
	"strings"
)

// ParsedInstruction represents a parsed assembly instruction
type ParsedInstruction struct {
	Line        int
	Mnemonic    string
	Operands    []string
	Label       string
	MachineCode []byte
}

// New - returns a new instance of the 64 assembler
func New(rawSource string) *Assembler {
	return &Assembler{
		rawSource: rawSource,
	}
}

// Assemble - assembles the raw assembly source code into machine code
func (a *Assembler) Assemble() ([]byte, error) {
	var machineCode []byte
	labels := make(map[string]int)

	// Split source into lines and parse
	lines := strings.Split(a.rawSource, "\n")
	var instructions []ParsedInstruction
	offset := 0

	// First pass: parse instructions and collect labels
	for lineNum, line := range lines {
		instr, err := parseLine(lineNum+1, line)
		if err != nil {
			return nil, err
		}

		if instr == nil {
			continue
		}

		// Record label position
		if instr.Label != "" {
			labels[instr.Label] = offset
		}

		// Add instruction if it has a mnemonic
		if instr.Mnemonic != "" {
			instructions = append(instructions, *instr)
		}
	}

	// Second pass: assemble instructions and calculate offsets
	offset = 0
	for i := range instructions {
		instr := &instructions[i]

		if err := assembleInstruction(instr); err != nil {
			return nil, err
		}

		machineCode = append(machineCode, instr.MachineCode...)
		offset += len(instr.MachineCode)
	}

	return machineCode, nil
}

// parseLine parses a single line of assembly code
func parseLine(lineNum int, line string) (*ParsedInstruction, error) {
	// Remove comments
	if idx := strings.Index(line, ";"); idx != -1 {
		line = line[:idx]
	}
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	instr := &ParsedInstruction{Line: lineNum}

	// Check for label
	if strings.HasSuffix(line, ":") {
		instr.Label = strings.TrimSuffix(line, ":")
		return instr, nil
	}

	// Split into mnemonic and operands
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, nil
	}

	instr.Mnemonic = strings.ToUpper(parts[0])

	// Parse operands
	if len(parts) > 1 {
		operandStr := strings.Join(parts[1:], "")
		instr.Operands = strings.Split(operandStr, ",")
		for i := range instr.Operands {
			instr.Operands[i] = strings.TrimSpace(instr.Operands[i])
		}
	}

	return instr, nil
}

// parseRegister parses a register operand
func parseRegister(operand string) (Register, bool) {
	reg, exists := RegistersByName[strings.ToLower(operand)]
	return reg, exists
}

// parseImmediate parses an immediate value
func parseImmediate(operand string) (int64, bool) {
	operand = strings.TrimSpace(operand)

	// Handle hex values (0x prefix)
	if strings.HasPrefix(operand, "0x") || strings.HasPrefix(operand, "0X") {
		val, err := strconv.ParseInt(operand[2:], 16, 64)
		if err == nil {
			return val, true
		}
	}

	// Handle decimal values
	val, err := strconv.ParseInt(operand, 10, 64)
	if err == nil {
		return val, true
	}

	return 0, false
}

// isLabel checks if an operand is a label reference
func isLabel(operand string) bool {
	if len(operand) == 0 {
		return false
	}
	if operand[0] >= '0' && operand[0] <= '9' {
		return false
	}
	if strings.HasPrefix(operand, "0x") || strings.HasPrefix(operand, "0X") {
		return false
	}
	for _, c := range operand {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// encodeModRM encodes the ModR/M byte
func encodeModRM(mod, reg, rm byte) byte {
	return (mod << 6) | ((reg & 0x7) << 3) | (rm & 0x7)
}

// assembleInstruction assembles a single instruction into machine code
func assembleInstruction(instr *ParsedInstruction) error {
	// Skip labels
	if instr.Label != "" && instr.Mnemonic == "" {
		return nil
	}

	//// Lookup instruction
	//instruction, exists := asm2.InstructionsByMnemonic[instr.Mnemonic]
	//if !exists {
	//	return fmt.Errorf("line %d: unknown instruction: %s", instr.Line, instr.Mnemonic)
	//}
	//
	//// Find matching instruction form
	//var matchedForm *asm2.InstructionForm
	//var operandValues []interface{}
	//
	//for i := range instruction.Forms {
	//	form := &instruction.Forms[i]
	//
	//	// Check if operand count matches
	//	expectedOps := len(form.Operands)
	//	if len(form.Operands) > 0 && form.Operands[0] == asm2.OperandNone {
	//		expectedOps = 0
	//	}
	//
	//	if len(instr.Operands) != expectedOps {
	//		continue
	//	}
	//
	//	// Try to match operands
	//	matched := true
	//	operandValues = make([]interface{}, len(instr.Operands))
	//
	//	for j, operand := range instr.Operands {
	//		expectedType := form.Operands[j]
	//
	//		// Try register
	//		if reg, ok := parseRegister(operand); ok {
	//			switch expectedType {
	//			case asm2.OperandReg8:
	//				if reg.Type == Register8 {
	//					operandValues[j] = reg
	//					continue
	//				}
	//			case asm2.OperandReg16:
	//				if reg.Type == Register16 {
	//					operandValues[j] = reg
	//					continue
	//				}
	//			case asm2.OperandReg32:
	//				if reg.Type == Register32 {
	//					operandValues[j] = reg
	//					continue
	//				}
	//			case asm2.OperandReg64:
	//				if reg.Type == Register64 {
	//					operandValues[j] = reg
	//					continue
	//				}
	//			}
	//		}
	//
	//		// Try immediate
	//		if imm, ok := parseImmediate(operand); ok {
	//			switch expectedType {
	//			case asm2.OperandImm8:
	//				if imm >= -128 && imm <= 255 {
	//					operandValues[j] = int8(imm)
	//					continue
	//				}
	//			case asm2.OperandImm16:
	//				if imm >= -32768 && imm <= 65535 {
	//					operandValues[j] = int16(imm)
	//					continue
	//				}
	//			case asm2.OperandImm32:
	//				if imm >= -2147483648 && imm <= 4294967295 {
	//					operandValues[j] = int32(imm)
	//					continue
	//				}
	//			case asm2.OperandImm64:
	//				operandValues[j] = imm
	//				continue
	//			case asm2.OperandRel8, asm2.OperandRel32:
	//				operandValues[j] = int32(imm)
	//				continue
	//			}
	//		}
	//
	//		// Try label reference (for jumps/calls)
	//		if isLabel(operand) {
	//			switch expectedType {
	//			case asm2.OperandRel8:
	//				operandValues[j] = "label:" + operand
	//				continue
	//			case asm2.OperandRel32:
	//				operandValues[j] = "label:" + operand
	//				continue
	//			}
	//		}
	//
	//		matched = false
	//		break
	//	}
	//
	//	if matched {
	//		matchedForm = form
	//		break
	//	}
	//}
	//
	//if matchedForm == nil {
	//	return fmt.Errorf("line %d: no matching form for %s with operands: %v",
	//		instr.Line, instr.Mnemonic, instr.Operands)
	//}
	//
	//// Generate machine code
	//var code []byte
	//
	//// Add REX prefix if needed
	//if matchedForm.REXPrefix != 0 {
	//	code = append(code, matchedForm.REXPrefix)
	//}
	//
	//// Add opcode
	//code = append(code, matchedForm.Opcode...)
	//
	//// Add ModR/M byte if needed
	//if matchedForm.ModRM {
	//	if len(operandValues) >= 2 {
	//		reg1, ok1 := operandValues[0].(Register)
	//		reg2, ok2 := operandValues[1].(Register)
	//		if ok1 && ok2 {
	//			// Register to register
	//			modrm := encodeModRM(0b11, reg2.Encoding, reg1.Encoding)
	//			code = append(code, modrm)
	//		}
	//	} else if len(operandValues) >= 1 {
	//		reg, ok := operandValues[0].(Register)
	//		if ok {
	//			// Single register operand (e.g., MUL, DIV, INC, DEC)
	//			var opcodeExt byte
	//			switch instr.Mnemonic {
	//			case "MUL":
	//				opcodeExt = 4
	//			case "IMUL":
	//				opcodeExt = 5
	//			case "DIV":
	//				opcodeExt = 6
	//			case "IDIV":
	//				opcodeExt = 7
	//			case "INC":
	//				opcodeExt = 0
	//			case "DEC":
	//				opcodeExt = 1
	//			case "NEG":
	//				opcodeExt = 3
	//			case "NOT":
	//				opcodeExt = 2
	//			case "SHL":
	//				opcodeExt = 4
	//			case "SHR":
	//				opcodeExt = 5
	//			case "SAR":
	//				opcodeExt = 7
	//			case "ROL":
	//				opcodeExt = 0
	//			case "ROR":
	//				opcodeExt = 1
	//			case "JMP":
	//				opcodeExt = 4
	//			case "CALL":
	//				opcodeExt = 2
	//			}
	//			modrm := encodeModRM(0b11, opcodeExt, reg.Encoding)
	//			code = append(code, modrm)
	//		}
	//	}
	//}
	//
	//// Add immediate value if needed
	//if matchedForm.Imm && len(operandValues) > 0 {
	//	// Find the immediate operand
	//	var immValue interface{}
	//	for i, opType := range matchedForm.Operands {
	//		if opType == asm2.OperandImm8 || opType == asm2.OperandImm16 ||
	//			opType == asm2.OperandImm32 || opType == asm2.OperandImm64 ||
	//			opType == asm2.OperandRel8 || opType == asm2.OperandRel32 {
	//			if i < len(operandValues) {
	//				immValue = operandValues[i]
	//				break
	//			}
	//		}
	//	}
	//
	//	if immValue != nil {
	//		switch v := immValue.(type) {
	//		case int8:
	//			code = append(code, byte(v))
	//		case int16:
	//			code = append(code, byte(v), byte(v>>8))
	//		case int32:
	//			code = append(code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
	//		case int64:
	//			code = append(code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
	//				byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
	//		case string:
	//			// Label placeholder - encode as 0x00000000 for now
	//			if strings.HasPrefix(v, "label:") {
	//				code = append(code, 0x00, 0x00, 0x00, 0x00)
	//			}
	//		}
	//	}
	//}
	//
	//instr.MachineCode = code
	return nil
}
