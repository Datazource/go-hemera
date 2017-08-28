# Hemera - Go Client
[Hemera](https://github.com/hemerajs/hemera) client for the language Go.

[![License MIT](https://img.shields.io/badge/License-MIT-blue.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/hemerajs/go-hemera.svg?branch=master)](http://travis-ci.org/hemerajs/go-hemera)

**Status:** Experimental

## Install

```
go get ./..
go get github.com/nats-io/gnatsd/server
```

### Example
```go

type MathPattern struct {
	Topic string
	Cmd   string
}

type RequestPattern struct {
	Topic    string
	Cmd      string
	A        int
	B        int
	Meta     server.Meta
	Delegate server.Delegate
}

type Response struct {
	Result int
}

nc, _ := nats.Connect(nats.DefaultURL)

hemera, _ := server.CreateHemera(nc, server.Timeout(2000), server.IndexingStrategy(DepthIndexing)...)

pattern := MathPattern{Topic: "math", Cmd: "add"}
hemera.Add(pattern, func(req *RequestPattern, reply server.Reply, context server.Context) {
	fmt.Printf("Request: %+v\n", req)
	result := Response{Result: req.A + req.B}
	reply.Send(result)
})

requestPattern := RequestPattern{
	Topic: "math",
	Cmd: "add",
	A: 1,
	B: 2,
	Meta: server.Meta{ "Test": 1 },
	Delegate: server.Delegate{ "Test": 2 },
}

res := &Response{} // Pointer to struct
hemera.Act(requestPattern, res)

log.Printf("Response %v", res)
```

## Pattern matching
We implemented two indexing strategys
- `depth order` match the entry with the most properties first.
- `insertion order` match the entry with the least properties first. `(default)`

### Examples

#### Depth order
```
a: AddPattern{ Topic: "order" }
b: AddPattern{ Topic: "order", Cmd: "create" }
c: AddPattern{ Topic: "order", Cmd: "create", Type: 3 }

ActPattern{ Topic: "order", Cmd: "create" } // b Matched
ActPattern{ Topic: "order" } // a Matched
ActPattern{ Topic: "order", Type: 3 } // c Matched
```

#### Insertion order
```
a: AddPattern{ Topic: "order" }
b: AddPattern{ Topic: "order", Cmd: "create" }
c: AddPattern{ Topic: "order", Cmd: "create", Type: 3 }

ActPattern{ Topic: "order", Cmd: "create" } // a Matched
ActPattern{ Topic: "order" } // a Matched
ActPattern{ Topic: "order", Type: 3 } // a Matched
```

## Benchmark
- `Lookup` on 10000 Pattern
- `List` on 10000 Pattern
- `Add` with struct of depth 4
```
BenchmarkLookupWeightDepth7-4             200000              7236 ns/op
BenchmarkLookupWeightDepth6-4              10000            139158 ns/op
BenchmarkLookupWeightDepth5-4               5000            281219 ns/op
BenchmarkLookupWeightDepth4-4               2000            705551 ns/op
BenchmarkLookupWeightDepth3-4               2000            557297 ns/op
BenchmarkLookupWeightDepth2-4               2000            690949 ns/op
BenchmarkLookupWeightDepth1-4               2000            682166 ns/op
BenchmarkListDepth100000-4                   500           2504608 ns/op
BenchmarkAddDepth-4                        10000            128326 ns/op
BenchmarkLookupWeightInsertion7-4         200000              7424 ns/op
BenchmarkLookupWeightInsertion6-4         200000              7020 ns/op
BenchmarkLookupWeightInsertion5-4         200000              6845 ns/op
BenchmarkLookupWeightInsertion4-4         200000              6480 ns/op
BenchmarkLookupWeightInsertion3-4         200000              6355 ns/op
BenchmarkLookupWeightInsertion2-4         200000              5895 ns/op
BenchmarkLookupWeightInsertion1-4           3000            468402 ns/op
BenchmarkListInsertion10000-4                500            2627245 ns/op
BenchmarkAddInsertion-4                    10000            734603 ns/op
PASS
```


## TODO
- [X] Setup nats server for testing
- [X] Implement Add and Act
- [X] Infer Response in Act
- [X] Create Context (trace, meta, delegate)
- [X] Use tree for pattern indexing
- [X] Support indexing by depth order
- [X] Support indexing by insetion order
- [ ] Clean request pattern from none primitive values
- [ ] Implement `remove` method
- [X] Implement basic pattern matching (router)

## Credits

- [Bloomrun](https://github.com/mcollina/bloomrun) the pattern matching library for NodeJs
