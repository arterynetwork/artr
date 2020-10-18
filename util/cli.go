package util

import "github.com/spf13/cobra"

func LineBreak() *cobra.Command {
	return &cobra.Command{Run: func(*cobra.Command, []string) {}}
}
