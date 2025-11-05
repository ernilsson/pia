package squeak

import (
	"github.com/ernilsson/pia/squeak/ast"
	"github.com/ernilsson/pia/squeak/token"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestParser_Next(t *testing.T) {
	tests := []struct {
		src      string
		expected ast.StatementNode
		err      error
	}{
		{
			src: "a;",
			expected: ast.ExpressionStatement{
				Expression: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
			},
		},
		{
			src: "{ break; continue; }",
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Break{},
					ast.Continue{},
				},
			},
		},
		{
			src: "import \"pia:response\";",
			expected: ast.Import{
				Source: ast.StringLiteral{
					String: "pia:response",
				},
			},
		},
		{
			src: "import some_variable;",
			expected: ast.Import{
				Source: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "some_variable",
					},
				},
			},
		},
		{
			src: "import true;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "import 15;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "import 15.4;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export true;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export 13;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export 134.5;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export \"some string\";",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export \"some string\" as string;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "string",
				},
				Value: ast.StringLiteral{
					String: "some string",
				},
			},
		},
		{
			src: "export my_var as alias;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "alias",
				},
				Value: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "my_var",
					},
				},
			},
		},
		{
			src: "export my_var;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "my_var",
				},
				Value: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "my_var",
					},
				},
			},
		},
		{
			src: `
			{
				import "pia:request";
			}
			`,
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Import{
						Source: ast.StringLiteral{
							String: "pia:request",
						},
					},
				},
			},
		},
		{
			src: `
			function add(a, b) {
				print(a + b);
				return 42.;
			}
			`,
			expected: ast.Function{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "add",
				},
				Params: []token.Token{
					{
						Type:   token.Identifier,
						Lexeme: "a",
					},
					{
						Type:   token.Identifier,
						Lexeme: "b",
					},
				},
				Body: ast.Block{
					Body: []ast.StatementNode{
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Level: 2,
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Infix{
										Operator: token.Token{
											Type:   token.Plus,
											Lexeme: "+",
										},
										LHS: ast.Variable{
											Level: 0,
											Name: token.Token{
												Type:   token.Identifier,
												Lexeme: "a",
											},
										},
										RHS: ast.Variable{
											Level: 0,
											Name: token.Token{
												Type:   token.Identifier,
												Lexeme: "b",
											},
										},
									},
								},
							},
						},
						ast.Return{
							Expression: ast.FloatLiteral{
								Float: 42,
							},
						},
					},
				},
			},
		},
		{
			src: `
			function clock() {
				# This is where we would put some logic!
			}
			`,
			expected: ast.Function{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "clock",
				},
				Params: []token.Token{},
				Body: ast.Block{
					Body: []ast.StatementNode{},
				},
			},
		},
		{
			src: "a + b;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
			},
		},
		{
			src: "a + b - 1;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Minus,
						Lexeme: "-",
					},
					LHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Plus,
							Lexeme: "+",
						},
						LHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
						},
						RHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
					},
					RHS: ast.IntegerLiteral{
						Integer: 1,
					},
				},
			},
		},
		{
			src: "name + \"is a good developer\";",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "name",
						},
					},
					RHS: ast.StringLiteral{
						String: "is a good developer",
					},
				},
			},
		},
		{
			src: "a + b * c;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Asterisk,
							Lexeme: "*",
						},
						LHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
						RHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "c",
							},
						},
					},
				},
			},
		},
		{
			src: "(a + b) * c;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Asterisk,
						Lexeme: "*",
					},
					LHS: ast.Grouping{
						Group: ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
							RHS: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "b",
								},
							},
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "c",
						},
					},
				},
			},
		},
		{
			src: "5 + -1 <= 6 * 5;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.LessEqual,
						Lexeme: "<=",
					},
					LHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Plus,
							Lexeme: "+",
						},
						LHS: ast.IntegerLiteral{Integer: 5},
						RHS: ast.Prefix{
							Operator: token.Token{
								Type:   token.Minus,
								Lexeme: "-",
							},
							Target: ast.IntegerLiteral{
								Integer: 1,
							},
						},
					},
					RHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Asterisk,
							Lexeme: "*",
						},
						LHS: ast.IntegerLiteral{Integer: 6},
						RHS: ast.IntegerLiteral{Integer: 5},
					},
				},
			},
		},
		{
			src: "\n5\n",
			// Since linefeed characters aren't much of a concern for the Squeak parser it makes sense that the error
			// actually appears on line 3, where we reach the end of the stream without having encountered a semicolon.
			err: SyntaxError{Line: 3},
		},
		{
			src: "\n5\n;\n",
			// This is an example of where the linefeed character is totally ignored (other than during line counting)
			// and it is therefor okay to defer the statement terminator (semicolon) to the next line (or several lines
			// down).
			expected: ast.ExpressionStatement{
				Expression: ast.IntegerLiteral{Integer: 5},
			},
		},
		{
			src: "",
			err: io.EOF,
		},
		{
			src: "\n\t\t\n# Hello world\n",
			err: io.EOF,
		},
		{
			src: "var name = \"crookdc\";",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.StringLiteral{
					String: "crookdc",
				},
			},
		},
		{
			src: "var name;",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
			},
		},
		{
			src: "var name ? \"crookdc\";",
			err: SyntaxError{
				Line: 1,
			},
		},
		{
			src: "var name = nil;",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.NilLiteral{},
			},
		},
		{
			src: "name = \"crookdc\";",
			expected: ast.ExpressionStatement{
				Expression: ast.Assignment{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "name",
					},
					Value: ast.StringLiteral{String: "crookdc"},
				},
			},
		},
		{
			src: "12.44 + 12;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.FloatLiteral{
						Float: 12.44,
					},
					RHS: ast.IntegerLiteral{
						Integer: 12,
					},
				},
			},
		},
		{
			src: "0.444456;",
			expected: ast.ExpressionStatement{
				Expression: ast.FloatLiteral{
					Float: 0.444456,
				},
			},
		},
		{
			src: "50.;",
			expected: ast.ExpressionStatement{
				Expression: ast.FloatLiteral{
					Float: 50,
				},
			},
		},
		{
			src: "run();",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "run",
						},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
					Args: []ast.ExpressionNode{},
				},
			},
		},
		{
			src: "index[4];",
			expected: ast.ExpressionStatement{
				Expression: ast.GetIndex{
					Target: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "index",
						},
					},
					Index: ast.IntegerLiteral{
						Integer: 4,
					},
				},
			},
		},
		{
			src: "indexed[add(a, b)];",
			expected: ast.ExpressionStatement{
				Expression: ast.GetIndex{
					Target: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "indexed",
						},
					},
					Index: ast.Call{
						Callee: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "add",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
							ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "b",
								},
							},
						},
					},
				},
			},
		},
		{
			src: "indexed[12;",
			err: SyntaxError{Line: 1},
		},
		{
			src: "var list = [1, 2, 3, true, false, \"crookdc\"];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: []ast.ExpressionNode{
						ast.IntegerLiteral{Integer: 1},
						ast.IntegerLiteral{Integer: 2},
						ast.IntegerLiteral{Integer: 3},
						ast.BooleanLiteral{Boolean: true},
						ast.BooleanLiteral{Boolean: false},
						ast.StringLiteral{String: "crookdc"},
					},
				},
			},
		},
		{
			src: "var list = [1 + 5, 9];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: []ast.ExpressionNode{
						ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.IntegerLiteral{
								Integer: 1,
							},
							RHS: ast.IntegerLiteral{
								Integer: 5,
							},
						},
						ast.IntegerLiteral{
							Integer: 9,
						},
					},
				},
			},
		},
		{
			src: "var list = [a];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: []ast.ExpressionNode{
						ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
						},
					},
				},
			},
		},
		{
			src: "developer.location.country;",
			expected: ast.ExpressionStatement{
				Expression: ast.GetProp{
					Target: ast.GetProp{
						Target: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "developer",
							},
						},
						Property: token.Token{
							Type:   token.Identifier,
							Lexeme: "location",
						},
					},
					Property: token.Token{
						Type:   token.Identifier,
						Lexeme: "country",
					},
				},
			},
		},
		{
			src: "developer.location.find(10);",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.GetProp{
						Target: ast.GetProp{
							Target: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "developer",
								},
							},
							Property: token.Token{
								Type:   token.Identifier,
								Lexeme: "location",
							},
						},
						Property: token.Token{
							Type:   token.Identifier,
							Lexeme: "find",
						},
					},
					Args: []ast.ExpressionNode{
						ast.IntegerLiteral{Integer: 10},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
				},
			},
		},
		{
			src: "developer.location.country = \"Sweden\";",
			expected: ast.ExpressionStatement{
				Expression: ast.SetProp{
					Target: ast.GetProp{
						Expression: ast.Expression{},
						Target: ast.GetProp{
							Target: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "developer",
								},
							},
							Property: token.Token{
								Type:   token.Identifier,
								Lexeme: "location",
							},
						},
						Property: token.Token{
							Type:   token.Identifier,
							Lexeme: "country",
						},
					},
					Property: token.Token{
						Type:   token.Identifier,
						Lexeme: "country",
					},
					Value: ast.StringLiteral{String: "Sweden"},
				},
			},
		},
		{
			src: "developer.age = 27;",
			expected: ast.ExpressionStatement{
				Expression: ast.SetProp{
					Target: ast.GetProp{
						Target: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "developer",
							},
						},
						Property: token.Token{
							Type:   token.Identifier,
							Lexeme: "age",
						},
					},
					Property: token.Token{
						Type:   token.Identifier,
						Lexeme: "age",
					},
					Value: ast.IntegerLiteral{Integer: 27},
				},
			},
		},
		{
			src: "developer.get();",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.GetProp{
						Target: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "developer",
							},
						},
						Property: token.Token{
							Type:   token.Identifier,
							Lexeme: "get",
						},
					},
					Args: []ast.ExpressionNode{},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
				},
			},
		},
		{
			src: `
			developer.birthday = function() {
				print("It is " + this.name + "'s birthday!");
			};
			`,
			expected: ast.ExpressionStatement{
				Expression: ast.SetProp{
					Target: ast.GetProp{
						Target: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "developer",
							},
						},
						Property: token.Token{
							Type:   token.Identifier,
							Lexeme: "birthday",
						},
					},
					Property: token.Token{
						Type:   token.Identifier,
						Lexeme: "birthday",
					},
					Value: ast.Method{
						Params: []token.Token{},
						Body: ast.Block{
							Body: []ast.StatementNode{
								ast.ExpressionStatement{
									Expression: ast.Call{
										Callee: ast.Variable{
											Level: 3,
											Name: token.Token{
												Type:   token.Identifier,
												Lexeme: "print",
											},
										},
										Operator: token.Token{
											Type:   token.LeftParenthesis,
											Lexeme: "(",
										},
										Args: []ast.ExpressionNode{
											ast.Infix{
												Operator: token.Token{
													Type:   token.Plus,
													Lexeme: "+",
												},
												LHS: ast.Infix{
													Operator: token.Token{
														Type:   token.Plus,
														Lexeme: "+",
													},
													LHS: ast.StringLiteral{
														String: "It is ",
													},
													RHS: ast.GetProp{
														Target: ast.Variable{
															Level: 1,
															Name: token.Token{
																Type:   token.Identifier,
																Lexeme: "this",
															},
														},
														Property: token.Token{
															Type:   token.Identifier,
															Lexeme: "name",
														},
													},
												},
												RHS: ast.StringLiteral{
													String: "'s birthday!",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "get_developers()[0].location;",
			expected: ast.ExpressionStatement{
				Expression: ast.GetProp{
					Target: ast.GetIndex{
						Target: ast.Call{
							Callee: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "get_developers",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{},
						},
						Index: ast.IntegerLiteral{
							Integer: 0,
						},
					},
					Property: token.Token{
						Type:   token.Identifier,
						Lexeme: "location",
					},
				},
			},
		},
		{
			src: "var list = [];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: make([]ast.ExpressionNode, 0),
				},
			},
		},
		{
			src: "run(5 + 1002, n);",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "run",
						},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
					Args: []ast.ExpressionNode{
						ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.IntegerLiteral{
								Integer: 5,
							},
							RHS: ast.IntegerLiteral{
								Integer: 1002,
							},
						},
						ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "n",
							},
						},
					},
				},
			},
		},
		{
			src: "factory()(5 + 1002, n)(n);",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Call{
						Callee: ast.Call{
							Callee: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "factory",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Infix{
								Operator: token.Token{
									Type:   token.Plus,
									Lexeme: "+",
								},
								LHS: ast.IntegerLiteral{
									Integer: 5,
								},
								RHS: ast.IntegerLiteral{
									Integer: 1002,
								},
							},
							ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "n",
								},
							},
						},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
					Args: []ast.ExpressionNode{
						ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "n",
							},
						},
					},
				},
			},
		},
		{
			src:      ";",
			expected: ast.Noop{},
		},
		{
			src: "while true {}",
			expected: ast.While{
				Condition: ast.BooleanLiteral{
					Boolean: true,
				},
				Body: ast.Block{
					Body: []ast.StatementNode{},
				},
			},
		},
		{
			src: "if a > b print(a); else print(b);",
			expected: ast.If{
				Condition: ast.Infix{
					Operator: token.Token{
						Type:   token.Greater,
						Lexeme: ">",
					},
					LHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
				Then: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "print",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
						},
					},
				},
				Else: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "print",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "b",
								},
							},
						},
					},
				},
			},
		},
		{
			src: "if a > b print(a);",
			expected: ast.If{
				Condition: ast.Infix{
					Operator: token.Token{
						Type:   token.Greater,
						Lexeme: ">",
					},
					LHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
				Then: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "print",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
						},
					},
				},
			},
		},
		{
			src: "if a if b print(b); else print(c);",
			expected: ast.If{
				Condition: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
				Then: ast.If{
					Condition: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
					Then: ast.ExpressionStatement{
						Expression: ast.Call{
							Callee: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "print",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{
								ast.Variable{
									Level: 1,
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "b",
									},
								},
							},
						},
					},
					Else: ast.ExpressionStatement{
						Expression: ast.Call{
							Callee: ast.Variable{
								Level: 1,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "print",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{
								ast.Variable{
									Level: 1,
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "c",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "{ var a; a + b; a = 2.; }",
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Declaration{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					ast.ExpressionStatement{
						Expression: ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.Variable{
								Level: 0,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
							RHS: ast.Variable{
								Level: 2,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "b",
								},
							},
						},
					},
					ast.ExpressionStatement{
						Expression: ast.Assignment{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
							Value: ast.FloatLiteral{
								Float: 2.0,
							},
						},
					},
				},
			},
		},
		{
			src: `
			{
				function get() {
					return name;
				}
			}
			`,
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Function{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "get",
						},
						Params: []token.Token{},
						Body: ast.Block{
							Body: []ast.StatementNode{
								ast.Return{
									Expression: ast.Variable{
										Level: 3,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: `
			{
				function get(name) {
					return name;
				}
			}
			`,
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Function{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "get",
						},
						Params: []token.Token{
							{
								Type:   token.Identifier,
								Lexeme: "name",
							},
						},
						Body: ast.Block{
							Body: []ast.StatementNode{
								ast.Return{
									Expression: ast.Variable{
										Level: 0,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: `
			{
				var name = "crookdc";
				function get() {
					name = "conker";
					return name;
				}
			}
			`,
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Declaration{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "name",
						},
						Initializer: ast.StringLiteral{
							String: "crookdc",
						},
					},
					ast.Function{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "get",
						},
						Params: []token.Token{},
						Body: ast.Block{
							Body: []ast.StatementNode{
								ast.ExpressionStatement{
									Expression: ast.Assignment{
										Level: 1,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "name",
										},
										Value: ast.StringLiteral{
											String: "conker",
										},
									},
								},
								ast.Return{
									Expression: ast.Variable{
										Level: 1,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: `
			{
				var name = "crookdc";
				var name = "crookdc2";
			}
			`,
			err: SyntaxError{Line: 4},
		},
		{
			src: `
			Object {
				status: Object {
					code: 200 + 1,	
					line: "Created"
				},
				success: true
			};
			`,
			expected: ast.ExpressionStatement{
				Expression: ast.ObjectLiteral{
					Properties: map[string]ast.ExpressionNode{
						"status": ast.ObjectLiteral{
							Properties: map[string]ast.ExpressionNode{
								"code": ast.Infix{
									Operator: token.Token{
										Type:   token.Plus,
										Lexeme: "+",
									},
									LHS: ast.IntegerLiteral{Integer: 200},
									RHS: ast.IntegerLiteral{Integer: 1},
								},
								"line": ast.StringLiteral{String: "Created"},
							},
						},
						"success": ast.BooleanLiteral{Boolean: true},
					},
				},
			},
		},
		{
			src: `
			Object {
				start: function() {
					print("working...");
				},
				end: function() {
					print("worked");
				}
			};
			`,
			expected: ast.ExpressionStatement{
				Expression: ast.ObjectLiteral{
					Properties: map[string]ast.ExpressionNode{
						"start": ast.Method{
							Params: []token.Token{},
							Body: ast.Block{
								Body: []ast.StatementNode{
									ast.ExpressionStatement{
										Expression: ast.Call{
											Callee: ast.Variable{
												Level: 3,
												Name: token.Token{
													Type:   token.Identifier,
													Lexeme: "print",
												},
											},
											Operator: token.Token{
												Type:   token.LeftParenthesis,
												Lexeme: "(",
											},
											Args: []ast.ExpressionNode{
												ast.StringLiteral{
													String: "working...",
												},
											},
										},
									},
								},
							},
						},
						"end": ast.Method{
							Params: []token.Token{},
							Body: ast.Block{
								Body: []ast.StatementNode{
									ast.ExpressionStatement{
										Expression: ast.Call{
											Callee: ast.Variable{
												Level: 3,
												Name: token.Token{
													Type:   token.Identifier,
													Lexeme: "print",
												},
											},
											Operator: token.Token{
												Type:   token.LeftParenthesis,
												Lexeme: "(",
											},
											Args: []ast.ExpressionNode{
												ast.StringLiteral{
													String: "worked",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: `
			{
				var name = "crickets";
				function get() {
					var name = "crookdc";
					name = "some other name";
					return name;
				}
			}
			`,
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Declaration{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "name",
						},
						Initializer: ast.StringLiteral{
							String: "crickets",
						},
					},
					ast.Function{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "get",
						},
						Params: []token.Token{},
						Body: ast.Block{
							Body: []ast.StatementNode{
								ast.Declaration{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "name",
									},
									Initializer: ast.StringLiteral{
										String: "crookdc",
									},
								},
								ast.ExpressionStatement{
									Expression: ast.Assignment{
										Level: 0,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "name",
										},
										Value: ast.StringLiteral{
											String: "some other name",
										},
									},
								},
								ast.Return{
									Expression: ast.Variable{
										Level: 0,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "name",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "while a < b and b >= 3 { print(a); }",
			expected: ast.While{
				Condition: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Less,
							Lexeme: "<",
						},
						LHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
						},
						RHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
					},
					RHS: ast.Infix{
						Operator: token.Token{
							Type:   token.GreaterEqual,
							Lexeme: ">=",
						},
						LHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
						RHS: ast.IntegerLiteral{
							Integer: 3,
						},
					},
				},
				Body: ast.Block{
					Body: []ast.StatementNode{
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Level: 2,
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Variable{
										Level: 2,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "a",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "while a < b { print(a); print(b); }",
			expected: ast.While{
				Condition: ast.Infix{
					Operator: token.Token{
						Type:   token.Less,
						Lexeme: "<",
					},
					LHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
				Body: ast.Block{
					Body: []ast.StatementNode{
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Level: 2,
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Variable{
										Level: 2,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "a",
										},
									},
								},
							},
						},
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Level: 2,
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Variable{
										Level: 2,
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "b",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "{}",
			expected: ast.Block{
				Body: []ast.StatementNode{},
			},
		},
		{
			src: "1 == 1 and b;",
			expected: ast.ExpressionStatement{
				Expression: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Equals,
							Lexeme: "==",
						},
						LHS: ast.IntegerLiteral{
							Integer: 1,
						},
						RHS: ast.IntegerLiteral{
							Integer: 1,
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
			},
		},
		{
			src: "1 == 1 and b or c;",
			expected: ast.ExpressionStatement{
				Expression: ast.Logical{
					Operator: token.Token{
						Type:   token.Or,
						Lexeme: "or",
					},
					LHS: ast.Logical{
						Operator: token.Token{
							Type:   token.And,
							Lexeme: "and",
						},
						LHS: ast.Infix{
							Operator: token.Token{
								Type:   token.Equals,
								Lexeme: "==",
							},
							LHS: ast.IntegerLiteral{
								Integer: 1,
							},
							RHS: ast.IntegerLiteral{
								Integer: 1,
							},
						},
						RHS: ast.Variable{
							Level: 1,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
					},
					RHS: ast.Variable{
						Level: 1,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "c",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src))
			assert.Nil(t, err)
			plx, err := NewPeekingLexer(lx)
			assert.Nil(t, err)

			ps := NewParser(plx)
			n, err := ps.Next()
			assert.ErrorIs(t, err, test.err)
			if err == nil {
				assert.Equal(t, test.expected, n)
			}
		})
	}

	t.Run("clears current statement if error occurs", func(t *testing.T) {
		src := `
		a +/ b; # This line contains an invalid evaluate 
		a + b;`
		lx, err := NewLexer(strings.NewReader(src))
		assert.Nil(t, err)
		plx, err := NewPeekingLexer(lx)
		assert.Nil(t, err)

		ps := NewParser(plx)
		_, err = ps.Next()
		assert.ErrorIs(t, err, SyntaxError{Line: 2})

		n, err := ps.Next()
		assert.Nil(t, err)
		assert.Equal(t, ast.ExpressionStatement{
			Expression: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
				RHS: ast.Variable{
					Level: 1,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "b",
					},
				},
			},
		}, n)
	})
}
