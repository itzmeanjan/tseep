# tseep
State of TCP in Go

## Motivation

All TCP servers written in golang mostly revolves around following idea

`Start a go-routine accepting TCP connections, each accepted connection is handled in its owned go-routine by spawning one, as soon as connection is received.`

No doubt this works well, but when handling thousands of concurrent connections go-scheduler needs to manage thousands of go-routines, where context switching proves to be quite expensive.

This situation can be avoided and TCP servers can be written in a better way where OS kernel is used for our benefit. Rather than all thousands of go-routines ( proactively ) waiting *( i.e. blocked )* for reading data from socket, doesn't matter whether client is going to send data, employing OS kernel in listening to **READ**, **WRITE** completion events on file descriptors and acting only when need to act --- is a lazier approach, which uses system resources less aggressively.
Thus enabling us in handling thousands of concurrent TCP connection at ease.

This way of writing TCP server helps solving **C10K**.

In this project I experiment with mainly aforementioned two kinds of writing TCP servers, while creating a simple key value store like Redis. I also run parallel benchmarking, stress testing with N-clients **( where 1 << 10 <= N <= 1 << 13 )** on solutions to relatively compare their performance, 

## Usage

> Before running any of following tests you've to probably increase open file limit on your system or you'll see `too many open files` error.

- [**v1**](#v1) - TCP server with one go-routine listening for new connections & each connection being handled in its own go-routine.

- [**v2**](#v2) - TCP server with one go-routine listening for new connections & another watching READ, WRITE events on accepted connection's file descriptors.

### :: v1 ::

- Run test with

```bash
pushd v1
go test -v
```

```bash
=== RUN   TestServerV1
--- PASS: TestServerV1 (0.00s)
PASS
ok  	github.com/itzmeanjan/tseep/v1	0.316s
```

- Run benchmarking, 8 rounds

```bash
go test -v -run=xxx -bench V1 -count 8
```

```bash
goos: darwin
goarch: amd64
pkg: github.com/itzmeanjan/tseep/v1
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkServerV1
BenchmarkServerV1-8   	   27032	     43917 ns/op	  23.52 MB/s	    3752 B/op	      52 allocs/op
BenchmarkServerV1-8   	   27034	     43985 ns/op	  23.49 MB/s	    3751 B/op	      52 allocs/op
BenchmarkServerV1-8   	   25768	     43859 ns/op	  23.55 MB/s	    3752 B/op	      52 allocs/op
BenchmarkServerV1-8   	   27397	     44153 ns/op	  23.40 MB/s	    3752 B/op	      52 allocs/op
BenchmarkServerV1-8   	   26668	     47325 ns/op	  21.83 MB/s	    3753 B/op	      52 allocs/op
BenchmarkServerV1-8   	   24280	     49255 ns/op	  20.97 MB/s	    3752 B/op	      52 allocs/op
BenchmarkServerV1-8   	   23754	     50374 ns/op	  20.51 MB/s	    3752 B/op	      52 allocs/op
BenchmarkServerV1-8   	   24038	     49641 ns/op	  20.81 MB/s	    3751 B/op	      52 allocs/op
PASS
ok  	github.com/itzmeanjan/tseep/v1	13.545s
```

- Run stress testing with 1k, 2k, 4k, 8k concurrent connections

```bash
go test -v -tags stress -run=8k # or 1k, 2k, 4k
popd
```

```bash
=== RUN   TestServerV1_Stress_8k
--- PASS: TestServerV1_Stress_8k (2.56s)
PASS
ok  	github.com/itzmeanjan/tseep/v1	2.723
```
