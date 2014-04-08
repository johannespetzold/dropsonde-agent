package main

import (
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {
	stopChan := make(chan struct{})

	//	go func() {
	//		err := agent.Start(stopChan)
	//		// log error
	//	}

	killChan := make(chan os.Signal)
	signal.Notify(killChan, os.Kill, os.Interrupt)

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
	threadDumpChan := make(chan os.Signal)
	signal.Notify(threadDumpChan, syscall.SIGUSR1)

	return threadDumpChan
}
