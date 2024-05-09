package plugin

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/server"
	"os"
	"os/exec"
	"path"
)

func startPlugin(plg *server.Plugin, serverAddr string) error {
	logsDir := server.LogsDir()
	cmd := exec.Command(plg.Path(), "--server", serverAddr)

	errLogs, err := os.OpenFile(path.Join(logsDir, fmt.Sprintf("%s.err.logs", plg.Config.Name)), os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	cmd.Stderr = errLogs

	outLogs, err := os.OpenFile(path.Join(logsDir, fmt.Sprintf("%s.out.logs", plg.Config.Name)), os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	cmd.Stdout = outLogs

	err = cmd.Start()
	if err != nil {
		return err
	}

	return nil
}
