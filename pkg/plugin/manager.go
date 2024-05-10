package plugin

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kaytu-io/kaytu/pkg/api/github"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/view"
	"github.com/schollz/progressbar/v3"
	"google.golang.org/grpc"
)

type RunningPlugin struct {
	Plugin server.Plugin
	Stream golang.Plugin_RegisterServer
}

type Manager struct {
	port    int
	started bool
	plugins []RunningPlugin
	stream  golang.Plugin_RegisterServer

	golang.PluginServer

	jobs          *view.JobsView
	optimizations *view.OptimizationsView

	NonInteractiveView *view.NonInteractiveView
}

func New() *Manager {
	return &Manager{
		port:    0,
		started: false,
	}
}

func (m *Manager) GetPlugin(name string) *RunningPlugin {
	for _, plg := range m.plugins {
		if plg.Plugin.Config.Name == name {
			return &plg
		}
	}
	return nil
}

func (m *Manager) StartPlugin(cmd string) error {
	plugins, err := server.GetPlugins()
	if err != nil {
		return err
	}

	for _, plg := range plugins {
		for _, c := range plg.Config.Commands {
			if cmd == c.Name {
				return startPlugin(plg, fmt.Sprintf("localhost:%d", m.port))
			}
		}
	}

	return errors.New("plugin not found")
}

func (m *Manager) StartServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", m.port))
	if err != nil {
		return err
	}

	m.port = lis.Addr().(*net.TCPAddr).Port

	grpcServer := grpc.NewServer()
	golang.RegisterPluginServer(grpcServer, m)
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (m *Manager) Register(stream golang.Plugin_RegisterServer) error {
	m.stream = stream
	if m.NonInteractiveView != nil {
		for {
			receivedMsg, err := stream.Recv()
			if err != nil {
				return err
			}

			switch {
			case receivedMsg.GetConf() != nil:
				conf := receivedMsg.GetConf()
				m.plugins = append(m.plugins, RunningPlugin{
					Plugin: server.Plugin{Config: conf},
					Stream: stream,
				})
			case receivedMsg.GetOi() != nil:
				m.NonInteractiveView.PublishItem(receivedMsg.GetOi())
			case receivedMsg.GetJob() != nil:
				m.NonInteractiveView.PublishJobs(receivedMsg.GetJob())
			case receivedMsg.GetErr() != nil:
				m.NonInteractiveView.PublishError(fmt.Errorf(receivedMsg.GetErr().Error))
			case receivedMsg.GetReady() != nil:
				m.NonInteractiveView.PublishResultsReady(receivedMsg.GetReady())
			}
		}
	} else {
		for {
			receivedMsg, err := stream.Recv()
			if err != nil {
				return err
			}

			switch {
			case receivedMsg.GetConf() != nil:
				conf := receivedMsg.GetConf()
				m.plugins = append(m.plugins, RunningPlugin{
					Plugin: server.Plugin{Config: conf},
					Stream: stream,
				})

			case receivedMsg.GetJob() != nil:
				m.jobs.Publish(receivedMsg.GetJob())

			case receivedMsg.GetOi() != nil:
				m.optimizations.SendItem(receivedMsg.GetOi())

			case receivedMsg.GetErr() != nil:
				m.jobs.PublishError(fmt.Errorf(receivedMsg.GetErr().Error))
			}
		}
	}
}

func (m *Manager) Install(addr string) error {
	name := addr
	cfg, err := server.GetConfig()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(addr, "github.com") {
		addr = fmt.Sprintf("github.com/kaytu-io/plugin-%s", addr)
	}

	addr = strings.TrimPrefix(addr, "github.com/")
	if strings.HasPrefix(name, "github.com") {
		name = strings.Split(addr, "/")[1]
	}

	fmt.Println("Installing plugin", addr)
	release, err := github.GetLatestRelease(addr)
	if err != nil {
		return err
	}
	fmt.Println("Latest release is", release.TagName)
	plugins := map[string]*server.Plugin{}
	for _, plg := range cfg.Plugins {
		plugins[plg.Config.Name] = plg
	}

	for _, asset := range release.Assets {
		pattern := fmt.Sprintf("plugin_([a-z0-9\\.]+)_%s_%s", runtime.GOOS, runtime.GOARCH)
		r, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		if r.MatchString(asset.Name) {
			version := strings.Split(asset.Name, "_")[1]
			if p, ok := plugins[name]; ok && p.Config.Version == version {
				fmt.Println("Plugin already exists")
				return nil
			}
			fmt.Println("Downloading the plugin...")

			resp, err := http.Get(asset.BrowserDownloadUrl)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			os.MkdirAll(server.PluginDir(), os.ModePerm)

			pluginExt := filepath.Ext(asset.Name)

			f, err := os.OpenFile(filepath.Join(server.PluginDir(), name+pluginExt), os.O_CREATE|os.O_RDWR, os.ModePerm)
			if err != nil {
				return err
			}

			bar := progressbar.DefaultBytes(resp.ContentLength)
			_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
			if err != nil {
				return err
			}

			err = f.Close()
			if err != nil {
				return err
			}

			plugin := server.Plugin{
				Config: &golang.RegisterConfig{
					Name:     name,
					Version:  "",
					Provider: "",
					Commands: nil,
				},
			}
			fmt.Println("Starting the plugin...")
			err = startPlugin(&plugin, fmt.Sprintf("localhost:%d", m.port))
			if err != nil {
				return err
			}

			fmt.Println("Waiting for plugin to load...")
			installed := false
			for i := 0; i < 30; i++ {
				for _, runningPlugin := range m.plugins {
					if runningPlugin.Plugin.Config.Name == name {
						installed = true
					}
				}

				if installed {
					break
				}
				time.Sleep(time.Second)
			}

			if !installed {
				return errors.New("plugin install timeout")
			}

			plugins[name] = &m.GetPlugin(name).Plugin
			break
		}
	}

	cfg.Plugins = nil
	for _, v := range plugins {
		cfg.Plugins = append(cfg.Plugins, v)
	}
	err = server.SetConfig(*cfg)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) SetUI(jobs *view.JobsView, optimizations *view.OptimizationsView) {
	m.jobs = jobs
	m.optimizations = optimizations

	optimizations.SetReEvaluateFunc(func(id string, items []*golang.PreferenceItem) {
		m.stream.Send(&golang.ServerMessage{
			ServerMessage: &golang.ServerMessage_ReEvaluate{
				ReEvaluate: &golang.ReEvaluate{
					Id:          id,
					Preferences: items,
				},
			},
		})
	})
}

func (m *Manager) SetJobsView(jobs *view.JobsView) {
	m.jobs = jobs
}

func (m *Manager) SetNonInteractiveView() {
	m.NonInteractiveView = view.NewNonInteractiveView()
}
