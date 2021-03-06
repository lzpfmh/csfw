# CoreStore library WIP = Work in Progress

[![Build Status](https://travis-ci.org/corestoreio/csfw.svg?branch=master)](https://travis-ci.org/corestoreio/csfw) [![wercker status](https://app.wercker.com/status/d7d0bdda415d2228b6fb5bb01681b5c4/s/master "wercker status")](https://app.wercker.com/project/bykey/d7d0bdda415d2228b6fb5bb01681b5c4) [![Appveyor Build status](https://ci.appveyor.com/api/projects/status/lrlnbpcjdy585mg1/branch/master?svg=true)](https://ci.appveyor.com/project/SchumacherFM/csfw/branch/master) [![GoDoc](http://godoc.org/github.com/corestoreio/csfw?status.svg)](http://godoc.org/github.com/corestoreio/csfw) [![Join the chat at https://gitter.im/corestoreio/csfw](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/corestoreio/csfw?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [goreportcard](http://goreportcard.com/report/Corestoreio/csfw)

eCommerce library which is compatible to Magento 1 and 2 database schema.

Magento is a trademark of [MAGENTO, INC.](http://www.magentocommerce.com/license/).

Min. Go Version: 1.7

## Usage

To properly use the CoreStore library some environment variables must be set
before running `go generate`. (TODO)

### Required settings

`CS_DSN` the environment variable for the MySQL connection.

```shell
$ export CS_DSN='magento1:magento1@tcp(localhost:3306)/magento1'
$ export CS_DSN='magento2:magento2@tcp(localhost:3306)/magento2'
```

```
$ go get github.com/corestoreio/csfw
$ export CS_DSN='see previous'
$ cd $GOPATH/src/github.com/corestoreio/csfw
$ go run codegen/tableToStruct/*.go
```

## Testing

Setup two databases. One for Magento 1 and one for Magento 2 and fill them with
the provided [testdata](https://github.com/corestoreio/csfw/tree/master/testData).

Create a DSN env var `CS_DSN` and point it to Magento 1 database. Run the tests.
Change the env var to let it point to Magento 2 database. Rerun the tests.

### Finding Allocations

Side note: There is a new testing approach: TBDD = Test n' Benchmark Driven
Development.

On the first run we got this result:

```
$ go test -run=😶 -bench=Benchmark_WithInitStoreByToken .
PASS
Benchmark_WithInitStoreByToken-4	  100000	     17297 ns/op	    9112 B/op	     203 allocs/op
ok  	github.com/corestoreio/csfw/store	2.569s

```

Quite shocking to use 203 allocs for just figuring out the current store view
within a request.

Now compile your tests into an executable binary:

```
$ go test -c
```

This compilation reduces the noise in the below output trace log.

```
$ GODEBUG=allocfreetrace=1 ./store -test.run=😶 -test.bench=Benchmark_WithInitStoreByToken -test.benchtime=10ms 2>trace.log
```

Now open the trace.log file (around 26MB) and investigate all the allocations
and refactor your code. Once finished you can achieve results like:

```
$ go test -run=NONE -bench=Benchmark_WithInitStoreByToken .
PASS
Benchmark_WithInitStoreByToken-4	 2000000	       826 ns/op	     128 B/op	       5 allocs/op
ok  	github.com/corestoreio/csfw/store	2.569s
```

### Profiling

```
$ go test -cpuprofile=cpu.out -benchmem -memprofile=mem.out -run=NONE -bench=NameOfBenchmark -v
$ go tool pprof packageName.test cpu.out
Entering interactive mode (type "help" for commands)
(pprof) top5
560ms of 1540ms total (36.36%)
Showing top 5 nodes out of 112 (cum >= 60ms)
      flat  flat%   sum%        cum   cum%
     180ms 11.69% 11.69%      400ms 25.97%  runtime.mallocgc
```

- `flat` is how much time is spent inside of a function.
- `cum` shows how much time is spent in a function, and also in any code called by a function.

For memory profile:

```
Sample value selection option (for heap profiles):
  -inuse_space      Display in-use memory size
  -inuse_objects    Display in-use object counts
  -alloc_space      Display allocated memory size
  -alloc_objects    Display allocated object counts

$ go tool pprof -alloc_objects packageName.test mem.out
```

#### Other development helpers

- [go get github.com/maruel/panicparse/cmd/pp](https://github.com/maruel/panicparse)
- [go get github.com/alecthomas/gometalinter](https://github.com/alecthomas/gometalinter)

## TODO

If you find an entry in the source like `TODO(CS)` that means you can ask `CS`
to get more information about how to implement and what to fix if the context of
the todo description isn't understandable.

- Create Magento 1+2 modules to setup test database and test Magento system.

## Contributing

Please have a look at the [contribution guidelines](https://github.com/corestoreio/corestore/blob/master/CONTRIBUTING.md).

## Acknowledgements

Some packages have been fully refactored but the initial idea has been extracted
from the geniuses listed below:

| Name | Package | License |
| -------|----------|-------|
| Steve Francia | [conv](http://github.com/corestoreio/csfw/tree/master/utils/conv) | MIT Copyright (c) 2014 |
| Steve Francia | [bufferpool](http://github.com/corestoreio/csfw/tree/master/utils/bufferpool) | [Simple Public License, V2.0](http://opensource.org/licenses/Simple-2.0) |
| Jonathan Novak, Tyler Smith, Michal Bohuslávek | [dbr](http://github.com/corestoreio/csfw/tree/master/storage/dbr) | The MIT License (MIT) 2014 |
| Martin Angers and Contributors. | [ctxthrottled](http://github.com/corestoreio/csfw/tree/master/net/ctxthrottled) | The MIT License (MIT) 2014 |
| Dave Cheney <dave AT cheney.net> | [errors](https://github.com/pkg/errors) | The MIT License (MIT) 2015 |
| Jad Dittmar | [finance](https://github.com/Confunctionist/finance) aka. [money](http://github.com/corestoreio/csfw/tree/master/storage/money) | Copyright (c) 2011 |
| Wenbin Xiao | [sqlparser](https://github.com/xwb1989/sqlparser) | Copyright 2015 BSD Style |
| Google Inc | [youtube/vitess\sqlparser](https://github.com/youtube/vitess) | Copyright 2012 BSD Style |
| Julien Schmidt | [ctxrouter](https://github.com/corestoreio/csfw/tree/master/net/ctxrouter) aka. [httprouter](https://github.com/julienschmidt/httprouter) | Copyright (c) 2013 MIT |
| Olivier Poitrey| [ctxmw.WithAccessLog](https://github.com/corestoreio/csfw/tree/master/net/ctxmw) & CORS | Copyright (c) 2014-2015  MIT License |
| Dave Grijalva| [csjwt](https://github.com/corestoreio/csfw/tree/master/util/csjwt) | Copyright (c) 2012 MIT License |
| Uber Technologies, Inc. | [log](https://github.com/corestoreio/csfw/tree/master/log) | Copyright (c) 2016 MIT License |
| 2013 The Go Authors | [singleflight](https://github.com/corestoreio/csfw/tree/master/sync/singleflight) | Copyright (c) 2013 BSD Style |

## Licensing

CoreStore is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/corestoreio/corestore/blob/master/LICENSE) for the full license text.

## Copyright

[Cyrill Schumacher](http://cyrillschumacher.com) - [PGP Key](https://keybase.io/cyrill)
