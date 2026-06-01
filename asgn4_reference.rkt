#lang typed/racket

(require typed/rackunit)

(define-type Env (Listof bind))
(struct bind ([name : Symbol] [value : Val]) #:transparent)
(define mt-env '())
(define extend-env cons)

(define-type ExprC (U NumC IdC AppC LamC ifC StringC))
(struct IdC ([id : Symbol]) #:transparent)
(struct StringC ([s : String])#:transparent)
(struct NumC ([n : Real]) #:transparent)
(struct LamC ([args : (Listof Symbol)] [body : ExprC]) #:transparent)
(struct AppC ([f : ExprC] [args : (Listof ExprC)]) #:transparent)
(struct ifC ([test : ExprC] [then : ExprC] [else : ExprC])#:transparent)


(define-type Val (U NumV CloV BoolV StringV PrimopV))
(struct NumV ([n : Real]) #:transparent)
(struct BoolV ([bool : Boolean]) #:transparent)
(struct StringV ([string : String]) #:transparent)
(struct CloV ([params : (Listof Symbol)] [body : ExprC] [env : Env]) #:transparent)
(struct PrimopV ([op : (U '+ '- '* '/ '<= 'equal? 'strlen 'substring 'error)]) #:transparent)

;defining the top-level environment
(define top-env (list (bind 'true (BoolV #t))
                      (bind 'false (BoolV #f))
                      (bind '+ (PrimopV '+))
                      (bind '- (PrimopV '-))
                      (bind '* (PrimopV '*))
                      (bind '/ (PrimopV '/))
                      (bind '<= (PrimopV '<=))
                      (bind 'equal? (PrimopV 'equal?))
                      (bind 'strlen (PrimopV 'strlen))
                      (bind 'substring (PrimopV 'substring))
                      (bind 'error (PrimopV 'error))))

;bunch of primitive operator functions
;(define (prim+ [l : Val] [r : Val]) : Val
;  (cond
;    [(and (NumV? l) (NumV? r)) (NumV (+ (NumV-n l) (NumV-n r)))]
;    [else (error 'prim+ "VEBG4: Invalid types")]))

;serialize
;takes a value and converts it into a string
(define (serialize [v : Val]) : String
  (match v
    [(NumV n) (~v n)]
    [(BoolV b) (if b "true" "false")]
    [(StringV s) (~v s)]
    [(CloV p b e) "#<procedure>"]
    [(PrimopV op) "#<primop>"]))

(check-equal? (serialize (NumV 3)) "3")
(check-equal? (serialize (BoolV #t)) "true")
(check-equal? (serialize (BoolV #f)) "false")
(check-equal? (serialize (StringV "hello")) "\"hello\"")
(check-equal? (serialize (CloV '(a) (NumC 0) mt-env)) "#<procedure>")
(check-equal? (serialize (PrimopV '+)) "#<primop>")

;zip
;takes two lists and zips them together into a list of binds
(define (zip [names : (Listof Symbol)] [values : (Listof Val)]) : (Listof bind)
  (match names
    ['() '()] 
    [(cons name r) (cons (bind name (first values)) (zip r (rest values)))]))

;env-lookup
;takes a symbol and returns its corresponding value in the environment
(define (env-lookup [for : Symbol] [env : Env]) : Val
  (cond
    [(empty? env) (error 'env-lookup "VEBG4: value not found: ~v" for)]
    [else (cond
            [(symbol=? for (bind-name (first env))) (bind-value (first env))]
            [else (env-lookup for (rest env))])]))

(define test-env (list (bind 'b (NumV 4)) (bind 'a (NumV 3))))
(check-equal? (env-lookup 'a test-env) (NumV 3))
(check-equal? (env-lookup 'b test-env) (NumV 4))
(check-exn #px"value not found"
           (lambda () (env-lookup 'c test-env)))

(check-equal? (extend-env (bind 'c (NumV 5)) test-env) (list (bind 'c (NumV 5)) (bind 'b (NumV 4)) (bind 'a (NumV 3))))

;interp
;evaluates an ExprC
(define (interp [expr : ExprC] [env : Env]) : Val
  (match expr
    [(NumC n) (NumV n)]
    [(IdC id) (env-lookup id env)]
    [(LamC params body) (CloV params body env)]
    [(StringC s) (StringV s)]
    [(ifC test then else) (define test-val (interp test env))
     (match test-val
      [(BoolV b)
       (cond
        [(equal? b #t) (interp then env)]
        [else (interp else env)])]
       [else (error 'interp "VEBG4 test condition not a predicate")])]
    [(AppC f args) (define funval (interp f env))
                   (define argvals (map (λ ([a : ExprC]) (interp a env)) args))
                   (match funval
                     [(CloV params body clo-env)
                      (cond
                        [(not (= (length params) (length argvals)))
                              (error 'interp "VEBG4: wrong number of arguments")]
                        [else (define binds (zip params argvals))
                              (define env2 (append binds clo-env))
                              (interp body env2)])]
                     ;[(CloV _ _ _ ) (error 'interp "VEBG4 wrong number of args to appc")]
                     [(PrimopV op)
                      (match (list op argvals)
                        [(list '+ (list (NumV l) (NumV r))) (NumV (+ l r))]
                        [(list '+ _) (error 'interp "VEBG4: + requires two numbers")]
                        [(list '- (list (NumV l) (NumV r))) (NumV (- l r))]
                        [(list '- _) (error 'interp "VEBG4: - requires two numbers")]
                        [(list '* (list (NumV l) (NumV r))) (NumV (* l r))]
                        [(list '* _) (error 'interp "VEBG4: * requires two numbers")]
                        [(list '/ (list (NumV l) (NumV r)))
                         (cond
                           [(= r 0) (error 'interp "VEBG4: division by 0 undefined")]
                           [else (NumV (/ l r))])]
                        [(list '/ _) (error 'interp "VEBG4: / requires two numbers")]
                        [(list '<= (list (NumV l) (NumV r))) (BoolV (<= l r))]
                        [(list '<= _) (error 'interp "VEBG4: <= requires two numbers")]
                        [(list'equal? (list (NumV l) (NumV r))) (BoolV(= l r))]
                        [(list'equal? (list (StringV l) (StringV r))) (BoolV(equal? l r))]
                        [(list'equal? (list (BoolV l) (BoolV r))) (BoolV(equal? l r))]
                        [(list'equal? (list a b)) (BoolV #f)]
                        [(list'equal? _) (error 'interp "VEBG wrong number of arguments for equal?")]
                        [(list 'strlen (list (StringV s))) (NumV (string-length s))]
                        [(list 'strlen _) (error 'interp "VEBG4 not a string")]
                        [(list 'substring (list (StringV s) (NumV start) (NumV stop)))
                         (cond
                           [(not (and (exact-nonnegative-integer? start) (exact-nonnegative-integer? stop)))
                            (error 'interp "VEBG4 substring called with non-naturals")]  
                           [(or (> start (string-length s)) (> stop (string-length s)))
                            (error 'interp "VEBG4 index out of bounds")]  
                           [(> start stop)
                            (error 'interp "VEBG4 stop before start")] 
                           [else (StringV (substring s
                                                     (assert start exact-nonnegative-integer?)
                                                     (assert stop exact-nonnegative-integer?)))])]
                        [(list 'substring _) (error 'interp "VEBG4 wrong number of arguments to substring")]
                        [(list 'error (list v)) (error 'interp "VEBG4 user-error ~e" (serialize v))])]
                     [other (error 'interp "VEBG4: incorrect function form")])]))



(check-equal? (interp (NumC 3) mt-env) (NumV 3))
(check-equal? (interp (IdC 'a) test-env) (NumV 3))
(check-equal?
 (interp
  (AppC (IdC 'f) (list (NumC 5)))
  (list (bind 'f (CloV '(x)
                       (IdC 'x)
                       mt-env))))
 (NumV 5))
(check-exn #px"incorrect function form"
           (lambda () (interp (AppC (NumC 3) (list (NumC 3))) mt-env)))

;parse
;turns a Sexpr into an ExprC to be interpreted
(define (parse [s : Sexp]) : ExprC
  (match s
    ;[(or '+ '- '* '/) (error 'parse "VEBG4: invalid operator position")]
    [(? real? n) (NumC n)]
    [(? symbol? s)
     (cond
       [(or (equal? s 'if) (equal? s 'fn)
            (equal? s 'given) (equal? s 'do)
            (equal? s '->) (equal? s '=)) (error 'parse "VEBG invalid identifier")]
       [else (IdC s)])]
    [(? string? s) (StringC s)]
    [(list 'if test then else) (ifC (parse test) (parse then) (parse else))]
    [(list 'if _ ...) (error 'parse "VEBG4: bad parameters for if")]
    [(list 'given (list (list ids '= exp) ...) 'do body)
     (cond
       [(contains-reserved? (cast ids (Listof Sexp))) (error 'parse "VEBG4: Invalid keyword")]
       [(= (length ids) (length (remove-duplicates ids)))
        (AppC
                                                     (LamC
                                                      (cast ids (Listof Symbol))
                                                      (parse body))
                                                      (map parse (cast exp (Listof Sexp))))]
       [else (error 'parse "VEBG4: Duplicate arguments")])]
    [(list 'given _ ...) (error 'parse "VEBG4: bad parameters for given")]
    [(list 'fn (list (? symbol? args) ...) '-> b)    
     (cond
       [(contains-reserved? (cast args (Listof Sexp))) (error 'parse "VEBG4: Invalid keyword")]
       [(= (length args) (length (remove-duplicates args)))
        (LamC (cast args (Listof Symbol)) (parse b))]
       [else (error 'parse "VEBG4: Duplicate arguments")])]
    [(list 'fn _ ...) (error 'parse "VEBG4: bad parameters for fn")]
    [(list fun args ...) (AppC (parse fun) (map parse (cast args (Listof Sexp))))]))


(define (contains-reserved? [l : (Listof Sexp) ])  : Boolean
       (not (empty? (filter (λ (s) (or (equal? s 'if)
                                       (equal? s 'fn)
                                       (equal? s 'given)
                                       (equal? s 'do)
                                       (equal? s '->)
                                       (equal? s '= ))) l))))


(check-equal? (parse '3) (NumC 3))
(check-equal? (parse 'x) (IdC 'x))
(check-equal? (parse '{abc 3}) (AppC (IdC 'abc) (list (NumC 3))))
(check-equal? (parse '{fn (a b c) -> {f 3}}) (LamC '(a b c) (AppC (IdC 'f) (list (NumC 3)))))
(check-equal? (parse '{+ 1 2}) (AppC (IdC '+) (list (NumC 1) (NumC 2))))
(check-exn #px"VEBG4"
           (lambda () (parse '{fn (fn fd fc) -> {fc 3}})))


;top interp
;takes an S expression written in VEBG4
;and returns the result of running the program
(define (top-interp [s : Sexp]) : String
  (serialize (interp (parse s) top-env)))


(check-equal? (top-interp '{{fn () -> 8}}) "8")
(check-equal? (top-interp '{{fn (x) -> x} 3}) "3")

;Correct use of +
(check-equal? (top-interp '{+ 1 2}) "3")
;Too many args + 
(check-exn #px"VEBG4"
           (lambda () (top-interp '{+ 1 2 2})))
;Too few args +
(check-exn #px"VEBG4"
           (lambda () (top-interp '{+ 1 })))
; Divison by 0 
(check-exn #px"VEBG4"
           (lambda () (top-interp '{/ 1 0})))


;Correct use of -
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {- x y}}) "2")
;Too many args -
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {- x y x}})))
;Too few args -
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {- x}})))


;Correct use of *
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {* x y}}) "15")
;Too many args *
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {* x y x}})))
;Too few args *
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {* x}})))


;Correct use of /
(check-equal? (top-interp '{given {[x = 6] [y = 3]} do {/ x y}}) "2")
;Too many args /
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {/ x y x}})))
;Too few args /
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {/ x}})))

;Correct use of <=
(check-equal? (top-interp '{given {[x = 6] [y = 3]} do {<= x y}}) "false")
;Too many args <=
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {<= x y x}})))
;Too few args <=
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {<= x}})))

;Correct use of equal?
(check-equal? (top-interp '{given {[x = 6] [y = 6]} do {equal? x y}}) "true")
(check-equal? (top-interp '{given {[x = "hello"] [y = "hello"]} do {equal? x y}}) "true")
(check-equal? (top-interp '{given {[x = true] [y = true]} do {equal? x y}}) "true")
(check-equal? (top-interp '{equal? {fn () -> 8} {fn () -> 8}}) "false")

;Too many args equal?
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {equal? x y x}})))
;Too few args equal?
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {equal? x}})))


;Correct use of strlen
(check-equal? (top-interp '{given {[x = "hello"]} do {strlen x}}) "5")
(check-equal? (top-interp '{given {[x = ""]} do {strlen x}}) "0")
(check-equal? (top-interp '{given {[x = " "]} do {strlen x}}) "1")

;Not a string, strlen
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 6]} do {strlen x}})))
(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = true]} do {strlen x}})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = true]} do {strlen x}})))



;Correct use of substring
(check-equal? (top-interp '{given {[x = "hello"]} do {substring x 0 2}}) "\"he\"" )
(check-equal? (top-interp '{given {[x = "hello"]} do {substring x 2 2}}) "\"\"" )

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 2 0}})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 1.5 3.5}})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 1 10}})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 1 5 1}})))





;Test 'if' and given form following to 'then' condition
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {if {<= y x} {+ x y} {* x y}}}) "8")

;Test 'if' and given form following to 'else' condition
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {if {<= x y} {+ x y} {* x y}}}) "15")

(check-exn #px"user-error"
           (lambda () (top-interp '{error 5})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {if {* x y} {+ x y} {* x y}}})))


;if condition not a predicate
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {<= x y}}) "false")


(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do { + if 4}})))


(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {if {+ x y} {* x y}}})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{{fn () -> 9} 17})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{{fn (x x) -> {+ x x}} 17})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{{fn x -> 5} 17})))


(check-exn #px"VEBG"
           (lambda () (top-interp '{given {x = 5 y = 3} do {if {+ x y} {* x y}}})))


(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[x = 5] [x = 3]} do {if {+ x y} {* x y}}})))

(check-exn #px"VEBG"
           (lambda () (top-interp '{given {[given = 5]} do {if {+ x y} {* x y}}})))