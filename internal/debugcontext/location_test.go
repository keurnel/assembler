package debugcontext

import "testing"

func TestLocation_String(t *testing.T) {
	// ==============================================================
	// FR-7.1: String() returns filePath:line:column or filePath:line.
	// ==============================================================
	t.Run("with column", func(t *testing.T) {
		loc := Loc("main.kasm", 12, 5)
		if loc.String() != "main.kasm:12:5" {
			t.Errorf("Expected 'main.kasm:12:5', got '%s'", loc.String())
		}
	})

	t.Run("without column", func(t *testing.T) {
		loc := Loc("main.kasm", 12, 0)
		if loc.String() != "main.kasm:12" {
			t.Errorf("Expected 'main.kasm:12', got '%s'", loc.String())
		}
	})
}

func TestLocation_Accessors(t *testing.T) {
	loc := Loc("test.kasm", 7, 3)

	if loc.FilePath() != "test.kasm" {
		t.Errorf("Expected FilePath 'test.kasm', got '%s'", loc.FilePath())
	}
	if loc.Line() != 7 {
		t.Errorf("Expected Line 7, got %d", loc.Line())
	}
	if loc.Column() != 3 {
		t.Errorf("Expected Column 3, got %d", loc.Column())
	}
}
