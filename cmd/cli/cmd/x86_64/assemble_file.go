package x86_64

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/keurnel/assembler/internal/lineMap"
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
		//	Debug information for pre-processing and assembly setup.
		//
		//
		//	==============================================================================

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		waitGroup := sync.WaitGroup{}

		debugInformation := kasm.SourceDebugInformationMake(source)
		canListen := debugInformation.CanListen()
		if canListen != nil {
			cmd.PrintErrln("Error: Failed to initialize debug information listener:", canListen)
			return
		}

		waitGroup.Add(1)
		go debugInformation.Listen(ctx, &waitGroup)

		// Example expand of line number 20
		debugInformation.ExpandLine(20, []int{21, 22, 23})
		println(debugInformation.LineNumberToOrigin(23))

		// Publish event o n the channel to expand line number 30 into 3 lines
		debugInformation.ExpansionChannel <- kasm.ExpansionEvent{
			LineNumber:         30,
			ExpandedLinesCount: 3,
		}

		// Send close after 5 seconds
		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()

		waitGroup.Wait()

		// Initialize a lineMap instance to track how lines transform during
		// pre-processing. This allows us to trace any line in the final
		// processed source back to its original location in the source file.
		//
		lineMapSource, err := lineMap.LoadSource(fullPath)
		if err != nil {
			cmd.PrintErrln("Error: Failed to load source file:", err)
			return
		}

		lm, err := lineMap.New(source, lineMapSource)
		if err != nil {
			cmd.PrintErrln("Error: Failed to initialize line map:", err)
			return
		}

		// Step 1: Handle inclusion of other `.kasm` files in the source code.
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

		// Snapshot after includes — lines may have expanded.
		err = lm.Update(source)
		if err != nil {
			cmd.PrintErrln("Error: Line map update failed after include processing:", err)
			return
		}

		// Step 2: Handle macros in the source code.
		//
		macros := kasm.PreProcessingMacroTable(source)
		kasm.PreProcessingColectMacroCalls(source, macros)
		source = kasm.PreProcessingReplaceMacroCalls(source, macros)

		// Snapshot after macro expansion — lines may have expanded or contracted.
		err = lm.Update(source)
		if err != nil {
			cmd.PrintErrln("Error: Line map update failed after macro processing:", err)
			return
		}

		// Step 3: Handle conditional assembly directives in the source code.
		//
		symbolTable := kasm.PreProcessingCreateSymbolTable(source, macros)
		source = kasm.PreProcessingHandleConditionals(source, symbolTable)

		// Snapshot after conditional processing — lines may have been removed.
		err = lm.Update(source)
		if err != nil {
			cmd.PrintErrln("Error: Line map update failed after conditional processing:", err)
			return
		}

		// Print history of line transformations for debugging
		lineHistory := lm.LineHistory(14)
		for i, lineChange := range lineHistory {
			println("Index:", i, lineChange.String())
		}

		return

		//	==============================================================================
		//
		//	Pre-processing of the assembly source code
		//
		//	==============================================================================

		//	==============================================================================
		//
		//	Assembling the source code into machine code
		//
		//	==============================================================================

		// Lexical analysis
		//
		lexer := kasm.LexerNew(source)
		tokens := lexer.Start()

		// Parsing
		//
		parser := kasm.ParserNew(tokens)
		err = parser.Parse()
		if err != nil {
			cmd.PrintErrln("Error: Failed to parse assembly source code:", err)
			return
		}

		// The lineMap instance `lm` can now be used to trace any line in the
		// assembled output back to its original source location:
		//   originalLine := lm.LineOrigin(processedLine)
		_ = lm

		return
	},
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
