package squeak

import (
	"errors"
	"fmt"
	"github.com/ernilsson/pia/squeak/token"
	"io"
	"unicode"
)

// LexerBufferLength is the default buffer length for Lexer values. This value should not be considered static across
// different versions of the squeak package and hence not relied upon as anything other than the default length of a
// Lexer buffer. The buffer length can be customized using the [squeak.BufferLength] option.
const LexerBufferLength = 512

var (
	ErrInvalidSourceReader = errors.New("invalid source reader")
	ErrInvalidSourceLexer  = errors.New("invalid source lexer")
)

// NewPeekingLexer returns a [squeak.PeekingLexer] that wraps the provided Lexer value. A non-nil error is returned only
// if the provided Lexer is invalid.
func NewPeekingLexer(lx *Lexer) (*PeekingLexer, error) {
	if lx == nil {
		return nil, fmt.Errorf("%w: nil", ErrInvalidSourceLexer)
	}
	return &PeekingLexer{
		lx:  lx,
		ptr: nil,
	}, nil
}

// PeekingLexer extends a value of [squeak.Lexer] with a Peek operation which is often necessary for a clean
// implementation of lexical analysis users such as parsers. See PeekingLexer.Peek.
type PeekingLexer struct {
	lx   *Lexer
	ptr  *token.Token
	line int
}

// Line reports the line of the currently peeked token if one exists. If there is no peeked token then it delegates the
// call to its underlying Lexer.
func (pl *PeekingLexer) Line() int {
	if pl.ptr == nil {
		return pl.lx.line
	}
	return pl.line
}

// Peek either returns the latest non-read peeked token or reads from the underlying [squeak.Lexer] and stores the read
// token for subsequent calls to Peek.
func (pl *PeekingLexer) Peek() (token.Token, error) {
	if pl.ptr != nil {
		return *pl.ptr, nil
	}
	n, err := pl.lx.Next()
	if err != nil {
		return token.Null, err
	}
	pl.ptr = &n
	pl.line = pl.lx.Line()
	return n, nil
}

// Discard clears the peeked token and discards it entirely. Subsequent calls to Peek or Next will return the next token
// as according to the underlying Lexer.
func (pl *PeekingLexer) Discard() {
	pl.ptr = nil
}

// Next either returns the latest peeked token and then marks it as read or reads directly from the underlying
// [squeak.Lexer].
func (pl *PeekingLexer) Next() (token.Token, error) {
	if pl.ptr == nil {
		return pl.lx.Next()
	}
	n := *pl.ptr
	pl.ptr = nil
	return n, nil
}

// LexerOpt represents an optional transformer of Lexer instances that are applied during construction.
type LexerOpt func(*Lexer)

// BufferLength allows the caller to change the underlying buffer length of a Lexer value from the default
// [squeak.LexerBufferLength].
func BufferLength(size int) LexerOpt {
	if size < 1 {
		panic(fmt.Sprintf("illegal buffer size: %d", size))
	}
	return func(lx *Lexer) {
		lx.buffer = make([]byte, size)
	}
}

// NewLexer returns a new Lexer instance which has been pre-wired to read from the provided [io.Reader]. Other parts of
// the lexer can only be customized by the use of [squeak.LexerOpt]. If the constructed Lexer value does not contain a
// non-nil [io.Reader] then an error is returned.
func NewLexer(src io.Reader, opts ...LexerOpt) (*Lexer, error) {
	lx := &Lexer{
		src: src,
		// Setting the cursor to the length of the buffer forces a read operation from the underlying reader when
		// Lexer.Next is called for the first time.
		cursor: 0,
		length: 0,
		buffer: make([]byte, LexerBufferLength),
		line:   1,
	}
	for _, opt := range opts {
		opt(lx)
	}
	if src == nil {
		// A nil source reader would cause a panic in Lexer.Next if we do not handle the error gracefully here. One
		// could argue that providing a nil reader warrants a panic since it indicates a severe logical fault on the
		// caller side. In short, this handling is up for discussion.
		return nil, fmt.Errorf("%w: nil", ErrInvalidSourceReader)
	}
	return lx, nil
}

// Lexer provides a means of lexical analysis (i.e. source code scanning) of code written according to the Squeak
// specification. A single Lexer can only ever read from one and the same source reader, which will most often be a
// file. The correct way of analyzing multiple streams of source code is to construct several Lexer values using
// [squeak.NewLexer].
type Lexer struct {
	src    io.Reader
	cursor int
	length int
	buffer []byte
	line   int
}

// Line reports the line number of the latest read token.
func (lx *Lexer) Line() int {
	return lx.line
}

// Next returns the next available token from the source code reader. If the source code contains an illegal token then
// a token of type [token.Illegal] is returned with a nil error. A [token.EOF] token is returned once the underlying
// source code reader has been exhausted. Any errors originating from the underlying source code reader is propagated to
// the caller.
func (lx *Lexer) Next() (token.Token, error) {
	err := lx.skip(ignored())
	if errors.Is(err, io.EOF) {
		return token.New(token.EOF)
	}
	if err != nil {
		return token.Null, err
	}
	c, err := lx.read(never)
	if err != nil {
		return token.Null, err
	}
	if unicode.IsDigit(rune(c)) {
		return lx.number()
	}
	if unicode.IsLetter(rune(c)) {
		return lx.word()
	}
	return lx.symbol()
}

func (lx *Lexer) number() (token.Token, error) {
	integer, err := lx.next(unicode.IsDigit)
	if err != nil {
		return token.Null, err
	}
	nxt, err := lx.read(never)
	if err != nil {
		return token.Null, err
	}
	switch nxt {
	case '.':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		fraction, err := lx.next(unicode.IsDigit)
		if errors.Is(err, io.EOF) {
			fraction = []byte("")
		} else if err != nil {
			return token.Null, err
		}
		return token.New(token.Float, token.Lexeme(fmt.Sprintf("%s.%s", integer, fraction)))
	default:
		return token.New(token.Integer, token.Lexeme(string(integer)))
	}
}

func (lx *Lexer) symbol() (token.Token, error) {
	c, err := lx.read(never)
	if err != nil {
		return token.Null, err
	}
	switch c {
	case ';':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Semicolon)
	case ':':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Colon)
	case '<':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		nxt, err := lx.read(never)
		if err != nil {
			return token.Null, err
		}
		if nxt != '=' {
			return token.New(token.Less)
		}
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.LessEqual)
	case '>':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		nxt, err := lx.read(never)
		if err != nil {
			return token.Null, err
		}
		if nxt != '=' {
			return token.New(token.Greater)
		}
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.GreaterEqual)
	case '+':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Plus)
	case '-':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Minus)
	case '*':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Asterisk)
	case '/':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Slash)
	case ',':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Comma)
	case '.':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Dot)
	case '(':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.LeftParenthesis)
	case ')':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.RightParenthesis)
	case '{':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.LeftBrace)
	case '}':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.RightBrace)
	case '[':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.LeftBracket)
	case ']':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.RightBracket)
	case '=':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		nxt, err := lx.read(never)
		if err != nil {
			return token.Null, err
		}
		if nxt != '=' {
			return token.New(token.Assign)
		}
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Equals)
	case '!':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		nxt, err := lx.read(never)
		if err != nil {
			return token.Null, err
		}
		if nxt != '=' {
			return token.New(token.Bang)
		}
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.NotEquals)
	case '&':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		nxt, err := lx.read(always)
		if err != nil {
			return token.Null, err
		}
		if nxt != '&' {
			return token.New(token.Illegal, token.Lexeme(string([]byte{c, nxt})))
		}
		return token.New(token.And)
	case '|':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		nxt, err := lx.read(always)
		if err != nil {
			return token.Null, err
		}
		if nxt != '|' {
			return token.New(token.Illegal, token.Lexeme(string([]byte{c, nxt})))
		}
		return token.New(token.Or)
	case '"':
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		literal, err := lx.next(func(r rune) bool {
			return r != '"'
		})
		if err != nil {
			return token.Null, err
		}
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.String, token.Lexeme(string(literal)))
	default:
		if err := lx.skip(amount(1)); err != nil {
			return token.Null, err
		}
		return token.New(token.Illegal, token.Lexeme(string(c)))
	}
}

func (lx *Lexer) word() (token.Token, error) {
	w, err := lx.next(identifier())
	if err != nil {
		return token.Null, err
	}
	switch string(w) {
	case "if":
		return token.New(token.If)
	case "else":
		return token.New(token.Else)
	case "and":
		return token.New(token.And)
	case "or":
		return token.New(token.Or)
	case "while":
		return token.New(token.While)
	case "return":
		return token.New(token.Return)
	case "break":
		return token.New(token.Break)
	case "continue":
		return token.New(token.Continue)
	case "function":
		return token.New(token.Function)
	case "var":
		return token.New(token.Var)
	case "nil":
		return token.New(token.Nil)
	case "import":
		return token.New(token.Import)
	case "as":
		return token.New(token.As)
	case "export":
		return token.New(token.Export)
	case "Object":
		return token.New(token.Object)
	case "true", "false":
		return token.New(token.Boolean, token.Lexeme(string(w)))
	default:
		return token.New(token.Identifier, token.Lexeme(string(w)))
	}
}

func identifier() func(rune) bool {
	return either(unicode.IsLetter, unicode.IsDigit, func(r rune) bool {
		return r == '_'
	})
}

func always(_ byte) bool {
	return true
}

func never(_ byte) bool {
	return false
}

func (lx *Lexer) read(proceed func(byte) bool) (c byte, err error) {
	if lx.cursor >= lx.length {
		lx.length, err = lx.src.Read(lx.buffer)
		if err != nil {
			return 0, err
		}
		lx.cursor = 0
	}
	c = lx.buffer[lx.cursor]
	if !proceed(c) {
		return
	}
	lx.cursor += 1
	if c == '\n' {
		lx.line++
	}
	return
}

func either(fns ...func(rune) bool) func(rune) bool {
	return func(r rune) bool {
		for _, fn := range fns {
			if fn(r) {
				return true
			}
		}
		return false
	}
}

func (lx *Lexer) next(fn func(rune) bool) ([]byte, error) {
	proceed := func(b byte) bool {
		return fn(rune(b))
	}
	c, err := lx.read(proceed)
	if err != nil {
		return nil, err
	}
	if !fn(rune(c)) {
		return nil, nil
	}
	word := []byte{c}
	for fn(rune(c)) {
		c, err = lx.read(proceed)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if fn(rune(c)) {
			word = append(word, c)
		}
	}
	return word, nil
}

func amount(n int) func(rune) bool {
	var c int
	return func(r rune) bool {
		if c == n {
			return false
		}
		c += 1
		return true
	}
}

func ignored() func(rune) bool {
	var delegate = unicode.IsSpace
	return func(r rune) bool {
		if r == '#' {
			delegate = comment()
		}
		n := delegate(r)
		if !n {
			delegate = unicode.IsSpace
		}
		return n
	}
}

func comment() func(rune) bool {
	var lf bool
	return func(r rune) bool {
		if lf {
			return unicode.IsSpace(r)
		}
		switch r {
		case '\n':
			lf = true
			return false
		default:
			return true
		}
	}
}

func (lx *Lexer) skip(fn func(rune) bool) error {
	proceed := func(b byte) bool {
		return fn(rune(b))
	}
	c, err := lx.read(proceed)
	if err != nil {
		return err
	}
	for fn(rune(c)) {
		c, err = lx.read(proceed)
		if err != nil {
			return err
		}
	}
	return nil
}
