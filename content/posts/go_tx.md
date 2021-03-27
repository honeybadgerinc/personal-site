---
title: Struct transactions
date: "2021-03-20"
description: A process for 'preparing' and 'committing' data changes
tldr: Use a SQL-style PREPARE/BEGIN and COMMIT flow for mutating data that needs some kind of validation step
tags: [go, sql, programming]
---

## Background

We had a problem with an existing API. Imagine that, for a given struct type,
there existed methods that were meant to be used in a certain order to: 
1. Verify that, given your business rules, you are 'allowed' to mutate your struct with some given data, and
2. Actually do the mutation.

And let's assume that you must code defensively to handle requests that aren't allowed by your business rules.

For example, you have this:
```go
type Foo struct {
    propertyA PropertyA
    propertyB PropertyB
    ...
    propertyN PropertyN
}
```

and you have existing methods that look something like this:
```go
func (*f Foo) CanMutateFoo(ops []*Operation) error {
    
    /*
        Verify that for your given Foo and inputs, this mutation is allowed under
        your business rules.

        Return an error if some part of your validation fails.
    */
}

func (*f Foo) MutateFoo(ops []*Operation) error {
    
    /*
        Mutate your Foo with the given inputs.
    */
}
```

There's a problem here. For a new user of this API, it might not be obvious (especially in the absence of good documentation)
that they are supposed to use `CanMutateFoo(...)` _before_ `MutateFoo(...)`. They could just say "Okay I have a Foo, and 
there's a handy `MutateFoo(...)` thing here that I guess I can just use. Easy peasy." Then they'll apply the mutation 
without first checking that it's allowed and *boom*, you now have bad data. Or even if you know you don't want the validation step,
it still might not be clear that it's strictly optional.

My team lead had a great idea. He noticed that this is basically the same workflow as a 
database transaction. First, you _prepare_ your data (with some validation perhaps). Then, you _commit_ the write. ~~And finally, 
you have a big glass of wine.~~

So let's create a new type to contain our new notion of a 'Foo transaction':
```go
type FooTx struct {
    Foo *Foo
    txErr error
    Ops []*Operation
}
```

with the following methods: 
```go

func (foo *Foo) Begin() *FooTx {
    return &FooTx{
        Foo : foo
        Ops : make([]*Operation, 0)
    }
}

func (fooTx *FooTx) Prepare(ops ...*Operation) error {

    for _, op := range ops {
        /*
            Validate that the ops are allowed, and return an error if not.
        */
        fooTx.Ops = append(fooTx.Ops, op)
    }
    return nil
}

func (fooTx *FooTx) Commit() error {
    if fooTx.txErr != nil {
        return txErr
    } else {
        for _, op := range fooTx.ops {
            /*
                Mutate fooTx.foo in some way with each op
            */
        }
    }
    fooTx.txErr = errors.New("Cannot commit the same transaction twice")

    return nil
}
```

Putting it all together: 
```go

func DoStuff(ops ...*Operations) error {
    foo := Foo{
        /* Initialize Foo */ 
    }

    fooTx := foo.Begin()

    err := fooTx.Prepare(ops)
    if err != nil {
        return err
    }

    err = fooTx.Commit()
    if err != nil {
        return err
    }
    
    doStuffWithYourNewFoo(foo)
}
```

But what happens when you want to stage changes but you don't care (for whatever reason) if those changes
are 'valid'? ~~You should just give up, honestly~~ We can take our `FooTx` one step further:

```go

func (fooTx *FooTx) VerifyAndPrepare(ops ...*Operation) error {

    for _, op := range ops {
        /*
            Validate that the ops are allowed, and return an error if not.
        */
        if op.propertyA != AllowableProp {
            return errors.New("This data is crap.")
        }
        fooTx.Ops = append(fooTx.Ops, op)
    }
    return nil
}

func (fooTx *FooTx) PrepareOnly(ops ...*Operation) {
    /*
        No validation of any kind. Just show me the money.
    */
    fooTx.Ops = append(fooTx.Ops, ...ops)
}
```

That way, we can short-circuit our business rules if we _know_ that we
_definitely_ want to apply this write.

All credit to Blain Smith for showing me this pattern.