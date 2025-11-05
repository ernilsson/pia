package squeak

import (
	"errors"
	"github.com/ernilsson/pia/squeak/token"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewLexer(t *testing.T) {
	t.Run("given nil source reader", func(t *testing.T) {
		_, err := NewLexer(nil)
		assert.True(t, errors.Is(err, ErrInvalidSourceReader))
	})
	t.Run("given non-nil source reader", func(t *testing.T) {
		src := "var a = b;"
		_, err := NewLexer(strings.NewReader(src))
		assert.Nil(t, err)
	})
}

func TestLexer_Line(t *testing.T) {
	tests := []struct {
		src      string
		expected int
	}{
		{
			src:      "  var = 4444;",
			expected: 1,
		},
		{
			src: `				# 1
			if (a == a) {		# 2
				return true;	# 3
			}					# 4
								# 5`,
			expected: 5,
		},
		{
			src:      "var\nit\nsnow\n\n\n\n",
			expected: 7,
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src))
			assert.Nil(t, err)
			for {
				tok, err := lx.Next()
				assert.Nil(t, err)
				if tok.Type == token.EOF {
					break
				}
			}
			assert.Equal(t, test.expected, lx.Line())
		})
	}
}

func TestLexer_Next(t *testing.T) {
	tests := []struct {
		src      string
		expected []token.Token
		bl       int
	}{
		{
			src: " var  = 512;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Var,
					Lexeme: "var",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.Integer,
					Lexeme: "512",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			Object {
				status: 200,
			};
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Object,
					Lexeme: "Object",
				},
				{
					Type:   token.LeftBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Identifier,
					Lexeme: "status",
				},
				{
					Type:   token.Colon,
					Lexeme: ":",
				},
				{
					Type:   token.Integer,
					Lexeme: "200",
				},
				{
					Type:   token.Comma,
					Lexeme: ",",
				},
				{
					Type:   token.RightBrace,
					Lexeme: "}",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "break; continue; import; import as; export; Object;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Break,
					Lexeme: "break",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.Continue,
					Lexeme: "continue",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.Import,
					Lexeme: "import",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.Import,
					Lexeme: "import",
				},
				{
					Type:   token.As,
					Lexeme: "as",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.Export,
					Lexeme: "export",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.Object,
					Lexeme: "Object",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			function add(a, b) {
				print(a + b);
			}
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Function,
					Lexeme: "function",
				},
				{
					Type:   token.Identifier,
					Lexeme: "add",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Comma,
					Lexeme: ",",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Identifier,
					Lexeme: "print",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightBrace,
					Lexeme: "}",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "a + nil;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.Nil,
					Lexeme: "nil",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "120.",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Float,
					Lexeme: "120.",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "if else",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.If,
					Lexeme: "if",
				},
				{
					Type:   token.Else,
					Lexeme: "else",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "120. + 13;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Float,
					Lexeme: "120.",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.Integer,
					Lexeme: "13",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "a + 12.55;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.Float,
					Lexeme: "12.55",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "0.33333;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Float,
					Lexeme: "0.33333",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "print(15);",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "print",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Integer,
					Lexeme: "15",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "a+b/and(}){*!=!.",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.Slash,
					Lexeme: "/",
				},
				{
					Type:   token.And,
					Lexeme: "and",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.RightBrace,
					Lexeme: "}",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				{
					Type:   token.Bang,
					Lexeme: "!",
				},
				{
					Type:   token.Dot,
					Lexeme: ".",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "true and false or false",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Boolean,
					Lexeme: "true",
				},
				{
					Type:   token.And,
					Lexeme: "and",
				},
				{
					Type:   token.Boolean,
					Lexeme: "false",
				},
				{
					Type:   token.Or,
					Lexeme: "or",
				},
				{
					Type:   token.Boolean,
					Lexeme: "false",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			# This makes no sense but it does not have to since this is a test
			# Will this work with two lines of comments?
			while (true) {
				return a[0];
			}
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.While,
					Lexeme: "while",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Boolean,
					Lexeme: "true",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Return,
					Lexeme: "return",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.LeftBracket,
					Lexeme: "[",
				},
				{
					Type:   token.Integer,
					Lexeme: "0",
				},
				{
					Type:   token.RightBracket,
					Lexeme: "]",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightBrace,
					Lexeme: "}",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "var name = \"crookdc\";",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Var,
					Lexeme: "var",
				},
				{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.String,
					Lexeme: "crookdc",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			if (a > b) {
				var c = 5;
			}
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.If,
					Lexeme: "if",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Greater,
					Lexeme: ">",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Var,
					Lexeme: "var",
				},
				{
					Type:   token.Identifier,
					Lexeme: "c",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.Integer,
					Lexeme: "5",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightBrace,
					Lexeme: "}",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "var developer = \"crookd\";",
			bl:  4,
			expected: []token.Token{
				{
					Type:   token.Var,
					Lexeme: "var",
				},
				{
					Type:   token.Identifier,
					Lexeme: "developer",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.String,
					Lexeme: "crookd",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "name + \"is a good developer\";",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.String,
					Lexeme: "is a good developer",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			# This function reports whether both a and b are positive
			var pos = function(a, b) {
				# Holy cow, this is a comment isn't it!
				return a > 0 and b > 0;
			};
			# Sometimes there are comments at the very end of the source code!
			# It's important that we cover those as well.`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Var,
					Lexeme: "var",
				},
				{
					Type:   token.Identifier,
					Lexeme: "pos",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.Function,
					Lexeme: "function",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Comma,
					Lexeme: ",",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Return,
					Lexeme: "return",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Greater,
					Lexeme: ">",
				},
				{
					Type:   token.Integer,
					Lexeme: "0",
				},
				{
					Type:   token.And,
					Lexeme: "and",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.Greater,
					Lexeme: ">",
				},
				{
					Type:   token.Integer,
					Lexeme: "0",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightBrace,
					Lexeme: "}",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src), BufferLength(test.bl))
			assert.Nil(t, err)
			var i int
			for {
				actual, err := lx.Next()
				assert.Nil(t, err)
				assert.Equal(t, test.expected[i], actual, "token index %d", i)
				i++
				if actual.Type == token.EOF {
					break
				}
			}
		})
	}
}

func TestNewPeekingLexer(t *testing.T) {
	t.Run("given nil lexer", func(t *testing.T) {
		_, err := NewPeekingLexer(nil)
		assert.True(t, errors.Is(err, ErrInvalidSourceLexer))
	})
	t.Run("given non-nil lexer", func(t *testing.T) {
		src := "var a = b;"
		lx, err := NewLexer(strings.NewReader(src))
		assert.Nil(t, err)
		_, err = NewPeekingLexer(lx)
		assert.Nil(t, err)
	})
}

func TestPeekingLexer_Line(t *testing.T) {
	src := "var \na\n = \nb;"

	lx, err := NewLexer(strings.NewReader(src))
	assert.Nil(t, err)

	plx, err := NewPeekingLexer(lx)
	assert.Nil(t, err)

	tok, err := plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Var, Lexeme: "var"}, tok)
	assert.Equal(t, 1, plx.Line())

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Var, Lexeme: "var"}, tok)
	assert.Equal(t, 1, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Var, Lexeme: "var"}, tok)
	assert.Equal(t, 1, plx.Line())

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)
	assert.Equal(t, 2, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)
	assert.Equal(t, 2, plx.Line())

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Assign, Lexeme: "="}, tok)
	assert.Equal(t, 3, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Assign, Lexeme: "="}, tok)
	assert.Equal(t, 3, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "b"}, tok)
	assert.Equal(t, 4, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Semicolon, Lexeme: ";"}, tok)
	assert.Equal(t, 4, plx.Line())
}

func TestPeekingLexer_Peek(t *testing.T) {
	src := "var a = b;"

	lx, err := NewLexer(strings.NewReader(src))
	assert.Nil(t, err)

	plx, err := NewPeekingLexer(lx)
	assert.Nil(t, err)

	tok, err := plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Var, Lexeme: "var"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Var, Lexeme: "var"}, tok)

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Var, Lexeme: "var"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)
}
