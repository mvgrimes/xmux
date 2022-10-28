package main

import (
	"fmt"
	"os"
	"os/exec"
)

func executeTmux(action stage, session string) {
	switch action {
	case liveSession:
		// tmux switch-client -t session
		runOrDie("tmux", "switch-client", "-t", session)

	case inactiveSession:
		// tmux run-shell "tmuxinator start {session}"
		runOrDie("tmux", "run-shell", fmt.Sprintf("tmuxinator start %s", session))

	case remoteSession:
		// tmux new -d -s "$host" "$HOME/bin/smux $host"
		// tmux switch-client -t "$host"
		exec.Command("tmux", "new", "-d", "-s", session, fmt.Sprintf("$HOME/bin/smux %s", session)).Run()
		runOrDie("tmux", "switch-client", "-t", session)
	}
}

func runOrDie(executable string, command ...string) {
	cmd := exec.Command(executable, command...)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running:", err)
		os.Exit(1)
	}
}

func runIgnoreError(executable string, command ...string) {
}
