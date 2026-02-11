package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "keurnel-asm",
	Short: "Keurnels assembler",
	Long:  `Keurnels assembler is a tool for assembling code.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddGroup(&cobra.Group{
		ID:    "arch",
		Title: "Architectures",
	})

	rootCmd.AddCommand(x8664Cmd)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
