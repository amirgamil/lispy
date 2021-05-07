# Lispy 
#### A simple lisp interpreter written in Go
Lispy is a tool and exercise to help me better understand lisp and more broadly functional programming. 

Spec
- [x] Arithmetic calculations with integers
- [x] Bindings to variables and state 
- [x] Logical and relational operators
- [x] Conditions
- [x] Functions
    - [ ] Thoroughly tested
- [ ] Refactor so bool, string have their own type in parser
- [ ] Support for strings
- [ ] Local bindings via let
- [ ] Tail-optimized recursion
- [ ] Reading input from a file
- [ ] Parsing multiple lines in the REPL


- Define new bindings to variables and functions with a universal keyword
`(define a 5)`
`(define function [name] () `

- Lists (in particular) and arrays are implemented as array under the hood for simplicity (as opposed to a List being implemented as a linked
list of cons cells in most Lisp dialects)