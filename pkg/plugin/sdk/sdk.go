package sdk

import (
	"errors"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"
)

type Plugin struct {
	jobMaxConcurrent int
	prc              Processor
	rootCmd          *cobra.Command
}

func New(prc Processor, jobMaxConcurrent int) *Plugin {
	plg := &Plugin{
		prc:              prc,
		jobMaxConcurrent: jobMaxConcurrent,
	}
	plg.rootCmd = &cobra.Command{
		Use:  "plugin",
		RunE: plg.runE,
	}
	return plg
}

func (p *Plugin) runE(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	serverFlag := cmd.Flags().Lookup("server")
	if serverFlag == nil || serverFlag.Value.String() == "" {
		return errors.New("server address not provided")
	}

	conn, err := grpc.NewClient(serverFlag.Value.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := golang.NewPluginClient(conn)
	rawStream, err := client.Register(ctx)
	if err != nil {
		return err
	}

	stream := NewStreamController(rawStream)
	stream.Start()

	p.prc.SetStream(ctx, stream)
	conf := p.prc.GetConfig(ctx)
	stream.Send(&golang.PluginMessage{
		PluginMessage: &golang.PluginMessage_Conf{
			Conf: &conf,
		},
	})
	jobQueue := NewJobQueue(p.jobMaxConcurrent, stream)
	jobQueue.Start(ctx)

	for {
		if ctx.Err() != nil {
			log.Printf("context error: %v", ctx.Err())
			return nil
		}

		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch {
		case msg.GetReEvaluate() != nil:
			p.prc.ReEvaluate(ctx, msg.GetReEvaluate())
		case msg.GetStart() != nil:
			startMsg := msg.GetStart()
			err = p.prc.StartProcess(ctx, startMsg.GetCommand(), startMsg.GetFlags(), startMsg.GetKaytuAccessToken(), jobQueue)
			if err != nil {
				stream.Send(&golang.PluginMessage{
					PluginMessage: &golang.PluginMessage_Err{
						Err: &golang.Error{
							Error: err.Error(),
						},
					},
				})
				stream.CloseSend()
				time.Sleep(time.Second)
				return err
			}
		}
	}
}

func (p *Plugin) Execute() {
	p.rootCmd.Flags().String("server", "", "")

	err := p.rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
