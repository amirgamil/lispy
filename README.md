# Lispy 
#### A simple lisp interpreter written in Go
Lispy is a dialect of Lisp inspired by Scheme and Clojure that was built to help me better understand lisp and more broadly functional programming. 

Spec
- [x] Arithmetic calculations with integers
- [x] Bindings to variables and state 
- [x] Logical and relational operators
- [x] Conditions
- [x] Reading Lispy code from a file
- [x] Functions
    - [x] Thoroughly tested
- [x] Local bindings via let
- [x] Macros
- [ ] Tail-optimized recursion


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


#### Bindings via define
Bindings to variables and functions in Lispy are done via the keyword define
```
lispy> (define a 5)
#user/define
lispy> a
5
```

#### Conditionals via if
Conditionals are handled with if statments that follow this pattern
```
(if (cond) true false)
```
True and false are S-expressions. For example
```
(if (>= age 20) (println "You're no longer a kid :(" ) (println "hell yeh, you living the good life"))
```

#### Functions
Function bodies in Lisp consist of one Sexp. For example, this would throw an error. Parameters are passed via square brackets.
```
lispy> (define doMultipleThings [x] 
                (+ x x)
                (- x x)
        )

```

If you'd like to execute multiple expressions, wrap it in a do statement like this

```
lispy> (define doMultipleThings [x] 
                ( do
                    (+ x x)
                    (- x x)
                )
                
        )

```
Naturally, recursion follows from this definition nicely, so we can define a recursive function like this
```
lispy> (define fact [n] (if (= n 0) 1 (* n fact(- n 1))))
#user/fact
lispy> (fact 4)
24
```

Note, to differentiate functions from variables since both use the `define` keyword, you should always have square brackets `[]` after the function name even if it takes no parameters. 
```
lispy> (define noParams [] 
                ( do
                    (+ 1 1)
                    (- 1 1)
                )
                
        )
lispy> (define manyParams [a b c] 
                ( do
                    (+ a b)
                    (- b c)
                )
                
        )

```

#### Quotes
Define quotes in Lispy with the `'` 
```
lispy> '(1 2 3)
(1 2 3)
lispy> 'atom
atom
```

Remember () = call to function, so when passing to cons, car, cdr, make sure you have a quote


