package sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
)

type Plugin struct {
	prc     Processor
	rootCmd *cobra.Command
}

func New(prc Processor) *Plugin {
	plg := &Plugin{
		prc: prc,
	}
	plg.rootCmd = &cobra.Command{
		Use:  "plugin",
		RunE: plg.runE,
	}
	return plg
}

func (p *Plugin) runE(cmd *cobra.Command, args []string) error {
	serverFlag := cmd.Flags().Lookup("server")
	if serverFlag == nil || serverFlag.Value.String() == "" {
		return errors.New("server address not provided")
	}

	conn, err := grpc.Dial(serverFlag.Value.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := golang.NewPluginClient(conn)
	stream, err := client.Register(context.Background())
	if err != nil {
		return err
	}
	p.prc.SetStream(stream)

	conf := p.prc.GetConfig()
	err = stream.Send(&golang.PluginMessage{
		PluginMessage: &golang.PluginMessage_Conf{
			Conf: &conf,
		},
	})
	if err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}

		switch {
		case msg.GetReEvaluate() != nil:
			p.prc.ReEvaluate(msg.GetReEvaluate())
		case msg.GetStart() != nil:
			startMsg := msg.GetStart()
			err = p.prc.StartProcess(startMsg.GetCommand(), startMsg.GetFlags(), startMsg.GetKaytuAccessToken())
			if err != nil {
				stream.Send(&golang.PluginMessage{
					PluginMessage: &golang.PluginMessage_Err{
						Err: &golang.Error{
							Error: err.Error(),
						},
					},
				})
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
