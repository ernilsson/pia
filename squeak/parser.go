package squeak

import (
	"errors"
	"fmt"
	"github.com/ernilsson/pia/squeak/ast"
	"github.com/ernilsson/pia/squeak/token"
	"io"
	"strconv"
	"strings"
)

type SyntaxError struct {
	Line int
}

func (s SyntaxError) Error() string {
	return fmt.Sprintf("syntax error on line %d", s.Line)
}

func Parse(r io.Reader) ([]ast.StatementNode, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseString(string(src))
}

// ParseString reads src all the way through and builds an AST containing multiple statements from it.
func ParseString(src string) ([]ast.StatementNode, error) {
	lx, err := NewLexer(strings.NewReader(src))
	if err != nil {
		return nil, err
	}
	plx, err := NewPeekingLexer(lx)
	if err != nil {
		return nil, err
	}
	program := make([]ast.StatementNode, 0)
	ps := NewParser(plx)
	for {
		stmt, err := ps.Next()
		if errors.Is(err, io.EOF) {
			return program, nil
		}
		if err != nil {
			return nil, err
		}
		program = append(program, stmt)
	}
}

func NewParser(lx *PeekingLexer) *Parser {
	return &Parser{
		lx: lx,
		stack: struct {
			slice []map[string]struct{}
			sp    int
		}{
			slice: []map[string]struct{}{
				make(map[string]struct{}), // Runtime environment
				make(map[string]struct{}), // Global user environment
			},
			sp: 1,
		},
	}
}

// Parser builds an abstract syntax tree from the tokens yielded by a Lexer.
type Parser struct {
	lx    *PeekingLexer
	stack struct {
		slice []map[string]struct{}
		sp    int
	}
}

func (ps *Parser) resolve(name token.Token) int {
	for i := range ps.stack.sp + 1 {
		scope := ps.stack.slice[ps.stack.sp-i]
		if _, ok := scope[name.Lexeme]; ok {
			return i
		}
	}
	// The value of the stack pointer is the current distance from the global environment. If a variable cannot be
	// resolved then it is assumed to exist within the global environment.
	return ps.stack.sp
}

func (ps *Parser) scope() (map[string]struct{}, bool) {
	if ps.stack.sp < 0 {
		return nil, false
	}
	return ps.stack.slice[ps.stack.sp], true
}

func (ps *Parser) declare(name token.Token) error {
	sc, ok := ps.scope()
	if !ok {
		return nil
	}
	if _, ok := sc[name.Lexeme]; ok {
		return fmt.Errorf("%w: %s", SyntaxError{Line: ps.lx.Line()}, name.Lexeme)
	}
	sc[name.Lexeme] = struct{}{}
	return nil
}

func (ps *Parser) begin() {
	if ps.stack.sp == len(ps.stack.slice)-1 {
		// If the stack is the largest it has been (and the stack pointer would exceed the length of the stack if
		// incremented) then it must first be extended.
		ps.stack.slice = append(ps.stack.slice, map[string]struct{}{})
	}
	ps.stack.sp += 1
	// Clear the old data, it belongs to an environment that has now gone out of scope.
	clear(ps.stack.slice[ps.stack.sp])
}

func (ps *Parser) end() {
	if ps.stack.sp < 0 {
		return
	}
	ps.stack.sp -= 1
}

// Next constructs and returns the next node in the abstract syntax tree for the underlying Lexer.
func (ps *Parser) Next() (stmt ast.StatementNode, err error) {
	defer func() {
		if err != nil {
			// If an error occurred for any reason during the parsing of the current statement then the parser should at
			// least try to fast-forward to the next statement. This counteracts cascading syntax errors that would be
			// fine if it wasn't for the initial error that triggered a chain reaction. However, it is possible that
			// another is encountered as the current statement is cleared, hence the call to errors.Join().
			err = errors.Join(err, ps.clear())
		}
	}()
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.EOF:
		return nil, io.EOF
	default:
		return ps.declaration()
	}
}

func (ps *Parser) declaration() (ast.StatementNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.Var:
		return ps.variable()
	default:
		return ps.statement()
	}
}

func (ps *Parser) variable() (ast.Declaration, error) {
	if _, err := ps.expect(token.Var); err != nil {
		return ast.Declaration{}, err
	}
	name, err := ps.expect(token.Identifier)
	if err != nil {
		return ast.Declaration{}, err
	}
	if err := ps.declare(name); err != nil {
		return ast.Declaration{}, err
	}
	stmt := ast.Declaration{
		Name:        name,
		Initializer: nil,
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.Declaration{}, err
	}
	if pk.Type == token.Semicolon {
		ps.lx.Discard()
		return stmt, nil
	}
	if _, err := ps.expect(token.Assign); err != nil {
		return ast.Declaration{}, err
	}
	init, err := ps.equality()
	if err != nil {
		return ast.Declaration{}, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.Declaration{}, err
	}
	stmt.Initializer = init
	return stmt, nil
}

func (ps *Parser) statement() (ast.StatementNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.LeftBrace:
		ps.begin()
		defer ps.end()
		return ps.block()
	case token.If:
		return ps.ifs()
	case token.While:
		return ps.while()
	case token.Semicolon:
		ps.lx.Discard()
		return ast.Noop{}, nil
	case token.Function:
		return ps.function()
	case token.Return:
		return ps.ret()
	case token.Break:
		ps.lx.Discard()
		if _, err := ps.expect(token.Semicolon); err != nil {
			return nil, err
		}
		return ast.Break{}, nil
	case token.Continue:
		ps.lx.Discard()
		if _, err := ps.expect(token.Semicolon); err != nil {
			return nil, err
		}
		return ast.Continue{}, nil
	case token.Import:
		return ps.imp()
	case token.Export:
		return ps.exp()
	default:
		return ps.expression()
	}
}

func (ps *Parser) function() (ast.Function, error) {
	if _, err := ps.expect(token.Function); err != nil {
		return ast.Function{}, err
	}
	// The function name must be declared before the new scope is registered since the function name would otherwise
	// be undefined in the surrounding environment.
	name, err := ps.expect(token.Identifier)
	if err != nil {
		return ast.Function{}, err
	}
	if err := ps.declare(name); err != nil {
		return ast.Function{}, err
	}
	ps.begin()
	defer ps.end()
	params, err := ps.tokens(token.LeftParenthesis, token.RightParenthesis)
	if err != nil {
		return ast.Function{}, err
	}
	// The parameters for the function should be declared as part of the scope of the function itself, meaning that
	// resolving any parameter name leads to a level of 0.
	for _, param := range params {
		if err := ps.declare(param); err != nil {
			return ast.Function{}, err
		}
	}
	body, err := ps.block()
	if err != nil {
		return ast.Function{}, err
	}
	return ast.Function{
		Name:   name,
		Params: params,
		Body:   body,
	}, nil
}

func (ps *Parser) ret() (ast.Return, error) {
	if _, err := ps.expect(token.Return); err != nil {
		return ast.Return{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.Return{}, err
	}
	switch pk.Type {
	case token.Semicolon:
		return ast.Return{}, nil
	default:
		expr, err := ps.logical()
		if err != nil {
			return ast.Return{}, err
		}
		if _, err := ps.expect(token.Semicolon); err != nil {
			return ast.Return{}, err
		}
		return ast.Return{
			Statement:  ast.Statement{},
			Expression: expr,
		}, nil
	}
}

func (ps *Parser) imp() (ast.Import, error) {
	if _, err := ps.expect(token.Import); err != nil {
		return ast.Import{}, err
	}
	expr, err := ps.primary()
	if err != nil {
		return ast.Import{}, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.Import{}, err
	}
	switch expr := expr.(type) {
	case ast.Variable, ast.StringLiteral:
		return ast.Import{
			Source: expr,
		}, nil
	default:
		return ast.Import{}, fmt.Errorf("%w: %T is not a valid import expression", ErrUnrecognizedExpression, expr)
	}
}

func (ps *Parser) exp() (ast.Export, error) {
	if _, err := ps.expect(token.Export); err != nil {
		return ast.Export{}, err
	}
	expr, err := ps.assignment()
	if err != nil {
		return ast.Export{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.Export{}, err
	}
	switch pk.Type {
	case token.Semicolon:
		ps.lx.Discard()
		switch expr := expr.(type) {
		case ast.Variable:
			return ast.Export{
				Name:  expr.Name,
				Value: expr,
			}, nil
		default:
			return ast.Export{}, fmt.Errorf("%w: %T as unnamed export", ErrUnrecognizedExpression, expr)
		}
	case token.As:
		ps.lx.Discard()
		name, err := ps.expect(token.Identifier)
		if err != nil {
			return ast.Export{}, err
		}
		if _, err := ps.expect(token.Semicolon); err != nil {
			return ast.Export{}, err
		}
		return ast.Export{
			Name:  name,
			Value: expr,
		}, nil
	default:
		return ast.Export{}, fmt.Errorf("%w: unexpected token %s", ErrRuntimeFault, pk.Lexeme)
	}
}

func (ps *Parser) while() (ast.StatementNode, error) {
	if _, err := ps.expect(token.While); err != nil {
		return nil, err
	}
	cnd, err := ps.logical()
	if err != nil {
		return nil, err
	}
	ps.begin()
	defer ps.end()
	body, err := ps.block()
	if err != nil {
		return nil, err
	}
	return ast.While{
		Condition: cnd,
		Body:      body,
	}, nil
}

func (ps *Parser) ifs() (ast.If, error) {
	if _, err := ps.expect(token.If); err != nil {
		return ast.If{}, err
	}
	cnd, err := ps.logical()
	if err != nil {
		return ast.If{}, err
	}
	then, err := ps.statement()
	if err != nil {
		return ast.If{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.If{}, err
	}
	st := ast.If{
		Condition: cnd,
		Then:      then,
		Else:      nil,
	}
	if pk.Type == token.Else {
		ps.lx.Discard()
		otherwise, err := ps.statement()
		if err != nil {
			return ast.If{}, err
		}
		st.Else = otherwise
	}
	return st, nil
}

func (ps *Parser) block() (ast.Block, error) {
	if _, err := ps.expect(token.LeftBrace); err != nil {
		return ast.Block{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.Block{}, err
	}
	body := make([]ast.StatementNode, 0)
	for {
		switch pk.Type {
		case token.RightBrace, token.EOF:
			ps.lx.Discard()
			return ast.Block{
				Body: body,
			}, nil
		default:
			st, err := ps.declaration()
			if err != nil {
				return ast.Block{}, err
			}
			body = append(body, st)
			pk, err = ps.lx.Peek()
			if err != nil {
				return ast.Block{}, err
			}
		}
	}
}

func (ps *Parser) clear() error {
	look := true
	for look {
		nxt, err := ps.lx.Next()
		if err != nil {
			return err
		}
		switch nxt.Type {
		case token.EOF, token.Semicolon, token.RightBrace:
			look = false
		default:
		}
	}
	return nil
}

func (ps *Parser) expression() (ast.ExpressionStatement, error) {
	expr, err := ps.assignment()
	if err != nil {
		return ast.ExpressionStatement{}, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.ExpressionStatement{}, err
	}
	return ast.ExpressionStatement{
		Expression: expr,
	}, nil
}

func (ps *Parser) assignment() (ast.ExpressionNode, error) {
	expr, err := ps.logical()
	if err != nil {
		return nil, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	if pk.Type != token.Assign {
		return expr, nil
	}
	ps.lx.Discard()
	switch expr := expr.(type) {
	case ast.Variable:
		val, err := ps.assignment()
		if err != nil {
			return nil, err
		}
		return ast.Assignment{
			Level: ps.resolve(expr.Name),
			Name:  expr.Name,
			Value: val,
		}, nil
	case ast.GetProp:
		val, err := ps.assignment()
		if err != nil {
			return nil, err
		}
		return ast.SetProp{
			Target:   expr,
			Property: expr.Property,
			Value:    val,
		}, nil
	case ast.GetIndex:
		val, err := ps.assignment()
		if err != nil {
			return nil, err
		}
		return ast.SetIndex{
			Target: expr,
			Value:  val,
		}, nil
	default:
		return nil, fmt.Errorf(
			"%w: invalid left hand side of assignment",
			SyntaxError{Line: ps.lx.Line()},
		)
	}
}

func (ps *Parser) logical() (ast.ExpressionNode, error) {
	lhs, err := ps.equality()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.And, token.Or:
			ps.lx.Discard()
			rhs, err := ps.equality()
			if err != nil {
				return nil, err
			}
			lhs = ast.Logical{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) equality() (ast.ExpressionNode, error) {
	lhs, err := ps.comparison()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Equals, token.NotEquals:
			ps.lx.Discard()
			rhs, err := ps.comparison()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) comparison() (ast.ExpressionNode, error) {
	lhs, err := ps.term()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Less, token.LessEqual, token.Greater, token.GreaterEqual:
			ps.lx.Discard()
			rhs, err := ps.term()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) term() (ast.ExpressionNode, error) {
	lhs, err := ps.factor()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Minus, token.Plus:
			ps.lx.Discard()
			rhs, err := ps.factor()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) factor() (ast.ExpressionNode, error) {
	lhs, err := ps.prefix()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Asterisk, token.Slash:
			ps.lx.Discard()
			rhs, err := ps.prefix()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) prefix() (ast.ExpressionNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.Bang, token.Minus:
		ps.lx.Discard()
		expr, err := ps.prefix()
		if err != nil {
			return nil, err
		}
		return ast.Prefix{
			Operator: pk,
			Target:   expr,
		}, nil
	default:
		return ps.call()
	}
}

func (ps *Parser) call() (ast.ExpressionNode, error) {
	expr, err := ps.primary()
	if err != nil {
		return nil, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	for {
		switch pk.Type {
		case token.LeftParenthesis:
			args, err := ps.expressions(pk.Type, token.Closers[pk.Type])
			if err != nil {
				return nil, err
			}
			expr = ast.Call{
				Callee:   expr,
				Operator: pk,
				Args:     args,
			}
		case token.LeftBracket:
			args, err := ps.expressions(pk.Type, token.Closers[pk.Type])
			if err != nil {
				return nil, err
			}
			if len(args) != 1 {
				return nil, fmt.Errorf(
					"%w: indexing requires exactly one argument",
					SyntaxError{Line: ps.lx.Line()},
				)
			}
			expr = ast.GetIndex{
				Target: expr,
				Index:  args[0],
			}
		case token.Dot:
			ps.lx.Discard()
			prop, err := ps.expect(token.Identifier, token.String)
			if err != nil {
				return nil, err
			}
			expr = ast.GetProp{
				Target:   expr,
				Property: prop,
			}
		default:
			return expr, nil
		}
		pk, err = ps.lx.Peek()
		if err != nil {
			return nil, err
		}
	}
}

func (ps *Parser) expressions(start, end token.Type) (exps []ast.ExpressionNode, err error) {
	exps = make([]ast.ExpressionNode, 0)
	if _, err := ps.expect(start); err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case end:
			ps.lx.Discard()
			return exps, nil
		case token.Comma:
			ps.lx.Discard()
		default:
			expr, err := ps.equality()
			if err != nil {
				return nil, err
			}
			exps = append(exps, expr)
		}
	}
}

func (ps *Parser) tokens(start, end token.Type) (tokens []token.Token, err error) {
	tokens = make([]token.Token, 0)
	if _, err := ps.expect(start); err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case end:
			ps.lx.Discard()
			return tokens, nil
		case token.Comma:
			ps.lx.Discard()
		default:
			tokens = append(tokens, pk)
			ps.lx.Discard()
		}
	}
}
func (ps *Parser) primary() (ast.ExpressionNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.Identifier:
		ps.lx.Discard()
		return ast.Variable{
			Level: ps.resolve(pk),
			Name:  pk,
		}, nil
	case token.String:
		ps.lx.Discard()
		return ast.StringLiteral{String: pk.Lexeme}, nil
	case token.Integer:
		ps.lx.Discard()
		i, err := strconv.Atoi(pk.Lexeme)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: invalid integer literal: %s",
				SyntaxError{ps.lx.Line()},
				pk.Lexeme,
			)
		}
		return ast.IntegerLiteral{Integer: i}, nil
	case token.Float:
		ps.lx.Discard()
		f, err := strconv.ParseFloat(pk.Lexeme, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: invalid float literal: %s",
				SyntaxError{ps.lx.Line()},
				pk.Lexeme,
			)
		}
		return ast.FloatLiteral{Float: f}, nil
	case token.Boolean:
		ps.lx.Discard()
		b, err := strconv.ParseBool(pk.Lexeme)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: invalid boolean literal: %s",
				SyntaxError{ps.lx.Line()},
				pk.Lexeme,
			)
		}
		return ast.BooleanLiteral{Boolean: b}, nil
	case token.LeftParenthesis:
		ps.lx.Discard()
		expr, err := ps.equality()
		if err != nil {
			return nil, err
		}
		if _, err := ps.expect(token.RightParenthesis); err != nil {
			return nil, err
		}
		return ast.Grouping{
			Group: expr,
		}, nil
	case token.LeftBracket:
		items, err := ps.expressions(token.LeftBracket, token.RightBracket)
		if err != nil {
			return nil, err
		}
		return ast.ListLiteral{
			Items: items,
		}, nil
	case token.Object:
		ps.lx.Discard()
		props, err := ps.keymap()
		if err != nil {
			return nil, err
		}
		return ast.ObjectLiteral{
			Properties: props,
		}, nil
	case token.Function:
		return ps.method()
	case token.Nil:
		ps.lx.Discard()
		return ast.NilLiteral{}, nil
	default:
		ps.lx.Discard()
		return nil, fmt.Errorf(
			"%w: unexpected token: %s",
			SyntaxError{Line: ps.lx.Line()},
			pk.Lexeme,
		)
	}
}

func (ps *Parser) method() (ast.Method, error) {
	if _, err := ps.expect(token.Function); err != nil {
		return ast.Method{}, err
	}
	params, err := ps.tokens(token.LeftParenthesis, token.RightParenthesis)
	if err != nil {
		return ast.Method{}, err
	}
	previous := ps.stack
	defer func() {
		ps.stack = previous
	}()
	ps.stack = struct {
		slice []map[string]struct{}
		sp    int
	}{
		slice: []map[string]struct{}{
			make(map[string]struct{}),
			make(map[string]struct{}),
			{
				"this": struct{}{},
			},
		},
		sp: 2,
	}
	ps.begin()
	for _, param := range params {
		if err := ps.declare(param); err != nil {
			return ast.Method{}, err
		}
	}
	body, err := ps.block()
	if err != nil {
		return ast.Method{}, err
	}
	ps.end()
	return ast.Method{
		Params: params,
		Body:   body,
	}, nil
}

func (ps *Parser) keymap() (map[string]ast.ExpressionNode, error) {
	if _, err := ps.expect(token.LeftBrace); err != nil {
		return nil, err
	}
	m := make(map[string]ast.ExpressionNode)
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Identifier:
			ps.lx.Discard()
			if _, err := ps.expect(token.Colon); err != nil {
				return nil, err
			}
			expr, err := ps.equality()
			if err != nil {
				return nil, err
			}
			m[pk.Lexeme] = expr
		default:
			return nil, fmt.Errorf(
				"%w: unexpected token %s",
				SyntaxError{
					Line: ps.lx.Line(),
				},
				pk.Lexeme,
			)
		}
		pk, err = ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		if pk.Type != token.Comma {
			break
		}
		ps.lx.Discard()
	}
	if _, err := ps.expect(token.RightBrace); err != nil {
		return nil, err
	}
	return m, nil
}

func (ps *Parser) expect(types ...token.Type) (token.Token, error) {
	tok, err := ps.lx.Next()
	if err != nil {
		return token.Token{}, err
	}
	for _, t := range types {
		if tok.Type == t {
			return tok, nil
		}
	}
	return token.Token{}, fmt.Errorf(
		"%w: unexpected token: %s",
		SyntaxError{ps.lx.Line()},
		tok.Lexeme,
	)
}
