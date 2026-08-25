package switchcmd

import (
	"fmt"
	"os"
	"os/exec"
)

func executeTmux(action stage, session string) {
	var err error

	switch action {
	case liveSession:
		err = exec.Command("tmux", "switch-client", "-t", tmuxSessionTarget(session)).Run()

	case inactiveSession:
		err = exec.Command("tmux", "run-shell", fmt.Sprintf("tmuxinator start %s", session)).Run()

	case remoteSession:
		exec.Command("tmux", "new", "-d", "-s", session, fmt.Sprintf("$HOME/bin/smux %s", session)).Run()

		err = exec.Command("tmux", "switch-client", "-t", tmuxSessionTarget(session)).Run()
	}

	if err != nil {
		fmt.Println("Error running:", err)
		os.Exit(1)
	}
}

func tmuxSessionTarget(session string) string {
	return session + ":"
}
