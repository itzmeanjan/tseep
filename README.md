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

- [**v1**](#-v1-) - TCP server with one go-routine listening for new connections & each connection being handled in its own go-routine.

- [**v2**](#-v2-) - TCP server with one go-routine listening for new connections & another watching READ, WRITE events on accepted connection's file descriptors

- [**v3**](#-v3-) - Also experimented with multiple go-routines watching different event loops; each newly accepted connection is delegated to any one of these watcher for rest of their life time. _In simple terms, this is a generic version of **v2**, where I use N go-routines for watching, where N > 1. In v2, N = 1._

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

- Run parallel benchmarking, 8 rounds

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

- Run stress testing with {1k, 2k, 4k, 8k} concurrent connections

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

### :: v2 ::

- Run test with

```bash
pushd v2
go test -v
```

```bash
=== RUN   TestServerV2
--- PASS: TestServerV2 (0.00s)
PASS
ok  	github.com/itzmeanjan/tseep/v2	0.861s
```

- Run benchmarking of 8 rounds using all CPU cores

```bash
go test -v -run=xxx -bench V2 -count 8
```

```bash
goos: darwin
goarch: amd64
pkg: github.com/itzmeanjan/tseep/v2
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkServerV2
BenchmarkServerV2-8   	   34479	     32831 ns/op	  62.93 MB/s	    6193 B/op	      72 allocs/op
BenchmarkServerV2-8   	   33703	     33781 ns/op	  61.16 MB/s	    6186 B/op	      72 allocs/op
BenchmarkServerV2-8   	   34296	     32974 ns/op	  62.66 MB/s	    6185 B/op	      72 allocs/op
BenchmarkServerV2-8   	   35888	     33468 ns/op	  61.73 MB/s	    6183 B/op	      72 allocs/op
BenchmarkServerV2-8   	   36135	     32762 ns/op	  63.06 MB/s	    6181 B/op	      72 allocs/op
BenchmarkServerV2-8   	   35479	     35970 ns/op	  57.44 MB/s	    6185 B/op	      72 allocs/op
BenchmarkServerV2-8   	   31009	     36539 ns/op	  56.54 MB/s	    6186 B/op	      72 allocs/op
BenchmarkServerV2-8   	   33430	     35943 ns/op	  57.48 MB/s	    6189 B/op	      72 allocs/op
PASS
ok  	github.com/itzmeanjan/tseep/v2	12.639s
```

- Run stress testing with {1k, 2k, 4k, 8k} concurrent connections

```bash
go test -v -tags stress -run=8k
popd
```

```bash
=== RUN   TestServerV2_Stress_8k
--- PASS: TestServerV2_Stress_8k (2.67s)
PASS
ok  	github.com/itzmeanjan/tseep/v2	3.234s
```

### :: v3 ::

- Run test with

```bash
pushd v3
go test -v
```

```bash
=== RUN   TestServerV3
--- PASS: TestServerV3 (0.00s)
PASS
ok  	github.com/itzmeanjan/tseep/v3	0.593s
```

- Run 8 rounds of parallel benchmarking, where **8** go-routines used for watching 8 kernel event loop, each managing a subset of total accepted connections

```bash
go test --run=xxx -bench V3 -count 8
```

```bash
goos: darwin
goarch: amd64
pkg: github.com/itzmeanjan/tseep/v3
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkServerV3-8   	   39541	     29209 ns/op	  70.73 MB/s	    5716 B/op	      74 allocs/op
BenchmarkServerV3-8   	   39259	     29116 ns/op	  70.96 MB/s	    5714 B/op	      74 allocs/op
BenchmarkServerV3-8   	   40550	     29216 ns/op	  70.71 MB/s	    5714 B/op	      74 allocs/op
BenchmarkServerV3-8   	   40640	     29507 ns/op	  70.02 MB/s	    5713 B/op	      74 allocs/op
BenchmarkServerV3-8   	   38982	     31441 ns/op	  65.71 MB/s	    5713 B/op	      74 allocs/op
BenchmarkServerV3-8   	   36420	     32439 ns/op	  63.69 MB/s	    5714 B/op	      74 allocs/op
BenchmarkServerV3-8   	   37038	     32846 ns/op	  62.90 MB/s	    5715 B/op	      74 allocs/op
BenchmarkServerV3-8   	   37333	     32611 ns/op	  63.35 MB/s	    5714 B/op	      74 allocs/op
PASS
ok  	github.com/itzmeanjan/tseep/v3	12.929s
```

- Doing stress testing with {1k, 2k, 4k, 8k} concurrent connections

```bash
go test -v -tags stress -run=8k
popd
```

```bash
=== RUN   TestServerV3_Stress_8k
--- PASS: TestServerV3_Stress_8k (2.55s)
PASS
ok  	github.com/itzmeanjan/tseep/v3	2.686s
```
