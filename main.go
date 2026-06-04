package main
//Need for serialize since serialize uses fmt.Println() and Sprintf()
import (
	"fmt"
)

// Bind structs contain string and Val
type Bind struct {
	name  string
	value Val
}

type Env []Bind // Env type are list of bindings
var top_env Env = []Bind{
	{"true", BoolV{bool_: true}},
	{"false", BoolV{bool_: false}},
	{"+", PrimopV{"+"}},
	{"-", PrimopV{"-"}},
	{"*", PrimopV{"*"}},
	{"/", PrimopV{"/"}},
	{"<=", PrimopV{"<="}},
	{"equal?", PrimopV{"equal?"}},
	{"strlen", PrimopV{"strlen"}},
	{"substring", PrimopV{"substring"}},
	{"error", PrimopV{"error"}}}

// we can define many structs to be apart of an interface, the interface has a method isVal() otherwise
// any struct is of type Val

type ExprC interface {
	isExpr()
}

type idC struct{ id string }
type StringC struct{ s string }
type NumC struct{ n float64 }
type LamC struct {
	args []string
	body ExprC
}
type AppC struct {
	f    ExprC
	args []ExprC
}
type ifC struct {
	test ExprC
	then ExprC
	els  ExprC
}

func (idC) isExpr()     {}
func (StringC) isExpr() {}
func (NumC) isExpr()    {}
func (LamC) isExpr()    {}
func (AppC) isExpr()    {}
func (ifC) isExpr()     {}

type Val interface {
	isVal()
}

type NumV struct{ num_ float64 }
type BoolV struct{ bool_ bool }
type StringV struct{ string_ string }
type PrimopV struct{ op string }

type CloV struct {
	params_ []string
	body_   ExprC
	env_    Env
}

// this is the only way to tell go that the structs belong to the interface
// In go, a struct satisifes an interface by implementing all of its methods
func (NumV) isVal()       {}
func (BoolV) isVal()     {}
func (StringV) isVal() {}
func (CloV) isVal()          {}
func (PrimopV) isVal()       {}
// serialize: Val -> string
// Converts the interpreted value into a string for printing
func serialize(v Val) string {
	switch v := v.(type) {
	case NumV:
		return fmt.Sprintf("%v", v.num_)
	case BoolV:
		if v.bool_ {
			return "true"
		}
		return "false"
	case StringV:
		return fmt.Sprintf("%q", v.string_)
	case CloV:
		return "#<procedure>"
	case PrimopV:
		return "#<primop>"
	default:
		return fmt.Sprintf("VEBG4: unknown value in serialize (given %T)", v)
	}
}



// envLookup looks up a name in an environment and returns the value it's bound to
func envLookup(name string, env Env) (Val, error) {
	if len(env) == 0 {
		return nil, fmt.Errorf("VEBG8 value not found: %s", name)
	}
	if env[0].name == name {
		return env[0].value, nil
	}
	return envLookup(name, env[1:])
}

// primEqual checks whether two numbers, strings, or booleans are equal
func primEqual(args []Val) (Val, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("equal? requires two values")
	}
	leftValue := args[0]
	rightValue := args[1]
	switch left := leftValue.(type) {
	case NumV:
		right, isNum := rightValue.(NumV)
		if !isNum {
			return BoolV{bool_: false}, nil
		}
		return BoolV{bool_: left.num_ == right.num_}, nil
	case StringV:
		right, isString := rightValue.(StringV)
		if !isString {
			return BoolV{bool_: false}, nil
		}
		return BoolV{bool_: left.string_ == right.string_}, nil
	case BoolV:
		right, isBool := rightValue.(BoolV)
		if !isBool {
			return BoolV{bool_: false}, nil
		}
		return BoolV{bool_: left.bool_ == right.bool_}, nil
	default:
		return BoolV{bool_: false}, nil
	}
}

// primStrlen returns the length of a string as a number
func primStrlen(args []Val) (Val, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strlen requires one value")
	}
	stringValue, isString := args[0].(StringV)
	if !isString {
		return nil, fmt.Errorf("not a string")
	}
	return NumV{num_: float64(len(stringValue.string_))}, nil
}

func zip(names []string, values []Val) Env {
	binds := make(Env, 0, len(names))

	for i := range names { // pythonic loop
		binds = append(binds, Bind{
			name:  names[i],
			value: values[i],
		})
	}

	return binds
}

func interp(e ExprC, env Env) (Val, error) {
	switch e := e.(type) {
	case NumC:
		return NumV{e.n}, nil

	case idC:
		return envLookup(e.id, env)

	case LamC:
		return CloV{params_: e.args, body_: e.body, env_: env}, nil

	case StringC:
		return StringV{e.s}, nil

	case ifC:
		test_val, err := interp(e.test, env)
		if err != nil {
			return nil, err
		}
		switch r := test_val.(type) {
		case BoolV:
			if r.bool_ {
				return interp(e.then, env)
			} else {
				return interp(e.els, env)
			}
		default:
			return nil, fmt.Errorf("VEBG4: if test condition is not a predicate, instead got %T", e)
		}
	case AppC:
		fun_val, err := interp(e.f, env)
		if err != nil {
			return nil, err
		}
		args := e.args
		argvals := make([]Val, 0, len(args)) // another way to do this argvals := []Val{} but this starts with a length of 0, we know we have at most len(argvals) elements
		//wrote the loop like this:  for i := 0; i < len(args); i++  gopls recommended modernizing
		for _, arg := range args { // the _ gives index, this is very pythonic way of writing loops
			val, err := interp(arg, env)
			if err != nil {
				return nil, err
			}
			argvals = append(argvals, val)
		}
		switch r := fun_val.(type) {
			case CloV:
				lenParams := len(r.params_)
				lenArgvals := len(argvals)
				if lenParams != lenArgvals {
					return nil, fmt.Errorf("VEBG8: wrong number of arguments: Expcted: %d, got: %d", lenParams, lenArgvals)
				} else {
					binds := zip(r.params_, argvals)
					env2 := append(binds, r.env_...)
					// without the elipsis operator append would attempt to put r.env_ as one element in binds,
					// but binds expects individual bindings, the "..." functions basically the exact same as the spread operator in js
					return interp(r.body_, env2)
				}
			default:
				return nil, fmt.Errorf("VEBG4: if test condition is not a predicate, instead got %T", e)
		}

	default:
		return nil, fmt.Errorf("VEBG4: interp takes an ExprC, got %T", e)
	}

}


// primSubstring : []Val -> Val
// Builds a substring given a stop and start index
func primSubstring(args []Val) Val {
   if len(args) != 3 {
       panic("VEBG4 substring called with bad argument types")
   }


   s, ok1 := args[0].(StringV)
   start, ok2 := args[1].(NumV)
   stop, ok3 := args[2].(NumV)


   if !ok1 || !ok2 || !ok3 {
       panic("VEBG4 substring called with bad argument types")
   }


   if start.num_ < 0 ||
       stop.num_ < 0 ||
       start.num_ != float64(int(start.num_)) ||
       stop.num_ != float64(int(stop.num_)) {


       panic("VEBG4 substring called with non-naturals")
   }


   if int(start.num_) > len(s.string_) ||
       int(stop.num_) > len(s.string_) {


       panic("VEBG4 index out of bounds")
   }


   if start.num_ > stop.num_ {
       panic("VEBG4 stop before start")
   }


   return StringV{
       string_: s.string_[int(start.num_):int(stop.num_)],
   }
}


// primError : []Val -> Val
// Raises a user error
func primError(args []Val) Val {
   if len(args) != 1 {
       panic("VEBG4 error requires one value")
   }
   panic("VEBG4 user-error " + serialize(args[0]))
}

func main() {
	
}