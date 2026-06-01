#lang typed/racket
(require typed/rackunit)

(define-type Env (Listof bind))
(struct bind ([name : Symbol] [value : Val]) #:transparent)
(define mt-env '())

(define-type ExprC (U NumC IdC AppC LamC ifC StringC))
(struct IdC ([id : Symbol]) #:transparent)
(struct StringC ([s : String]) #:transparent)
(struct NumC ([n : Real]) #:transparent)
(struct LamC ([args : (Listof Symbol)] [body : ExprC]) #:transparent)
(struct AppC ([f : ExprC] [args : (Listof ExprC)]) #:transparent)
(struct ifC ([test : ExprC] [then : ExprC] [else : ExprC]) #:transparent)


(define-type Val (U NumV CloV BoolV StringV PrimopV))
(struct NumV ([n : Real]) #:transparent)
(struct BoolV ([bool : Boolean]) #:transparent)
(struct StringV ([string : String]) #:transparent)
(struct CloV ([params : (Listof Symbol)] [body : ExprC] [env : Env]) #:transparent)

(define-type Primop (U '+ '- '* '/ '<= 'equal? 'strlen 'substring 'error))
(struct PrimopV ([op : Primop]) #:transparent)

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

;top-interp Sexp -> String
;parses then evaluates then serializes a VEBG4 program.
(define (top-interp [s : Sexp]) : String
  (serialize (interp (parse s) top-env)))

;serialize Val -> String
;Converts an interpreted value into a string
(define (serialize [v : Val]) : String
  (match v
    [(NumV n) (~v n)]
    [(BoolV b) (if b "true" "false")]
    [(StringV s) (~v s)]
    [(CloV p b e) "#<procedure>"]
    [(PrimopV op) "#<primop>"]))

;extend-env bind Env -> Env
;Adds one binding to the front of an environment.
(define extend-env cons)

;env-lookup Symbol Env -> Val
;Searches the environment for a symbol and returns the value it's bound to
(define (env-lookup [for : Symbol] [env : Env]) : Val
  (cond
    [(empty? env) (error 'env-lookup "VEBG4: value not found: ~v" for)]
    [else (cond
            [(symbol=? for (bind-name (first env))) (bind-value (first env))]
            [else (env-lookup for (rest env))])]))

;zip (Listof Symbol) (Listof Val) -> (Listof bind)
;Pairs parameter names with argument values to make new environment bindings.
(define (zip [names : (Listof Symbol)] [values : (Listof Val)]) : (Listof bind)
  (match names
    ['() '()]
    [(cons name r) (cons (bind name (first values)) (zip r (rest values)))]))

;check-arity Symbol (Listof Val) real -> (Listof Val)
;Checks that a primitive has the expected number of arguments
(define (check-arity [op : Symbol] [args : (Listof Val)] [expected : Real]) : (Listof Val)
  (cond
    [(= (length args) expected) args]
    [else (error 'interp "VEBG4: wrong number of arguments to ~a" op)]))

;prim+ (Listof Val) -> NumV
;Adds two numbers
(define (prim+ [args : (Listof Val)]) : NumV
  (match args
    [(list (NumV l) (NumV r)) (NumV (+ l r))]
    [_ (error 'interp "VEBG4: + requires numbers")]))

;prim- (Listof Val) -> NumV
;Subtracts two numbers
(define (prim- [args : (Listof Val)]) : NumV
  (match args
    [(list (NumV l) (NumV r)) (NumV (- l r))]
    [_ (error 'interp "VEBG4: - requires numbers")]))

;prim* (Listof Val) -> NumV
;Multiplies two numbers.
(define (prim* [args : (Listof Val)]) : NumV
  (match args
    [(list (NumV l) (NumV r)) (NumV (* l r))]
    [_ (error 'interp "VEBG4: * requires numbers")]))

;prim/ (Listof Val) -> NumV
;Divides the first number by the second.
(define (prim/ [args : (Listof Val)]) : NumV
  (match args
    [(list (NumV l) (NumV r))
     (cond
       [(= r 0) (error 'interp "VEBG4: division by 0 undefined")]
       [else (NumV (/ l r))])]
    [_ (error 'interp "VEBG4: / requires numbers")]))

;prim<= (Listof Val) -> BoolV
;checks if the first number is less than or equal to second.
(define (prim<= [args : (Listof Val)]) : BoolV
  (match args
    [(list (NumV l) (NumV r)) (BoolV (<= l r))]
    [_ (error 'interp "VEBG4: <= requires numbers")]))

;prim-equal? (Listof Val) -> BoolV
;checks if numbers, booleans, or strings are equal
(define (prim-equal? [args : (Listof Val)]) : BoolV
  (match args
    [(list (NumV l) (NumV r)) (BoolV (= l r))]
    [(list (StringV l) (StringV r)) (BoolV (equal? l r))]
    [(list (BoolV l) (BoolV r)) (BoolV (equal? l r))]
    [(list a b) (BoolV #f)]
    [_ (error 'interp "VEBG4: equal? requires two values")]))

;prim-strlen (Listof Val) -> NumV
;Finds length of a string
(define (prim-strlen [args : (Listof Val)]) : NumV
  (match args
    [(list (StringV s)) (NumV (string-length s))]
    [_ (error 'interp "VEBG4 not a string")]))

;prim-substring (Listof Val) -> StringV
;builds a substring given a stop and start index
(define (prim-substring [args : (Listof Val)]) : StringV
  (match args
    [(list (StringV s) (NumV start) (NumV stop))
     (cond
       [(not (and (exact-nonnegative-integer? start) (exact-nonnegative-integer? stop)))
        (error 'interp "VEBG4 substring called with non-naturals")]
       [(or (> start (string-length s)) (> stop (string-length s)))
        (error 'interp "VEBG4 index out of bounds")]
       [(> start stop)
        (error 'interp "VEBG4 stop before start")]
       [else (StringV (substring s start stop))])]
    [_ (error 'interp "VEBG4 substring called with bad argument types")]))

;prim-error (Listof Val) -> Val
;Raises a user error
(define (prim-error [args : (Listof Val)]) : Val
  (match args
    [(list v) (error 'interp "VEBG4 user-error ~e" (serialize v))]
    [_ (error 'interp "VEBG4 error requires one value")]))

;apply-primop Symbol (Listof Val) -> Val
;calls the relevant primative operator function depending on the operator
(define (apply-primop [op : Symbol] [args : (Listof Val)]) : Val
  (cond
    [(symbol=? op '+) (prim+ (check-arity '+ args 2))]
    [(symbol=? op '-) (prim- (check-arity '- args 2))]
    [(symbol=? op '*) (prim* (check-arity '* args 2))]
    [(symbol=? op '/) (prim/ (check-arity '/ args 2))]
    [(symbol=? op '<=) (prim<= (check-arity '<= args 2))]
    [(symbol=? op 'equal?) (prim-equal? (check-arity 'equal? args 2))]
    [(symbol=? op 'strlen) (prim-strlen (check-arity 'strlen args 1))]
    [(symbol=? op 'substring) (prim-substring (check-arity 'substring args 3))]
    [(symbol=? op 'error) (prim-error (check-arity 'error args 1))]
    [else (error 'interp "VEBG4: unknown primitive")]))

;interp ExprC Env -> Val
;interpretes an expression into a value
(define (interp [expr : ExprC] [env : Env]) : Val
  (match expr
    [(NumC n) (NumV n)]
    [(IdC id) (env-lookup id env)]
    [(LamC params body) (CloV params body env)]
    [(StringC s) (StringV s)]
    [(ifC test then else)
     (define test-val (interp test env))
     (match test-val
       [(BoolV b)
        (cond
          [(equal? b #t) (interp then env)]
          [else (interp else env)])]
       [else (error 'interp "VEBG4 test condition not a predicate")])]
    [(AppC f args)
     (define funval (interp f env))
     (define argvals (map (λ ([a : ExprC]) (interp a env)) args))
     (match funval
       [(CloV params body clo-env)
        (cond
          [(not (= (length params) (length argvals)))
           (error 'interp "VEBG4: wrong number of arguments")]
          [else (define binds (zip params argvals))
                (define env2 (append binds clo-env))
                (interp body env2)])]
       [(PrimopV op) (apply-primop op argvals)]
       [other (error 'interp "VEBG4: incorrect function form")])]))

;reserved-symbol? Sexp -> Boolean
;checks if a given Sexp is a VEBG4 reserved keyword
(define (reserved-symbol? [s : Sexp]) : Boolean
  (cond
    [(equal? s 'if) #t]
    [(equal? s 'fn) #t]
    [(equal? s 'given) #t]
    [(equal? s 'do) #t]
    [(equal? s '->) #t]
    [(equal? s '=) #t]
    [else #f]))

;contains-reserved? (Listof Sexp) -> Boolean
;checks if a list of Sexp contains VEBG4 reserved keywords
(define (contains-reserved? [l : (Listof Sexp)]) : Boolean
  (match l
    ['() #f]
    [(cons f r)
     (cond
       [(reserved-symbol? f) #t]
       [else (contains-reserved? r)])]))

;valid-ids? (Listof Sexp) -> Boolean
;Checks that a list of ids is not reserved and doesn't contain duplicates
(define (valid-ids? [ids : (Listof Sexp)]) : Boolean
  (cond
    [(contains-reserved? ids) #f]
    [(not (= (length ids) (length (remove-duplicates ids)))) #f]
    [else #t]))

;parse-fn (Listof Sexp) Sexp -> ExprC
;parses a fn form
(define (parse-fn [args : (Listof Sexp)] [body : Sexp]) : ExprC
  (cond
    [(valid-ids? args)
     (LamC (cast args (Listof Symbol)) (parse body))]
    [else (error 'parse "VEBG4: invalid or duplicate arguments")]))

;parse-given (Listof Sexp) (Listof Sexp) Sexp -> ExprC
;Desugars given form into AppC and lamC
(define (parse-given [ids : (Listof Sexp)] [exps : (Listof Sexp)] [body : Sexp]) : ExprC
  (cond
    [(valid-ids? ids)
     (AppC
      (LamC (cast ids (Listof Symbol)) (parse body))
      (map parse exps))]
    [else (error 'parse "VEBG4: invalid or duplicate arguments")]))

;parse Sexp -> ExprC
;Parses Sexp into AST ExprC's
(define (parse [s : Sexp]) : ExprC
  (match s
    [(? real? n) (NumC n)]
    [(? symbol? s)
     (cond
       [(reserved-symbol? s) (error 'parse "VEBG invalid identifier")]
       [else (IdC s)])]
    [(? string? s) (StringC s)]
    [(list 'if test then else) (ifC (parse test) (parse then) (parse else))]
    [(list 'if _ ...) (error 'parse "VEBG4: bad parameters for if")]
    [(list 'given (list (list ids '= exp) ...) 'do body)
     (parse-given (cast ids (Listof Sexp)) (cast exp (Listof Sexp)) body)]
    [(list 'given _ ...) (error 'parse "VEBG4: bad parameters for given")]
    [(list 'fn (list (? symbol? args) ...) '-> b)
     (parse-fn (cast args (Listof Sexp)) b)]
    [(list 'fn _ ...) (error 'parse "VEBG4: bad parameters for fn")]
    [(list fun args ...) (AppC (parse fun) (map parse (cast args (Listof Sexp))))]))


(define test-env (list (bind 'b (NumV 4)) (bind 'a (NumV 3))))


(check-equal? (serialize (NumV 3)) "3")
(check-equal? (serialize (BoolV #t)) "true")
(check-equal? (serialize (BoolV #f)) "false")
(check-equal? (serialize (StringV "hello")) "\"hello\"")
(check-equal? (serialize (CloV '(a) (NumC 0) mt-env)) "#<procedure>")
(check-equal? (serialize (PrimopV '+)) "#<primop>")

(check-equal? (zip '(a b) (list (NumV 1) (NumV 2)))
              (list (bind 'a (NumV 1)) (bind 'b (NumV 2))))

(check-equal? (env-lookup 'a test-env) (NumV 3))
(check-equal? (env-lookup 'b test-env) (NumV 4))
(check-exn #px"value not found: 'c"
           (lambda () (env-lookup 'c test-env)))

(check-equal? (extend-env (bind 'c (NumV 5)) test-env)
              (list (bind 'c (NumV 5)) (bind 'b (NumV 4)) (bind 'a (NumV 3))))

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

(check-equal? (parse '3) (NumC 3))
(check-equal? (parse 'x) (IdC 'x))
(check-equal? (parse '{abc 3}) (AppC (IdC 'abc) (list (NumC 3))))
(check-equal? (parse '{fn (a b c) -> {f 3}})
              (LamC '(a b c) (AppC (IdC 'f) (list (NumC 3)))))
(check-equal? (parse '{+ 1 2}) (AppC (IdC '+) (list (NumC 1) (NumC 2))))
(check-exn #px"invalid or duplicate arguments"
           (lambda () (parse '{fn (fn fd fc) -> {fc 3}})))

(check-equal? (top-interp '{{fn () -> 8}}) "8")
(check-equal? (top-interp '{{fn (x) -> x} 3}) "3")

;Correct use of +
(check-equal? (top-interp '{+ 1 2}) "3")
;Too many args +
(check-exn #px"wrong number of arguments to \\+"
           (lambda () (top-interp '{+ 1 2 2})))
;Too few args +
(check-exn #px"wrong number of arguments to \\+"
           (lambda () (top-interp '{+ 1})))
;Divison by 0
(check-exn #px"division by 0 undefined"
           (lambda () (top-interp '{/ 1 0})))

;Correct use of -
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {- x y}}) "2")
;Too many args -
(check-exn #px"wrong number of arguments to -"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {- x y x}})))
;Too few args -
(check-exn #px"wrong number of arguments to -"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {- x}})))

;Correct use of *
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {* x y}}) "15")
;Too many args *
(check-exn #px"wrong number of arguments to \\*"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {* x y x}})))
;Too few args *
(check-exn #px"wrong number of arguments to \\*"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {* x}})))

;Correct use of /
(check-equal? (top-interp '{given {[x = 6] [y = 3]} do {/ x y}}) "2")
;Too many args /
(check-exn #px"wrong number of arguments to /"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {/ x y x}})))
;Too few args /
(check-exn #px"wrong number of arguments to /"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {/ x}})))

;Correct use of <=
(check-equal? (top-interp '{given {[x = 6] [y = 3]} do {<= x y}}) "false")
;Too many args <=
(check-exn #px"wrong number of arguments to <="
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {<= x y x}})))
;Too few args <=
(check-exn #px"wrong number of arguments to <="
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {<= x}})))

;Correct use of equal?
(check-equal? (top-interp '{given {[x = 6] [y = 6]} do {equal? x y}}) "true")
(check-equal? (top-interp '{given {[x = "hello"] [y = "hello"]} do {equal? x y}}) "true")
(check-equal? (top-interp '{given {[x = true] [y = true]} do {equal? x y}}) "true")
(check-equal? (top-interp '{equal? {fn () -> 8} {fn () -> 8}}) "false")
;Too many args equal?
(check-exn #px"wrong number of arguments to equal\\?"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {equal? x y x}})))
;Too few args equal?
(check-exn #px"wrong number of arguments to equal\\?"
           (lambda () (top-interp '{given {[x = 6] [y = 3]} do {equal? x}})))

;Correct use of strlen
(check-equal? (top-interp '{given {[x = "hello"]} do {strlen x}}) "5")
(check-equal? (top-interp '{given {[x = ""]} do {strlen x}}) "0")
(check-equal? (top-interp '{given {[x = " "]} do {strlen x}}) "1")
;Not a string, strlen
(check-exn #px"not a string"
           (lambda () (top-interp '{given {[x = 6]} do {strlen x}})))
(check-exn #px"not a string"
           (lambda () (top-interp '{given {[x = true]} do {strlen x}})))

;Correct use of substring
(check-equal? (top-interp '{given {[x = "hello"]} do {substring x 0 2}}) "\"he\"")
(check-equal? (top-interp '{given {[x = "hello"]} do {substring x 2 2}}) "\"\"")
(check-exn #px"stop before start"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 2 0}})))
(check-exn #px"substring called with non-naturals"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 1.5 3.5}})))
(check-exn #px"index out of bounds"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 1 10}})))
(check-exn #px"wrong number of arguments to substring"
           (lambda () (top-interp '{given {[x = "Hello"]} do {substring x 1 5 1}})))

;Test 'if' and given form following to 'then' condition
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {if {<= y x} {+ x y} {* x y}}}) "8")
;Test 'if' and given form following to 'else' condition
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {if {<= x y} {+ x y} {* x y}}}) "15")

(check-exn #px"user-error \"5\""
           (lambda () (top-interp '{error 5})))

(check-exn #px"test condition not a predicate"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {if {* x y} {+ x y} {* x y}}})))

;if condition not a predicate
(check-equal? (top-interp '{given {[x = 5] [y = 3]} do {<= x y}}) "false")

(check-exn #px"invalid identifier"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {+ if 4}})))

(check-exn #px"bad parameters for if"
           (lambda () (top-interp '{given {[x = 5] [y = 3]} do {if {+ x y} {* x y}}})))

(check-exn #px"wrong number of arguments"
           (lambda () (top-interp '{{fn () -> 9} 17})))

(check-exn #px"invalid or duplicate arguments"
           (lambda () (top-interp '{{fn (x x) -> {+ x x}} 17})))

(check-exn #px"bad parameters for fn"
           (lambda () (top-interp '{{fn x -> 5} 17})))

(check-exn #px"bad parameters for given"
           (lambda () (top-interp '{given {x = 5 y = 3} do {if {+ x y} {* x y}}})))

(check-exn #px"invalid or duplicate arguments"
           (lambda () (top-interp '{given {[x = 5] [x = 3]} do {if {+ x y} {* x y}}})))

(check-exn #px"invalid or duplicate arguments"
           (lambda () (top-interp '{given {[given = 5]} do {if {+ x y} {* x y}}})))

(check-exn #px"invalid or duplicate arguments"
           (lambda () (top-interp '{given {[do = 5]} do {if {+ x y} {* x y}}})))

(check-exn #px"invalid or duplicate arguments"
           (lambda () (top-interp '{given {[-> = 5]} do {if {+ x y} {* x y}}})))


(check-exn #px"invalid or duplicate arguments"
           (lambda () (top-interp '{given {[= = 5]} do {if {+ x y} {* x y}}})))

(check-exn #px" \\+ requires numbers"
           (lambda () (top-interp '{given {[x = "hello"] [y = "why"]} do {+ x y}})))

(check-exn #px"- requires numbers"
           (lambda () (top-interp '{given {[x = "hello"] [y = "why"]} do {- x y}})))

(check-exn #px"/ requires numbers"
           (lambda () (top-interp '{given {[x = "hello"] [y = "why"]} do {/ x y}})))

(check-exn #px" \\* requires numbers"
           (lambda () (top-interp '{given {[x = "hello"] [y = "why"]} do {* x y}})))

(check-exn #px" <= requires numbers"
           (lambda () (top-interp '{given {[x = "hello"] [y = "why"]} do {<= x y}})))

(check-exn #px"wrong number of arguments to equal\\?"
           (lambda () (top-interp '{given {[x = "hello"] [y = "why"]} do {equal? x y x}})))

(check-exn #px"error requires one value"
           (lambda () (prim-error '())))

(check-exn #px"substring called with bad argument types"
           (lambda () (top-interp '{substring "hello" true 3})))

(check-exn #px"unknown primitive"
           (lambda () (apply-primop 'not-a-primop '())))

(check-exn #px"equal\\? requires two values"
           (lambda () (prim-equal? '())))
