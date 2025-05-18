package analysis

import (
	"context"
	"strings"

	analysisv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/analysis/v1/analysisv1grpc"
	analysisv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/analysis/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	persist   bool
	component string
)

func NewEventAnalysisComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "recompute analysis data of selected event",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			computeEvent(args[0])
		},
	}
	cmd.Flags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	cmd.Flags().StringVar(&component,
		"component", "all", "component to recompute (all, racegraph)")

	cmd.Flags().BoolVar(&persist,
		"persist", true, "persist recomputed data")

	return cmd
}

func computeEvent(arg string) {
	log.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	persistMode := analysisv1.AnalysisPersistMode_ANALYSIS_PERSIST_MODE_OFF
	if persist {
		persistMode = analysisv1.AnalysisPersistMode_ANALYSIS_PERSIST_MODE_ON
	}
	comp := parseComponent(component)
	req := analysisv1.ComputeAnalysisRequest{
		EventSelector: util.ResolveEvent(arg),
		Component:     comp,
		PersistMode:   persistMode,
	}

	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := analysisv1grpc.NewAnalysisServiceClient(conn)
	if _, err := c.ComputeAnalysis(ctx, &req); err != nil {
		log.Error("error computing analysis for event ",
			log.ErrorField(err), log.String("event", arg))
		return
	}
	log.Info("Event analysis recomputed.")
}

func parseComponent(component string) analysisv1.AnalysisComponent {
	switch strings.ToLower(component) {
	case "all":
		return analysisv1.AnalysisComponent_ANALYSIS_COMPONENT_ALL
	case "racegraph":
		return analysisv1.AnalysisComponent_ANALYSIS_COMPONENT_RACEGRAPH
	default:
		return analysisv1.AnalysisComponent_ANALYSIS_COMPONENT_UNSPECIFIED
	}
}
