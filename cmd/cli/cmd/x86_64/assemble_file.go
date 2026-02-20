package x86_64

import (
	"log/slog"
	"os"

	"github.com/keurnel/assembler/architecture/x86_64"
	"github.com/keurnel/assembler/internal/asm"
	"github.com/keurnel/assembler/internal/keurnel_asm"
	"github.com/spf13/cobra"
)

var AssembleFileCmd = &cobra.Command{
	Use:     "assemble-file <assembly-file>",
	GroupID: "file-operations",
	Short:   "Assemble an 64 assembly file into a binary file.",
	Long:    `Assemble an 64 assembly file into a binary file.`,
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

		// Read the raw assembly source code from the specified file and create a new instance of the 64 assembler
		//
		sourceBytes, err := os.ReadFile(fullPath)
		if err != nil {
			cmd.PrintErrln("Error: Failed to read assembly file:", err)
			return
		}
		source := string(sourceBytes)

		assemblerContext := x86_64.AssemblerNew(source)

		// Assemble file using the assembler context.
		//
		binary, err := assembleFile(assemblerContext.RawSource(), assemblerContext)

		println(binary)

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
//	// the 64 assembler.
//	//
//	assembler := 64.New(source)
//	machineCode, err := assembler.Assemble()
//	if err != nil {
//		return "", err
//	}
//
//	// Return the assembled machine code as a string.
//	//
//	return string(machineCode), nil
//}
