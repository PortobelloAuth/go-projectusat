# Testing the memory impact of the default parser

This directory contains two subdirectories with simple cli commands that run a suite
of addresses through `goprojectusat.Normalize()`. One of them (`defaultparser`) uses
the default parser. The other (`customparser`) uses a stub parser that always returns
the same address. Each of them use `"runtime/pprof"` to write a profile describing
their own memory usage.

## Building

```sh
go build -o "./customparser" ./test/cmd/customparser
go build -o "./defaultparser" ./test/cmd/defaultparser
```

## Running

```sh
./customparser
./defaultparser
```

## Reviewing

```sh
go tool pprof customparser_mem_profile.prof 
go tool pprof defaultparser_mem_profile.prof
```

## Recorded results

The resulting pprof memory profiles show that the `defaultparser` executable uses `~58MB`
of memory while the `customparser` executable uses `~1MB` of memory. This is a good
indication that our logic for preventing loading the default parser in to memory when it
is not in use is working properly.

```sh
% go tool pprof customparser_mem_profile.prof 
File: customparser
Type: inuse_space
Time: 2026-10-10 13:58:52 MST
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top -alloc_space
Ignore expression matched no samples
Active filters:
   ignore=alloc_space
Showing nodes accounting for 1025.19kB, 100% of 1025.19kB total
Showing top 10 nodes out of 20
      flat  flat%   sum%        cum   cum%
     513kB 50.04% 50.04%      513kB 50.04%  runtime.mallocgc
  512.19kB 49.96%   100%   512.19kB 49.96%  regexp.onePassCopy
         0     0%   100%   512.19kB 49.96%  github.com/PortobelloAuth/go-projectusat/pkg/addresstypes/military.init
         0     0%   100%   512.19kB 49.96%  regexp.Compile (inline)
         0     0%   100%   512.19kB 49.96%  regexp.MustCompile
         0     0%   100%   512.19kB 49.96%  regexp.compile
         0     0%   100%   512.19kB 49.96%  regexp.compileOnePass
         0     0%   100%      513kB 50.04%  runtime.allocm
         0     0%   100%   512.19kB 49.96%  runtime.doInit
         0     0%   100%   512.19kB 49.96%  runtime.doInit1
(pprof) exit
```

```sh
% go tool pprof defaultparser_mem_profile.prof          
File: defaultparser
Type: inuse_space
Time: 2026-10-10 14:00:34 MST
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top -alloc_space
Ignore expression matched no samples
Active filters:
   ignore=alloc_space
Showing nodes accounting for 57.89MB, 100% of 57.89MB total
Showing top 10 nodes out of 64
      flat  flat%   sum%        cum   cum%
   27.69MB 47.84% 47.84%    27.69MB 47.84%  github.com/bits-and-blooms/bitset.(*BitSet).ReadFrom
   18.56MB 32.05% 79.89%    20.06MB 34.65%  github.com/poetic-systems/zipcity/internal/zipcities.Decode
    6.10MB 10.53% 90.42%     6.10MB 10.53%  github.com/poetic-systems/zipcity.New.func1
       3MB  5.19% 95.61%        3MB  5.19%  runtime.mallocgc
    1.50MB  2.59% 98.20%     1.50MB  2.59%  bufio.(*Scanner).Text (inline)
    0.54MB  0.93% 99.14%     0.54MB  0.93%  github.com/PortobelloAuth/go-projectusat/pkg/businesswords.init.Collect[go.shape.string,go.shape.struct { Primary string; Short string; Alt []string }].Insert[go.shape.map[go.shape.string]go.shape.struct { Primary string; Short string; Alt []string },go.shape.string,go.shape.struct { Primary string; Short string; Alt []string }]-range1
    0.50MB  0.86%   100%    28.19MB 48.70%  github.com/poetic-systems/zipcity/internal/bloomdata.(*BloomData).init.func1 (inline)
         0     0%   100%    54.35MB 93.88%  github.com/PortobelloAuth/go-projectusat.Normalize
         0     0%   100%     6.10MB 10.53%  github.com/PortobelloAuth/go-projectusat/pkg/address/parser.(*Parser).Parse
         0     0%   100%    48.25MB 83.35%  github.com/PortobelloAuth/go-projectusat/pkg/address/parser.New
(pprof) exit
```

## Executable size

```sh
-rwxr-xr-x   1 user  staff    35M Oct 10 13:49 customparser
-rwxr-xr-x   1 user  staff    35M Oct 10 13:50 defaultparser
```

As expected, both executables are roughly the same size; each still contains the embedded files.
