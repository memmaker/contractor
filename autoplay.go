package main

import (
	"contractor/foundation"
	"contractor/game"
	"contractor/ui_console"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"
)

// runAutoplay plays data_atom/playtests/<name>.rec scripts. Returns the number of failed runs.
//
//	contractor autoplay [-headless] [-seed N] [-max-turns N] [-delay 40ms] name|path.rec ...
func runAutoplay(config *foundation.Configuration, args []string) int {
	flags := flag.NewFlagSet("autoplay", flag.ExitOnError)
	headless := flags.Bool("headless", false, "no window, run as fast as possible")
	seed := flags.Int64("seed", 1, "random seed")
	maxTurns := flags.Int("max-turns", 5000, "fail after this many turns")
	delay := flags.Duration("delay", 40*time.Millisecond, "visual mode: time per step")
	cpuProfile := flags.String("cpuprofile", "", "write a CPU profile to this file")
	flags.Parse(args)
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Println(err)
			return 1
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		interrupted := make(chan os.Signal, 1) // ctrl-c still leaves a usable profile
		signal.Notify(interrupted, os.Interrupt)
		go func() { <-interrupted; pprof.StopCPUProfile(); os.Exit(130) }()
	}

	if config.PlayerName == "" {
		config.PlayerName = "Autoplay"
	}
	failed := 0
	var summary []string
	for _, name := range flags.Args() {
		file := name
		if !strings.HasSuffix(file, ".rec") {
			file = filepath.Join(config.DataRootDir, "playtests", name+".rec")
		}
		rand.Seed(*seed)
		gameState := game.NewGameState(config)
		run := game.NewAutoplay(gameState, file, os.Stdout)
		run.MaxTurns = *maxTurns
		start := time.Now()
		if *headless {
			gameState.UIReady(game.NewHeadlessUI())
			for !run.Step() {
			}
		} else {
			ui := ui_console.NewTextUI(gameState, chooseGraphicsMode(), config)
			ui.Autoplay(run.Step, run.Reading, *delay, 25**delay)
			ui.StartGameLoop()
		}
		status := "PASS"
		if !run.Passed {
			status = "FAIL"
			failed++
		}
		summary = append(summary, fmt.Sprintf("%s  %-36s %8.2fs  %s", status, run.Name, time.Since(start).Seconds(), firstLine(run.Result)))
	}
	fmt.Println(strings.Join(summary, "\n"))
	return failed
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
