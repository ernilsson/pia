package ast

import (
	"github.com/ernilsson/pia/squeak/token"
)

// Node represent any type of node in an AST. It does not define any functional behaviours. This is by design as the
// Node abstraction is little more than a means of categorizing data. It is not considered incorrect to implement the
// Node.Node method as nothing but a panic since it should never be called.
type Node interface {
	Node()
}

// ExpressionNode is specialization of [ast.Node] that does not provide any additional functional behaviors directly.
// The extra method ExpressionNode.ExpressionNode() is never meant to be called and the correct behavior of any concrete
// implementation is to panic when this method is invoked. Values that implement the ExpressionNode interface agrees to
// be processed as expressions in an AST.
type ExpressionNode interface {
	Node
	ExpressionNode()
}

// Expression is the default concrete implementation of [ast.ExpressionNode] and should be the first tool to reach for
// when defining a new expression. Its intended usage is to be embedded in structs that themselves provide the necessary
// expressive data.
type Expression struct{}

// ExpressionNode does nothing but panic. An explanation as to why is given in the documentation for
// [ast.ExpressionNode].
func (e Expression) ExpressionNode() {
	panic("Expression.ExpressionNode's behavior is classed as undefined and should never be invoked")
}

// Node does nothing but panic. An explanation as to why is given in the documentation for [ast.Node].
func (e Expression) Node() {
	panic("Expression.Node's behavior is classed as undefined and should never be invoked")
}

// StatementNode is specialization of [ast.Node] that does not provide any additional functional behaviors directly. The
// extra method StatementNode.StatementNode() is never meant to be called and the correct behavior of any concrete
// implementation is to panic when this method is invoked. Values that implement the StatementNode interface agrees to
// be processed as statements in an AST.
type StatementNode interface {
	Node
	StatementNode()
}

// Statement is the default concrete implementation of [ast.StatementNode] and should be the first tool to reach for
// when defining a new statement. Its intended usage is to be embedded in structs that themselves provide the necessary
// data to represent the statement.
type Statement struct{}

// StatementNode does nothing but panic. An explanation as to why is given in the documentation for [ast.StatementNode].
func (s Statement) StatementNode() {
	panic("Statement.StatementNode's behavior is classed as undefined and should never be invoked")
}

// Node does nothing but panic. An explanation as to why is given in the documentation for [ast.Node].
func (s Statement) Node() {
	panic("Statement.Node's behavior is classed as undefined and should never be invoked")
}

// ExpressionStatement represents an expression that exists in isolation within a Squeak script, meaning that it is not
// defined as part of a statement and will thus be considered a statement by itself.
type ExpressionStatement struct {
	Statement
	Expression ExpressionNode
}

// Declaration represents a variable declaration with an optional initializer. Since Initializer is optional it must always be
// nil-checked before use.
type Declaration struct {
	Statement
	Name        token.Token
	Initializer ExpressionNode
}

// Block represents a collection of statements that can be regarded as a single statement where each statement within
// the block should be executed in order during the execution of the block itself.
type Block struct {
	Statement
	Body []StatementNode
}

// If represents the branching structure that allows for conditionally executing a statement.
type If struct {
	Statement
	Condition ExpressionNode
	Then      StatementNode
	Else      StatementNode
}

// While represents the common looping control structure which causes the Squeak interpreter to continually evaluate the
// loop Body until the Condition returns a falsy value. Through forms of de-sugaring, the Squeak for-loop is also
// represented in part using the While loop.
type While struct {
	Statement
	Condition ExpressionNode
	Body      Block
}

// Noop represents a statement that should be ignored by the interpreter. Unlike other statements, the Noop statement
// does not have any side effect.
type Noop struct {
	Statement
}

// Function represents a stored function in a Squeak script.
type Function struct {
	Statement
	Name   token.Token
	Params []token.Token
	Body   Block
}

// Method is a type of function that is bound to an instance of an object. A method is different from a function in that
// it does not capture the surrounding environment upon definition, and it can be represented as an expression rather
// than a statement.
type Method struct {
	Expression
	Params []token.Token
	Body   Block
}

// Return represents a statement which allows a value from within a block to be returned to the caller of said block.
type Return struct {
	Statement
	Expression ExpressionNode
}

// Break represents a break statement, which immediately stops the execution of a loop.
type Break struct {
	Statement
}

// Continue represents a continue statement, which skips the remainder of the loop body for the current iteration and
// jumps immediately to the next iteration of the loop.
type Continue struct {
	Statement
}

// Import represents an import statement, which executes code from another file within the same environment as the
// current scope.
type Import struct {
	Statement
	Source ExpressionNode
}

// Export represents an export statement. Export statements allows an imported script to share named data with its
// importer.
type Export struct {
	Statement
	Name  token.Token
	Value ExpressionNode
}

// Variable represents an expression in the format of just an identifier.
type Variable struct {
	Expression
	// Level corresponds to the depth of which the resolved value exists within the current callstack.
	Level int
	Name  token.Token
}

type GetIndex struct {
	Expression
	Target ExpressionNode
	Index  ExpressionNode
}

type SetIndex struct {
	Expression
	Target GetIndex
	Value  ExpressionNode
}

type GetProp struct {
	Expression
	Target   ExpressionNode
	Property token.Token
}

type SetProp struct {
	Expression
	Target   GetProp
	Property token.Token
	Value    ExpressionNode
}

// IntegerLiteral represents an expression which holds a primitive integer literal.
type IntegerLiteral struct {
	Expression
	Integer int
}

// FloatLiteral represents an expression which holds a primitive float literal.
type FloatLiteral struct {
	Expression
	Float float64
}

// StringLiteral represents an expression which holds a string literal.
type StringLiteral struct {
	Expression
	String string
}

// BooleanLiteral represents an expression which holds a boolean literal.
type BooleanLiteral struct {
	Expression
	Boolean bool
}

// NilLiteral represents a literal nil expression, which in turn represents the absence of a value.
type NilLiteral struct {
	Expression
}

// ListLiteral represents a literal list expression with zero or more items declared within it.
type ListLiteral struct {
	Expression
	Items []ExpressionNode
}

// ObjectLiteral represents a literal object defined by the properties it holds.
type ObjectLiteral struct {
	Expression
	Properties map[string]ExpressionNode
}

// Assignment represents the assignment of a value to a variable without also declaring said variable.
type Assignment struct {
	Expression
	Level int
	Name  token.Token
	Value ExpressionNode
}

// Grouping represents an expression held together as a unit.
type Grouping struct {
	Expression
	Group ExpressionNode
}

// Prefix represents an expression with a single operand where the operator is located before the operand.
type Prefix struct {
	Expression
	Operator token.Token
	Target   ExpressionNode
}

// Infix represents an expression with two operands where the operator is located in between the operands.
type Infix struct {
	Expression
	Operator token.Token
	LHS      ExpressionNode
	RHS      ExpressionNode
}

// Logical represents a binary logical expression comprised of two operands and an operator.
type Logical struct {
	Expression
	Operator token.Token
	LHS      ExpressionNode
	RHS      ExpressionNode
}

// Call represents an invocation of a function identified by the result of evaluating Callee.
type Call struct {
	Expression
	Callee   ExpressionNode
	Operator token.Token
	Args     []ExpressionNode
}
