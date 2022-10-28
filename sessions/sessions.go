package sessions

import (
	"bufio"
	"fmt"
	// "log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type session struct {
	ts   int
	name string
}

func GetActiveSession() string {
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error getting active session:", err)
		os.Exit(1)
	}
	return strings.TrimRight(string(output), "\n")
}

// Sorted by ts
func GetActiveSessions() []string {
	sessions := make([]session, 0)
	activeSession := GetActiveSession()

	// tmux ls -F '#{session_activity} #{session_name}' -f "#{!=:#{session_name},$(tmux display-message -p '#S')}"
	cmd := exec.Command("tmux", "ls", "-F", "#{session_activity} #{session_name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error getting active sessions:", err)
		os.Exit(1)
	}

	for _, line := range strings.Split(string(output), "\n") {
		// log.Printf("- %s", line)

		s := strings.Split(line, " ")
		if len(s) < 2 {
			continue
		}
		if s[1] == activeSession {
			continue
		}

		ts, err := strconv.Atoi(s[0])
		if err != nil {
			fmt.Println("Error reading timestamp from tmux session:", line)
			os.Exit(1)
		}

		// log.Printf("'%s' == '%s'", activeSession, s[1])
		sessions = append(sessions, session{ts: ts, name: s[1]})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[j].ts <= sessions[i].ts
	})

	return sessionsToString(sessions)
}

func getActiveSessions() map[string]int {
	sessions := make(map[string]int)

	cmd := exec.Command("tmux", "ls", "-F", "#S")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error getting active sessions:", err)
		os.Exit(1)
	}

	for _, line := range strings.Split(string(output), "\n") {
		sessions[line] = 1
	}

	return sessions
}

func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user's home directory:", err)
		os.Exit(1)
	}
	return homeDir
}

func GetInactiveSessions() []string {
	activeSessions := getActiveSessions()
	sessions := make([]string, 0)

	file, err := os.Open(fmt.Sprintf("%s/%s", getHomeDir(), ".config/tmuxinator"))
	if err != nil {
		fmt.Println("Error reading tmuxinator config:", err)
		os.Exit(1)
	}
	defer file.Close()

	names, _ := file.Readdirnames(0)
	for _, name := range names {
		l := len(name)
		if name[0] == '.' {
			continue
		}
		if name[l-4:l] != ".yml" {
			continue
		}
		session := name[0 : l-4]
		if _, exists := activeSessions[session]; exists {
			continue
		}
		// log.Printf("dir: %s", session)
		sessions = append(sessions, session)
	}

	sort.Strings(sessions)
	return sessions
}

// cat ~/.ssh/known_hosts \
// | perl -nE 's/^ \[? ( [A-Za-z][\w\.-]+ ) .*$/$1/x and print' \
func GetRemoteSessions() []string {
	file, err := os.Open(fmt.Sprintf("%s/%s", getHomeDir(), ".ssh/known_hosts"))
	if err != nil {
		fmt.Println("Error reading .ssh/known_hosts:", err)
		os.Exit(1)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	hosts := make([]string, 0)
	re := regexp.MustCompile(`^\[?([A-Za-z][\w\.-]+)`)
	for fileScanner.Scan() {
		line := fileScanner.Text()

		match := re.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		host := match[1]
		if host == "" {
			continue
		}

		hosts = append(hosts, host)

	}

	return hosts
}

func sessionsToString(sessions []session) []string {
	s := make([]string, 0)
	for _, v := range sessions {
		s = append(s, v.name)
	}
	return s
}
