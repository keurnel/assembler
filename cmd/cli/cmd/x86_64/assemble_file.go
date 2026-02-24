package x86_64

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
		if err := runAssembleFile(cmd, args); err != nil {
			cmd.PrintErrln("Error:", err)
		}
	},
}

// runAssembleFile orchestrates the full assembly pipeline: resolve the file,
// load architecture instructions, run pre-processing, and assemble.
func runAssembleFile(cmd *cobra.Command, args []string) error {
	fullPath, err := resolveFilePath(args)
	if err != nil {
		return err
	}

	loadArchitectureInstructions()

	source, err := readSourceFile(fullPath)
	if err != nil {
		return err
	}

	runDebugInformationDemo(source)

	tracker, err := lineMap.Track(fullPath)
	if err != nil {
		return fmt.Errorf("failed to initialise line tracker: %w", err)
	}

	source = preProcess(source, tracker)

	// Print history of line transformations for debugging.
	for i, change := range tracker.History(0) {
		println("Index:", i, change.String())
	}

	_ = source
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

// runDebugInformationDemo runs the temporary debug information listener demo.
// TODO: remove or integrate properly once the debug pipeline is finalised.
func runDebugInformationDemo(source string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	debugInfo := kasm.SourceDebugInformationMake(source)
	if err := debugInfo.CanListen(); err != nil {
		println("Warning: debug listener unavailable:", err.Error())
		return
	}

	wg.Add(1)
	go debugInfo.Listen(ctx, &wg)

	debugInfo.ExpandLine(20, []int{21, 22, 23})
	println(debugInfo.LineNumberToOrigin(23))

	debugInfo.ExpansionChannel <- kasm.ExpansionEvent{
		LineNumber:         30,
		ExpandedLinesCount: 3,
	}

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	wg.Wait()
}

// preProcess runs the three pre-processing phases (includes, macros,
// conditionals) and snapshots each transformation in the tracker.
func preProcess(source string, tracker *lineMap.Tracker) string {
	source = preProcessIncludes(source, tracker)
	source = preProcessMacros(source, tracker)
	source = preProcessConditionals(source, tracker)
	return source
}

// preProcessIncludes handles %include directives, detects circular inclusions,
// and snapshots the result with source file annotations.
func preProcessIncludes(source string, tracker *lineMap.Tracker) string {
	source, inclusions := kasm.PreProcessingHandleIncludes(source)

	seen := make(map[string]bool, len(inclusions))
	trackerInclusions := make([]lineMap.Inclusion, 0, len(inclusions))
	for _, inc := range inclusions {
		if seen[inc.IncludedFilePath] {
			panic(fmt.Sprintf("pre-processing error: circular inclusion of '%s' at line %d",
				inc.IncludedFilePath, inc.LineNumber))
		}
		seen[inc.IncludedFilePath] = true
		trackerInclusions = append(trackerInclusions, lineMap.Inclusion{
			FilePath:   inc.IncludedFilePath,
			LineNumber: inc.LineNumber,
		})
	}

	tracker.SnapshotWithInclusions(source, trackerInclusions)
	return source
}

// preProcessMacros builds the macro table, collects calls, expands them,
// and snapshots the result.
func preProcessMacros(source string, tracker *lineMap.Tracker) string {
	macros := kasm.PreProcessingMacroTable(source)
	kasm.PreProcessingColectMacroCalls(source, macros)
	source = kasm.PreProcessingReplaceMacroCalls(source, macros)

	tracker.Snapshot(source)
	return source
}

// preProcessConditionals evaluates %ifdef / %ifndef / %else / %endif blocks,
// and snapshots the result.
func preProcessConditionals(source string, tracker *lineMap.Tracker) string {
	macros := kasm.PreProcessingMacroTable(source)
	symbolTable := kasm.PreProcessingCreateSymbolTable(source, macros)
	source = kasm.PreProcessingHandleConditionals(source, symbolTable)

	tracker.Snapshot(source)
	return source
}
