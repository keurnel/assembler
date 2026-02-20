package x86_64

import (
	"log/slog"
	"os"

	"github.com/keurnel/assembler/internal/asm"
	"github.com/keurnel/assembler/internal/keurnel_asm"
	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/architecture/x86/_64"
	"github.com/keurnel/assembler/v0/kasm"
	"github.com/spf13/cobra"
)

var AssembleFileCmd = &cobra.Command{
	Use:     "assemble-file <assembly-file>",
	GroupID: "file-operations",
	Short:   "Assemble an _64 assembly file into a binary file.",
	Long:    `Assemble an _64 assembly file into a binary file.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			cmd.PrintErrln("Error: No assembly file provided.")
			return
		}

		assemblyFile := args[0]
		if assemblyFile == "" {
			cmd.PrintErrln("Error: Assembly file path is empty.")
			return
		}

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			cmd.PrintErrln("Error: Unable to get current working directory:", err)
			return
		}

		fullPath := cwd + string(os.PathSeparator) + assemblyFile
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			cmd.PrintErrln("Error: Assembly file does not exist at path:", fullPath)
			return
		}

		// Read content of singular assembly file
		//
		sourceBytes, err := os.ReadFile(fullPath)
		if err != nil {
			cmd.PrintErrln("Error: Failed to read assembly file:", err)
			return
		}
		source := string(sourceBytes)

		//	==============================================================================
		//
		//	Loading architecture instructions
		//
		//	==============================================================================

		groups := make(map[string]architecture.InstructionGroup)
		for groupName, instructions := range _64.Instructions() {
			groups[groupName] = *architecture.FromSlice(groupName, instructions)
		}

		//	==============================================================================
		//
		//	Verifying the instructions to ensure correctness and validity
		//
		//	==============================================================================

		// todo: implement instruction verification logic here

		//	==============================================================================
		//
		//	Pre-processing of the assembly source code
		//
		//	==============================================================================

		// Handle inclusion of other `.kasm` files in the source code.
		//
		source, inclusions := kasm.PreProcessingHandleIncludes(source)
		inclusionPaths := make(map[string]bool)
		for _, inclusion := range inclusions {
			if inclusionPaths[inclusion.IncludedFilePath] {
				message := "pre-processing error: Circular inclusion detected for file '" + inclusion.IncludedFilePath + "' at line " + string(inclusion.LineNumber)
				panic(message)
			}
			inclusionPaths[inclusion.IncludedFilePath] = true
		}

		// Handle macros in the source code.
		//
		macros := kasm.PreProcessingMacroTable(source)
		kasm.PreProcessingColectMacroCalls(source, macros)
		source = kasm.PreProcessingReplaceMacroCalls(source, macros)

		// Handle conditional assembly macros in the source code.
		//
		symbolTable := kasm.PreProcessingCreateSymbolTable(source, macros)
		source = kasm.PreProcessingHandleConditionals(source, symbolTable)

		println(source)

		//	==============================================================================
		//
		//	Assembling the source code into machine code
		//
		//	==============================================================================

		return
	},
}

func assembleFile(source string, ctx asm.Architecture) (string, error) {

	lexer := keurnel_asm.LexerNew(source, &ctx)
	lexer.Process()

	parser := keurnel_asm.ParserNew(lexer)
	err := parser.Parse()
	if err != nil {
		slog.Error("Parsing failed:", "error", err)
		os.Exit(1)
	}

	// Print each group
	//
	for identifier, group := range parser.Groups() {
		println("Group Identifier:", identifier)
		println("Group Type:", group.Type)
		println("Group uses:")
		for _, ns := range group.Uses {
			println("  -", ns)
		}
		println("Instructions:")
		for _, instr := range group.Instructions {
			println("  Mnemonic:", instr.Mnemonic)
			println("  Operands:")
			for _, operand := range instr.Operands {
				println("    -", operand)
			}
		}

		// Child groups (for namespaces)
		if group.HasChildren() {
			println("Child Groups:")
			for childIdentifier, childGroup := range group.Children {
				println("  Child Group Identifier:", childIdentifier)
				println("  Child Group Type:", childGroup.Type)
				println("  Instructions:")
				for _, instr := range childGroup.Instructions {
					println("    Mnemonic:", instr.Mnemonic)
					println("    Operands:")
					for _, operand := range instr.Operands {
						println("      -", operand)
					}
				}
			}
		}

		println()
	}

	semanticAnalyzer := keurnel_asm.SemanticAnalyzerNew(parser)
	err = semanticAnalyzer.Analyze()
	if err != nil {
		slog.Error("Semantic analysis failed:", "error", err)
		os.Exit(1)
	}

	return "", nil
}

//func assembleFile(filePath string) (string, error) {
//
//	// Read content of singular assembly file
//	//
//	assembly, err := os.ReadFile(filePath)
//	if err != nil {
//		return "", err
//	}
//
//	// When the assembly file is empty, return an error
//	// indicating that the assembly file is empty and
//	// cannot be assembled.
//	//
//	if len(assembly) == 0 {
//		return "", errors.New("Assemble error: Assembly file is empty")
//	}
//
//	// Perform pre-processing steps on the assembly code
//	//
//	source := asm.PreProcessingRemoveComments(string(assembly))
//	source = asm.PreProcessingTrimWhitespace(source)
//	source = asm.PreProcessingRemoveEmptyLines(source)
//
//	// Assemble the pre-processed assembly code into machine code using
//	// the _64 assembler.
//	//
//	assembler := _64.New(source)
//	machineCode, err := assembler.Assemble()
//	if err != nil {
//		return "", err
//	}
//
//	// Return the assembled machine code as a string.
//	//
//	return string(machineCode), nil
//}
