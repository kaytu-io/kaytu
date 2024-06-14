package plugin

import (
	"context"
	"errors"
	"fmt"
	githubAPI "github.com/google/go-github/v62/github"
	"github.com/kaytu-io/kaytu/controller"
	"github.com/kaytu-io/kaytu/pkg/version"
	"github.com/rogpeppe/go-internal/semver"
	"github.com/schollz/progressbar/v3"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"github.com/kaytu-io/kaytu/pkg/server"
	"github.com/kaytu-io/kaytu/view"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

var (
	patternVersionRegex, _ = regexp.Compile(fmt.Sprintf("plugin_([a-z0-9\\.]+)_%s_%s", runtime.GOOS, runtime.GOARCH))
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

	jobs                      *controller.Jobs
	optimizations             *controller.Optimizations[golang.OptimizationItem]
	pluginCustomOptimizations *controller.Optimizations[golang.ChartOptimizationItem]

	overviewPage *view.PluginCustomOverviewPage
	detailsPage  *view.PluginCustomResourceDetailsPage

	NonInteractiveView *view.NonInteractiveView
	lis                net.Listener
	grpcServer         *grpc.Server
}

func New() *Manager {
	return &Manager{
		port:    0,
		started: false,
	}
}

func (m *Manager) SetListenPort(port int) {
	m.port = port
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
				_, err := startPlugin(plg, fmt.Sprintf("localhost:%d", m.port))
				return err
			}
		}
	}

	return errors.New("plugin not found")
}

func (m *Manager) StartServer() error {
	var err error

	m.lis, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", m.port))
	if err != nil {
		return err
	}

	m.port = m.lis.Addr().(*net.TCPAddr).Port

	m.grpcServer = grpc.NewServer()
	golang.RegisterPluginServer(m.grpcServer, m)
	go func() {
		err = m.grpcServer.Serve(m.lis)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (m *Manager) StopServer() error {
	m.grpcServer.Stop()
	return m.lis.Close()
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
				m.NonInteractiveView.Optimizations.SendItem(receivedMsg.GetOi())
			case receivedMsg.GetCoi() != nil:
				m.NonInteractiveView.PluginCustomOptimizations.SendItem(receivedMsg.GetCoi())
			case receivedMsg.GetJob() != nil:
				m.NonInteractiveView.PublishJobs(receivedMsg.GetJob())
			case receivedMsg.GetErr() != nil:
				m.NonInteractiveView.PublishError(fmt.Errorf(receivedMsg.GetErr().Error))
			case receivedMsg.GetReady() != nil:
				m.NonInteractiveView.PublishResultsReady(receivedMsg.GetReady())
			case receivedMsg.GetNonInteractive() != nil:
				m.NonInteractiveView.PublishNonInteractiveExport(receivedMsg.GetNonInteractive())
			case receivedMsg.GetUpdateChart() != nil:
				if m.NonInteractiveView == nil {
					return errors.New("custom optimizations controller not set - is plugin running in default ui mode?")
				}
				updateChart := receivedMsg.GetUpdateChart()
				if updateChart.GetOverviewChart() != nil {
					m.NonInteractiveView.SetChartDefinition(updateChart.GetOverviewChart())
				}
				if updateChart.GetDevicesChart() != nil {
					m.NonInteractiveView.SetDevicesChartDefinition(updateChart.GetDevicesChart())
				}
			}
		}
	} else {
		for {
			receivedMsg, err := stream.Recv()
			if err != nil {
				if m.jobs != nil {
					m.jobs.PublishError(err)
				}
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
				if m.optimizations == nil {
					return errors.New("default optimizations controller not set - is plugin running in custom ui mode?")
				}
				m.optimizations.SendItem(receivedMsg.GetOi())

			case receivedMsg.GetCoi() != nil:
				if m.pluginCustomOptimizations == nil {
					return errors.New("custom optimizations controller not set - is plugin running in default ui mode?")
				}
				m.pluginCustomOptimizations.SendItem(receivedMsg.GetCoi())

			case receivedMsg.GetUpdateChart() != nil:
				if m.pluginCustomOptimizations == nil {
					return errors.New("custom optimizations controller not set - is plugin running in default ui mode?")
				}
				updateChart := receivedMsg.GetUpdateChart()
				if updateChart.GetOverviewChart() != nil && m.overviewPage != nil {
					m.overviewPage.SetChartDefinition(updateChart.GetOverviewChart())
				}
				if updateChart.GetDevicesChart() != nil && m.detailsPage != nil {
					m.detailsPage.SetChartDefinition(updateChart.GetDevicesChart())
				}

			case receivedMsg.GetErr() != nil:
				if m.jobs != nil {
					m.jobs.PublishError(fmt.Errorf(receivedMsg.GetErr().Error))
				}

			case receivedMsg.GetSummary() != nil:
				if m.pluginCustomOptimizations == nil {
					return errors.New("default optimizations controller not set - is plugin running in custom ui mode?")
				}
				m.pluginCustomOptimizations.SetResultSummary(receivedMsg.GetSummary().Message)
			}
		}
	}
}

func (m *Manager) Install(addr, token string, unsafe, pluginDebugMode bool) error {
	cfg, err := server.GetConfig()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(addr, "github.com") {
		addr = fmt.Sprintf("github.com/kaytu-io/plugin-%s", addr)
	}
	addr = strings.TrimPrefix(addr, "github.com/")
	owner, repository, _ := strings.Cut(addr, "/")

	var tc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(context.Background(), ts)
	}
	approved, err := m.isPluginApproved(tc, addr)
	if err != nil {
		return err
	}

	if !approved && !unsafe {
		return fmt.Errorf("plugin not approved. either use --unsafe or make a pull request on github.com/kaytu-io/kaytu to approve your plugin")
	}

	api := githubAPI.NewClient(tc)
	plugins := map[string]*server.Plugin{}
	for _, plg := range cfg.Plugins {
		plugins[plg.Config.Name] = plg
	}

	var release *githubAPI.RepositoryRelease
	if !pluginDebugMode {
		release, _, err = api.Repositories.GetLatestRelease(context.Background(), owner, repository)
		if err != nil {
			return err
		}
	} else {
		release = &githubAPI.RepositoryRelease{}
		installed := false
		for i := 0; i < 30; i++ {
			for _, runningPlugin := range m.plugins {
				if runningPlugin.Plugin.Config.Name == addr {
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

		plugins[addr] = &m.GetPlugin(addr).Plugin
	}

	for _, asset := range release.Assets {
		if asset.ID != nil && asset.Name != nil && patternVersionRegex.MatchString(*asset.Name) {
			assetVersion := strings.Split(*asset.Name, "_")[1]
			if p, ok := plugins[addr]; ok && p.Config.Version == assetVersion {
				return nil
			}
			os.Stderr.WriteString(fmt.Sprintf("Installing plugin %s, version %s\n", addr, assetVersion))
			os.Stderr.WriteString("Downloading the plugin...")

			rc, url, err := api.Repositories.DownloadReleaseAsset(context.Background(), owner, repository, *asset.ID, nil)
			if err != nil {
				return err
			}

			if len(url) > 0 {
				resp, err := http.Get(url)
				if err != nil {
					return err
				}

				defer resp.Body.Close()

				rc = resp.Body
			}

			os.MkdirAll(server.PluginDir(), os.ModePerm)

			pluginExt := filepath.Ext(*asset.Name)
			if runtime.GOOS != "windows" {
				pluginExt = ""
			}
			f, err := os.OpenFile(filepath.Join(server.PluginDir(), strings.ReplaceAll(addr, "/", "_")+pluginExt), os.O_CREATE|os.O_RDWR, os.ModePerm)
			if err != nil {
				return err
			}

			bar := progressbar.DefaultBytes(int64(asset.GetSize()))
			_, err = io.Copy(io.MultiWriter(f, bar), rc)
			if err != nil {
				return err
			}

			err = f.Close()
			if err != nil {
				return err
			}

			plugin := server.Plugin{
				Config: &golang.RegisterConfig{
					Name:     addr,
					Version:  "",
					Provider: "",
					Commands: nil,
				},
			}
			os.Stderr.WriteString("Starting the plugin...")
			runningCmd, err := startPlugin(&plugin, fmt.Sprintf("localhost:%d", m.port))
			if err != nil {
				return err
			}
			defer runningCmd.Process.Kill()
			defer func() {
				m.plugins = nil
			}()

			os.Stderr.WriteString("Waiting for plugin to load...")
			installed := false
			for i := 0; i < 30; i++ {
				for _, runningPlugin := range m.plugins {
					if runningPlugin.Plugin.Config.Name == addr {
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

			plugins[addr] = &m.GetPlugin(addr).Plugin

			if semver.Compare("v"+version.VERSION, plugins[addr].Config.MinKaytuVersion) == -1 {
				return fmt.Errorf("plugin requires kaytu version %s, please update your Kaytu CLI", plugins[addr].Config.MinKaytuVersion)
			}
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

func (m *Manager) SetDefaultUI(jobs *controller.Jobs, optimizations *controller.Optimizations[golang.OptimizationItem]) {
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

func (m *Manager) SetCustomUI(jobs *controller.Jobs, optimizations *controller.Optimizations[golang.ChartOptimizationItem],
	overviewPage *view.PluginCustomOverviewPage, detailsPage *view.PluginCustomResourceDetailsPage) {
	m.jobs = jobs
	m.pluginCustomOptimizations = optimizations
	m.overviewPage = overviewPage
	m.detailsPage = detailsPage

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

func (m *Manager) SetNonInteractiveView() {
	m.NonInteractiveView = view.NewNonInteractiveView()
}

func (m *Manager) isPluginApproved(tc *http.Client, pluginName string) (bool, error) {
	if pluginName == "kaytu-io/plugin-aws" {
		return true, nil
	}
	api := githubAPI.NewClient(tc)
	fileContent, _, resp, err := api.Repositories.GetContents(context.Background(), "kaytu-io", "kaytu", "approved_plugins", nil)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return false, err
	}
	plugins := strings.Split(content, "\n")
	for _, plugin := range plugins {
		if plugin == pluginName {
			return true, nil
		}
	}
	return false, nil
}

func (m *Manager) Uninstall(pluginName string) error {
	fmt.Println(fmt.Sprintf("Uninstalling plugin %s", pluginName))
	cfg, err := server.GetConfig()
	if err != nil {
		return err
	}

	plugins := map[string]*server.Plugin{}
	installed := false
	for _, plg := range cfg.Plugins {
		if pluginName == plg.Config.Name {
			installed = true
			continue
		}
		plugins[plg.Config.Name] = plg
	}
	if !installed {
		return fmt.Errorf("plugin not found")
	}

	pluginFile := filepath.Join(server.PluginDir(), strings.ReplaceAll(pluginName, "/", "_"))

	err = os.Remove(pluginFile)
	if err != nil {
		return err
	}

	cfg.Plugins = nil
	for _, v := range plugins {
		cfg.Plugins = append(cfg.Plugins, v)
	}
	err = server.SetConfig(*cfg)
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Plugin %s uninstalled", pluginName))
	return nil
}
