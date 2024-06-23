//go:build windows
// +build windows

package plugin

import (
	"context"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/server"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func startPlugin(ctx context.Context, plg *server.Plugin, serverAddr string) (*exec.Cmd, error) {
	logsDir := server.LogsDir()
	cmd := exec.CommandContext(ctx, plg.Path(), "--server", serverAddr)

	errLogs, err := os.OpenFile(filepath.Join(logsDir, fmt.Sprintf("%s.err.logs", strings.ReplaceAll(plg.Config.Name, "/", "_"))), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer errLogs.Close()
	cmd.Stderr = errLogs

	outLogs, err := os.OpenFile(filepath.Join(logsDir, fmt.Sprintf("%s.out.logs", strings.ReplaceAll(plg.Config.Name, "/", "_"))), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer outLogs.Close()
	cmd.Stdout = outLogs

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
