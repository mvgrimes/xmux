package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/cmd/bar"
	"github.com/mvgrimes/xmux/cmd/popup"
	switchcmd "github.com/mvgrimes/xmux/cmd/switch"
	versioncmd "github.com/mvgrimes/xmux/cmd/version"
	"github.com/mvgrimes/xmux/cmd/watch"
)

var version = "0.2.4"

func main() {
	root := &cobra.Command{
		Use:          "xmux",
		Short:        "tmux session launcher and service monitor",
		SilenceUsage: true,
		RunE:         switchcmd.Run,
	}
	root.AddCommand(switchcmd.NewCommand())
	root.AddCommand(watch.NewCommand())
	root.AddCommand(bar.NewCommand())
	root.AddCommand(popup.NewCommand())
	root.AddCommand(versioncmd.NewCommand(version))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
