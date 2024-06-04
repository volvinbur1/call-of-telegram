package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/volvinbur1/call-of-telegram/internal/cli/broadcast"
)

func NewCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "github.com/volvinbur1/call-of-telegram",
		Short: "github.com/volvinbur1/call-of-telegram broadcasts messages in telegram",
		Long:  "github.com/volvinbur1/call-of-telegram is a cli tool which can send provided message to users of specified community (group/channel) in telegram.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				panic("CLI RootCmd help failed: " + err.Error())
			}
		},
	}

	rootCmd.AddCommand(broadcast.NewCmd())

	return rootCmd
}

func Execute(rootCmd *cobra.Command) error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("CLI RootCmd execution failed: %s", err)
	}

	return nil
}
