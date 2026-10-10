package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	goprojectusat "github.com/PortobelloAuth/go-projectusat"
	"github.com/PortobelloAuth/go-projectusat/pkg/address"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser"
	"github.com/PortobelloAuth/go-projectusat/pkg/address/parser/parsertest"
)

func main() {
	f, err := os.Create("customparser_mem_profile.prof")
	if err != nil {
		log.Fatalf("could not create memory profile: %v", err)
	}
	defer f.Close()

	stub := parser.ParsingFn(func(source string) (*address.Address, error) {
		return &address.Address{
			PrimaryNumber: "1015",
			StreetName:    "NORTH",
			StreetSuffix:  "AVENUE",
		}, nil
	})

	for _, c := range parsertest.Cases {
		out, err := goprojectusat.Normalize(c.Input, goprojectusat.WithCustomAddressParser(stub))
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}

		if len(c.Want) > 0 && out != c.Want {
			fmt.Printf("Want: %s\nGot:  %s\n\n", c.Want, out)
		}
	}

	time.Sleep(20 * time.Second)

	runtime.GC() // Run garbage collection first for accurate, active memory mapping
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatalf("could not write memory profile: %v", err)
	}
}
