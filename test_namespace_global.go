package main

import (
	"fmt"
	"io/ioutil"

	"github.com/keurnel/assembler/internal/keurnel_asm"
)

func main() {
	// Read the file
	content, err := ioutil.ReadFile("main.kasm")
	if err != nil {
		panic(err)
	}

	// Create lexer and tokenize
	lexer := keurnel_asm.LexerNew(string(content))
	lexer.Process()

	// Create parser and parse
	parser := keurnel_asm.ParserNew(lexer)
	parser.Parse()

	fmt.Println("=== PARSED STRUCTURE ===\n")

	for name, group := range parser.Groups() {
		printGroup(name, group, 0)
	}
}

func printGroup(name string, group keurnel_asm.InstructionGroup, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	typeStr := "Unknown"
	switch group.Type {
	case 0:
		typeStr = "DIRECTIVE"
	case 1:
		typeStr = "LABEL"
	case 2:
		typeStr = "GLOBAL"
	case 3:
		typeStr = "NAMESPACE"
	}

	fmt.Printf("%sGroup: %s (Type: %s)\n", indentStr, name, typeStr)

	// Print instructions
	if len(group.Instructions) > 0 {
		fmt.Printf("%s  Instructions:\n", indentStr)
		for _, instr := range group.Instructions {
			fmt.Printf("%s    %s", indentStr, instr.Mnemonic)
			if len(instr.Operands) > 0 {
				fmt.Printf(" %v", instr.Operands)
			}
			fmt.Println()
		}
	}

	// Print children
	if len(group.Children) > 0 {
		fmt.Printf("%s  Children:\n", indentStr)
		for childName, child := range group.Children {
			printGroup(childName, child, indent+2)
		}
	}

	fmt.Println()
}
