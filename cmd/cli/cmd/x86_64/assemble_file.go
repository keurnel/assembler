package x86_64

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/keurnel/assembler/internal/debugcontext"
	"github.com/keurnel/assembler/internal/lineMap"
	"github.com/keurnel/assembler/v0/architecture"
	"github.com/keurnel/assembler/v0/architecture/x86/_64"
	"github.com/keurnel/assembler/v0/kasm"
	"github.com/keurnel/assembler/v0/kasm/dependency_graph"
	"github.com/keurnel/assembler/v0/kasm/preProcessing"
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
	AssembleFileCmd.Flags().Bool("dependency-graph-dot", false, "Print the dependency graph in Graphviz DOT format and exit")
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

	// FR-11.4.2: When --dependency-graph-dot is set, build the dependency
	// graph, print its DOT representation to stdout, and exit early.
	if dot, _ := cmd.Flags().GetBool("dependency-graph-dot"); dot {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("unable to get working directory: %w", err)
		}
		graph := dependency_graph.New(source, cwd, fullPath)
		fmt.Println(graph.ToDot())
		return nil
	}

	// Create the debug context for this assembly invocation (FR-9.1).
	debugCtx := debugcontext.NewDebugContext(fullPath)

	tracker, err := lineMap.Track(fullPath)
	if err != nil {
		return fmt.Errorf("failed to initialise line tracker: %w", err)
	}

	source = preProcess(source, fullPath, tracker, debugCtx)

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
	semanticErrors := kasm.AnalyserNew(program, instrTable).
		WithDebugContext(debugCtx).
		WithLineMapper(tracker).
		Analyse()

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

	// Code generation phase: encode the validated AST into machine code
	// (FR-9.1, FR-9.2).
	generator := kasm.GeneratorNew(program, instrTable).WithDebugContext(debugCtx)
	output, codegenErrors := generator.Generate()

	// Print debug context entries when verbose mode is enabled (codegen phase).
	if verbose {
		for _, e := range debugCtx.Entries() {
			cmd.PrintErrln(e.String())
		}
	}

	// Abort if code generation recorded any errors (FR-9.3).
	if len(codegenErrors) > 0 {
		if !verbose {
			for _, e := range debugCtx.Errors() {
				cmd.PrintErrln(e.String())
			}
		}
		return fmt.Errorf("assembly aborted: %d error(s) during code generation", len(codegenErrors))
	}

	// FR-9.4: Write the binary output. Default name is the input file with
	// the extension replaced by .bin.
	outputPath := strings.TrimSuffix(fullPath, filepath.Ext(fullPath)) + ".bin"
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
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
func preProcess(source string, rootFilePath string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	source = preProcessIncludes(source, rootFilePath, tracker, debugCtx)
	if debugCtx.HasErrors() {
		return source
	}

	// Write the include-resolved source for debugging (before macro/conditional expansion).
	os.WriteFile("preprocessed.kasm", []byte(source), 0644)

	source = preProcessMacros(source, tracker, debugCtx)
	if debugCtx.HasErrors() {
		return source
	}

	source = preProcessConditionals(source, tracker, debugCtx)
	return source
}

// preProcessIncludes handles %include directives, detects circular inclusions,
// and snapshots the result with source file annotations.
//
// The rootFilePath is added to the seen set before the first invocation of
// PreProcessingHandleIncludes so that a file cannot include itself indirectly
// through a chain that leads back to the root (FR-1.6.6).
func preProcessIncludes(source string, rootFilePath string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	debugCtx.SetPhase("pre-processing/includes")

	cwd, err := os.Getwd()
	if err != nil {
		debugCtx.Error(debugCtx.Loc(0, 0), fmt.Sprintf("unable to get working directory: %v", err))
		return source
	}

	dependencyGraph := dependency_graph.New(source, cwd, rootFilePath)

	// FR-11.4.1: Log the text representation of the dependency graph via
	// debugCtx.Trace so it appears in verbose mode output.
	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("dependency graph:\n%s", dependencyGraph.String()))

	if !dependencyGraph.Acyclic() {
		// FR-11.3.3: Use CyclePath() to enrich the error message with the
		// full chain of files involved in the cycle.
		cyclePath := dependencyGraph.CyclePath()
		if cyclePath != nil {
			debugCtx.Error(debugCtx.Loc(0, 0),
				fmt.Sprintf("circular inclusion detected in dependency graph: %s",
					strings.Join(cyclePath, " → ")))
		} else {
			debugCtx.Error(debugCtx.Loc(0, 0), "circular inclusion detected in dependency graph")
		}
		return source
	}

	// FR-1.6.6: Seed the seen set with the root file path so that any
	// included file that re-includes the root is caught.
	seen := map[string]bool{rootFilePath: true}
	totalInclusions := 0

	// FR-1.7.6 / FR-1.7.7: Identify shared dependencies from the dependency
	// graph and hoist them into a shared-inclusions block at the top of the
	// source. Shared dependencies are files included by more than one parent.
	sharedDeps := dependencyGraph.SharedDependencies()
	var sharedBlock strings.Builder
	sharedBlock.WriteString("; =======================================\n")
	sharedBlock.WriteString("; Begin shared inclusions\n")
	sharedBlock.WriteString("; =======================================\n\n")
	for _, sharedPath := range sharedDeps {
		content := strings.TrimSpace(dependencyGraph.NodeSource(sharedPath))
		sharedBlock.WriteString(fmt.Sprintf("; FILE: %s\n%s\n; END FILE: %s\n",
			sharedPath, content, sharedPath))
		// Pre-seed the seen set so that all %include directives referencing
		// this shared file are silently stripped during normal resolution.
		// Add both the absolute path and the cwd-relative path so that
		// directives using either form are matched.
		seen[sharedPath] = true
		if rel, err := filepath.Rel(cwd, sharedPath); err == nil {
			seen[rel] = true
		}
		totalInclusions++
	}
	sharedBlock.WriteString("\n; =======================================\n")
	sharedBlock.WriteString("; End shared inclusions\n")
	sharedBlock.WriteString("; =======================================\n\n\n")

	source = sharedBlock.String() + source

	// Standard Library block: placeholder for future standard library content.
	stdLibBlock := "; =======================================\n" +
		"; Start Standard Library\n" +
		"; =======================================\n" +
		"; =======================================\n" +
		"; End Standard Library\n" +
		"; =======================================\n"

	source = stdLibBlock + source

	// FR-1.5: Recursively resolve includes. Each iteration inlines one level
	// of %include directives. The loop continues until no new inclusions are
	// found (the source is fully resolved) or a circular inclusion is detected.
	for {
		var inclusions []preProcessing.Inclusion
		source, inclusions = preProcessing.HandleIncludes(source, seen)

		if len(inclusions) == 0 {
			break
		}

		trackerInclusions := make([]lineMap.Inclusion, 0, len(inclusions))

		// FR-1.6.3: Check each included file path against the seen set.
		for _, inc := range inclusions {
			if seen[inc.IncludedFilePath] {
				// FR-1.6.4 / FR-1.6.7: Report circular inclusion error and abort.
				debugCtx.Error(
					debugCtx.Loc(inc.LineNumber, 0),
					fmt.Sprintf("circular inclusion of '%s'", inc.IncludedFilePath),
				)
				return source
			}
		}

		// FR-1.6.5: No circular inclusion detected — add all newly included
		// file paths to the seen set before the next recursive invocation.
		for _, inc := range inclusions {
			seen[inc.IncludedFilePath] = true
			trackerInclusions = append(trackerInclusions, lineMap.Inclusion{
				FilePath:   inc.IncludedFilePath,
				LineNumber: inc.LineNumber,
			})
		}

		tracker.SnapshotWithInclusions(source, trackerInclusions)
		totalInclusions += len(inclusions)
	}

	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("included %d file(s)", totalInclusions))
	return source
}

// preProcessMacros builds the macro table, collects calls, expands them,
// and snapshots the result.
func preProcessMacros(source string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	debugCtx.SetPhase("pre-processing/macros")

	macros := preProcessing.MacroTable(source)
	preProcessing.CollectMacroCalls(source, macros)
	source = preProcessing.ReplaceMacroCalls(source, macros)

	tracker.Snapshot(source)
	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("expanded %d macro(s)", len(macros)))
	return source
}

// preProcessConditionals evaluates %ifdef / %ifndef / %else / %endif blocks,
// and snapshots the result.
func preProcessConditionals(source string, tracker *lineMap.Tracker, debugCtx *debugcontext.DebugContext) string {
	debugCtx.SetPhase("pre-processing/conditionals")

	macros := preProcessing.MacroTable(source)
	symbolTable := preProcessing.CreateSymbolTable(source, macros)
	source = preProcessing.HandleConditionals(source, symbolTable)

	tracker.Snapshot(source)
	debugCtx.Trace(debugCtx.Loc(0, 0), fmt.Sprintf("evaluated conditionals with %d symbol(s)", len(symbolTable)))
	return source
}
