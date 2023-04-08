package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/wait"
)

var allSignals = []os.Signal{
	syscall.SIGABRT,
	syscall.SIGALRM,
	syscall.SIGBUS,
	syscall.SIGCHLD,
	syscall.SIGCONT,
	syscall.SIGEMT,
	syscall.SIGFPE,
	syscall.SIGHUP,
	syscall.SIGILL,
	syscall.SIGINFO,
	syscall.SIGINT,
	syscall.SIGIO,
	syscall.SIGIOT,
	syscall.SIGKILL,
	syscall.SIGPIPE,
	syscall.SIGPROF,
	syscall.SIGQUIT,
	syscall.SIGSEGV,
	syscall.SIGSTOP,
	syscall.SIGSYS,
	syscall.SIGTERM,
	syscall.SIGTRAP,
	syscall.SIGTSTP,
	syscall.SIGTTIN,
	syscall.SIGTTOU,
	syscall.SIGURG,
	syscall.SIGVTALRM,
	syscall.SIGWINCH,
	syscall.SIGXCPU,
	syscall.SIGXFSZ,
}

func main() {
	fs := flag.NewFlagSet("wait", flag.ExitOnError)

	loggerConfig := logger.Flags(fs, "logger")

	protocol := flags.String(fs, "", "wait", "Protocol", "Dial protocol (udp or tcp)", "tcp", nil)
	address := flags.String(fs, "", "wait", "Address", "Dial address, e.g. host:port", "", nil)
	timeout := flags.Duration(fs, "", "wait", "Timeout", "Timeout of retries", time.Second*10, nil)
	next := flags.String(fs, "", "wait", "Next", "Action to execute after", "", nil)

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))

	network := strings.TrimSpace(*protocol)
	if len(network) == 0 {
		logger.Fatal(errors.New("protocol is required"))
	}

	addr := strings.TrimSpace(*address)
	if len(addr) == 0 {
		logger.Fatal(errors.New("address is required"))
	}

	if !wait.Wait(network, addr, *timeout) {
		os.Exit(1)
	}

	action := strings.TrimSpace(*next)
	if len(action) == 0 {
		return
	}

	command := exec.Command(action)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	go func() {
		signalsChan := make(chan os.Signal, 1)
		defer close(signalsChan)

		signal.Notify(signalsChan, allSignals...)
		defer signal.Stop(signalsChan)

		for signal := range signalsChan {
			if err := command.Process.Signal(signal); err != nil {
				logger.Error("sending `%s` signal: %s", signal, err)
			}
		}
	}()

	logger.Fatal(command.Run())
}
