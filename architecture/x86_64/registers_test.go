package x86_64_test

import (
	"testing"

	"github.com/keurnel/assembler/architecture/x86_64"
)

// TestRegister64Bit tests all 64-bit general purpose registers
func TestRegister64Bit(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"RAX", x86_64.RAX, "rax", 0},
		{"RCX", x86_64.RCX, "rcx", 1},
		{"RDX", x86_64.RDX, "rdx", 2},
		{"RBX", x86_64.RBX, "rbx", 3},
		{"RSP", x86_64.RSP, "rsp", 4},
		{"RBP", x86_64.RBP, "rbp", 5},
		{"RSI", x86_64.RSI, "rsi", 6},
		{"RDI", x86_64.RDI, "rdi", 7},
		{"R8", x86_64.R8, "r8", 8},
		{"R9", x86_64.R9, "r9", 9},
		{"R10", x86_64.R10, "r10", 10},
		{"R11", x86_64.R11, "r11", 11},
		{"R12", x86_64.R12, "r12", 12},
		{"R13", x86_64.R13, "r13", 13},
		{"R14", x86_64.R14, "r14", 14},
		{"R15", x86_64.R15, "r15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.Register64 {
				t.Errorf("Register type = %v, want Register64", tt.reg.Type)
			}
		})
	}
}

// TestRegister32Bit tests all 32-bit general purpose registers
func TestRegister32Bit(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"EAX", x86_64.EAX, "eax", 0},
		{"ECX", x86_64.ECX, "ecx", 1},
		{"EDX", x86_64.EDX, "edx", 2},
		{"EBX", x86_64.EBX, "ebx", 3},
		{"ESP", x86_64.ESP, "esp", 4},
		{"EBP", x86_64.EBP, "ebp", 5},
		{"ESI", x86_64.ESI, "esi", 6},
		{"EDI", x86_64.EDI, "edi", 7},
		{"R8D", x86_64.R8D, "r8d", 8},
		{"R9D", x86_64.R9D, "r9d", 9},
		{"R10D", x86_64.R10D, "r10d", 10},
		{"R11D", x86_64.R11D, "r11d", 11},
		{"R12D", x86_64.R12D, "r12d", 12},
		{"R13D", x86_64.R13D, "r13d", 13},
		{"R14D", x86_64.R14D, "r14d", 14},
		{"R15D", x86_64.R15D, "r15d", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.Register32 {
				t.Errorf("Register type = %v, want Register32", tt.reg.Type)
			}
		})
	}
}

// TestRegister16Bit tests all 16-bit general purpose registers
func TestRegister16Bit(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"AX", x86_64.AX, "ax", 0},
		{"CX", x86_64.CX, "cx", 1},
		{"DX", x86_64.DX, "dx", 2},
		{"BX", x86_64.BX, "bx", 3},
		{"SP", x86_64.SP, "sp", 4},
		{"BP", x86_64.BP, "bp", 5},
		{"SI", x86_64.SI, "si", 6},
		{"DI", x86_64.DI, "di", 7},
		{"R8W", x86_64.R8W, "r8w", 8},
		{"R9W", x86_64.R9W, "r9w", 9},
		{"R10W", x86_64.R10W, "r10w", 10},
		{"R11W", x86_64.R11W, "r11w", 11},
		{"R12W", x86_64.R12W, "r12w", 12},
		{"R13W", x86_64.R13W, "r13w", 13},
		{"R14W", x86_64.R14W, "r14w", 14},
		{"R15W", x86_64.R15W, "r15w", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.Register16 {
				t.Errorf("Register type = %v, want Register16", tt.reg.Type)
			}
		})
	}
}

// TestRegister8BitLow tests all 8-bit low byte general purpose registers
func TestRegister8BitLow(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"AL", x86_64.AL, "al", 0},
		{"CL", x86_64.CL, "cl", 1},
		{"DL", x86_64.DL, "dl", 2},
		{"BL", x86_64.BL, "bl", 3},
		{"SPL", x86_64.SPL, "spl", 4},
		{"BPL", x86_64.BPL, "bpl", 5},
		{"SIL", x86_64.SIL, "sil", 6},
		{"DIL", x86_64.DIL, "dil", 7},
		{"R8B", x86_64.R8B, "r8b", 8},
		{"R9B", x86_64.R9B, "r9b", 9},
		{"R10B", x86_64.R10B, "r10b", 10},
		{"R11B", x86_64.R11B, "r11b", 11},
		{"R12B", x86_64.R12B, "r12b", 12},
		{"R13B", x86_64.R13B, "r13b", 13},
		{"R14B", x86_64.R14B, "r14b", 14},
		{"R15B", x86_64.R15B, "r15b", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.Register8 {
				t.Errorf("Register type = %v, want Register8", tt.reg.Type)
			}
		})
	}
}

// TestRegister8BitHigh tests all 8-bit high byte legacy registers
func TestRegister8BitHigh(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"AH", x86_64.AH, "ah", 4},
		{"CH", x86_64.CH, "ch", 5},
		{"DH", x86_64.DH, "dh", 6},
		{"BH", x86_64.BH, "bh", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.Register8 {
				t.Errorf("Register type = %v, want Register8", tt.reg.Type)
			}
		})
	}
}

// TestSegmentRegisters tests all segment registers
func TestSegmentRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"ES", x86_64.ES, "es", 0},
		{"CS", x86_64.CS, "cs", 1},
		{"SS", x86_64.SS, "ss", 2},
		{"DS", x86_64.DS, "ds", 3},
		{"FS", x86_64.FS, "fs", 4},
		{"GS", x86_64.GS, "gs", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterSegment {
				t.Errorf("Register type = %v, want RegisterSegment", tt.reg.Type)
			}
		})
	}
}

// TestControlRegisters tests all control registers
func TestControlRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"CR0", x86_64.CR0, "cr0", 0},
		{"CR1", x86_64.CR1, "cr1", 1},
		{"CR2", x86_64.CR2, "cr2", 2},
		{"CR3", x86_64.CR3, "cr3", 3},
		{"CR4", x86_64.CR4, "cr4", 4},
		{"CR5", x86_64.CR5, "cr5", 5},
		{"CR6", x86_64.CR6, "cr6", 6},
		{"CR7", x86_64.CR7, "cr7", 7},
		{"CR8", x86_64.CR8, "cr8", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterControl {
				t.Errorf("Register type = %v, want RegisterControl", tt.reg.Type)
			}
		})
	}
}

// TestDebugRegisters tests all debug registers
func TestDebugRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"DR0", x86_64.DR0, "dr0", 0},
		{"DR1", x86_64.DR1, "dr1", 1},
		{"DR2", x86_64.DR2, "dr2", 2},
		{"DR3", x86_64.DR3, "dr3", 3},
		{"DR4", x86_64.DR4, "dr4", 4},
		{"DR5", x86_64.DR5, "dr5", 5},
		{"DR6", x86_64.DR6, "dr6", 6},
		{"DR7", x86_64.DR7, "dr7", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterDebug {
				t.Errorf("Register type = %v, want RegisterDebug", tt.reg.Type)
			}
		})
	}
}

// TestMMXRegisters tests all MMX registers
func TestMMXRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"MM0", x86_64.MM0, "mm0", 0},
		{"MM1", x86_64.MM1, "mm1", 1},
		{"MM2", x86_64.MM2, "mm2", 2},
		{"MM3", x86_64.MM3, "mm3", 3},
		{"MM4", x86_64.MM4, "mm4", 4},
		{"MM5", x86_64.MM5, "mm5", 5},
		{"MM6", x86_64.MM6, "mm6", 6},
		{"MM7", x86_64.MM7, "mm7", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterMMX {
				t.Errorf("Register type = %v, want RegisterMMX", tt.reg.Type)
			}
		})
	}
}

// TestXMMRegisters tests all XMM registers
func TestXMMRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"XMM0", x86_64.XMM0, "xmm0", 0},
		{"XMM1", x86_64.XMM1, "xmm1", 1},
		{"XMM2", x86_64.XMM2, "xmm2", 2},
		{"XMM3", x86_64.XMM3, "xmm3", 3},
		{"XMM4", x86_64.XMM4, "xmm4", 4},
		{"XMM5", x86_64.XMM5, "xmm5", 5},
		{"XMM6", x86_64.XMM6, "xmm6", 6},
		{"XMM7", x86_64.XMM7, "xmm7", 7},
		{"XMM8", x86_64.XMM8, "xmm8", 8},
		{"XMM9", x86_64.XMM9, "xmm9", 9},
		{"XMM10", x86_64.XMM10, "xmm10", 10},
		{"XMM11", x86_64.XMM11, "xmm11", 11},
		{"XMM12", x86_64.XMM12, "xmm12", 12},
		{"XMM13", x86_64.XMM13, "xmm13", 13},
		{"XMM14", x86_64.XMM14, "xmm14", 14},
		{"XMM15", x86_64.XMM15, "xmm15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterXMM {
				t.Errorf("Register type = %v, want RegisterXMM", tt.reg.Type)
			}
		})
	}
}

// TestYMMRegisters tests all YMM registers
func TestYMMRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"YMM0", x86_64.YMM0, "ymm0", 0},
		{"YMM1", x86_64.YMM1, "ymm1", 1},
		{"YMM2", x86_64.YMM2, "ymm2", 2},
		{"YMM3", x86_64.YMM3, "ymm3", 3},
		{"YMM4", x86_64.YMM4, "ymm4", 4},
		{"YMM5", x86_64.YMM5, "ymm5", 5},
		{"YMM6", x86_64.YMM6, "ymm6", 6},
		{"YMM7", x86_64.YMM7, "ymm7", 7},
		{"YMM8", x86_64.YMM8, "ymm8", 8},
		{"YMM9", x86_64.YMM9, "ymm9", 9},
		{"YMM10", x86_64.YMM10, "ymm10", 10},
		{"YMM11", x86_64.YMM11, "ymm11", 11},
		{"YMM12", x86_64.YMM12, "ymm12", 12},
		{"YMM13", x86_64.YMM13, "ymm13", 13},
		{"YMM14", x86_64.YMM14, "ymm14", 14},
		{"YMM15", x86_64.YMM15, "ymm15", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterYMM {
				t.Errorf("Register type = %v, want RegisterYMM", tt.reg.Type)
			}
		})
	}
}

// TestZMMRegisters tests all ZMM registers
func TestZMMRegisters(t *testing.T) {
	tests := []struct {
		name     string
		reg      x86_64.Register
		wantName string
		wantEnc  byte
	}{
		{"ZMM0", x86_64.ZMM0, "zmm0", 0},
		{"ZMM1", x86_64.ZMM1, "zmm1", 1},
		{"ZMM2", x86_64.ZMM2, "zmm2", 2},
		{"ZMM3", x86_64.ZMM3, "zmm3", 3},
		{"ZMM4", x86_64.ZMM4, "zmm4", 4},
		{"ZMM5", x86_64.ZMM5, "zmm5", 5},
		{"ZMM6", x86_64.ZMM6, "zmm6", 6},
		{"ZMM7", x86_64.ZMM7, "zmm7", 7},
		{"ZMM8", x86_64.ZMM8, "zmm8", 8},
		{"ZMM9", x86_64.ZMM9, "zmm9", 9},
		{"ZMM10", x86_64.ZMM10, "zmm10", 10},
		{"ZMM11", x86_64.ZMM11, "zmm11", 11},
		{"ZMM12", x86_64.ZMM12, "zmm12", 12},
		{"ZMM13", x86_64.ZMM13, "zmm13", 13},
		{"ZMM14", x86_64.ZMM14, "zmm14", 14},
		{"ZMM15", x86_64.ZMM15, "zmm15", 15},
		{"ZMM16", x86_64.ZMM16, "zmm16", 16},
		{"ZMM17", x86_64.ZMM17, "zmm17", 17},
		{"ZMM18", x86_64.ZMM18, "zmm18", 18},
		{"ZMM19", x86_64.ZMM19, "zmm19", 19},
		{"ZMM20", x86_64.ZMM20, "zmm20", 20},
		{"ZMM21", x86_64.ZMM21, "zmm21", 21},
		{"ZMM22", x86_64.ZMM22, "zmm22", 22},
		{"ZMM23", x86_64.ZMM23, "zmm23", 23},
		{"ZMM24", x86_64.ZMM24, "zmm24", 24},
		{"ZMM25", x86_64.ZMM25, "zmm25", 25},
		{"ZMM26", x86_64.ZMM26, "zmm26", 26},
		{"ZMM27", x86_64.ZMM27, "zmm27", 27},
		{"ZMM28", x86_64.ZMM28, "zmm28", 28},
		{"ZMM29", x86_64.ZMM29, "zmm29", 29},
		{"ZMM30", x86_64.ZMM30, "zmm30", 30},
		{"ZMM31", x86_64.ZMM31, "zmm31", 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reg.Name != tt.wantName {
				t.Errorf("Register name = %v, want %v", tt.reg.Name, tt.wantName)
			}
			if tt.reg.Encoding != tt.wantEnc {
				t.Errorf("Register encoding = %v, want %v", tt.reg.Encoding, tt.wantEnc)
			}
			if tt.reg.Type != x86_64.RegisterZMM {
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
		expectedReg x86_64.Register
		shouldExist bool
	}{
		// 64-bit
		{"lookup rax", "rax", x86_64.RAX, true},
		{"lookup r15", "r15", x86_64.R15, true},
		// 32-bit
		{"lookup eax", "eax", x86_64.EAX, true},
		{"lookup r15d", "r15d", x86_64.R15D, true},
		// 16-bit
		{"lookup ax", "ax", x86_64.AX, true},
		{"lookup r15w", "r15w", x86_64.R15W, true},
		// 8-bit low
		{"lookup al", "al", x86_64.AL, true},
		{"lookup r15b", "r15b", x86_64.R15B, true},
		// 8-bit high
		{"lookup ah", "ah", x86_64.AH, true},
		{"lookup bh", "bh", x86_64.BH, true},
		// Segment
		{"lookup fs", "fs", x86_64.FS, true},
		{"lookup gs", "gs", x86_64.GS, true},
		// Control
		{"lookup cr0", "cr0", x86_64.CR0, true},
		{"lookup cr8", "cr8", x86_64.CR8, true},
		// Debug
		{"lookup dr0", "dr0", x86_64.DR0, true},
		{"lookup dr7", "dr7", x86_64.DR7, true},
		// MMX
		{"lookup mm0", "mm0", x86_64.MM0, true},
		{"lookup mm7", "mm7", x86_64.MM7, true},
		// XMM
		{"lookup xmm0", "xmm0", x86_64.XMM0, true},
		{"lookup xmm15", "xmm15", x86_64.XMM15, true},
		// YMM
		{"lookup ymm0", "ymm0", x86_64.YMM0, true},
		{"lookup ymm15", "ymm15", x86_64.YMM15, true},
		// ZMM
		{"lookup zmm0", "zmm0", x86_64.ZMM0, true},
		{"lookup zmm31", "zmm31", x86_64.ZMM31, true},
		// Non-existent
		{"lookup invalid", "invalid", x86_64.Register{}, false},
		{"lookup r16", "r16", x86_64.Register{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, exists := x86_64.RegistersByName[tt.lookupName]
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
	allRegisters := []x86_64.Register{
		// 64-bit
		x86_64.RAX, x86_64.RCX, x86_64.RDX, x86_64.RBX, x86_64.RSP, x86_64.RBP, x86_64.RSI, x86_64.RDI,
		x86_64.R8, x86_64.R9, x86_64.R10, x86_64.R11, x86_64.R12, x86_64.R13, x86_64.R14, x86_64.R15,
		// 32-bit
		x86_64.EAX, x86_64.ECX, x86_64.EDX, x86_64.EBX, x86_64.ESP, x86_64.EBP, x86_64.ESI, x86_64.EDI,
		x86_64.R8D, x86_64.R9D, x86_64.R10D, x86_64.R11D, x86_64.R12D, x86_64.R13D, x86_64.R14D, x86_64.R15D,
		// 16-bit
		x86_64.AX, x86_64.CX, x86_64.DX, x86_64.BX, x86_64.SP, x86_64.BP, x86_64.SI, x86_64.DI,
		x86_64.R8W, x86_64.R9W, x86_64.R10W, x86_64.R11W, x86_64.R12W, x86_64.R13W, x86_64.R14W, x86_64.R15W,
		// 8-bit low
		x86_64.AL, x86_64.CL, x86_64.DL, x86_64.BL, x86_64.SPL, x86_64.BPL, x86_64.SIL, x86_64.DIL,
		x86_64.R8B, x86_64.R9B, x86_64.R10B, x86_64.R11B, x86_64.R12B, x86_64.R13B, x86_64.R14B, x86_64.R15B,
		// 8-bit high
		x86_64.AH, x86_64.CH, x86_64.DH, x86_64.BH,
		// Segment
		x86_64.ES, x86_64.CS, x86_64.SS, x86_64.DS, x86_64.FS, x86_64.GS,
		// Control
		x86_64.CR0, x86_64.CR1, x86_64.CR2, x86_64.CR3, x86_64.CR4, x86_64.CR5, x86_64.CR6, x86_64.CR7, x86_64.CR8,
		// Debug
		x86_64.DR0, x86_64.DR1, x86_64.DR2, x86_64.DR3, x86_64.DR4, x86_64.DR5, x86_64.DR6, x86_64.DR7,
		// MMX
		x86_64.MM0, x86_64.MM1, x86_64.MM2, x86_64.MM3, x86_64.MM4, x86_64.MM5, x86_64.MM6, x86_64.MM7,
		// XMM
		x86_64.XMM0, x86_64.XMM1, x86_64.XMM2, x86_64.XMM3, x86_64.XMM4, x86_64.XMM5, x86_64.XMM6, x86_64.XMM7,
		x86_64.XMM8, x86_64.XMM9, x86_64.XMM10, x86_64.XMM11, x86_64.XMM12, x86_64.XMM13, x86_64.XMM14, x86_64.XMM15,
		// YMM
		x86_64.YMM0, x86_64.YMM1, x86_64.YMM2, x86_64.YMM3, x86_64.YMM4, x86_64.YMM5, x86_64.YMM6, x86_64.YMM7,
		x86_64.YMM8, x86_64.YMM9, x86_64.YMM10, x86_64.YMM11, x86_64.YMM12, x86_64.YMM13, x86_64.YMM14, x86_64.YMM15,
		// ZMM
		x86_64.ZMM0, x86_64.ZMM1, x86_64.ZMM2, x86_64.ZMM3, x86_64.ZMM4, x86_64.ZMM5, x86_64.ZMM6, x86_64.ZMM7,
		x86_64.ZMM8, x86_64.ZMM9, x86_64.ZMM10, x86_64.ZMM11, x86_64.ZMM12, x86_64.ZMM13, x86_64.ZMM14, x86_64.ZMM15,
		x86_64.ZMM16, x86_64.ZMM17, x86_64.ZMM18, x86_64.ZMM19, x86_64.ZMM20, x86_64.ZMM21, x86_64.ZMM22, x86_64.ZMM23,
		x86_64.ZMM24, x86_64.ZMM25, x86_64.ZMM26, x86_64.ZMM27, x86_64.ZMM28, x86_64.ZMM29, x86_64.ZMM30, x86_64.ZMM31,
	}

	for _, reg := range allRegisters {
		t.Run("register_"+reg.Name, func(t *testing.T) {
			found, exists := x86_64.RegistersByName[reg.Name]
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
		regType   x86_64.RegisterType
		registers []x86_64.Register
	}{
		{
			name:    "64-bit GPRs",
			regType: x86_64.Register64,
			registers: []x86_64.Register{
				x86_64.RAX, x86_64.RCX, x86_64.RDX, x86_64.RBX, x86_64.RSP, x86_64.RBP, x86_64.RSI, x86_64.RDI,
				x86_64.R8, x86_64.R9, x86_64.R10, x86_64.R11, x86_64.R12, x86_64.R13, x86_64.R14, x86_64.R15,
			},
		},
		{
			name:    "32-bit GPRs",
			regType: x86_64.Register32,
			registers: []x86_64.Register{
				x86_64.EAX, x86_64.ECX, x86_64.EDX, x86_64.EBX, x86_64.ESP, x86_64.EBP, x86_64.ESI, x86_64.EDI,
				x86_64.R8D, x86_64.R9D, x86_64.R10D, x86_64.R11D, x86_64.R12D, x86_64.R13D, x86_64.R14D, x86_64.R15D,
			},
		},
		{
			name:    "Segment registers",
			regType: x86_64.RegisterSegment,
			registers: []x86_64.Register{
				x86_64.ES, x86_64.CS, x86_64.SS, x86_64.DS, x86_64.FS, x86_64.GS,
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
	types := []x86_64.RegisterType{
		x86_64.Register8,
		x86_64.Register16,
		x86_64.Register32,
		x86_64.Register64,
		x86_64.RegisterMMX,
		x86_64.RegisterXMM,
		x86_64.RegisterYMM,
		x86_64.RegisterZMM,
		x86_64.RegisterSegment,
		x86_64.RegisterControl,
		x86_64.RegisterDebug,
	}

	seen := make(map[x86_64.RegisterType]bool)
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
