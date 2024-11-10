package track

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/track/list"
	"github.com/mpapenbr/iracelog-cli/cmd/track/pit"
	"github.com/mpapenbr/iracelog-cli/cmd/track/transfer"
)

func NewTrackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Commands regarding tracks. ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(list.NewTrackListCmd())
	cmd.AddCommand(pit.NewPitEditCmd())
	cmd.AddCommand(transfer.NewTransferDataCmd())

	return cmd
}
