package pit

import (
	"context"
	"strconv"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/track/v1/trackv1grpc"
	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	editPitEntry float32
	editPitExit  float32
)

func NewPitEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pit",
		Short: "edit pit data.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			editPitData(args[0])
		},
	}
	cmd.Flags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")

	cmd.Flags().Float32Var(&editPitEntry,
		"pit-entry", -1, "track pos (float range 0-1) of pit entry")

	cmd.Flags().Float32Var(&editPitExit,
		"pit-exit", -1, "track pos (float range 0-1) of pit exit")

	cmd.MarkFlagsRequiredTogether("pit-entry", "pit-exit", "token")
	return cmd
}

//nolint:funlen // by design
func editPitData(arg string) {
	log.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	var trackId int
	if trackId, err = strconv.Atoi(arg); err != nil {
		log.Error("could not convert track id", log.ErrorField(err), log.String("track", arg))
		return
	}
	validate := func(val float32) bool {
		return val >= 0 && val <= 1
	}
	if !validate(editPitEntry) || !validate(editPitExit) {
		log.Error("pit values must be in range 0-1",
			log.Float32("entry", editPitEntry),
			log.Float32("exit", editPitExit))
		return
	}

	t := trackv1grpc.NewTrackServiceClient(conn)
	req := trackv1.GetTrackRequest{
		Id: uint32(trackId),
	}
	var trackResp *trackv1.GetTrackResponse
	trackResp, err = t.GetTrack(context.Background(), &req)
	if err != nil {
		log.Error("could not get track", log.ErrorField(err), log.Int("trackId", trackId))
		return
	}

	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	pitLaneLength := func(entry, exit float32) float32 {
		if exit > entry {
			return (exit - entry) * trackResp.Track.Length
		} else {
			return (1.0 - entry + exit) * trackResp.Track.Length
		}
	}
	updateReq := trackv1.UpdatePitInfoRequest{
		Id: uint32(trackId),
		PitInfo: &trackv1.PitInfo{
			Entry:      editPitEntry,
			Exit:       editPitExit,
			LaneLength: pitLaneLength(editPitEntry, editPitExit),
		},
	}
	if _, err := t.UpdatePitInfo(ctx, &updateReq); err != nil {
		log.Error("could not update track", log.ErrorField(err), log.Int("trackId", trackId))
		return
	}
	log.Info("PitInfo updated.")
}
