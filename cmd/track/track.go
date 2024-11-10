package track

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/track/pit"
)

func NewTrackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Commands regarding tracks. ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(pit.NewPitEditCmd())

	return cmd
}
