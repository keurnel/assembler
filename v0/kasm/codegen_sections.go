package kasm

import "sort"

// ---------------------------------------------------------------------------
// Internal types (FR-3, sectionBuffer)
// ---------------------------------------------------------------------------

// sectionBuffer accumulates bytes for a single section during code generation.
type sectionBuffer struct {
	name string
	data []byte
	size int // for .bss, tracks reserved size without emitting data
}

// sectionOrder defines the deterministic layout order of sections in the
// final binary (FR-3.3): .text first, then .data, then .bss. Unknown
// sections are placed after .bss in alphabetical order.
var sectionOrder = map[string]int{
	".text": 0,
	".data": 1,
	".bss":  2,
}

// ---------------------------------------------------------------------------
// Section management
// ---------------------------------------------------------------------------

// switchSection sets the current section to the given name. If the section
// does not yet exist, it is created (FR-3.1).
func (g *Generator) switchSection(name string) {
	g.current = name
	if _, exists := g.sections[name]; !exists {
		g.sections[name] = &sectionBuffer{
			name: name,
			data: make([]byte, 0),
		}
	}
}

// ensureSection guarantees that a current section is active. If no section
// has been declared yet, a default ".text" section is created (FR-3.2).
func (g *Generator) ensureSection(line, column int) {
	if g.current == "" {
		g.switchSection(".text")
	}
}

// currentSection returns the active section buffer, or nil if none is set.
func (g *Generator) currentSection() *sectionBuffer {
	if g.current == "" {
		return nil
	}
	return g.sections[g.current]
}

// sectionCount returns the number of sections that contain data or have
// reserved space.
func (g *Generator) sectionCount() int {
	count := 0
	for _, sec := range g.sections {
		if len(sec.data) > 0 || sec.size > 0 {
			count++
		}
	}
	return count
}

// assemble concatenates all section buffers in deterministic order to produce
// the final binary output (FR-3.3). The .bss section does not emit bytes
// (FR-3.4).
func (g *Generator) assemble() []byte {
	names := make([]string, 0, len(g.sections))
	for name := range g.sections {
		names = append(names, name)
	}

	// Sort by defined order, then alphabetically for unknown sections.
	sort.Slice(names, func(i, j int) bool {
		oi, oki := sectionOrder[names[i]]
		oj, okj := sectionOrder[names[j]]
		if !oki {
			oi = len(sectionOrder)
		}
		if !okj {
			oj = len(sectionOrder)
		}
		if oi != oj {
			return oi < oj
		}
		return names[i] < names[j]
	})

	var output []byte
	for _, name := range names {
		sec := g.sections[name]
		// FR-3.4: .bss does not emit bytes.
		if name == ".bss" {
			continue
		}
		output = append(output, sec.data...)
	}
	return output
}
