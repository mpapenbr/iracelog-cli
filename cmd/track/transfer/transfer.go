package transfer

import (
	"context"
	"errors"
	"io"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/track/v1/trackv1grpc"
	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	sourceAddr     string
	sourceInsecure bool
	dryRun         bool
)

func NewTransferDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfers track data from another iRacelog instance",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			runTransfer()
		},
	}
	cmd.Flags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	cmd.Flags().StringVar(&sourceAddr,
		"source-addr",
		"",
		"gRPC server address of the source iRacelog instance")
	//nolint:errcheck // by design
	cmd.MarkFlagRequired("source-addr")
	cmd.Flags().BoolVar(&sourceInsecure,
		"source-insecure",
		false,
		"connect gRPC address without TLS (development only)")
	cmd.Flags().BoolVar(&dryRun,
		"dry-run",
		false,
		"just check, do not transfer data")

	return cmd
}

func runTransfer() {
	log.Info("connect source server", log.String("addr", sourceAddr))
	source, err := util.NewClient(
		sourceAddr,
		util.WithTLSEnabled(!sourceInsecure))
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer source.Close()

	log.Info("connect dest server", log.String("addr", config.DefaultCliArgs().Addr))
	dest, err := util.NewClient(
		config.DefaultCliArgs().Addr,
		util.WithCliArgs(config.DefaultCliArgs()))
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer dest.Close()

	transferData := transferData{
		source: source,
		dest:   dest,
		dryRun: dryRun,
	}
	transferData.transfer()
}

type transferData struct {
	source *grpc.ClientConn
	dest   *grpc.ClientConn
	dryRun bool
}

func (t *transferData) transfer() {
	log.Info("transfer track data")
	sourceTracks := t.readTracks(t.source)
	destTracks := t.readTracks(t.dest)
	log.Info("got tracks",
		log.Int("source", len(sourceTracks)),
		log.Int("dest", len(destTracks)))

	destLookup := t.createLookup(destTracks)
	trackService := trackv1grpc.NewTrackServiceClient(t.dest)
	for _, st := range sourceTracks {
		//nolint:nestif	 // false positive
		if _, ok := destLookup[st.Id]; !ok {
			log.Debug("transfer track",
				log.Uint32("id", st.Id),
				log.String("name", st.Name),
				log.String("config", st.Config),
				log.Float32("pitEntry", st.PitInfo.Entry),
				log.Float32("pitExit", st.PitInfo.Exit),
				log.Float32("pitLane", st.PitInfo.LaneLength),
			)
			if st.PitInfo.LaneLength == 0 {
				log.Info("track has no pit lane length, skipping",
					log.Uint32("id", st.Id),
					log.String("name", st.Name))
				continue
			}
			if !t.dryRun {
				md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
				ctx := metadata.NewOutgoingContext(context.Background(), md)
				req := trackv1.EnsureTrackRequest{
					Track: st,
				}
				log.Info("transferring track",
					log.Uint32("id", st.Id),
					log.String("name", st.Name))
				if _, err := trackService.EnsureTrack(ctx, &req); err != nil {
					log.Error("could not transfer track", log.ErrorField(err))
				}
			}
		}
	}
}

func (t *transferData) createLookup(tracks []*trackv1.Track) map[uint32]*trackv1.Track {
	ret := map[uint32]*trackv1.Track{}
	for _, t := range tracks {
		ret[t.Id] = t
	}
	return ret
}

func (t *transferData) readTracks(conn *grpc.ClientConn) []*trackv1.Track {
	log.Debug("read tracks")
	req := trackv1.GetTracksRequest{}
	c := trackv1grpc.NewTrackServiceClient(conn)
	reqCtx := context.Background()

	r, err := c.GetTracks(reqCtx, &req)
	if err != nil {
		log.Error("could not get tracks", log.ErrorField(err))
		return []*trackv1.Track{}
	}
	ret := []*trackv1.Track{}
	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error("error fetching track", log.ErrorField(err))
			break
		} else {
			ret = append(ret, resp.Track)
		}
	}
	return ret
}
