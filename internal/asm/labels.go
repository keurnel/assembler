package asm

// Label
//
// Each label represents a named position in the assembly code. Labels are used as targets for
// jump and call instructions, allowing the program to control the flow of execution. The Label
// struct contains the identifier (name) of the label and its offset in the machine code, which
// is the byte position where the label is defined in the assembled machine code.
//
// A Label is similar to a bookmark in a book that allows you to quickly navigate to a specific page. In assembly
// language, labels serve a similar purpose by allowing you to reference specific points in the code, such as the
// beginning of a loop, the start of a function, or a specific instruction. This makes it easier to write and read
// assembly code, as you can use meaningful names for labels instead of relying on byte offsets or line numbers.
//
// For example, consider the following assembly code snippet:
//
//		0:	start:
//		1:    mov rax, 1
//	 	3:	  call stop
//	 	4: stop:
//		5:   mov rax, 0
//		6:   ret
//
// In this example, "start" and "stop" are labels that represent specific positions in the code. The "call stop" instruction
// makes a call to the "stop" label, which allows the program to jump to that point in the code when executed. The Label struct
// would contain the identifier "start" with an offset of 0 and the identifier "stop" with an offset of 4, indicating their
// positions in the machine code.
type Label struct {
	Identifier string
	Offset     int
}

// IsLabel - checks if a given line of assembly code is a label definition. A label definition can be
// identified by a line that ends with a colon (":") and does not contain any instruction or operands.
// for example, "start:" is a valid label definition, while "mov rax, 1" is not. This function returns
// true if the line is a label definition, and false otherwise.
func IsLabel(line string, architecture Architecture) bool {
	line = trimComments(line)
	return false
}

// containsInstruction - checks if a line of assembly code contains any instruction from the provided list of instructions.
// This function iterates through the list of instructions and checks if any of them are present in the line. If it finds
// an instruction, it returns true; otherwise, it returns false after checking all instructions.
func containsInstruction(line string, instructions []string) bool {
	return false
}

// trimComments - removes any comments from a line of assembly code. In assembly language, comments are typically
// denoted by a semicolon (";"). This function checks if the line contains a semicolon and, if so, returns
// the portion of the line before the semicolon, effectively removing the comment. If there is no semicolon in the line,
// it returns the line unchanged.
func trimComments(line string) string {
	if idx := indexOf(line, ';'); idx != -1 {
		return line[:idx]
	}
	return line
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
