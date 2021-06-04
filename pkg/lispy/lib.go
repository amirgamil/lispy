//necessary evil to include non-go file in a package to expose the API other apps can use to run lispy code
const lib = `
(define caar [x] (car (car x)))
(define cadr [x] (car (cdr x)))
(define cdar [x] (cdr (car x)))
(define cddr [x] (cdr (cdr x)))


; basic expressions
(define sqrt [x] (# x 0.5))
(define square [x] (* x x))
(define inc [x] (+ x 1))
(define dec [x] (- x 1))
(define abs [x] 
    (if (>= x 0) x (* x -1))
)
(define neg [x] (- 0 x))
(define ! [x] (if x false true))
(define neg? [x] (< x 0))
(define pos? [x] (> x 0))
(define zero? [x] (= x 0))
(define divisible? [a b] (= (% a b) 0))
(define even? [x] (zero? (% x 2)))
(define odd? [x] (! (even? x)))
(define nil? [x] (= x ()))
(define list? [x] (= (type x) "list"))
(define int? [x] (= (type x) "int"))
(define float? [x] (= (type x) "float"))
(define symbol? [x] (= (type x) "symbol"))

; list methods
(define range [start stop step]
    (if (< start stop)
        (cons start (range (+ start step) stop step))
        ()
    )
)


(define reduce [arr func current]
    (if (nil? arr)
        current
        (reduce (cdr arr) func (func current (car arr)))
    )
)


(define max [arr]
    (if (nil? arr) 
        0
        (reduce arr (fn [a b] (if (< a b) b a)) (car arr))
    )
)


(define min [arr]
    (if (nil? arr) 
        0
        (reduce arr (fn [a b] (if (> a b) b a)) (car arr))
    )
)

(define sum [arr]
    (if (nil? arr)
        0
        (reduce arr + 0)
    )
)

; defines a list from 0...x-1 
(define seq [x] (range 0 x 1))


(define map [arr func] 
    (if (nil? arr)
        ()
        (cons (func (car arr)) (map (cdr arr) func))
    )
)

(define filter [arr func]
    (if (nil? arr)
        ()
        (if (func (car arr))
            (cons (car arr) (filter (cdr arr) func))
            (filter (cdr arr) func)
        )
    )
)

; O(n) operation, loop through entire list and add to end
(define append [arr el]
    (if (nil? arr)
        (list el)
        (cons (car arr) (append (cdr arr) el))
    )
)

; O(n^2) since each append is O(n)
(define reverse [arr]
    (if (nil? arr)
        ()
        (append (reverse (cdr arr)) (car arr))
    )
)


(define each [arr func]
    (if (nil? arr)
        ()
        (
            do
            (println (func (car arr)))
            (each (cdr arr) func)
        )
    )
)

; generate unique, not previously defined symbol
(define gensym []
    (symbol (str "var" (* 10000 (rand))))
)

; get nth item in list (0-indexed)
(define nth [arr n]
    (if (= n 0)
        (car arr)
        (nth (cdr arr) (dec n))
    )
)

; get size of list
(define size [arr]
    (do
        (define iterSize [n arr]
            (if (nil? arr)
                n
                (iterSize (inc n) (cdr arr))
            )
        )
        (iterSize 0 arr)
    )
)

; get index of item in list
(define index [arr item]
    (do 
        (define getIndex [index arr item]
            (if (nil? arr)
                -1
                (if (= (car arr) item)
                    index
                    (getIndex (inc index) (cdr arr) item)
                )
            )
        )
        (getIndex 0 arr item)
    )
)


; get last item in list
(define last [arr]
    (if (nil? (cdr arr))
        (car arr)
        (last (cdr arr))
    )
)

; appends arr2 to the end of arr1
(define join [arr1 arr2]
    (if (nil? arr2)
        arr1
        (join (append  arr1 (car arr2)) (cdr arr2))
    )
)

; adds element to the front of the array
(define addToFront [el arr]
    (do
        (define helper [arr]
            (if (nil? arr)
                ()
                (cons (car arr) (helper (cdr arr)))
            )
        )
        (helper (cons el arr))
    )
)


; macros


; (when (precondition) (postcondition))
(macro when [terms]
    (list 'if (car terms) (cadr terms))
)

; local bindings within lexical scope
(macro let [terms]
    (do
        (define declname (caar terms))
        (define declval (cdar terms))
        (define body (cdr terms))
        (list 
            (list 'fn [declname] body)
            declval
        )
        
    )
)
; ex: (quasiquote (1 2 (unquote (+ 3 4)))) => (1 2 7) 
; note, by design, don't include ' before it
(macro quasiquote [terms]
    ; note we do cons 'list so that map is called when evaluating the macro-expansion, not on the first call
    (cons 'list 
        (map (car terms)
            (fn [term] 
                (if (list? term)
                    (if (= (car term) 'unquote) 
                        (cadr term)
                        (list 'quasiquote term)
                    )
                    (list 'quote term)
                )
            )
         )
    ) 
)

; special form of a funcCall where the last argument is a list that should be treated as parameters
; e.g. (apply fn 1 2 (3 4))
(define apply [& terms]
    (do
        (define funcCall (car terms))
        (define helper [args]
            (if (nil? args)
                ()
                (if (list? (car args))
                    (cons (caar args) (helper (cdar args)))
                    (cons (car args) (helper (cdr args)))
                )
            )
        )

        (applyTo funcCall (helper (cdr terms)))
    )
)


; (cond (precondition) (postcondition) (precondition2) (postcondition2)...)
(macro cond [terms]
    (if (nil? terms)
        ()
        (list 'if (car terms) (cadr terms) (cons 'cond (cddr terms)))
    )
)


;(switch val (case1 result1) (case2 result2))
(macro switch [statements]
    (do
        (define val (gensym))
        (define match [conditions]
            (if (nil? conditions)
                (list)
                (list 'if (list '= val (caar conditions)) (cdar conditions) (match (cdr conditions)))
            )
        )
        (let (val (car statements))
            (match (cdr statements))
        )
    )
)

; thread-first
; inserts first form as the first argument (second in list) of the second form, and so forth
(macro -> [terms]
    (do
        (define apply-partials [partials expr]
            (if (nil? partials)
                expr
                (if (symbol? (car partials))
                    (list (car partials) (apply-partials (cdr partials) expr))
                    ; if it's a list with other parameters, insert expr (recursive call) 
                    ; as second parameter into partial (note need to use cons to ensure same list for func args)
                    (cons (caar partials) (cons (apply-partials (cdr partials) expr) (cdar partials)))
                )
            )
        )
        (apply-partials (reverse (cdr terms)) (car terms))
    )
)

; thread-last
; same as -> but inserts first form as last argument (last in list) of second form, and so forth
(macro ->> [terms]
    (do
        (define apply-partials [partials expr]
            (if (nil? partials)
                expr
                (if (symbol? (car partials))
                    (list (car partials) (apply-partials (cdr partials) expr))
                    ; if it's a list with other parameters, insert expr (recursive call) 
                    ; as last form 
                    (cons (caar partials) (append (cdar partials) (apply-partials (cdr partials) expr)))
                )
            )
        )
        (apply-partials (reverse (cdr terms)) (car terms))
    )
)



; immutable key-value hashmap 
; O(n) lookup with O(1) insert
; ex: (comp "key1" "val1" "key2" "val2")
(macro hash-map [terms]
    (if (nil? terms)
        ()
        (list 'cons (list 'cons (car terms) (cadr terms)) (cons 'hash-map (cddr terms)))
    )
)

; O(n) recursive lookup
(define get [hm key]
    (if (nil? hm)
        ()
        (if (= key (caar hm))
            (car (cdar hm))
            (get (cdr hm) key)
        )
    )
)


; hash-maps are immutable, add returns a new hash-map with the new key, val pair 
; it does not modify existing ones
(define add [hm key val]
    (if (nil? (get hm key))
        (cons (cons key val) hm)
    )
)

; remove returns a hash-map with the key-value pair provided removed (if it exists in the hash-map)
(define remove [hm key]
    (do
        (define val (get hm key))
        (define helper [hm]
            (if (nil? hm)
                ()
                (if (= (car (cdar hm)) val)
                    (helper (cdr hm))
                    (cons (car hm) (helper (cdr hm)))
                )
            )
        )
        (helper hm)
    )
)

; return list of keys in the hash-map
(define keys [hm]
    (if (nil? hm)
        ()
        (cons (caar hm) (keys (cdr hm)))
    )
)

; return list of values in the hash-map
(define values [hm]
    (if (nil? hm)
        ()
        (cons (car (cdar hm)) (values (cdr hm)))
    )
)

`