package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	asm2 "github.com/keurnel/assembler/architecture/x86_64/internal/asm"
)

// Assembler represents the x86_64 assembler
type Assembler struct {
	instructions []ParsedInstruction
	machineCode  []byte
	labels       map[string]int // label -> offset in machine code
}

// ParsedInstruction represents a parsed assembly instruction
type ParsedInstruction struct {
	Line        int
	Mnemonic    string
	Operands    []string
	Label       string
	MachineCode []byte
}

// NewAssembler creates a new assembler instance
func NewAssembler() *Assembler {
	return &Assembler{
		instructions: make([]ParsedInstruction, 0),
		machineCode:  make([]byte, 0),
		labels:       make(map[string]int),
	}
}

// ParseLine parses a single line of assembly code
func (a *Assembler) ParseLine(lineNum int, line string) (*ParsedInstruction, error) {
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

// ParseRegister parses a register operand
func ParseRegister(operand string) (asm2.Register, bool) {
	reg, exists := asm2.RegistersByName[strings.ToLower(operand)]
	return reg, exists
}

// ParseImmediate parses an immediate value
func ParseImmediate(operand string) (int64, bool) {
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

// IsLabel checks if an operand is a label reference
func IsLabel(operand string) bool {
	// Labels are alphanumeric identifiers (not starting with digit, no 0x prefix)
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

// EncodeModRM encodes the ModR/M byte
func EncodeModRM(mod, reg, rm byte) byte {
	return (mod << 6) | ((reg & 0x7) << 3) | (rm & 0x7)
}

// AssembleInstruction assembles a single instruction into machine code
func (a *Assembler) AssembleInstruction(instr *ParsedInstruction) error {
	// Skip labels
	if instr.Label != "" && instr.Mnemonic == "" {
		return nil
	}

	// Lookup instruction
	instruction, exists := asm2.InstructionsByMnemonic[instr.Mnemonic]
	if !exists {
		return fmt.Errorf("line %d: unknown instruction: %s", instr.Line, instr.Mnemonic)
	}

	// Find matching instruction form
	var matchedForm *asm2.InstructionForm
	var operandValues []interface{}

	for i := range instruction.Forms {
		form := &instruction.Forms[i]

		// Check if operand count matches
		expectedOps := len(form.Operands)
		if form.Operands[0] == asm2.OperandNone {
			expectedOps = 0
		}

		if len(instr.Operands) != expectedOps {
			continue
		}

		// Try to match operands
		matched := true
		operandValues = make([]interface{}, len(instr.Operands))

		for j, operand := range instr.Operands {
			expectedType := form.Operands[j]

			// Try register
			if reg, ok := ParseRegister(operand); ok {
				switch expectedType {
				case asm2.OperandReg8:
					if reg.Type == asm2.Register8 {
						operandValues[j] = reg
						continue
					}
				case asm2.OperandReg16:
					if reg.Type == asm2.Register16 {
						operandValues[j] = reg
						continue
					}
				case asm2.OperandReg32:
					if reg.Type == asm2.Register32 {
						operandValues[j] = reg
						continue
					}
				case asm2.OperandReg64:
					if reg.Type == asm2.Register64 {
						operandValues[j] = reg
						continue
					}
				}
			}

			// Try immediate
			if imm, ok := ParseImmediate(operand); ok {
				switch expectedType {
				case asm2.OperandImm8:
					if imm >= -128 && imm <= 255 {
						operandValues[j] = int8(imm)
						continue
					}
				case asm2.OperandImm16:
					if imm >= -32768 && imm <= 65535 {
						operandValues[j] = int16(imm)
						continue
					}
				case asm2.OperandImm32:
					if imm >= -2147483648 && imm <= 4294967295 {
						operandValues[j] = int32(imm)
						continue
					}
				case asm2.OperandImm64:
					operandValues[j] = imm
					continue
				case asm2.OperandRel8, asm2.OperandRel32:
					operandValues[j] = int32(imm)
					continue
				}
			}

			// Try label reference (for jumps/calls)
			if IsLabel(operand) {
				switch expectedType {
				case asm2.OperandRel8:
					// Use placeholder - will be resolved later
					operandValues[j] = "label:" + operand
					continue
				case asm2.OperandRel32:
					// Use placeholder - will be resolved later
					operandValues[j] = "label:" + operand
					continue
				}
			}

			matched = false
			break
		}

		if matched {
			matchedForm = form
			break
		}
	}

	if matchedForm == nil {
		return fmt.Errorf("line %d: no matching form for %s with operands: %v",
			instr.Line, instr.Mnemonic, instr.Operands)
	}

	// Generate machine code
	var code []byte

	// Add REX prefix if needed
	if matchedForm.REXPrefix != 0 {
		code = append(code, matchedForm.REXPrefix)
	}

	// Add opcode
	code = append(code, matchedForm.Opcode...)

	// Add ModR/M byte if needed
	if matchedForm.ModRM {
		if len(operandValues) >= 2 {
			reg1, ok1 := operandValues[0].(asm2.Register)
			reg2, ok2 := operandValues[1].(asm2.Register)
			if ok1 && ok2 {
				// Register to register
				modrm := EncodeModRM(0b11, reg2.Encoding, reg1.Encoding)
				code = append(code, modrm)
			}
		} else if len(operandValues) >= 1 {
			reg, ok := operandValues[0].(asm2.Register)
			if ok {
				// Single register operand (e.g., MUL, DIV, INC, DEC)
				// Use opcode extension in reg field
				var opcodeExt byte
				switch instr.Mnemonic {
				case "MUL":
					opcodeExt = 4
				case "IMUL":
					opcodeExt = 5
				case "DIV":
					opcodeExt = 6
				case "IDIV":
					opcodeExt = 7
				case "INC":
					opcodeExt = 0
				case "DEC":
					opcodeExt = 1
				case "NEG":
					opcodeExt = 3
				case "NOT":
					opcodeExt = 2
				case "SHL":
					opcodeExt = 4
				case "SHR":
					opcodeExt = 5
				case "SAR":
					opcodeExt = 7
				case "ROL":
					opcodeExt = 0
				case "ROR":
					opcodeExt = 1
				case "JMP":
					opcodeExt = 4
				case "CALL":
					opcodeExt = 2
				}
				modrm := EncodeModRM(0b11, opcodeExt, reg.Encoding)
				code = append(code, modrm)
			}
		}
	}

	// Add immediate value if needed
	if matchedForm.Imm && len(operandValues) > 0 {
		// Find the immediate operand
		var immValue interface{}
		for i, opType := range matchedForm.Operands {
			if opType == asm2.OperandImm8 || opType == asm2.OperandImm16 ||
				opType == asm2.OperandImm32 || opType == asm2.OperandImm64 ||
				opType == asm2.OperandRel8 || opType == asm2.OperandRel32 {
				if i < len(operandValues) {
					immValue = operandValues[i]
					break
				}
			}
		}

		if immValue != nil {
			switch v := immValue.(type) {
			case int8:
				code = append(code, byte(v))
			case int16:
				code = append(code, byte(v), byte(v>>8))
			case int32:
				code = append(code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
			case int64:
				code = append(code, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
					byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
			case string:
				// Label placeholder - encode as 0x00000000 for now
				if strings.HasPrefix(v, "label:") {
					code = append(code, 0x00, 0x00, 0x00, 0x00)
				}
			}
		}
	}

	instr.MachineCode = code
	return nil
}

// Assemble assembles all instructions
func (a *Assembler) Assemble() error {
	offset := 0

	// First pass: record label positions
	for i := range a.instructions {
		instr := &a.instructions[i]

		if instr.Label != "" {
			a.labels[instr.Label] = offset
		}

		if instr.Mnemonic != "" {
			if err := a.AssembleInstruction(instr); err != nil {
				return err
			}
			offset += len(instr.MachineCode)
		}
	}

	// Build final machine code
	for i := range a.instructions {
		a.machineCode = append(a.machineCode, a.instructions[i].MachineCode...)
	}

	return nil
}

// PrintMachineCode prints the generated machine code
func (a *Assembler) PrintMachineCode() {
	fmt.Println("Machine Code:")
	fmt.Println("=============")

	offset := 0
	for i := range a.instructions {
		instr := &a.instructions[i]

		if instr.Label != "" && instr.Mnemonic == "" {
			fmt.Printf("%s:\n", instr.Label)
			continue
		}

		if len(instr.MachineCode) > 0 {
			hexCode := hex.EncodeToString(instr.MachineCode)

			// Format as pairs
			var formatted string
			for i := 0; i < len(hexCode); i += 2 {
				if i > 0 {
					formatted += " "
				}
				formatted += hexCode[i : i+2]
			}

			operandsStr := ""
			if len(instr.Operands) > 0 {
				operandsStr = " " + strings.Join(instr.Operands, ", ")
			}

			fmt.Printf("%04x: %-20s  %s%s\n", offset, formatted,
				strings.ToLower(instr.Mnemonic), operandsStr)
			offset += len(instr.MachineCode)
		}
	}

	fmt.Println()
	fmt.Printf("Total bytes: %d\n", len(a.machineCode))
	fmt.Println("\nRaw hex:")
	fmt.Println(hex.EncodeToString(a.machineCode))
}

// ReadAssemblyFile reads and parses an assembly file
func (a *Assembler) ReadAssemblyFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		instr, err := a.ParseLine(lineNum, line)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}

		if instr != nil {
			a.instructions = append(a.instructions, *instr)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}

// ReadAssemblyString reads assembly code from a string
func (a *Assembler) ReadAssemblyString(code string) error {
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		instr, err := a.ParseLine(i+1, line)
		if err != nil {
			return fmt.Errorf("line %d: %w", i+1, err)
		}

		if instr != nil {
			a.instructions = append(a.instructions, *instr)
		}
	}

	return nil
}

func main() {
	fmt.Println("x86_64 Assembler")
	fmt.Println("================")
	fmt.Println()

	// Check command line arguments
	if len(os.Args) < 2 {
		// Demo mode - assemble some example instructions
		fmt.Println("Demo Mode - Assembling example code:")
		fmt.Println()

		exampleCode := `
; Example x86_64 assembly code
mov rax, rbx          ; Move rbx to rax
add rax, 0x10         ; Add 16 to rax
sub rbx, rax          ; Subtract rax from rbx
xor rcx, rcx          ; Zero out rcx
push rax              ; Push rax onto stack
pop rdx               ; Pop into rdx
inc rsi               ; Increment rsi
dec rdi               ; Decrement rdi
cmp rax, rbx          ; Compare rax and rbx
je 0x10               ; Jump if equal
nop                   ; No operation
syscall               ; System call
ret                   ; Return
`

		fmt.Println(exampleCode)
		fmt.Println()

		assembler := NewAssembler()

		if err := assembler.ReadAssemblyString(exampleCode); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing assembly: %v\n", err)
			os.Exit(1)
		}

		if err := assembler.Assemble(); err != nil {
			fmt.Fprintf(os.Stderr, "Error assembling: %v\n", err)
			os.Exit(1)
		}

		assembler.PrintMachineCode()

		fmt.Println("\nUsage: x86_64 <file.asm>")
		fmt.Println("  Assemble the specified assembly file")
		return
	}

	// File mode - assemble the specified file
	filename := os.Args[1]
	fmt.Printf("Assembling: %s\n", filename)
	fmt.Println()

	assembler := NewAssembler()

	if err := assembler.ReadAssemblyFile(filename); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	if err := assembler.Assemble(); err != nil {
		fmt.Fprintf(os.Stderr, "Error assembling: %v\n", err)
		os.Exit(1)
	}

	assembler.PrintMachineCode()

	// Optionally write binary output
	if len(os.Args) > 2 && os.Args[2] == "-o" && len(os.Args) > 3 {
		outputFile := os.Args[3]
		if err := os.WriteFile(outputFile, assembler.machineCode, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nBinary output written to: %s\n", outputFile)
	}
}
