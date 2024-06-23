package snapshot

import (
	"context"
	"errors"
	"io"

	livedatav1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/livedata/v1/livedatav1grpc"
	livedatav1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/livedata/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var startFrom string

func NewLiveSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "receives live snapshot data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			liveSnapshot(args[0])
		},
	}

	cmd.Flags().StringVar(&startFrom,
		"start-from", "current", "begin|current")
	return cmd
}

func liveSnapshot(eventArg string) {
	sel := util.ResolveEvent(eventArg)
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer conn.Close()

	req := livedatav1.LiveSnapshotDataRequest{
		Event: sel,
	}
	if startFrom != "" {
		switch startFrom {
		case "begin":
			req.StartFrom = livedatav1.SnapshotStartMode_SNAPSHOT_START_MODE_BEGIN
		default:
			req.StartFrom = livedatav1.SnapshotStartMode_SNAPSHOT_START_MODE_CURRENT
		}
	}
	c := livedatav1grpc.NewLiveDataServiceClient(conn)
	r, err := c.LiveSnapshotData(context.Background(), &req)
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
			log.Error("error fetching live speedmap", log.ErrorField(err))
			break
		} else {
			log.Debug("got snapshot: ",
				log.Time("ts", resp.Timestamp.AsTime()),
				log.Time("recstamp", resp.SnapshotData.RecordStamp.AsTime()))
		}
	}
}
