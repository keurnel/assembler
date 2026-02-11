package cmd

import (
	"github.com/keurnel/assembler/cmd/cli/cmd/x86_64"
	"github.com/spf13/cobra"
)

var x8664Cmd = &cobra.Command{
	Use:     "x86_64",
	GroupID: "arch",
	Short:   "x86_64 architecture",
	Long:    `Functions related to the x86_64 architecture.`,
}

func init() {
	x8664Cmd.AddGroup(&cobra.Group{
		ID:    "file-operations",
		Title: "File Operations",
	})

	x8664Cmd.AddCommand(x86_64.AssembleFileCmd)
}
