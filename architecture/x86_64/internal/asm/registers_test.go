package asm_test

import (
	"testing"

	"github.com/keurnel/assembler/architecture/x86_64/internal/asm"
)

// TestRegister64Bit tests all 64-bit general purpose registers
func TestRegister64Bit(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"RAX", asm.RAX, "rax", 0},
		{"RCX", asm.RCX, "rcx", 1},
		{"RDX", asm.RDX, "rdx", 2},
		{"RBX", asm.RBX, "rbx", 3},
		{"RSP", asm.RSP, "rsp", 4},
		{"RBP", asm.RBP, "rbp", 5},
		{"RSI", asm.RSI, "rsi", 6},
		{"RDI", asm.RDI, "rdi", 7},
		{"R8", asm.R8, "r8", 8},
		{"R9", asm.R9, "r9", 9},
		{"R10", asm.R10, "r10", 10},
		{"R11", asm.R11, "r11", 11},
		{"R12", asm.R12, "r12", 12},
		{"R13", asm.R13, "r13", 13},
		{"R14", asm.R14, "r14", 14},
		{"R15", asm.R15, "r15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.Register64 {
				t.Errorf("Register type = %v, want Register64", tt.reg.Type)
			}
		})
	}
}

// TestRegister32Bit tests all 32-bit general purpose registers
func TestRegister32Bit(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"EAX", asm.EAX, "eax", 0},
		{"ECX", asm.ECX, "ecx", 1},
		{"EDX", asm.EDX, "edx", 2},
		{"EBX", asm.EBX, "ebx", 3},
		{"ESP", asm.ESP, "esp", 4},
		{"EBP", asm.EBP, "ebp", 5},
		{"ESI", asm.ESI, "esi", 6},
		{"EDI", asm.EDI, "edi", 7},
		{"R8D", asm.R8D, "r8d", 8},
		{"R9D", asm.R9D, "r9d", 9},
		{"R10D", asm.R10D, "r10d", 10},
		{"R11D", asm.R11D, "r11d", 11},
		{"R12D", asm.R12D, "r12d", 12},
		{"R13D", asm.R13D, "r13d", 13},
		{"R14D", asm.R14D, "r14d", 14},
		{"R15D", asm.R15D, "r15d", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.Register32 {
				t.Errorf("Register type = %v, want Register32", tt.reg.Type)
			}
		})
	}
}

// TestRegister16Bit tests all 16-bit general purpose registers
func TestRegister16Bit(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"AX", asm.AX, "ax", 0},
		{"CX", asm.CX, "cx", 1},
		{"DX", asm.DX, "dx", 2},
		{"BX", asm.BX, "bx", 3},
		{"SP", asm.SP, "sp", 4},
		{"BP", asm.BP, "bp", 5},
		{"SI", asm.SI, "si", 6},
		{"DI", asm.DI, "di", 7},
		{"R8W", asm.R8W, "r8w", 8},
		{"R9W", asm.R9W, "r9w", 9},
		{"R10W", asm.R10W, "r10w", 10},
		{"R11W", asm.R11W, "r11w", 11},
		{"R12W", asm.R12W, "r12w", 12},
		{"R13W", asm.R13W, "r13w", 13},
		{"R14W", asm.R14W, "r14w", 14},
		{"R15W", asm.R15W, "r15w", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.Register16 {
				t.Errorf("Register type = %v, want Register16", tt.reg.Type)
			}
		})
	}
}

// TestRegister8BitLow tests all 8-bit low byte general purpose registers
func TestRegister8BitLow(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"AL", asm.AL, "al", 0},
		{"CL", asm.CL, "cl", 1},
		{"DL", asm.DL, "dl", 2},
		{"BL", asm.BL, "bl", 3},
		{"SPL", asm.SPL, "spl", 4},
		{"BPL", asm.BPL, "bpl", 5},
		{"SIL", asm.SIL, "sil", 6},
		{"DIL", asm.DIL, "dil", 7},
		{"R8B", asm.R8B, "r8b", 8},
		{"R9B", asm.R9B, "r9b", 9},
		{"R10B", asm.R10B, "r10b", 10},
		{"R11B", asm.R11B, "r11b", 11},
		{"R12B", asm.R12B, "r12b", 12},
		{"R13B", asm.R13B, "r13b", 13},
		{"R14B", asm.R14B, "r14b", 14},
		{"R15B", asm.R15B, "r15b", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.Register8 {
				t.Errorf("Register type = %v, want Register8", tt.reg.Type)
			}
		})
	}
}

// TestRegister8BitHigh tests all 8-bit high byte legacy registers
func TestRegister8BitHigh(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"AH", asm.AH, "ah", 4},
		{"CH", asm.CH, "ch", 5},
		{"DH", asm.DH, "dh", 6},
		{"BH", asm.BH, "bh", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.Register8 {
				t.Errorf("Register type = %v, want Register8", tt.reg.Type)
			}
		})
	}
}

// TestSegmentRegisters tests all segment registers
func TestSegmentRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"ES", asm.ES, "es", 0},
		{"CS", asm.CS, "cs", 1},
		{"SS", asm.SS, "ss", 2},
		{"DS", asm.DS, "ds", 3},
		{"FS", asm.FS, "fs", 4},
		{"GS", asm.GS, "gs", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterSegment {
				t.Errorf("Register type = %v, want RegisterSegment", tt.reg.Type)
			}
		})
	}
}

// TestControlRegisters tests all control registers
func TestControlRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"CR0", asm.CR0, "cr0", 0},
		{"CR1", asm.CR1, "cr1", 1},
		{"CR2", asm.CR2, "cr2", 2},
		{"CR3", asm.CR3, "cr3", 3},
		{"CR4", asm.CR4, "cr4", 4},
		{"CR5", asm.CR5, "cr5", 5},
		{"CR6", asm.CR6, "cr6", 6},
		{"CR7", asm.CR7, "cr7", 7},
		{"CR8", asm.CR8, "cr8", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterControl {
				t.Errorf("Register type = %v, want RegisterControl", tt.reg.Type)
			}
		})
	}
}

// TestDebugRegisters tests all debug registers
func TestDebugRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"DR0", asm.DR0, "dr0", 0},
		{"DR1", asm.DR1, "dr1", 1},
		{"DR2", asm.DR2, "dr2", 2},
		{"DR3", asm.DR3, "dr3", 3},
		{"DR4", asm.DR4, "dr4", 4},
		{"DR5", asm.DR5, "dr5", 5},
		{"DR6", asm.DR6, "dr6", 6},
		{"DR7", asm.DR7, "dr7", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterDebug {
				t.Errorf("Register type = %v, want RegisterDebug", tt.reg.Type)
			}
		})
	}
}

// TestMMXRegisters tests all MMX registers
func TestMMXRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"MM0", asm.MM0, "mm0", 0},
		{"MM1", asm.MM1, "mm1", 1},
		{"MM2", asm.MM2, "mm2", 2},
		{"MM3", asm.MM3, "mm3", 3},
		{"MM4", asm.MM4, "mm4", 4},
		{"MM5", asm.MM5, "mm5", 5},
		{"MM6", asm.MM6, "mm6", 6},
		{"MM7", asm.MM7, "mm7", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterMMX {
				t.Errorf("Register type = %v, want RegisterMMX", tt.reg.Type)
			}
		})
	}
}

// TestXMMRegisters tests all XMM registers
func TestXMMRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"XMM0", asm.XMM0, "xmm0", 0},
		{"XMM1", asm.XMM1, "xmm1", 1},
		{"XMM2", asm.XMM2, "xmm2", 2},
		{"XMM3", asm.XMM3, "xmm3", 3},
		{"XMM4", asm.XMM4, "xmm4", 4},
		{"XMM5", asm.XMM5, "xmm5", 5},
		{"XMM6", asm.XMM6, "xmm6", 6},
		{"XMM7", asm.XMM7, "xmm7", 7},
		{"XMM8", asm.XMM8, "xmm8", 8},
		{"XMM9", asm.XMM9, "xmm9", 9},
		{"XMM10", asm.XMM10, "xmm10", 10},
		{"XMM11", asm.XMM11, "xmm11", 11},
		{"XMM12", asm.XMM12, "xmm12", 12},
		{"XMM13", asm.XMM13, "xmm13", 13},
		{"XMM14", asm.XMM14, "xmm14", 14},
		{"XMM15", asm.XMM15, "xmm15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterXMM {
				t.Errorf("Register type = %v, want RegisterXMM", tt.reg.Type)
			}
		})
	}
}

// TestYMMRegisters tests all YMM registers
func TestYMMRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"YMM0", asm.YMM0, "ymm0", 0},
		{"YMM1", asm.YMM1, "ymm1", 1},
		{"YMM2", asm.YMM2, "ymm2", 2},
		{"YMM3", asm.YMM3, "ymm3", 3},
		{"YMM4", asm.YMM4, "ymm4", 4},
		{"YMM5", asm.YMM5, "ymm5", 5},
		{"YMM6", asm.YMM6, "ymm6", 6},
		{"YMM7", asm.YMM7, "ymm7", 7},
		{"YMM8", asm.YMM8, "ymm8", 8},
		{"YMM9", asm.YMM9, "ymm9", 9},
		{"YMM10", asm.YMM10, "ymm10", 10},
		{"YMM11", asm.YMM11, "ymm11", 11},
		{"YMM12", asm.YMM12, "ymm12", 12},
		{"YMM13", asm.YMM13, "ymm13", 13},
		{"YMM14", asm.YMM14, "ymm14", 14},
		{"YMM15", asm.YMM15, "ymm15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterYMM {
				t.Errorf("Register type = %v, want RegisterYMM", tt.reg.Type)
			}
		})
	}
}

// TestZMMRegisters tests all ZMM registers
func TestZMMRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      asm.Register
		wantName string
		wantEnc  byte
	}{
		{"ZMM0", asm.ZMM0, "zmm0", 0},
		{"ZMM1", asm.ZMM1, "zmm1", 1},
		{"ZMM2", asm.ZMM2, "zmm2", 2},
		{"ZMM3", asm.ZMM3, "zmm3", 3},
		{"ZMM4", asm.ZMM4, "zmm4", 4},
		{"ZMM5", asm.ZMM5, "zmm5", 5},
		{"ZMM6", asm.ZMM6, "zmm6", 6},
		{"ZMM7", asm.ZMM7, "zmm7", 7},
		{"ZMM8", asm.ZMM8, "zmm8", 8},
		{"ZMM9", asm.ZMM9, "zmm9", 9},
		{"ZMM10", asm.ZMM10, "zmm10", 10},
		{"ZMM11", asm.ZMM11, "zmm11", 11},
		{"ZMM12", asm.ZMM12, "zmm12", 12},
		{"ZMM13", asm.ZMM13, "zmm13", 13},
		{"ZMM14", asm.ZMM14, "zmm14", 14},
		{"ZMM15", asm.ZMM15, "zmm15", 15},
		{"ZMM16", asm.ZMM16, "zmm16", 16},
		{"ZMM17", asm.ZMM17, "zmm17", 17},
		{"ZMM18", asm.ZMM18, "zmm18", 18},
		{"ZMM19", asm.ZMM19, "zmm19", 19},
		{"ZMM20", asm.ZMM20, "zmm20", 20},
		{"ZMM21", asm.ZMM21, "zmm21", 21},
		{"ZMM22", asm.ZMM22, "zmm22", 22},
		{"ZMM23", asm.ZMM23, "zmm23", 23},
		{"ZMM24", asm.ZMM24, "zmm24", 24},
		{"ZMM25", asm.ZMM25, "zmm25", 25},
		{"ZMM26", asm.ZMM26, "zmm26", 26},
		{"ZMM27", asm.ZMM27, "zmm27", 27},
		{"ZMM28", asm.ZMM28, "zmm28", 28},
		{"ZMM29", asm.ZMM29, "zmm29", 29},
		{"ZMM30", asm.ZMM30, "zmm30", 30},
		{"ZMM31", asm.ZMM31, "zmm31", 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != asm.RegisterZMM {
				t.Errorf("Register type = %v, want RegisterZMM", tt.reg.Type)
			}
		})
	}
}

// TestRegistersByName tests the RegistersByName lookup map
func TestRegistersByName(t *testing.T) {
	tests := []struct {
		name        string
		lookupName  string
		expectedReg asm.Register
		shouldExist bool
	}{
		// 64-bit
		{"lookup rax", "rax", asm.RAX, true},
		{"lookup r15", "r15", asm.R15, true},
		// 32-bit
		{"lookup eax", "eax", asm.EAX, true},
		{"lookup r15d", "r15d", asm.R15D, true},
		// 16-bit
		{"lookup ax", "ax", asm.AX, true},
		{"lookup r15w", "r15w", asm.R15W, true},
		// 8-bit low
		{"lookup al", "al", asm.AL, true},
		{"lookup r15b", "r15b", asm.R15B, true},
		// 8-bit high
		{"lookup ah", "ah", asm.AH, true},
		{"lookup bh", "bh", asm.BH, true},
		// Segment
		{"lookup fs", "fs", asm.FS, true},
		{"lookup gs", "gs", asm.GS, true},
		// Control
		{"lookup cr0", "cr0", asm.CR0, true},
		{"lookup cr8", "cr8", asm.CR8, true},
		// Debug
		{"lookup dr0", "dr0", asm.DR0, true},
		{"lookup dr7", "dr7", asm.DR7, true},
		// MMX
		{"lookup mm0", "mm0", asm.MM0, true},
		{"lookup mm7", "mm7", asm.MM7, true},
		// XMM
		{"lookup xmm0", "xmm0", asm.XMM0, true},
		{"lookup xmm15", "xmm15", asm.XMM15, true},
		// YMM
		{"lookup ymm0", "ymm0", asm.YMM0, true},
		{"lookup ymm15", "ymm15", asm.YMM15, true},
		// ZMM
		{"lookup zmm0", "zmm0", asm.ZMM0, true},
		{"lookup zmm31", "zmm31", asm.ZMM31, true},
		// Non-existent
		{"lookup invalid", "invalid", asm.Register{}, false},
		{"lookup r16", "r16", asm.Register{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, exists := asm.RegistersByName[tt.lookupName]
			if exists != tt.shouldExist {
				t.Errorf("RegistersByName[%q] exists = %v, want %v", tt.lookupName, exists, tt.shouldExist)
			}
			if tt.shouldExist {
				if reg.Name != tt.expectedReg.Name {
					t.Errorf("RegistersByName[%q].Name = %v, want %v", tt.lookupName, reg.Name, tt.expectedReg.Name)
				}
				if reg.Type != tt.expectedReg.Type {
					t.Errorf("RegistersByName[%q].Type = %v, want %v", tt.lookupName, reg.Type, tt.expectedReg.Type)
				}
				if reg.Encoding != tt.expectedReg.Encoding {
					t.Errorf("RegistersByName[%q].Encoding = %v, want %v", tt.lookupName, reg.Encoding, tt.expectedReg.Encoding)
				}
			}
		})
	}
}

// TestRegistersByNameCompleteness tests that all defined registers are in the lookup map
func TestRegistersByNameCompleteness(t *testing.T) {
	allRegisters := []asm.Register{
		// 64-bit
		asm.RAX, asm.RCX, asm.RDX, asm.RBX, asm.RSP, asm.RBP, asm.RSI, asm.RDI,
		asm.R8, asm.R9, asm.R10, asm.R11, asm.R12, asm.R13, asm.R14, asm.R15,
		// 32-bit
		asm.EAX, asm.ECX, asm.EDX, asm.EBX, asm.ESP, asm.EBP, asm.ESI, asm.EDI,
		asm.R8D, asm.R9D, asm.R10D, asm.R11D, asm.R12D, asm.R13D, asm.R14D, asm.R15D,
		// 16-bit
		asm.AX, asm.CX, asm.DX, asm.BX, asm.SP, asm.BP, asm.SI, asm.DI,
		asm.R8W, asm.R9W, asm.R10W, asm.R11W, asm.R12W, asm.R13W, asm.R14W, asm.R15W,
		// 8-bit low
		asm.AL, asm.CL, asm.DL, asm.BL, asm.SPL, asm.BPL, asm.SIL, asm.DIL,
		asm.R8B, asm.R9B, asm.R10B, asm.R11B, asm.R12B, asm.R13B, asm.R14B, asm.R15B,
		// 8-bit high
		asm.AH, asm.CH, asm.DH, asm.BH,
		// Segment
		asm.ES, asm.CS, asm.SS, asm.DS, asm.FS, asm.GS,
		// Control
		asm.CR0, asm.CR1, asm.CR2, asm.CR3, asm.CR4, asm.CR5, asm.CR6, asm.CR7, asm.CR8,
		// Debug
		asm.DR0, asm.DR1, asm.DR2, asm.DR3, asm.DR4, asm.DR5, asm.DR6, asm.DR7,
		// MMX
		asm.MM0, asm.MM1, asm.MM2, asm.MM3, asm.MM4, asm.MM5, asm.MM6, asm.MM7,
		// XMM
		asm.XMM0, asm.XMM1, asm.XMM2, asm.XMM3, asm.XMM4, asm.XMM5, asm.XMM6, asm.XMM7,
		asm.XMM8, asm.XMM9, asm.XMM10, asm.XMM11, asm.XMM12, asm.XMM13, asm.XMM14, asm.XMM15,
		// YMM
		asm.YMM0, asm.YMM1, asm.YMM2, asm.YMM3, asm.YMM4, asm.YMM5, asm.YMM6, asm.YMM7,
		asm.YMM8, asm.YMM9, asm.YMM10, asm.YMM11, asm.YMM12, asm.YMM13, asm.YMM14, asm.YMM15,
		// ZMM
		asm.ZMM0, asm.ZMM1, asm.ZMM2, asm.ZMM3, asm.ZMM4, asm.ZMM5, asm.ZMM6, asm.ZMM7,
		asm.ZMM8, asm.ZMM9, asm.ZMM10, asm.ZMM11, asm.ZMM12, asm.ZMM13, asm.ZMM14, asm.ZMM15,
		asm.ZMM16, asm.ZMM17, asm.ZMM18, asm.ZMM19, asm.ZMM20, asm.ZMM21, asm.ZMM22, asm.ZMM23,
		asm.ZMM24, asm.ZMM25, asm.ZMM26, asm.ZMM27, asm.ZMM28, asm.ZMM29, asm.ZMM30, asm.ZMM31,
	}

	for _, reg := range allRegisters {
		t.Run("register_"+reg.Name, func(t *testing.T) {
			found, exists := asm.RegistersByName[reg.Name]
			if !exists {
				t.Errorf("Register %q not found in RegistersByName", reg.Name)
				return
			}
			if found.Name != reg.Name {
				t.Errorf("RegistersByName[%q].Name = %v, want %v", reg.Name, found.Name, reg.Name)
			}
			if found.Type != reg.Type {
				t.Errorf("RegistersByName[%q].Type = %v, want %v", reg.Name, found.Type, reg.Type)
			}
			if found.Encoding != reg.Encoding {
				t.Errorf("RegistersByName[%q].Encoding = %v, want %v", reg.Name, found.Encoding, reg.Encoding)
			}
		})
	}
}

// TestRegisterEncodingUniqueness tests that registers of the same type have unique encodings
func TestRegisterEncodingUniqueness(t *testing.T) {
	testCases := []struct {
		name      string
		regType   asm.RegisterType
		registers []asm.Register
	}{
		{
			name:    "64-bit GPRs",
			regType: asm.Register64,
			registers: []asm.Register{
				asm.RAX, asm.RCX, asm.RDX, asm.RBX, asm.RSP, asm.RBP, asm.RSI, asm.RDI,
				asm.R8, asm.R9, asm.R10, asm.R11, asm.R12, asm.R13, asm.R14, asm.R15,
			},
		},
		{
			name:    "32-bit GPRs",
			regType: asm.Register32,
			registers: []asm.Register{
				asm.EAX, asm.ECX, asm.EDX, asm.EBX, asm.ESP, asm.EBP, asm.ESI, asm.EDI,
				asm.R8D, asm.R9D, asm.R10D, asm.R11D, asm.R12D, asm.R13D, asm.R14D, asm.R15D,
			},
		},
		{
			name:    "Segment registers",
			regType: asm.RegisterSegment,
			registers: []asm.Register{
				asm.ES, asm.CS, asm.SS, asm.DS, asm.FS, asm.GS,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encodings := make(map[byte]string)
			for _, reg := range tc.registers {
				if reg.Type != tc.regType {
					t.Errorf("Register %q has type %v, want %v", reg.Name, reg.Type, tc.regType)
				}
				if existing, found := encodings[reg.Encoding]; found {
					t.Errorf("Duplicate encoding %d for registers %q and %q", reg.Encoding, existing, reg.Name)
				}
				encodings[reg.Encoding] = reg.Name
			}
		})
	}
}

// TestRegisterTypeConstants tests that register type constants are unique
func TestRegisterTypeConstants(t *testing.T) {
	types := []asm.RegisterType{
		asm.Register8,
		asm.Register16,
		asm.Register32,
		asm.Register64,
		asm.RegisterMMX,
		asm.RegisterXMM,
		asm.RegisterYMM,
		asm.RegisterZMM,
		asm.RegisterSegment,
		asm.RegisterControl,
		asm.RegisterDebug,
	}

	seen := make(map[asm.RegisterType]bool)
	for _, rt := range types {
		if seen[rt] {
			t.Errorf("Duplicate RegisterType value: %v", rt)
		}
		seen[rt] = true
	}

	if len(seen) != len(types) {
		t.Errorf("Expected %d unique RegisterType values, got %d", len(types), len(seen))
	}
}
