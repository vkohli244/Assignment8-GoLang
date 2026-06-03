package main
import "fmt"

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

func (i idC) isExpr()     {}
func (s StringC) isExpr() {}
func (n NumC) isExpr()    {}
func (l LamC) isExpr()    {}
func (a AppC) isExpr()    {}
func (i ifC) isExpr()     {}

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
func (num_ NumV) isVal()       {}
func (bool_ BoolV) isVal()     {}
func (string_ StringV) isVal() {}
func (c CloV) isVal()          {}
func (p PrimopV) isVal()       {}

// envLookup looks up a name in an environment and returns the value it's bound to
func envLookup(name string, env Env) (Val, error) {
	if len(env) == 0 {
		return nil, fmt.Errorf("value not found: %s", name)
	}
	if env[0].name == name {
		return env[0].value, nil
	}
	return envLookup(name, env[1:])
}

func main() {

}
