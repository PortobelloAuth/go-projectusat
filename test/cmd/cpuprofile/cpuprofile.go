// cpuprofile runs every parsertest case through goprojectusat.Normalize many
// times under the CPU profiler, so the per-parse hot path (not the one-off
// zipcity load) dominates the profile. See test/cmd/README.md.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
)

func main() {
	iterations := flag.Int("n", 50, "passes over parsertest.Cases")
	out := flag.String("o", "cpu_profile.prof", "CPU profile output path")
	flag.Parse()

	// Load the default parser (and zipcity) before profiling starts, so the
	// profile shows steady-state parsing rather than initialization.
	start := time.Now()
	if _, err := goprojectusat.Normalize(parsertest.Cases[0].Input); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("warm-up (parser load + 1 parse): %s\n", time.Since(start))

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("could not create CPU profile: %v", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatalf("could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	start = time.Now()
	for range *iterations {
		for _, c := range parsertest.Cases {
			_, _ = goprojectusat.Normalize(c.Input)
		}
	}
	elapsed := time.Since(start)
	parses := *iterations * len(parsertest.Cases)
	fmt.Printf("%d parses in %s (%s/parse)\n", parses, elapsed, elapsed/time.Duration(parses))
}
