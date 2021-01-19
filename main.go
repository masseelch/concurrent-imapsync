package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	minThreadCount     = 1
	defaultThreadCount = 2
)

var (
	cyan  = color.New(color.FgCyan).SprintfFunc()
	green = color.New(color.FgHiGreen).SprintfFunc()
	red   = color.New(color.FgHiRed).SprintfFunc()
)

type (
	credentials struct {
		host     string
		user     string
		password string
	}
	job struct {
		source credentials
		target credentials
	}
)

func main() {
	// Mak sure imapsync is installed.
	if _, err := exec.LookPath("imapsync"); err != nil {
		panic(err)
	}

	// Source and amount of parallel executions
	var s string
	var t int

	// Set by flag.
	flag.StringVar(&s, "source", "accounts.txt", "File containing the accounts data")
	flag.IntVar(&t, "threads", defaultThreadCount, "Amount of parallel processes to use")

	flag.Parse()

	// Validate
	if _, err := os.Stat(s); err != nil {
		panic(err)
	}

	if t < minThreadCount {
		t = minThreadCount
	}

	// The accounts we have to transfer.
	jobs := make(chan job, t)

	// Sync the go routines.
	wg := new(sync.WaitGroup)

	// Start the workers.
	for w := 1; w <= t; w++ {
		go syncMailboxWorker(wg, jobs)
	}

	// Read in the account data.
	f, err := os.Open(s)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		wg.Add(1)
		jobs <- jobFromString(sc.Text())
	}

	wg.Wait()

	close(jobs)
}

func syncMailboxWorker(wg *sync.WaitGroup, jobs <-chan job /* todo - We need a channel for errors and logs here. */) {
	for j := range jobs {
		fmt.Println("Started", cyan(j.source.user))

		cmd := exec.Command(
			"imapsync",
			fmt.Sprintf("--host1 '%s'", j.source.host),
			fmt.Sprintf("--user1 '%s'", j.source.user),
			fmt.Sprintf("--password1 '%s'", j.source.password),
			fmt.Sprintf("--host2 '%s'", j.target.host),
			fmt.Sprintf("--user2 '%s'", j.target.user),
			fmt.Sprintf("--password2 '%s'", j.target.password),
			fmt.Sprintf("--logdir '%s'", j.logDir()),
			fmt.Sprintf("--logfile '%s'", j.logFile()),
		)
		if err := cmd.Run(); err != nil {
			p, err2 := filepath.Abs(filepath.Join(j.logDir(), j.logFile()))
			if err2 != nil {
				panic(err2)
			}

			f, err2 := os.OpenFile(p, os.O_WRONLY|os.O_APPEND, 0666)
			if err2 != nil {
				panic(err2)
			}
			defer f.Close()

			f.WriteString("\n\nAn error occured:\n")
			f.WriteString(err.Error())

			fmt.Printf("%s %s\n\tLogs can be seen in %s\n", red("ERROR"), cyan(j.source.user), filepath.Join(j.logDir(), j.logFile()))
		} else {
			fmt.Printf("%s %s\n\tLogs can be seen in %s\n", green("FINISHED"), cyan(j.source.user), filepath.Join(j.logDir(), j.logFile()))
		}

		wg.Done()
	}
}

func jobFromString(s string) job { // todo - add error
	d := strings.Split(s, "|")

	return job{
		source: credentials{
			host:     d[0],
			user:     d[1],
			password: d[2],
		},
		target: credentials{
			host:     d[3],
			user:     d[4],
			password: d[5],
		},
	}
}

func (j job) logDir() string {
	return "logs"
}

func (j job) logFile() string {
	return fmt.Sprintf("%s_TO_%s.log", j.source.user, j.target.user)
}
