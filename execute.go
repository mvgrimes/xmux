package main

import (
	"fmt"
	"os"
	"os/exec"
)

func executeTmux(action stage, session string) {
	var err error

	switch action {
	case liveSession:
		err = exec.Command("tmux", "switch-client", "-t", session).Run()

	case inactiveSession:
		err = exec.Command("tmux", "run-shell", fmt.Sprintf("tmuxinator start %s", session)).Run()

	case remoteSession:
		exec.Command("tmux", "new", "-d", "-s", session, fmt.Sprintf("$HOME/bin/smux %s", session)).Run()
		err = exec.Command("tmux", "switch-client", "-t", session).Run()
	}

	if err != nil {
		fmt.Println("Error running:", err)
		os.Exit(1)
	}
}
