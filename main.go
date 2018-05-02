package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/alphagov/paas-cf-conduit/logging"
	"github.com/spf13/cobra"
)

var (
	NonInteractive   bool
	ConduitReuse     bool
	ConduitAppName   string
	ConduitOrg       string
	ConduitSpace     string
	ConduitLocalPort int64
	ApiEndpoint      string
	ApiToken         string
	ApiInsecure      bool
	shutdown         chan struct{}
)

func init() {
	shutdown = make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 3)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		<-sig
		close(shutdown)
		for range sig {
			log.Println("...shutting down")
		}
	}()
}

func retry(fn func() error) error {
	delayBetweenRetries := 500 * time.Millisecond
	maxRetries := 10
	try := 0
	for {
		try++
		err := fn()
		if err == nil {
			return nil
		}
		if try > maxRetries {
			return err
		}
		time.Sleep(delayBetweenRetries)
	}
}

func main() {
	if terminal.IsTerminal(int(os.Stdout.Fd())) && terminal.IsTerminal(int(os.Stderr.Fd())) {
		NonInteractive = false
	} else {
		NonInteractive = true
	}
	cmd := &cobra.Command{Use: "cf"}
	cmd.PersistentFlags().BoolVarP(&logging.Verbose, "verbose", "", false, "verbose output")
	cmd.PersistentFlags().BoolVarP(&NonInteractive, "no-interactive", "", NonInteractive, "disable progress indicator and status output")
	cmd.PersistentFlags().StringVarP(&ConduitOrg, "org", "o", "", "target org (defaults to currently targeted org)")
	cmd.PersistentFlags().StringVarP(&ConduitSpace, "space", "s", "", "target space (defaults to currently targeted space)")
	cmd.PersistentFlags().BoolVarP(&ConduitReuse, "reuse", "r", false, "speed up multiple invocations of conduit by not destroying the tunnelling app")
	cmd.PersistentFlags().MarkHidden("reuse")
	cmd.PersistentFlags().StringVarP(&ConduitAppName, "app-name", "n", fmt.Sprintf("__conduit_%d__", os.Getpid()), "app name to use for tunnelling app (must not exist)")
	cmd.PersistentFlags().MarkHidden("app-name")
	cmd.PersistentFlags().Int64VarP(&ConduitLocalPort, "local-port", "p", 7080, "start selecting local ports from")
	cmd.PersistentFlags().StringVar(&ApiEndpoint, "endpoint", "", "set API endpoint")
	cmd.PersistentFlags().MarkHidden("endpoint")
	cmd.PersistentFlags().StringVar(&ApiToken, "token", "", "set API token")
	cmd.PersistentFlags().MarkHidden("token")
	cmd.PersistentFlags().BoolVar(&ApiInsecure, "insecure", false, "allow insecure API endpoint")
	cmd.PersistentFlags().MarkHidden("insecure")
	cmd.AddCommand(ConnectService)
	cmd.AddCommand(Uninstall)
	plugin.Start(&Plugin{cmd})
}
