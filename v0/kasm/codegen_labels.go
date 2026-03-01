package kasm

import "fmt"

// ---------------------------------------------------------------------------
// Internal types (FR-4, labelEntry)
// ---------------------------------------------------------------------------

// labelEntry tracks a label's resolved address within a section.
type labelEntry struct {
	name    string
	section string
	offset  int
	line    int
	column  int
}

// ---------------------------------------------------------------------------
// Label collection (Pass 1 — FR-4.1)
// ---------------------------------------------------------------------------

// collectLabel records a label declaration and its current byte offset within
// the active section. Duplicate labels within the same section produce a
// CodegenError (FR-4.1).
func (g *Generator) collectLabel(s *LabelStmt) {
	sec := g.currentSection()
	if sec == nil {
		return
	}

	key := g.labelKey(g.current, s.Name)
	if prev, exists := g.labels[key]; exists {
		g.addError(
			fmt.Sprintf("duplicate label '%s' in section '%s', previously declared at %d:%d",
				s.Name, g.current, prev.line, prev.column),
			s.Line, s.Column,
		)
		return
	}

	g.labels[key] = labelEntry{
		name:    s.Name,
		section: g.current,
		offset:  sec.size,
		line:    s.Line,
		column:  s.Column,
	}
}

// ---------------------------------------------------------------------------
// Label resolution (Pass 2 — FR-4.2, FR-4.3)
// ---------------------------------------------------------------------------

// resolveLabel looks up a label by name in the current section and returns
// its byte offset. If the label is not found in the current section, a
// CodegenError is recorded and -1 is returned (FR-4.3, FR-4.5).
func (g *Generator) resolveLabel(name string, line, column int) (int, bool) {
	key := g.labelKey(g.current, name)
	if entry, exists := g.labels[key]; exists {
		return entry.offset, true
	}

	// FR-4.5: Cross-section references are not allowed.
	for _, entry := range g.labels {
		if entry.name == name && entry.section != g.current {
			g.addError(
				fmt.Sprintf("cross-section reference to label '%s' (declared in '%s', used in '%s')",
					name, entry.section, g.current),
				line, column,
			)
			return -1, false
		}
	}

	g.addError(
		fmt.Sprintf("unresolved label '%s'", name),
		line, column,
	)
	return -1, false
}

// labelKey produces a section-scoped key for the label table. Labels are
// scoped per section (FR-4.5).
func (g *Generator) labelKey(section, name string) string {
	return section + "\x00" + name
}
