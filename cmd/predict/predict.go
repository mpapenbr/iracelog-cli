package predict

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/predict/event"
	"github.com/mpapenbr/iracelog-cli/cmd/predict/live"
	"github.com/mpapenbr/iracelog-cli/cmd/predict/param"
)

func NewPredictCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "predict",
		Short: "Commands regarding predicting race development. ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(event.NewPredictEventCmd())
	cmd.AddCommand(param.NewPredictByParamCmd())
	cmd.AddCommand(live.NewPredictLiveEventCmd())
	return cmd
}
