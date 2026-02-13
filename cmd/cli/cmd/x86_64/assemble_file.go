package x86_64

import (
	"os"

	"github.com/keurnel/assembler/architecture/x86_64"
	"github.com/keurnel/assembler/internal/keurnel_asm"
	"github.com/spf13/cobra"
)

var AssembleFileCmd = &cobra.Command{
	Use:     "assemble-file <assembly-file>",
	GroupID: "file-operations",
	Short:   "Assemble an x86_64 assembly file into a binary file.",
	Long:    `Assemble an x86_64 assembly file into a binary file.`,
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

		// Read the raw assembly source code from the specified file and create a new instance of the x86_64 assembler
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
		binary, err := assembleFile(assemblerContext)

		println(binary)

		return
	},
}

func assembleFile(ctx *x86_64.Assembler) (string, error) {

	lexer := keurnel_asm.LexerNew(ctx.RawSource())
	lexer.Process()

	parser := keurnel_asm.ParserNew(lexer)
	parser.Parse()

	// Print each group
	for identifier, group := range parser.Groups() {
		println("Group Identifier:", identifier)
		println("Group Type:", group.Type)
		println("Instructions:")
		for _, instr := range group.Instructions {
			println("  Mnemonic:", instr.Mnemonic)
			println("  Operands:")
			for _, operand := range instr.Operands {
				println("    -", operand)
			}
		}
		println()
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
//	// the x86_64 assembler.
//	//
//	assembler := x86_64.New(source)
//	machineCode, err := assembler.Assemble()
//	if err != nil {
//		return "", err
//	}
//
//	// Return the assembled machine code as a string.
//	//
//	return string(machineCode), nil
//}
