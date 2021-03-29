---
title: Testing in Go
date: "2021-03-27"
description: Use Go's built-in tools for testing and code coverage
tags: [go, programming]
---

Go's tooling makes testing your business logic relatively straightforward. From [Go's docs](https://pkg.go.dev/testing),
just create a new `*xxx_test.go` file; create a `func TestXxx(t *testing.T)`; run your test(s) and use the builtin testing APIs to 
fail a test when your business logic returns the wrong output. Easy.

A perfect place to start (besides the Go docs), is [Dave Cheney's blog](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests). Note that
the article is not limited to only table-drive tests.

I'm going to walk through some ideas I've found useful when writing and designing tests in Go. I hope they may serve as food for thought when
you're writing your own tests. Much of what is here is covered elsewhere, but I'll share my thoughts about how tests should be structured.

# Your business logic

Imagine you have some business data encapsulated like so:
```go
type BusinessDocument struct {
	ID   int
	Name string
}
```
And the VP of Product has informed you that your application must support updating BusinessDocuments with data
from external clients. So you write the following function:
```go
func MutateBusinessDocument(bd *BusinessDocument, newID int, newName string) error {
	if bd == nil {
		return errors.New("business document cannot be nil")
	}
	if newID <= 0 {
		return errors.New("new ID must be strictly positive")
	}
	if len(newName) == 0 {
		return errors.New("newName must not be empty")
	}

	bd.ID = newID
	bd.Name = newName

	return nil
}
```
As you can see, there are some rules:
1. Assume your business document is not nil. If your business document is nil, you can throw an error and assume that some other service has failed, _and_
2. Business documents must have IDs that are strictly positive, _and_
3. They cannot have blank names.
If all these rules are followed, then go ahead and update the document. Lovely.

Now, we test. We test and test and test and then, ~~we have a glass of wine and watch The Expanse~~ we test some more.

Go's `testing` library is great. If you haven't already, go ahead and read the official docs.

# Normal Unit Tests

In our contrived example, there are no external dependencies, so there's nothing to mock and no complex setup.

Let's begin by stubbing out our test function in `xxx_test.go`
```go
func TestMutateBusinessDocumentSimple(t *testing.T) {
    // Test code will go here
}
```
The Go test tool expects some things about the names and signatures of our test functions. Principally,
that they begin with `Test`, and that they accept one parameter: `t *testing.T`. Here, the `t`, is simply
a struct that is used by the testing tool for running tests, failing or skipping tests, and logging output.
(There is a very similar `b *testing.B` for benchmark tests, which we'll see next.)

Firstly, let's test the happy path.

We'll start by creating a new `BusinessDocument` with some random data:
```go
func TestMutateBusinessDocumentSimple(t *testing.T) {
    haveID := rand.Intn(100)
    wantID := rand.Intn(100)
    
    haveName := randomString(100)
    wantName := randomString(100)
    
    have := BusinessDocument{
        ID:   haveID,
        Name: haveName,
	}
}
```

And then we'll actually invoke our new function, and then verify that its new state is what we expected:
```go
func TestMutateBusinessDocument(t *testing.T) {
    haveID := rand.Intn(100)
    wantID := rand.Intn(100)
    
    haveName := randomString(100)
    wantName := randomString(100)
    
    have := BusinessDocument{
        ID:   haveID,
        Name: haveName,
	}
    
    // Invoke our new function
    err := MutateBusinessDocument(&have, wantID, wantName)

    // And check that we mutate it 'correctly'
    if err != nil {
        t.Fatalf("Should not get an error")
	}
	if have.ID != wantID {
		t.Fatalf("IDs do not match")	
    }
	if have.Name != wantName {
        t.Fatalf("IDs do not match")
	}    
}
```

And run it with `go test` (it should pass).

That's all well and good, but we should also verify that our function will fail if we feed it bad data. This is where table
tests come in might handy. (required reading: [Dave Cheney on table-driven tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests))

Create a new function that we'll put our error-checking tests in, and create your test table:
```go
func TestMutateBusinessDocument(t *testing.T) {
		tests := []struct {
			name    string
			bd      *BusinessDocument
			newID   int
			newName string
			err     error
		}{}
}
```
Each struct member of our `tests` slice contains all the data that we're going to use and verify.

```go
func TestMutateBusinessDocumentErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		bd      *BusinessDocument
		newID   int
		newName string
		err     error
	}{
        {"business document is nil", nil, 0, "", errors.New("business document cannot be nil")},
		{"newID is not strictly positive", &BusinessDocument{}, 0, 0, errors.New("new ID must be strictly positive")},
		{"newName is empty", &BusinessDocument{}, 1, "", errors.New("newName must not be empty")},
    }
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := MutateBusinessDocument(test.bd, test.newID, test.newName)
			if err == nil {
				t.Fatal("Err must not be nil")
			} else {
				if err.Error() != test.err.Error() {
					t.Fatal("Errors must match")
				}
			}
		})
	}
}
```
Above, we're going to invoke `MutateBusinessDocument` with the data in our `tests` slice. Since we're only checking our
error-handling, we expect that `MutateBusinessDocument` should _always_ return an error, and that the error we get
is the error we expect.

This is glorious. It's easy to write tests against concrete implementations of our business logic for both the happy path
and the unhappy path.

Next, we'll see how to use subtests to group our tests by whatever categories we like, and how that grouping can really
help in finding a failing or broken test.

The `t` from `t *testing.T` has a `Run(...)` method. We can create one or more subtests like so
```go
func TestMutateBusinessDocument(t *testing.T) {
	t.Run("mutate business document", func(t *testing.T) {
        // test code goes here
	})

	t.Run("check error handling", func(t *testing.T) {
        // test code goes here
	})
}
```

One thing I like to do is group into two sets: happy path; and unhappy path (aka, error handling), with the name of the API
under test as the name of the function. So let's move all of our testing logic into the `t.Run(...)` blocks. That leaves us with:
```go
func TestMutateBusinessDocument(t *testing.T) {
	t.Run("update ID and name", func(t *testing.T) {
		haveID := rand.Intn(100)
		wantID := rand.Intn(100)

		haveName := randomString(100)
		wantName := randomString(100)

		have := BusinessDocument{
			ID:   haveID,
			Name: haveName,
		}

		// The actual code under test
		err := MutateBusinessDocument(&have, wantID, wantName)
		if err != nil {
			t.Fatalf("Should not get an error")
		}

		if have.ID != wantID {
			t.Fatalf("IDs do not match")
		}

		if have.Name != wantName {
			t.Fatalf("IDs do not match")
		}
	})

	t.Run("check error handling", func(t *testing.T) {
		tests := []struct {
			name    string
			bd      *BusinessDocument
			newID   int
			newName string
			err     error
		}{
			{"business document is nil", nil, 0, randomString(100), errors.New("business document cannot be nil")},
			{"newID is not strictly positive", &BusinessDocument{}, 0, randomString(100), errors.New("new ID must be strictly positive")},
			{"newName is empty", &BusinessDocument{}, 1, "", errors.New("newName must not be empty")},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				// run our business function
				err := MutateBusinessDocument(test.bd, test.newID, test.newName)
				if err != nil {
					t.Fatal("Err must not be nil")
				} else {
					if err.Error() != test.err.Error() {
						t.Fatal("Errors must match")
					}
				}
			})
		}
	})
}
```
Sweet. So, if you run these tests with `go test`, nothing remarkable happens. You might think, as I did, what's the point
of grouping tests like this?

It becomes helpful when you have a failing test.

Let's intentionally break and then run the `"check error handling"` test by changing
```go
err := MutateBusinessDocument(test.bd, test.newID, test.newName)
    if err != nil {
        ...
```
to
```go
err := MutateBusinessDocument(test.bd, test.newID, test.newName)
    if err == nil {
        ...
```

And apply the run juice:
```bash
$ go test -run=.*/check
--- FAIL: TestMutateBusinessDocument (0.00s)
    --- FAIL: TestMutateBusinessDocument/check_error_handling (0.00s)
        --- FAIL: TestMutateBusinessDocument/check_error_handling/business_document_is_nil (0.00s)
            main_test.go:85: Err must not be nil
        --- FAIL: TestMutateBusinessDocument/check_error_handling/newID_is_not_strictly_positive (0.00s)
            main_test.go:85: Err must not be nil
        --- FAIL: TestMutateBusinessDocument/check_error_handling/newName_is_empty (0.00s)
            main_test.go:85: Err must not be nil
FAIL
exit status 1
FAIL    acme.com/your_fun_business_app 0.002s
```
From the output, we can see the exact name of each failing test. The syntax is roughly `test_func/outer_test/inner_test`. This makes it
easy (or at least, less difficult) to find the failing test by just searching on its name. I also find the tree structure of the output way easier
to read than a bunch of lines that aren't indented.

# Benchmark tests

Here's where we can really give it the juice.

Benchmark tests are run against functions of the form `func BenchmarkXxx(*testing.B)` (again, from the official [Go docs](https://pkg.go.dev/testing#hdr-Benchmarks)).

For our example, we can write this benchmark test:
```go
func BenchmarkMutateBusinessDocument(b *testing.B) {
	bd := BusinessDocument{
		ID:   rand.Intn(100),
		Name: randomString(100),
	}

	// Reports memory allocations
	b.ReportAllocs()

	// Zeroes the benchmark timer. Use this if you have to do any testing setup beforehand.
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We can ignore errors. We're only interested in benchmarking its performance.
		_ = MutateBusinessDocument(&bd, 1, "new name")
	}
}
```
And run it with `go test -bench=BenchmarkMutateBusinessDocument` (or, `$ go test -bench=.`. The dot regex will run all benchmark tests.)

There's a fair amount to upack there. 
- The 'b' in `b *testing.B` is a struct in the `testing` library that holds some state about the specific 
benchmark test that is using it;
- `b.ReportAllocs()` records memory allocations and writes the output to stdout;
- `b.ResetTimer()` does exactly as advertised; it resets the benchmark timer. It should be used when there's any sort of setup
that needs to happen before you invoke your method under test (like instantiating a new struct, setting up DB mocks, etc.)
  - There are also `b.StopTimer()` and `b.StartTimer()` methods available. Those methods are intended for use _inside your benchmark for-loop_. If 
  you need to generate new data before each invocation of your method under test, then call `b.StopTimer()` -> do your setup -> `b.StartTimer()`.
  When setting up your test data _outside_ the for-loop, just use `b.ResetTimer()` on the line right above the loop declaration (as I did above).
- `b.N` (in the for-loop) represents the number of times that the testing framework will run your benchmark test. It is chosen
by the framework itself. You should not reassign it (not that you ever would).

Let's run our test:
```bash
$ go test -bench=.
goos: linux
goarch: amd64
pkg: acme.com/your_fun_business_app
cpu: Intel(R) Core(TM) i5-5200U CPU @ 2.20GHz
BenchmarkMutateBusinessDocument-4       1000000000               0.3855 ns/op          0 B/op          0 allocs/op
PASS
ok      acme.com/your_fun_business_app 0.433s
```

And discuss what we've got.
- **goos**: "...the running program's operating system target: one of darwin, freebsd, linux, and so on. To view possible combinations of GOOS and GOARCH, run `go tool dist list`". ([Go docs](https://golang.org/pkg/runtime/#pkg-constants))
- **goarch**: "GOARCH is the running program's architecture target: one of 386, amd64, arm, s390x, and so on." ([Go docs](https://golang.org/pkg/runtime/#pkg-constants))
- **pkg**: This one is a little misleading. It's not the name of the package that your code is in; rather, it's the name of the _module_, defined in your `go.mod` file.
- **cpu**: self-explanatory

After `cpu`, we can see that our `BenchmarkMutateBusinessDocument` ran a bunch of times. The number to the right of the name is the value of our `b.N` from earlier.
To the right of that, we see that our function under test ran at 0.3855 ns (not bad (???)). Next right is the number of bytes that our function allocated
per invocation, which in our case is zero (WOOOOOOO!!!). Last right is the number of times our function allocated memory per invocation, which is also zero.

(Note that here, the word "allocate" means "allocate on the heap." For a while I found this confusing, because isn't all memory _technically_ allocated _somewhere_?
But in the world of Go, when we want to write performant code, that means allocating as little heap memory as possible to cut down or 
altogether eliminate latency as a result of GC sweeps (among other optimizations). Thus, we should try to avoid allocations where possible. 
And remember: heap allocations are slow; stack allocations are fast.)

## Begin sidebar: heap allocations, and you

The best thing I ever read to understand heap allocations in Go is this ([again, from Go's docs](https://golang.org/doc/faq#stack_or_heap)):
```
...if the compiler cannot prove that the variable is not referenced after the 
function returns, then the compiler must allocate the variable on the 
garbage-collected heap to avoid dangling pointer errors.
```
Read Dave Cheney's [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gophercon-2019.html). It's awesome.
## End sidebar

With all of that being said, a lot of what I've written could be considered matters of taste.
- Maybe you think table tests are overkill
- Subtests seem like an odd way to group things. Why not just write new test functions?
- Why are we benchmarking such a simple API? 
- Is this just more [CI theater](https://arxiv.org/abs/1907.01602)?

I think a lot of that is fair. Even if you have 100% coverage, you might still not feel confident in your implementation
for any number of reasons. And no amount of unit and integration tests are going to replace user acceptance tests, or, you know,
the project manager's sign off.

These are my thoughts on testing in Go, which I humbly submit for your consideration.