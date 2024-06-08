package analysis

import (
	"context"
	"errors"
	"io"

	livedatav1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/livedata/v1/livedatav1grpc"
	analysisv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/analysis/v1"
	livedatav1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/livedata/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewLiveAnalysisSelectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "selector",
		Short: "receives live analysis data (using selector)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			liveAnalysisDataWithSelector(args[0])
		},
	}
	cmd.Flags().StringSliceVarP(&config.DefaultCliArgs().Components,
		"components", "c", []string{"caroccupancies", "raceorder"},
		"requested components (comma separated)")
	cmd.Flags().IntVar(&tailNum,
		"tail", 2,
		"request tail entries for components carlaps, racegraph")
	return cmd
}

var tailNum int

func liveAnalysisDataWithSelector(eventArg string) {
	eventSel := util.ResolveEvent(eventArg)
	sel := resolveAnalysisSelector()

	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer conn.Close()
	req := livedatav1.LiveAnalysisSelRequest{
		Event:    eventSel,
		Selector: sel,
	}
	c := livedatav1grpc.NewLiveDataServiceClient(conn)
	r, err := c.LiveAnalysisSel(context.Background(), &req)
	if err != nil {
		log.Error("could not get live data", log.ErrorField(err))
		return
	}

	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error("error fetching live state", log.ErrorField(err))
			return
		} else {
			resolveOutput(resp, sel)
		}
	}
}

//nolint:whitespace // linter+editor conflict
func resolveOutput(resp *livedatav1.LiveAnalysisSelResponse,
	sel *livedatav1.AnalysisSelector,
) {
	for _, comp := range sel.Components {
		switch comp {
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_OCCUPANCIES:
			log.Debug("carinfos", log.Int("size", len(resp.CarOccupancies)))
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_PITS:
			log.Debug("carpits", log.Int("size", len(resp.CarPits)))
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_STINTS:
			log.Debug("carstints", log.Int("size", len(resp.CarStints)))
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_LAPS:
			log.Debug("carlaps", log.Int("size", len(resp.CarLaps)))
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_COMPUTE_STATES:
			log.Debug("computeStates", log.Int("size", len(resp.CarComputeStates)))
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_RACE_ORDER:
			log.Debug("raceorder", log.Int("size", len(resp.RaceOrder)))
		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_RACE_GRAPH:
			racegraphOutput(resp.RaceGraph)

		case livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_UNSPECIFIED:
			// do nothing
		}
	}
}

//nolint:lll // better readability
func resolveAnalysisSelector() *livedatav1.AnalysisSelector {
	selector := &livedatav1.AnalysisSelector{}
	selComps := []livedatav1.AnalysisComponent{}
	for _, comp := range config.DefaultCliArgs().Components {
		switch comp {
		case "caroccupancies":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_OCCUPANCIES)
		case "carpits":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_PITS)
		case "carstints":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_STINTS)
		case "carlaps":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_LAPS)
		case "carcomputestates":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_COMPUTE_STATES)
		case "raceorder":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_RACE_ORDER)
		case "racegraph":
			selComps = append(selComps, livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_RACE_GRAPH)
		default:
			log.Warn("ignoring unknown component in selector: ",
				log.String("component", comp))
		}
	}
	selector.Components = selComps
	selector.CarLapsNumTail = uint32(tailNum)
	selector.RaceGraphNumTail = uint32(tailNum)
	return selector
}

func racegraphOutput(resp []*analysisv1.RaceGraph) {
	m := toMap(resp)
	for k, v := range m {
		log.Debug("RaceGraph",
			log.String("CarClass", k),
			log.Int("size", len(v)),
			log.Int("lapNo", int(v[0].LapNo)),
			log.Int("gaps", len(v[0].Gaps)))
	}
}

func toMap(in []*analysisv1.RaceGraph) map[string][]*analysisv1.RaceGraph {
	ret := make(map[string][]*analysisv1.RaceGraph, 0)
	for _, r := range in {
		if _, ok := ret[r.CarClass]; !ok {
			ret[r.CarClass] = make([]*analysisv1.RaceGraph, 0)
		}
		ret[r.CarClass] = append(ret[r.CarClass], r)
	}
	return ret
}
