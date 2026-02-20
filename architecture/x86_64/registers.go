package x86_64

// RegisterType represents the type/size of a register
type RegisterType int

const (
	Register8       RegisterType = iota // 8-bit register
	Register16                          // 16-bit register
	Register32                          // 32-bit register
	Register64                          // _64-bit register
	RegisterMMX                         // MMX register (_64-bit)
	RegisterXMM                         // XMM register (128-bit)
	RegisterYMM                         // YMM register (256-bit)
	RegisterZMM                         // ZMM register (512-bit)
	RegisterSegment                     // Segment register
	RegisterControl                     // Control register
	RegisterDebug                       // Debug register
)

// Register represents an _64 register
type Register struct {
	Name     string
	Type     RegisterType
	Encoding byte // Register encoding value
}

// General Purpose Registers - _64-bit
var (
	RAX = Register{Name: "rax", Type: Register64, Encoding: 0}
	RCX = Register{Name: "rcx", Type: Register64, Encoding: 1}
	RDX = Register{Name: "rdx", Type: Register64, Encoding: 2}
	RBX = Register{Name: "rbx", Type: Register64, Encoding: 3}
	RSP = Register{Name: "rsp", Type: Register64, Encoding: 4}
	RBP = Register{Name: "rbp", Type: Register64, Encoding: 5}
	RSI = Register{Name: "rsi", Type: Register64, Encoding: 6}
	RDI = Register{Name: "rdi", Type: Register64, Encoding: 7}
	R8  = Register{Name: "r8", Type: Register64, Encoding: 8}
	R9  = Register{Name: "r9", Type: Register64, Encoding: 9}
	R10 = Register{Name: "r10", Type: Register64, Encoding: 10}
	R11 = Register{Name: "r11", Type: Register64, Encoding: 11}
	R12 = Register{Name: "r12", Type: Register64, Encoding: 12}
	R13 = Register{Name: "r13", Type: Register64, Encoding: 13}
	R14 = Register{Name: "r14", Type: Register64, Encoding: 14}
	R15 = Register{Name: "r15", Type: Register64, Encoding: 15}
)

// General Purpose Registers - 32-bit
var (
	EAX  = Register{Name: "eax", Type: Register32, Encoding: 0}
	ECX  = Register{Name: "ecx", Type: Register32, Encoding: 1}
	EDX  = Register{Name: "edx", Type: Register32, Encoding: 2}
	EBX  = Register{Name: "ebx", Type: Register32, Encoding: 3}
	ESP  = Register{Name: "esp", Type: Register32, Encoding: 4}
	EBP  = Register{Name: "ebp", Type: Register32, Encoding: 5}
	ESI  = Register{Name: "esi", Type: Register32, Encoding: 6}
	EDI  = Register{Name: "edi", Type: Register32, Encoding: 7}
	R8D  = Register{Name: "r8d", Type: Register32, Encoding: 8}
	R9D  = Register{Name: "r9d", Type: Register32, Encoding: 9}
	R10D = Register{Name: "r10d", Type: Register32, Encoding: 10}
	R11D = Register{Name: "r11d", Type: Register32, Encoding: 11}
	R12D = Register{Name: "r12d", Type: Register32, Encoding: 12}
	R13D = Register{Name: "r13d", Type: Register32, Encoding: 13}
	R14D = Register{Name: "r14d", Type: Register32, Encoding: 14}
	R15D = Register{Name: "r15d", Type: Register32, Encoding: 15}
)

// General Purpose Registers - 16-bit
var (
	AX   = Register{Name: "ax", Type: Register16, Encoding: 0}
	CX   = Register{Name: "cx", Type: Register16, Encoding: 1}
	DX   = Register{Name: "dx", Type: Register16, Encoding: 2}
	BX   = Register{Name: "bx", Type: Register16, Encoding: 3}
	SP   = Register{Name: "sp", Type: Register16, Encoding: 4}
	BP   = Register{Name: "bp", Type: Register16, Encoding: 5}
	SI   = Register{Name: "si", Type: Register16, Encoding: 6}
	DI   = Register{Name: "di", Type: Register16, Encoding: 7}
	R8W  = Register{Name: "r8w", Type: Register16, Encoding: 8}
	R9W  = Register{Name: "r9w", Type: Register16, Encoding: 9}
	R10W = Register{Name: "r10w", Type: Register16, Encoding: 10}
	R11W = Register{Name: "r11w", Type: Register16, Encoding: 11}
	R12W = Register{Name: "r12w", Type: Register16, Encoding: 12}
	R13W = Register{Name: "r13w", Type: Register16, Encoding: 13}
	R14W = Register{Name: "r14w", Type: Register16, Encoding: 14}
	R15W = Register{Name: "r15w", Type: Register16, Encoding: 15}
)

// General Purpose Registers - 8-bit (low byte)
var (
	AL   = Register{Name: "al", Type: Register8, Encoding: 0}
	CL   = Register{Name: "cl", Type: Register8, Encoding: 1}
	DL   = Register{Name: "dl", Type: Register8, Encoding: 2}
	BL   = Register{Name: "bl", Type: Register8, Encoding: 3}
	SPL  = Register{Name: "spl", Type: Register8, Encoding: 4}
	BPL  = Register{Name: "bpl", Type: Register8, Encoding: 5}
	SIL  = Register{Name: "sil", Type: Register8, Encoding: 6}
	DIL  = Register{Name: "dil", Type: Register8, Encoding: 7}
	R8B  = Register{Name: "r8b", Type: Register8, Encoding: 8}
	R9B  = Register{Name: "r9b", Type: Register8, Encoding: 9}
	R10B = Register{Name: "r10b", Type: Register8, Encoding: 10}
	R11B = Register{Name: "r11b", Type: Register8, Encoding: 11}
	R12B = Register{Name: "r12b", Type: Register8, Encoding: 12}
	R13B = Register{Name: "r13b", Type: Register8, Encoding: 13}
	R14B = Register{Name: "r14b", Type: Register8, Encoding: 14}
	R15B = Register{Name: "r15b", Type: Register8, Encoding: 15}
)

// General Purpose Registers - 8-bit (high byte, legacy)
var (
	AH = Register{Name: "ah", Type: Register8, Encoding: 4}
	CH = Register{Name: "ch", Type: Register8, Encoding: 5}
	DH = Register{Name: "dh", Type: Register8, Encoding: 6}
	BH = Register{Name: "bh", Type: Register8, Encoding: 7}
)

// Segment Registers
var (
	ES = Register{Name: "es", Type: RegisterSegment, Encoding: 0}
	CS = Register{Name: "cs", Type: RegisterSegment, Encoding: 1}
	SS = Register{Name: "ss", Type: RegisterSegment, Encoding: 2}
	DS = Register{Name: "ds", Type: RegisterSegment, Encoding: 3}
	FS = Register{Name: "fs", Type: RegisterSegment, Encoding: 4}
	GS = Register{Name: "gs", Type: RegisterSegment, Encoding: 5}
)

// Control Registers
var (
	CR0 = Register{Name: "cr0", Type: RegisterControl, Encoding: 0}
	CR1 = Register{Name: "cr1", Type: RegisterControl, Encoding: 1}
	CR2 = Register{Name: "cr2", Type: RegisterControl, Encoding: 2}
	CR3 = Register{Name: "cr3", Type: RegisterControl, Encoding: 3}
	CR4 = Register{Name: "cr4", Type: RegisterControl, Encoding: 4}
	CR5 = Register{Name: "cr5", Type: RegisterControl, Encoding: 5}
	CR6 = Register{Name: "cr6", Type: RegisterControl, Encoding: 6}
	CR7 = Register{Name: "cr7", Type: RegisterControl, Encoding: 7}
	CR8 = Register{Name: "cr8", Type: RegisterControl, Encoding: 8}
)

// Debug Registers
var (
	DR0 = Register{Name: "dr0", Type: RegisterDebug, Encoding: 0}
	DR1 = Register{Name: "dr1", Type: RegisterDebug, Encoding: 1}
	DR2 = Register{Name: "dr2", Type: RegisterDebug, Encoding: 2}
	DR3 = Register{Name: "dr3", Type: RegisterDebug, Encoding: 3}
	DR4 = Register{Name: "dr4", Type: RegisterDebug, Encoding: 4}
	DR5 = Register{Name: "dr5", Type: RegisterDebug, Encoding: 5}
	DR6 = Register{Name: "dr6", Type: RegisterDebug, Encoding: 6}
	DR7 = Register{Name: "dr7", Type: RegisterDebug, Encoding: 7}
)

// MMX Registers
var (
	MM0 = Register{Name: "mm0", Type: RegisterMMX, Encoding: 0}
	MM1 = Register{Name: "mm1", Type: RegisterMMX, Encoding: 1}
	MM2 = Register{Name: "mm2", Type: RegisterMMX, Encoding: 2}
	MM3 = Register{Name: "mm3", Type: RegisterMMX, Encoding: 3}
	MM4 = Register{Name: "mm4", Type: RegisterMMX, Encoding: 4}
	MM5 = Register{Name: "mm5", Type: RegisterMMX, Encoding: 5}
	MM6 = Register{Name: "mm6", Type: RegisterMMX, Encoding: 6}
	MM7 = Register{Name: "mm7", Type: RegisterMMX, Encoding: 7}
)

// XMM Registers (128-bit SSE)
var (
	XMM0  = Register{Name: "xmm0", Type: RegisterXMM, Encoding: 0}
	XMM1  = Register{Name: "xmm1", Type: RegisterXMM, Encoding: 1}
	XMM2  = Register{Name: "xmm2", Type: RegisterXMM, Encoding: 2}
	XMM3  = Register{Name: "xmm3", Type: RegisterXMM, Encoding: 3}
	XMM4  = Register{Name: "xmm4", Type: RegisterXMM, Encoding: 4}
	XMM5  = Register{Name: "xmm5", Type: RegisterXMM, Encoding: 5}
	XMM6  = Register{Name: "xmm6", Type: RegisterXMM, Encoding: 6}
	XMM7  = Register{Name: "xmm7", Type: RegisterXMM, Encoding: 7}
	XMM8  = Register{Name: "xmm8", Type: RegisterXMM, Encoding: 8}
	XMM9  = Register{Name: "xmm9", Type: RegisterXMM, Encoding: 9}
	XMM10 = Register{Name: "xmm10", Type: RegisterXMM, Encoding: 10}
	XMM11 = Register{Name: "xmm11", Type: RegisterXMM, Encoding: 11}
	XMM12 = Register{Name: "xmm12", Type: RegisterXMM, Encoding: 12}
	XMM13 = Register{Name: "xmm13", Type: RegisterXMM, Encoding: 13}
	XMM14 = Register{Name: "xmm14", Type: RegisterXMM, Encoding: 14}
	XMM15 = Register{Name: "xmm15", Type: RegisterXMM, Encoding: 15}
)

// YMM Registers (256-bit AVX)
var (
	YMM0  = Register{Name: "ymm0", Type: RegisterYMM, Encoding: 0}
	YMM1  = Register{Name: "ymm1", Type: RegisterYMM, Encoding: 1}
	YMM2  = Register{Name: "ymm2", Type: RegisterYMM, Encoding: 2}
	YMM3  = Register{Name: "ymm3", Type: RegisterYMM, Encoding: 3}
	YMM4  = Register{Name: "ymm4", Type: RegisterYMM, Encoding: 4}
	YMM5  = Register{Name: "ymm5", Type: RegisterYMM, Encoding: 5}
	YMM6  = Register{Name: "ymm6", Type: RegisterYMM, Encoding: 6}
	YMM7  = Register{Name: "ymm7", Type: RegisterYMM, Encoding: 7}
	YMM8  = Register{Name: "ymm8", Type: RegisterYMM, Encoding: 8}
	YMM9  = Register{Name: "ymm9", Type: RegisterYMM, Encoding: 9}
	YMM10 = Register{Name: "ymm10", Type: RegisterYMM, Encoding: 10}
	YMM11 = Register{Name: "ymm11", Type: RegisterYMM, Encoding: 11}
	YMM12 = Register{Name: "ymm12", Type: RegisterYMM, Encoding: 12}
	YMM13 = Register{Name: "ymm13", Type: RegisterYMM, Encoding: 13}
	YMM14 = Register{Name: "ymm14", Type: RegisterYMM, Encoding: 14}
	YMM15 = Register{Name: "ymm15", Type: RegisterYMM, Encoding: 15}
)

// ZMM Registers (512-bit AVX-512)
var (
	ZMM0  = Register{Name: "zmm0", Type: RegisterZMM, Encoding: 0}
	ZMM1  = Register{Name: "zmm1", Type: RegisterZMM, Encoding: 1}
	ZMM2  = Register{Name: "zmm2", Type: RegisterZMM, Encoding: 2}
	ZMM3  = Register{Name: "zmm3", Type: RegisterZMM, Encoding: 3}
	ZMM4  = Register{Name: "zmm4", Type: RegisterZMM, Encoding: 4}
	ZMM5  = Register{Name: "zmm5", Type: RegisterZMM, Encoding: 5}
	ZMM6  = Register{Name: "zmm6", Type: RegisterZMM, Encoding: 6}
	ZMM7  = Register{Name: "zmm7", Type: RegisterZMM, Encoding: 7}
	ZMM8  = Register{Name: "zmm8", Type: RegisterZMM, Encoding: 8}
	ZMM9  = Register{Name: "zmm9", Type: RegisterZMM, Encoding: 9}
	ZMM10 = Register{Name: "zmm10", Type: RegisterZMM, Encoding: 10}
	ZMM11 = Register{Name: "zmm11", Type: RegisterZMM, Encoding: 11}
	ZMM12 = Register{Name: "zmm12", Type: RegisterZMM, Encoding: 12}
	ZMM13 = Register{Name: "zmm13", Type: RegisterZMM, Encoding: 13}
	ZMM14 = Register{Name: "zmm14", Type: RegisterZMM, Encoding: 14}
	ZMM15 = Register{Name: "zmm15", Type: RegisterZMM, Encoding: 15}
	ZMM16 = Register{Name: "zmm16", Type: RegisterZMM, Encoding: 16}
	ZMM17 = Register{Name: "zmm17", Type: RegisterZMM, Encoding: 17}
	ZMM18 = Register{Name: "zmm18", Type: RegisterZMM, Encoding: 18}
	ZMM19 = Register{Name: "zmm19", Type: RegisterZMM, Encoding: 19}
	ZMM20 = Register{Name: "zmm20", Type: RegisterZMM, Encoding: 20}
	ZMM21 = Register{Name: "zmm21", Type: RegisterZMM, Encoding: 21}
	ZMM22 = Register{Name: "zmm22", Type: RegisterZMM, Encoding: 22}
	ZMM23 = Register{Name: "zmm23", Type: RegisterZMM, Encoding: 23}
	ZMM24 = Register{Name: "zmm24", Type: RegisterZMM, Encoding: 24}
	ZMM25 = Register{Name: "zmm25", Type: RegisterZMM, Encoding: 25}
	ZMM26 = Register{Name: "zmm26", Type: RegisterZMM, Encoding: 26}
	ZMM27 = Register{Name: "zmm27", Type: RegisterZMM, Encoding: 27}
	ZMM28 = Register{Name: "zmm28", Type: RegisterZMM, Encoding: 28}
	ZMM29 = Register{Name: "zmm29", Type: RegisterZMM, Encoding: 29}
	ZMM30 = Register{Name: "zmm30", Type: RegisterZMM, Encoding: 30}
	ZMM31 = Register{Name: "zmm31", Type: RegisterZMM, Encoding: 31}
)

// RegistersByName is a map for looking up registers by their name
var RegistersByName = map[string]Register{
	// _64-bit
	"rax": RAX, "rcx": RCX, "rdx": RDX, "rbx": RBX,
	"rsp": RSP, "rbp": RBP, "rsi": RSI, "rdi": RDI,
	"r8": R8, "r9": R9, "r10": R10, "r11": R11,
	"r12": R12, "r13": R13, "r14": R14, "r15": R15,
	// 32-bit
	"eax": EAX, "ecx": ECX, "edx": EDX, "ebx": EBX,
	"esp": ESP, "ebp": EBP, "esi": ESI, "edi": EDI,
	"r8d": R8D, "r9d": R9D, "r10d": R10D, "r11d": R11D,
	"r12d": R12D, "r13d": R13D, "r14d": R14D, "r15d": R15D,
	// 16-bit
	"ax": AX, "cx": CX, "dx": DX, "bx": BX,
	"sp": SP, "bp": BP, "si": SI, "di": DI,
	"r8w": R8W, "r9w": R9W, "r10w": R10W, "r11w": R11W,
	"r12w": R12W, "r13w": R13W, "r14w": R14W, "r15w": R15W,
	// 8-bit low
	"al": AL, "cl": CL, "dl": DL, "bl": BL,
	"spl": SPL, "bpl": BPL, "sil": SIL, "dil": DIL,
	"r8b": R8B, "r9b": R9B, "r10b": R10B, "r11b": R11B,
	"r12b": R12B, "r13b": R13B, "r14b": R14B, "r15b": R15B,
	// 8-bit high
	"ah": AH, "ch": CH, "dh": DH, "bh": BH,
	// Segment
	"es": ES, "cs": CS, "ss": SS, "ds": DS, "fs": FS, "gs": GS,
	// Control
	"cr0": CR0, "cr1": CR1, "cr2": CR2, "cr3": CR3,
	"cr4": CR4, "cr5": CR5, "cr6": CR6, "cr7": CR7, "cr8": CR8,
	// Debug
	"dr0": DR0, "dr1": DR1, "dr2": DR2, "dr3": DR3,
	"dr4": DR4, "dr5": DR5, "dr6": DR6, "dr7": DR7,
	// MMX
	"mm0": MM0, "mm1": MM1, "mm2": MM2, "mm3": MM3,
	"mm4": MM4, "mm5": MM5, "mm6": MM6, "mm7": MM7,
	// XMM
	"xmm0": XMM0, "xmm1": XMM1, "xmm2": XMM2, "xmm3": XMM3,
	"xmm4": XMM4, "xmm5": XMM5, "xmm6": XMM6, "xmm7": XMM7,
	"xmm8": XMM8, "xmm9": XMM9, "xmm10": XMM10, "xmm11": XMM11,
	"xmm12": XMM12, "xmm13": XMM13, "xmm14": XMM14, "xmm15": XMM15,
	// YMM
	"ymm0": YMM0, "ymm1": YMM1, "ymm2": YMM2, "ymm3": YMM3,
	"ymm4": YMM4, "ymm5": YMM5, "ymm6": YMM6, "ymm7": YMM7,
	"ymm8": YMM8, "ymm9": YMM9, "ymm10": YMM10, "ymm11": YMM11,
	"ymm12": YMM12, "ymm13": YMM13, "ymm14": YMM14, "ymm15": YMM15,
	// ZMM
	"zmm0": ZMM0, "zmm1": ZMM1, "zmm2": ZMM2, "zmm3": ZMM3,
	"zmm4": ZMM4, "zmm5": ZMM5, "zmm6": ZMM6, "zmm7": ZMM7,
	"zmm8": ZMM8, "zmm9": ZMM9, "zmm10": ZMM10, "zmm11": ZMM11,
	"zmm12": ZMM12, "zmm13": ZMM13, "zmm14": ZMM14, "zmm15": ZMM15,
	"zmm16": ZMM16, "zmm17": ZMM17, "zmm18": ZMM18, "zmm19": ZMM19,
	"zmm20": ZMM20, "zmm21": ZMM21, "zmm22": ZMM22, "zmm23": ZMM23,
	"zmm24": ZMM24, "zmm25": ZMM25, "zmm26": ZMM26, "zmm27": ZMM27,
	"zmm28": ZMM28, "zmm29": ZMM29, "zmm30": ZMM30, "zmm31": ZMM31,
}
