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
type NumC struct{ n int }
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

type NumV struct{ num_ int }
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
func (NumV) isVal()    {}
func (BoolV) isVal()   {}
func (StringV) isVal() {}
func (CloV) isVal()    {}
func (PrimopV) isVal() {}

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
		return fmt.Sprintf("VEBG8: unknown value in serialize (given %T)", v)
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

// checkArity checks that a primitive has the expected number of arguments.
func checkArity(op string, args []Val, expected int) ([]Val, error) {
	if len(args) == expected {
		return args, nil
	}
	return nil, fmt.Errorf("VEBG8: wrong number of arguments to %s", op)
}

// primPlus adds two numbers.
func primPlus(args []Val) (Val, error) {
	left, leftOk := args[0].(NumV)
	right, rightOk := args[1].(NumV)

	if leftOk && rightOk {
		return NumV{num_: left.num_ + right.num_}, nil
	}

	return nil, fmt.Errorf("VEBG8: + requires numbers, got %T and %T", args[0], args[1])
}

// primMinus subtracts the second number from the first.
func primMinus(args []Val) (Val, error) {
	left, leftOk := args[0].(NumV)
	right, rightOk := args[1].(NumV)

	if leftOk && rightOk {
		return NumV{num_: left.num_ - right.num_}, nil
	}

	return nil, fmt.Errorf("VEBG8: - requires numbers, got %T and %T", args[0], args[1])
}

// primMult multiplies two numbers.
func primMult(args []Val) (Val, error) {
	left, leftOk := args[0].(NumV)
	right, rightOk := args[1].(NumV)

	if leftOk && rightOk {
		return NumV{num_: left.num_ * right.num_}, nil
	}

	return nil, fmt.Errorf("VEBG8: * requires numbers, got %T and %T", args[0], args[1])
}

// primDiv divides the first number by  second
func primDiv(args []Val) (Val, error) {
	left, leftOk := args[0].(NumV)
	right, rightOk := args[1].(NumV)

	if leftOk && rightOk {
		if right.num_ == 0 {
			return nil, fmt.Errorf("VEBG8: division by 0 undefined")
		}
		return NumV{num_: left.num_ / right.num_}, nil
	}

	return nil, fmt.Errorf("VEBG8: / requires numbers, got %T and %T", args[0], args[1])
}

// primLessEqual checks if the first number is less than or equal to the second
func primLessEqual(args []Val) (Val, error) {
	left, leftOk := args[0].(NumV)
	right, rightOk := args[1].(NumV)

	if leftOk && rightOk {
		return BoolV{bool_: left.num_ <= right.num_}, nil
	}

	return nil, fmt.Errorf("VEBG8: <= requires numbers, got %T and %T", args[0], args[1])
}

// primEqual checks whether two numbers, strings, or booleans are equal
func primEqual(args []Val) (Val, error) {
	leftNum, leftNumOk := args[0].(NumV)
	rightNum, rightNumOk := args[1].(NumV)
	if leftNumOk && rightNumOk {
		return BoolV{bool_: leftNum.num_ == rightNum.num_}, nil
	}

	leftString, leftStringOk := args[0].(StringV)
	rightString, rightStringOk := args[1].(StringV)
	if leftStringOk && rightStringOk {
		return BoolV{bool_: leftString.string_ == rightString.string_}, nil
	}

	leftBool, leftBoolOk := args[0].(BoolV)
	rightBool, rightBoolOk := args[1].(BoolV)
	if leftBoolOk && rightBoolOk {
		return BoolV{bool_: leftBool.bool_ == rightBool.bool_}, nil
	}

	return BoolV{bool_: false}, nil
}

// primStrlen returns the length of a string as a number
func primStrlen(args []Val) (Val, error) {
	stringValue, isString := args[0].(StringV)
	if isString {
		return NumV{num_: len(stringValue.string_)}, nil
	}

	return nil, fmt.Errorf("VEBG8: not a string, got %T", args[0])
}

// primSubstring : []Val -> (Val, error)
// Builds a substring given a stop and start index
func primSubstring(args []Val) (Val, error) {
	str, StrOk := args[0].(StringV)
	start, StartOk := args[1].(NumV)
	stop, StopOk3 := args[2].(NumV)
	strLength  := len(str.string_)

	if StrOk && StartOk && StopOk3 {
		if start.num_ < 0 ||
			stop.num_ < 0 {
			return nil, fmt.Errorf("VEBG8 substring called with non-naturals, got %v and %v", start.num_, stop.num_)
		}

		if start.num_ > strLength ||
			stop.num_ > strLength {
			return nil, fmt.Errorf("VEBG8 index out of bounds, string length %d, got %v and %v", strLength, start.num_, stop.num_)
		}
		if start.num_ > stop.num_ {
			return nil, fmt.Errorf("VEBG8 stop before start, got start %v and stop %v", start.num_, stop.num_)
		}

		return StringV{
			string_: str.string_[start.num_:stop.num_],
		}, nil
	}

	return nil, fmt.Errorf("VEBG8 substring called with bad argument types, got %T, %T, and %T", args[0], args[1], args[2])
}

// primError : []Val -> (Val, error)
// Raises a user error
func primError(args []Val) (Val, error) {
	return nil, fmt.Errorf("VEBG4 user-error %s", serialize(args[0]))
}

// applyPrimop calls the relevant primitive operator function for an operator.
func applyPrimop(op string, args []Val) (Val, error) {
	switch op {
	case "+":
		checked, err := checkArity("+", args, 2)
		if err != nil {
			return nil, err
		}
		return primPlus(checked)
	case "-":
		checked, err := checkArity("-", args, 2)
		if err != nil {
			return nil, err
		}
		return primMinus(checked)
	case "*":
		checked, err := checkArity("*", args, 2)
		if err != nil {
			return nil, err
		}
		return primMult(checked)
	case "/":
		checked, err := checkArity("/", args, 2)
		if err != nil {
			return nil, err
		}
		return primDiv(checked)
	case "<=":
		checked, err := checkArity("<=", args, 2)
		if err != nil {
			return nil, err
		}
		return primLessEqual(checked)
	case "equal?":
		checked, err := checkArity("equal?", args, 2)
		if err != nil {
			return nil, err
		}
		return primEqual(checked)
	case "strlen":
		checked, err := checkArity("strlen", args, 1)
		if err != nil {
			return nil, err
		}
		return primStrlen(checked)
	case "substring":
		checked, err := checkArity("substring", args, 3)
		if err != nil {
			return nil, err
		}
		return primSubstring(checked)
	case "error":
		checked, err := checkArity("error", args, 1)
		if err != nil {
			return nil, err
		}
		return primError(checked)
	default:
		return nil, fmt.Errorf("VEBG8: unknown primitive")
	}
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
			return nil, fmt.Errorf("VEBG8: if test condition is not a predicate, instead got %T", e)
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
		case PrimopV:
			return applyPrimop(r.op, argvals)
		default:
			return nil, fmt.Errorf("VEBG8: application expected a function, instead got %T", fun_val)
		}

	default:
		return nil, fmt.Errorf("VEBG8: interp takes an ExprC, got %T", e)
	}

}

func main() {

}
