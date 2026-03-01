package x86_64

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/internal/lineMap"
	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/architecture/x86/_64"
	"github.com/keurnel/assembler/v0/kasm"
	"github.com/keurnel/assembler/v0/kasm/dependency_graph"
	"github.com/keurnel/assembler/v0/kasm/profile"
	"github.com/spf13/cobra"
)

var AssembleFileCmd = &cobra.Command{
	Use:     "assemble-file <assembly-file>",
	GroupID: "file-operations",
	Short:   "Assemble an _64 assembly file into a binary file.",
	Long:    `Assemble an _64 assembly file into a binary file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAssembleFile(cmd, args); err != nil {
			cmd.PrintErrln("Error:", err)
		}
	},
}

func init() {
	AssembleFileCmd.Flags().BoolP("verbose", "v", false, "Show debug context logs (trace, info, warning) during assembly")
}

// runAssembleFile orchestrates the full assembly pipeline: resolve the file,
// load architecture instructions, run pre-processing, and assemble.
func runAssembleFile(cmd *cobra.Command, args []string) error {
	fullPath, err := resolveFilePath(args)
	if err != nil {
		return err
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	loadArchitectureInstructions()

	source, err := readSourceFile(fullPath)
	if err != nil {
		return err
	}

	// Create the debug context for this assembly invocation (FR-9.1).
	debugCtx := debugcontext.NewDebugContext(fullPath)

	tracker, err := lineMap.Track(fullPath)
	if err != nil {
		return fmt.Errorf("failed to initialise line tracker: %w", err)
	}

	source = preProcess(source, tracker, debugCtx)

	// Print debug context entries when verbose mode is enabled.
	if verbose {
		for _, e := range debugCtx.Entries() {
			cmd.PrintErrln(e.String())
		}
	}

	// Abort if pre-processing recorded any errors (FR-9.5).
	if debugCtx.HasErrors() {
		if !verbose {
			// Errors were not yet printed; print them now.
			for _, e := range debugCtx.Errors() {
				cmd.PrintErrln(e.String())
			}
		}
		return fmt.Errorf("assembly aborted: %d error(s) during pre-processing", len(debugCtx.Errors()))
	}

	// Lexer phase: tokenise the pre-processed source using the x86_64
	// architecture profile. Because the profile is constructed once and is
	// immutable (FR-1.1.5), it can be reused across invocations.
	archProfile := profile.NewX8664Profile()
	tokens := kasm.LexerNew(source, archProfile).WithDebugContext(debugCtx).Start()

	// Abort if lexer recorded any errors.
	if debugCtx.HasErrors() {
		if !verbose {
			for _, e := range debugCtx.Errors() {
				cmd.PrintErrln(e.String())
			}
		}
		return fmt.Errorf("assembly aborted: %d error(s) during lexing", len(debugCtx.Errors()))
	}

	// Parser phase: transform the token slice into an AST.
	program, parseErrors := kasm.ParserNew(tokens).WithDebugContext(debugCtx).Parse()

	// Print debug context entries when verbose mode is enabled (parser phase).
	if verbose {
		for _, e := range debugCtx.Entries() {
			cmd.PrintErrln(e.String())
		}
	}

	// Abort if parsing recorded any errors.
	if len(parseErrors) > 0 {
		if !verbose {
			for _, e := range debugCtx.Errors() {
				cmd.PrintErrln(e.String())
			}
		}
		return fmt.Errorf("assembly aborted: %d error(s) during parsing", len(parseErrors))
	}

	// Semantic analysis phase: validate the AST against the architecture's
	// instruction metadata. The instruction table is flattened from all
	// architecture groups into a single map keyed by upper-case mnemonic.
	instrTable := buildInstructionTable()
	semanticErrors := kasm.AnalyserNew(program, instrTable).WithDebugContext(debugCtx).Analyse()

	// Print debug context entries when verbose mode is enabled (semantic phase).
	if verbose {
		for _, e := range debugCtx.Entries() {
			cmd.PrintErrln(e.String())
		}
	}

	// Abort if semantic analysis recorded any errors.
	if len(semanticErrors) > 0 {
		if !verbose {
			for _, e := range debugCtx.Errors() {
				cmd.PrintErrln(e.String())
			}
		}
		return fmt.Errorf("assembly aborted: %d error(s) during semantic analysis", len(semanticErrors))
	}

	// Render program
	//
	for _, stmt := range program.Statements {

		switch s := stmt.(type) {
		case *kasm.InstructionStmt:
			fmt.Printf("Instruction: %s\n", s.Mnemonic)
			for i, op := range s.Operands {
				fmt.Printf("  Operand %d: %T\n", i+1, op)
			}
		case *kasm.LabelStmt:
			fmt.Printf("Label: %s\n", s.Name)
		case *kasm.SectionStmt:
			fmt.Printf("Section: %s\n", s.Name)
			fmt.Printf("type: %s\n", s.Type)
		}

	}

	return nil
}

// resolveFilePath validates the CLI arguments and returns the absolute path
// to the assembly file.
func resolveFilePath(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no assembly file provided")
	}
	if args[0] == "" {
		return "", fmt.Errorf("assembly file path is empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current working directory: %w", err)
	}

	fullPath := filepath.Join(cwd, args[0])
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("assembly file does not exist at path: %s", fullPath)
	}

	return fullPath, nil
}

// readSourceFile reads the assembly source file and returns its content.
func readSourceFile(path string) (string, error) {
	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read assembly file: %w", err)
	}
	return string(sourceBytes), nil
}

// loadArchitectureInstructions loads and indexes the x86_64 instruction set.
func loadArchitectureInstructions() map[string]architecture.InstructionGroup {
	groups := make(map[string]architecture.InstructionGroup)
	for groupName, instructions := range _64.Instructions() {
		groups[groupName] = *architecture.FromSlice(groupName, instructions)
	}
	return groups
}

// buildInstructionTable flattens all architecture instruction groups into a
// single map keyed by upper-case mnemonic, suitable for the semantic analyser.
// If two groups contain the same mnemonic, the last one wins (AR-3.2).
func buildInstructionTable() map[string]architecture.Instruction {
	table := make(map[string]architecture.Instruction)
	for _, instructions := range _64.Instructions() {
		for _, instr := range instructions {
			table[instr.Mnemonic] = instr
		}
	}
	return table
}

// preProcess runs the three pre-processing phases (includes, macros,
// conditionals) and snapshots each transformation in the tracker.
// Each phase sets its debug context phase and records errors instead of panicking.
func preProcess(source string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	source = preProcessIncludes(source, tracker, debugCtx)
	if debugCtx.HasErrors() {
		return source
	}

	source = preProcessMacros(source, tracker, debugCtx)
	if debugCtx.HasErrors() {
		return source
	}

	source = preProcessConditionals(source, tracker, debugCtx)
	return source
}

// preProcessIncludes handles %include directives, detects circular inclusions,
// and snapshots the result with source file annotations.
func preProcessIncludes(source string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	debugCtx.SetPhase("pre-processing/includes")

	cwd, err := os.Getwd()
	if err != nil {
		debugCtx.Error(debugCtx.Loc(0, 0), fmt.Sprintf("unable to get working directory: %v", err))
		return source
	}

	dependencyGraph := dependency_graph.New(source, cwd)

	if !dependencyGraph.Acyclic() {
		debugCtx.Error(debugCtx.Loc(0, 0), "circular inclusion detected in dependency graph")
		return source
	}

	println(dependencyGraph.Acyclic())

	source, inclusions := kasm.PreProcessingHandleIncludes(source)

	seen := make(map[string]bool, len(inclusions))
	trackerInclusions := make([]lineMap.Inclusion, 0, len(inclusions))
	for _, inc := range inclusions {
		if seen[inc.IncludedFilePath] {
			debugCtx.Error(
				debugCtx.Loc(inc.LineNumber, 0),
				fmt.Sprintf("circular inclusion of '%s'", inc.IncludedFilePath),
			)
			return source
		}
		seen[inc.IncludedFilePath] = true
		trackerInclusions = append(trackerInclusions, lineMap.Inclusion{
			FilePath:   inc.IncludedFilePath,
			LineNumber: inc.LineNumber,
		})
	}

	tracker.SnapshotWithInclusions(source, trackerInclusions)
	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("included %d file(s)", len(inclusions)))
	return source
}

// preProcessMacros builds the macro table, collects calls, expands them,
// and snapshots the result.
func preProcessMacros(source string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	debugCtx.SetPhase("pre-processing/macros")

	macros := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingCollectMacroCalls(source, macros)
	source = kasm.PreProcessingReplaceMacroCalls(source, macros)

	tracker.Snapshot(source)
	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("expanded %d macro(s)", len(macros)))
	return source
}

// preProcessConditionals evaluates %ifdef / %ifndef / %else / %endif blocks,
// and snapshots the result.
func preProcessConditionals(source string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	debugCtx.SetPhase("pre-processing/conditionals")

	macros := kasm.PreProcessingMacroTable(source)
	symbolTable := kasm.PreProcessingCreateSymbolTable(source, macros)
	source = kasm.PreProcessingHandleConditionals(source, symbolTable)

	tracker.Snapshot(source)
	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("evaluated conditionals with %d symbol(s)", len(symbolTable)))
	return source
}
