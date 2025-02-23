package demo

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/demo/replayloop"
)

func NewDemoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Collection of commands for demo purposes.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(replayloop.NewReplayLoopCmd())
	return cmd
}
