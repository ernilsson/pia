package squeak

import (
	"errors"
	"fmt"
	"github.com/ernilsson/pia/squeak/ast"
	"github.com/ernilsson/pia/squeak/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

var (
	ErrRuntimeFault            = errors.New("runtime error")
	ErrNotCallable             = fmt.Errorf("%w: not callable", ErrRuntimeFault)
	ErrUnrecognizedExpression  = fmt.Errorf("%w: unrecognized expression", ErrRuntimeFault)
	ErrUnrecognizedStatement   = fmt.Errorf("%w: unrecognized statement", ErrRuntimeFault)
	ErrUnrecognizedOperator    = fmt.Errorf("%w: unrecognized operator", ErrRuntimeFault)
	ErrObjectNotDeclared       = fmt.Errorf("%w: variable not declared", ErrRuntimeFault)
	ErrUnrecognizedOperandType = fmt.Errorf("%w: unrecognized operand type", ErrRuntimeFault)
	ErrIllegalArgument         = fmt.Errorf("%w: illegal argument", ErrRuntimeFault)
	ErrIllegalOperation        = fmt.Errorf("%w: illegal operation", ErrRuntimeFault)
	ErrFailedAssertion         = fmt.Errorf("%w: assertion failed", ErrRuntimeFault)
)

var ()

type unwinder struct {
	source token.Token
	value  Object
}

type EnvironmentOpt func(*Environment)

func Parent(parent *Environment) EnvironmentOpt {
	return func(env *Environment) {
		env.parent = parent
	}
}

func Prefill(k string, v Object) EnvironmentOpt {
	return func(env *Environment) {
		if env.tbl == nil {
			env.tbl = make(map[string]Object)
		}
		env.tbl[k] = v
	}
}

func NewEnvironment(opts ...EnvironmentOpt) *Environment {
	env := &Environment{
		tbl: make(map[string]Object),
	}
	for _, opt := range opts {
		opt(env)
	}
	return env
}

// Environment is a table of contents for runtime variables that exposes an API to interface with the current
// environment correctly. It also supports the concepts of hierarchical environments which enables scoping of variables.
type Environment struct {
	parent *Environment
	tbl    map[string]Object
}

// Resolve returns the current value stored within the environment for the provided key. If the key cannot be resolved
// for the immediate scope (the table of variables that is stored within the environment) then the parent environment is
// invoked to resolve the same key within its immediate scope. This call chain continues until the key is successfully
// resolved or the next parent is a nil value, in which case a non-nil error is returned.
func (env *Environment) Resolve(k string, lvl int) (Object, error) {
	scope := env
	for range lvl {
		scope = scope.parent
	}
	obj, ok := scope.tbl[k]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrObjectNotDeclared, k)
	}
	return obj, nil
}

// Declare sets the provided value for the provided key in the immediate scope.
func (env *Environment) Declare(k string, v Object) {
	env.tbl[k] = v
}

// Assign sets a new value for an already declared variable in the immediate scope. If the key cannot be resolved in the
// immediate scope then the parent is invoked. This call chain continues until the assignment is successful or until the
// next parent is nil, in which case it returns a non-nil error.
func (env *Environment) Assign(k string, v Object, lvl int) error {
	scope := env
	for range lvl {
		scope = scope.parent
	}
	_, ok := scope.tbl[k]
	if !ok {
		return fmt.Errorf("%w: %s", ErrObjectNotDeclared, k)
	}
	scope.tbl[k] = v
	return nil
}

// NewInterpreter constructs an interpreter with a prefilled runtime in its global scope. The caller is responsible for
// supplying a valid [io.Writer] which is used as the standard output stream. While the caller is allowed to provide a
// nil [io.Writer], it is discouraged as any usage of the standard output will result in a panic.
func NewInterpreter(wd string, out io.Writer) *Interpreter {
	runtime := NewEnvironment(
		Prefill("print", PrintBuiltin{}),
		Prefill("println", PrintlnBuiltin{}),
		Prefill("clone", CloneBuiltin{}),
		Prefill("panic", PanicBuiltin{}),
		Prefill("assert", AssertBuiltin{}),
	)
	global := NewEnvironment(Parent(runtime))
	return &Interpreter{
		wd:      wd,
		exports: make(map[string]Object),
		runtime: runtime,
		global:  global,
		scope:   global,
		out:     out,
	}
}

type Interpreter struct {
	wd      string
	exports map[string]Object
	runtime *Environment
	global  *Environment
	scope   *Environment
	out     io.Writer
}

func (in *Interpreter) Execute(program []ast.StatementNode) error {
	for _, stmt := range program {
		uw, err := in.execute(stmt)
		if err != nil {
			return err
		}
		if uw != nil {
			// Unwinders should never bubble up all the way here, if they have then that means the unwinder never passed
			// through a caller that could correctly handle it, which is considered an erroneous state.
			return fmt.Errorf("%w: unexpected unwinder: %s", ErrRuntimeFault, uw.source.Lexeme)
		}
	}
	return nil
}

func (in *Interpreter) Declare(name string, obj Object) {
	in.runtime.Declare(name, obj)
}

func (in *Interpreter) Resolve(name string, lvl int) (Object, bool) {
	obj, err := in.scope.Resolve(name, lvl)
	return obj, err == nil
}

// execute runs the provided statement node within the current context of the interpreter. Statements do not generally
// evaluate to a value. Some statements such as [ast.Return] changes the control flow drastically, those cases are not
// handled by this method. Instead, whenever an unwinding statement is encountered then a non-nil value of unwinder is
// returned which is expected to be processed properly by some caller in the call stack.
func (in *Interpreter) execute(stmt ast.StatementNode) (*unwinder, error) {
	switch stmt := stmt.(type) {
	case ast.ExpressionStatement:
		_, err := in.evaluate(stmt.Expression)
		return nil, err
	case ast.Declaration:
		return nil, in.declaration(stmt)
	case ast.Block:
		return in.block(NewEnvironment(Parent(in.scope)), stmt.Body)
	case ast.If:
		return in.branching(stmt)
	case ast.While:
		return in.loop(stmt)
	case ast.Noop:
		// In the future it might be a good idea to restructure the AST so that it does not contain any [ast.Noop].
		return nil, nil
	case ast.Function:
		in.scope.Declare(stmt.Name.Lexeme, Function{
			declaration: stmt,
			closure:     in.scope,
		})
		return nil, nil
	case ast.Return, ast.Break, ast.Continue:
		// Perhaps these three types, which all share the common behaviour of unwinding the call stack of the
		// interpreter in one way or the other should be grouped with another 'subtype' of [ast.Statement] such as for
		// example ast.Unwinder. However, for as long as we only have three of these statement types I do not see any
		// harm in handling them directly like we do now rather than abstracting things away. On the contrary, I believe
		// that abstracting it away prematurely would just cause confusion.
		return in.unwinder(stmt)
	case ast.Import:
		fn, err := in.evaluate(stmt.Source)
		if err != nil {
			return nil, err
		}
		if _, ok := fn.(String); !ok {
			return nil, fmt.Errorf("%w: %s is not a valid import value", ErrIllegalArgument, fn.String())
		}
		loc := filepath.Join(in.wd, fn.String())
		// TODO: Here is where any special Pia imports should be handled.
		src, err := os.ReadFile(loc)
		if err != nil {
			return nil, err
		}
		stmts, err := ParseString(string(src))
		if err != nil {
			return nil, err
		}
		child := NewInterpreter(filepath.Dir(loc), in.out)
		if err := child.Execute(stmts); err != nil {
			return nil, err
		}
		// Declare any exported variables to the current scope. This allows users to limit the scope in which exported
		// variables exist on the importing side.
		for k, v := range child.exports {
			in.scope.Declare(k, v)
		}
		return nil, nil
	case ast.Export:
		val, err := in.evaluate(stmt.Value)
		if err != nil {
			return nil, err
		}
		in.exports[stmt.Name.Lexeme] = val
		return nil, nil
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnrecognizedStatement, stmt)
	}
}

func (in *Interpreter) declaration(stmt ast.Declaration) error {
	if stmt.Initializer == nil {
		in.scope.Declare(stmt.Name.Lexeme, nil)
		return nil
	}
	val, err := in.evaluate(stmt.Initializer)
	if err != nil {
		return err
	}
	in.scope.Declare(stmt.Name.Lexeme, val)
	return nil
}

func (in *Interpreter) branching(stmt ast.If) (*unwinder, error) {
	cnd, err := in.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	if in.truthy(cnd) {
		return in.execute(stmt.Then)
	}
	if stmt.Else != nil {
		return in.execute(stmt.Else)
	}
	return nil, nil
}

func (in *Interpreter) loop(stmt ast.While) (*unwinder, error) {
	cnd, err := in.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	for in.truthy(cnd) {
		uw, err := in.execute(stmt.Body)
		if err != nil {
			return nil, err
		}
		if uw != nil {
			if uw.source.Type == token.Break {
				break
			}
			if uw.source.Type != token.Continue {
				return uw, nil
			}
		}
		cnd, err = in.evaluate(stmt.Condition)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (in *Interpreter) unwinder(stmt ast.StatementNode) (*unwinder, error) {
	switch stmt := stmt.(type) {
	case ast.Return:
		uw := &unwinder{
			source: token.Token{
				Type:   token.Return,
				Lexeme: "return",
			},
		}
		if stmt.Expression == nil {
			return uw, nil
		}
		// The return statement is the only unwinding statement that can also evaluate an evaluate which shall be
		// returned to the caller. If the return statement does indeed contain an evaluate then it is set on the value
		// field in the unwinder value.
		val, err := in.evaluate(stmt.Expression)
		if err != nil {
			return nil, err
		}
		uw.value = val
		return uw, nil
	case ast.Break:
		return &unwinder{
			source: token.Token{
				Type:   token.Break,
				Lexeme: "break",
			},
		}, nil
	case ast.Continue:
		return &unwinder{
			source: token.Token{
				Type:   token.Continue,
				Lexeme: "continue",
			},
		}, nil
	default:
		panic(fmt.Errorf("unwinder executor called with invalid statement type: %T", stmt))
	}
}

func (in *Interpreter) block(scope *Environment, block []ast.StatementNode) (*unwinder, error) {
	prev := in.scope
	defer func() {
		in.scope = prev
	}()
	in.scope = scope
	for _, stmt := range block {
		uw, err := in.execute(stmt)
		if err != nil {
			return nil, err
		}
		if uw != nil {
			return uw, nil
		}
	}
	return nil, nil
}

func (in *Interpreter) evaluate(expr ast.ExpressionNode) (Object, error) {
	switch expr := expr.(type) {
	case ast.IntegerLiteral:
		return Number{float64(expr.Integer)}, nil
	case ast.FloatLiteral:
		return Number{expr.Float}, nil
	case ast.StringLiteral:
		return String{expr.String}, nil
	case ast.BooleanLiteral:
		return Boolean{expr.Boolean}, nil
	case ast.NilLiteral:
		return nil, nil
	case ast.ListLiteral:
		return in.list(expr)
	case ast.Grouping:
		return in.evaluate(expr.Group)
	case ast.Prefix:
		return in.prefix(expr)
	case ast.Infix:
		return in.infix(expr)
	case ast.Variable:
		return in.scope.Resolve(expr.Name.Lexeme, expr.Level)
	case ast.Assignment:
		val, err := in.evaluate(expr.Value)
		if err != nil {
			return nil, err
		}
		if err := in.scope.Assign(expr.Name.Lexeme, val, expr.Level); err != nil {
			return nil, err
		}
		return val, nil
	case ast.SetProp:
		return in.setProp(expr)
	case ast.SetIndex:
		return in.setIndex(expr)
	case ast.Logical:
		return in.logical(expr)
	case ast.Call:
		return in.call(expr)
	case ast.GetProp:
		return in.getProp(expr)
	case ast.GetIndex:
		return in.getIndex(expr)
	case ast.ObjectLiteral:
		obj := &ObjectInstance{
			Properties: make(map[string]Object),
		}
		for k, v := range expr.Properties {
			val, err := in.evaluate(v)
			if err != nil {
				return nil, err
			}
			obj.Properties[k] = val
		}
		return obj, nil
	case ast.Method:
		return &ObjectInstanceMethod{declaration: expr}, nil
	default:
		return nil, fmt.Errorf("%w: %T", ErrUnrecognizedExpression, expr)
	}
}

func (in *Interpreter) list(node ast.ListLiteral) (Object, error) {
	items := make([]Object, 0, len(node.Items))
	for _, expr := range node.Items {
		item, err := in.evaluate(expr)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &List{slice: items}, nil
}

func (in *Interpreter) getProp(node ast.GetProp) (Object, error) {
	obj, err := in.evaluate(node.Target)
	if err != nil {
		return nil, err
	}
	i, ok := obj.(Instance)
	if !ok {
		return nil, fmt.Errorf("%w: %T cannot invoke property getter", ErrIllegalArgument, obj)
	}
	// If the property does not exist on the instance then a nil value is returned. This allows the users to do
	// presence checks using the getter as an expression.
	p := i.Get(node.Property.Lexeme)
	switch p := p.(type) {
	case Method:
		return p.Bind(i)
	default:
		return p, nil
	}
}

func (in *Interpreter) getIndex(node ast.GetIndex) (Object, error) {
	obj, err := in.evaluate(node.Target)
	if err != nil {
		return nil, err
	}
	l, ok := obj.(*List)
	if !ok {
		return nil, fmt.Errorf("%w: %T cannot invoke indexing", ErrIllegalArgument, obj)
	}
	val, err := in.evaluate(node.Index)
	if err != nil {
		return nil, err
	}
	idx, ok := val.(Number)
	if !ok {
		return nil, fmt.Errorf("%w: %T cannot be used as index", ErrIllegalArgument, val)
	}
	if idx.value < 0 || int(idx.value) >= len(l.slice) {
		return nil, fmt.Errorf("%w: index out of range", ErrIllegalArgument)
	}
	return l.slice[int(idx.value)], nil
}

func (in *Interpreter) setIndex(expr ast.SetIndex) (Object, error) {
	val, err := in.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	obj, err := in.evaluate(expr.Target.Target)
	if err != nil {
		return nil, err
	}
	switch obj := obj.(type) {
	case *List:
		arg, err := in.evaluate(expr.Target.Index)
		if err != nil {
			return nil, err
		}
		index, ok := arg.(Number)
		if !ok {
			return nil, fmt.Errorf("%w: list index must be number", ErrIllegalArgument)
		}
		if int(index.value) >= len(obj.slice) || index.value < 0 {
			return nil, fmt.Errorf("%w: %d index is out of range", ErrIllegalArgument, int(index.value))
		}
		obj.slice[int(index.value)] = val
		return val, nil
	default:
		return nil, fmt.Errorf(
			"%w: %T cannot store indexed items",
			ErrIllegalArgument,
			obj,
		)
	}
}

func (in *Interpreter) setProp(expr ast.SetProp) (Object, error) {
	val, err := in.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	obj, err := in.evaluate(expr.Target.Target)
	if err != nil {
		return nil, err
	}
	switch obj := obj.(type) {
	case Instance:
		return obj.Put(expr.Property.Lexeme, val), nil
	default:
		return nil, fmt.Errorf(
			"%w: %T cannot invoke property setter",
			ErrIllegalArgument,
			obj,
		)
	}
}

func (in *Interpreter) call(node ast.Call) (Object, error) {
	switch node.Operator.Type {
	case token.LeftParenthesis:
		return in.function(node)
	default:
		return nil, fmt.Errorf("%w: %s is not a call operator", ErrUnrecognizedOperator, node.Operator.Lexeme)
	}
}

func (in *Interpreter) function(node ast.Call) (Object, error) {
	fn, err := in.evaluate(node.Callee)
	if err != nil {
		return nil, err
	}
	switch fn := fn.(type) {
	case Callable:
		if fn.Arity() != len(node.Args) {
			return nil, fmt.Errorf(
				"function accepts %d parameters but was provided %d arguments",
				fn.Arity(),
				len(node.Args),
			)
		}
		var args []Object
		for _, expr := range node.Args {
			arg, err := in.evaluate(expr)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return fn.Call(in, args...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrNotCallable, fn)
	}
}

func (in *Interpreter) logical(node ast.Logical) (Object, error) {
	left, err := in.evaluate(node.LHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.And:
		if in.falsy(left) {
			return left, nil
		}
		return in.evaluate(node.RHS)
	case token.Or:
		if in.truthy(left) {
			return left, nil
		}
		return in.evaluate(node.RHS)
	default:
		return nil, fmt.Errorf("%w: %s as logical operator", ErrUnrecognizedOperator, node.Operator.Lexeme)
	}
}

func (in *Interpreter) prefix(node ast.Prefix) (Object, error) {
	obj, err := in.evaluate(node.Target)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Bang:
		return Boolean{!in.truthy(obj)}, nil
	case token.Minus:
		return in.multiply(obj, Number{-1})
	default:
		return nil, fmt.Errorf("%w: %s as prefix operator", ErrUnrecognizedOperator, node.Operator.Lexeme)
	}
}

func (in *Interpreter) truthy(obj Object) bool {
	if obj == nil {
		return false
	}
	switch obj := obj.(type) {
	case Boolean:
		return obj.value
	default:
		return true
	}
}

func (in *Interpreter) falsy(obj Object) bool {
	return !in.truthy(obj)
}

func (in *Interpreter) infix(node ast.Infix) (Object, error) {
	lhs, err := in.evaluate(node.LHS)
	if err != nil {
		return nil, err
	}
	rhs, err := in.evaluate(node.RHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Plus:
		// Addition evaluation lets the left hand side evaluate operand control whether the addition should be
		// considered a concatenation or an addition of numbers.
		switch lhs := lhs.(type) {
		case String:
			return in.concat(lhs, rhs)
		case Number:
			return in.add(lhs, rhs)
		default:
			return nil, fmt.Errorf(
				"%w: cannot add %T",
				ErrUnrecognizedOperandType,
				lhs,
			)
		}
	case token.Minus:
		return in.subtract(lhs, rhs)
	case token.Asterisk:
		return in.multiply(lhs, rhs)
	case token.Slash:
		return in.divide(lhs, rhs)
	case token.Less:
		return in.isLessThan(lhs, rhs)
	case token.LessEqual:
		lt, err := in.isLessThan(lhs, rhs)
		if err != nil {
			return nil, err
		}
		if lt.value {
			return lt, nil
		}
		return in.isEqual(lhs, rhs)
	case token.Greater:
		return in.isGreaterThan(lhs, rhs)
	case token.GreaterEqual:
		gt, err := in.isGreaterThan(lhs, rhs)
		if err != nil {
			return nil, err
		}
		if gt.value {
			return gt, nil
		}
		return in.isEqual(lhs, rhs)
	case token.Equals:
		return in.isEqual(lhs, rhs)
	case token.NotEquals:
		eq, err := in.isEqual(lhs, rhs)
		eq.value = !eq.value
		return eq, err
	default:
		return nil, fmt.Errorf("%w: %s as infix operator", ErrUnrecognizedOperator, node.Operator.Lexeme)
	}
}

func (in *Interpreter) concat(lhs, rhs Object) (String, error) {
	lhn, ok := lhs.(String)
	if !ok {
		return String{}, fmt.Errorf("%w: %T is not a String", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(String)
	if !ok {
		return String{}, fmt.Errorf("%w: %T is not a String", ErrUnrecognizedOperandType, rhs)
	}
	return String{lhn.value + rhn.value}, nil
}

func (in *Interpreter) add(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, rhs)
	}
	return Number{lhn.value + rhn.value}, nil
}

func (in *Interpreter) subtract(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, rhs)
	}
	return Number{lhn.value - rhn.value}, nil
}

func (in *Interpreter) multiply(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, rhs)
	}
	return Number{lhn.value * rhn.value}, nil
}

func (in *Interpreter) divide(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, rhs)
	}
	if rhn.value == 0 {
		// Division by zero is undefined and counts as an erroneous input.
		return Number{}, fmt.Errorf("%w: division by zero", ErrIllegalArgument)
	}
	return Number{lhn.value / rhn.value}, nil
}

func (in *Interpreter) isLessThan(lhs, rhs Object) (Boolean, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, rhs)
	}
	return Boolean{lhn.value < rhn.value}, nil
}

func (in *Interpreter) isGreaterThan(lhs, rhs Object) (Boolean, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a Number", ErrUnrecognizedOperandType, rhs)
	}
	return Boolean{lhn.value > rhn.value}, nil
}

func (in *Interpreter) isEqual(lhs, rhs Object) (Boolean, error) {
	if lhs == nil && rhs == nil {
		return Boolean{true}, nil
	}
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return Boolean{false}, nil
	}
	return Boolean{lhs == rhs}, nil
}
