package cmd

import "github.com/spf13/cobra"

var x8664Cmd = &cobra.Command{
	Use:     "x86_64",
	GroupID: "arch",
	Short:   "x86_64 architecture",
	Long:    `Functions related to the x86_64 architecture.`,
}

func init() {

}
