package profile

// ArchitectureProfile represents a validated, immutable vocabulary for a
// specific hardware architecture. If an ArchitectureProfile value exists, it
// is guaranteed to hold three non-nil maps — registers, instructions, and
// keywords — all keyed by lower-case strings. There is no partially-initialised
// or mutable state.
//
// The profile must not be modified after construction. The lexer stores the
// reference — mutations would corrupt classification. Because the profile is
// immutable, it is safe for concurrent use: multiple lexer instances may share
// the same profile without synchronisation.
type ArchitectureProfile interface {
	// Registers returns the set of recognised register names (lower-case).
	Registers() map[string]bool
	// Instructions returns the set of recognised instruction mnemonics (lower-case).
	Instructions() map[string]bool
	// Keywords returns the set of reserved language keywords (lower-case).
	Keywords() map[string]bool
}

// defaultKeywords returns a fresh map containing the language-level reserved
// keywords shared across all architecture profiles. Because the helper returns
// a fresh map each time (not a shared reference), callers may extend it with
// profile-specific keywords without affecting other profiles.
func defaultKeywords() map[string]bool {
	return map[string]bool{
		"namespace": true,
	}
}

// emptyProfile is an ArchitectureProfile with empty maps. It satisfies all
// three map methods with valid (empty) maps, so the lexer operates correctly —
// it simply classifies every word as an identifier.
type emptyProfile struct {
	registers    map[string]bool
	instructions map[string]bool
	keywords     map[string]bool
}

// NewEmptyProfile returns an ArchitectureProfile with empty maps. Because the
// empty profile satisfies all three map methods with valid (empty) maps, the
// lexer operates correctly — it simply classifies every word as an identifier.
// Intended for tests that need to verify classification falls through to
// TokenIdentifier for all words.
func NewEmptyProfile() ArchitectureProfile {
	return &emptyProfile{
		registers:    make(map[string]bool),
		instructions: make(map[string]bool),
		keywords:     make(map[string]bool),
	}
}

func (p *emptyProfile) Registers() map[string]bool    { return p.registers }
func (p *emptyProfile) Instructions() map[string]bool { return p.instructions }
func (p *emptyProfile) Keywords() map[string]bool     { return p.keywords }

// FromArchitecture builds an ArchitectureProfile from architecture
// instruction groups and a register map, bridging the v0/architecture package
// to the lexer. Because this helper lower-cases all mnemonics and merges the
// default keyword set, callers do not need to normalise data themselves.
func FromArchitecture(groups map[string][]string, registers map[string]bool, extraKeywords ...string) ArchitectureProfile {
	instructions := make(map[string]bool)
	for _, mnemonics := range groups {
		for _, m := range mnemonics {
			instructions[toLower(m)] = true
		}
	}

	kw := defaultKeywords()
	for _, k := range extraKeywords {
		kw[toLower(k)] = true
	}

	// Copy registers and lower-case them
	regs := make(map[string]bool, len(registers))
	for r := range registers {
		regs[toLower(r)] = true
	}

	return &staticProfile{
		registers:    regs,
		instructions: instructions,
		keywords:     kw,
	}
}

// toLower is a minimal ASCII lower-case helper to avoid importing strings
// in this file. The profile construction functions only deal with ASCII
// register/instruction names.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
