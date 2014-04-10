package main

import (
	"dropsonde-agent/agent"
	"dropsonde-agent/emitter"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {
	log.Printf("Agent starting, use CTL-C to quit\n")

	defer log.Printf("Agent stopped\n")

	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug logging.")
	flag.Parse()

	if debug {
		emitter.DefaultEmitter = emitter.NewLoggingEmitter()
	}

	stopChan := make(chan struct{})

	go func() {
		err := agent.Run(stopChan)
		if err != nil {
			log.Fatalf("failed to run agent: %v", err)
		}
	}()

	killChan := make(chan os.Signal, 2)
	signal.Notify(killChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-RegisterGoRoutineDumpSignalChannel():
			DumpGoRoutine()
		case <-killChan:
			close(stopChan)
			return
		}
	}
}

func DumpGoRoutine() {
	goRoutineProfiles := pprof.Lookup("goroutine")
	goRoutineProfiles.WriteTo(os.Stdout, 2)
}

func RegisterGoRoutineDumpSignalChannel() chan os.Signal {
	threadDumpChan := make(chan os.Signal, 16)
	signal.Notify(threadDumpChan, syscall.SIGUSR1)

	return threadDumpChan
}
