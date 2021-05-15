# Lispy 
#### A simple lisp interpreter written in Go
Lispy is a tool and exercise to help me better understand lisp and more broadly functional programming. 

Spec
- [x] Arithmetic calculations with integers
- [x] Bindings to variables and state 
- [x] Logical and relational operators
- [x] Conditions
- [x] Reading Lispy code from a file
- [x] Functions
    - [x] Thoroughly tested
- [ ] Support for strings
- [ ] Local bindings via let
- [ ] Tail-optimized recursion
- [ ] Parsing multiple lines in the REPL


- Define new bindings to variables and functions with a universal keyword
`(define a 5)`
`(define function [name] () `

- Lists (in particular) and arrays are implemented as array under the hood for simplicity (as opposed to a List being implemented as a linked
list of cons cells in most Lisp dialects)



### Running Lispy
To run Lispy, you have two options.
1. You can launch a repl by running `make` in the outer directory
2. If you want to run a specific file, you can run `./run <path/to/file>`. For context, run is an executable with a small
script to run a passed in file. Note don't include the `<>` when passing a path (I included it for clarity).

There is no distinction between statements and expressions -> everything is an expression! A function declaration will return the name of the functiion. A function will return the last expression of the body.

Function bodies in Lisp consist of one Sexp. For example, this would throw an error. 
```
lispy > (define doMultipleThings [x] 
                (+ x x)
                (- x x)
        )

```

If you'd like to execute multiple expressions, wrap it in a do statement like this

```
lispy > (define doMultipleThings [x] 
                ( do
                    (+ x x)
                    (- x x)
                )
                
        )

```

