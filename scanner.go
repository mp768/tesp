package main

import (
	"fmt"
	"log"
	"os"
)

type Token_Type byte

const (
	TOKEN_NONE Token_Type = iota
	TOKEN_EOF
	TOKEN_ERROR

	TOKEN_PRINT
	TOKEN_PRINTLN
	TOKEN_VAR
	TOKEN_ASSIGN
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_AND
	TOKEN_OR
	TOKEN_SWITCH
	TOKEN_FOR
	TOKEN_BREAK
	TOKEN_FUNC
	TOKEN_RETURN

	TOKEN_LEFT_PAREN
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_LEFT_BRACKET
	TOKEN_RIGHT_BRACKET
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_SEMICOLON

	TOKEN_COLON
	TOKEN_COLON_EQUAL
	TOKEN_EQUAL
	TOKEN_EQUAL_EQUAL
	TOKEN_NOT
	TOKEN_NOT_EQUAL
	TOKEN_GREATER
	TOKEN_GREATER_EQUAL
	TOKEN_LESS
	TOKEN_LESS_EQUAL

	TOKEN_TYPE_STRING
	TOKEN_TYPE_INT
	TOKEN_TYPE_DECIMAL
	TOKEN_TYPE_UINT
	TOKEN_TYPE_BOOL

	// Literal
	TOKEN_STRING
	TOKEN_INT
	TOKEN_DECIMAL
	TOKEN_UINT
	TOKEN_BOOL
	TOKEN_IDENTIFER
)

type Scanner struct {
	chars   []byte
	current uint
	start   uint
	line    uint
}

type Token struct {
	t_type Token_Type
	lexeme string
	line   uint
}

var scanner Scanner

func is_Token_of_type(token Token, t_type Token_Type) bool {
	return token.t_type == t_type
}

func make_Token(t_type Token_Type) Token {
	var token = Token{}
	token.t_type = t_type
	token.lexeme = string(scanner.chars[scanner.start:scanner.current])
	token.line = scanner.line
	return token
}

func make_Token_len(t_type Token_Type, start uint, current uint) Token {
	var token = Token{}
	token.t_type = t_type
	token.lexeme = string(scanner.chars[start:current])
	token.line = scanner.line
	return token
}

func error_token(msg string) Token {
	var token = Token{}
	token.t_type = TOKEN_ERROR
	token.lexeme = msg
	token.line = scanner.line
	return token
}

func is_at_end() bool {
	if len(scanner.chars) <= int(scanner.current) {
		return true
	}

	return scanner.chars[scanner.current] == '\000'
}

func advance() (result byte) {
	result = peek()
	scanner.current++
	return
}

func new_Scanner(file_path string) {
	chars, err := os.ReadFile(file_path)
	if err != nil {
		log.Panic("custom", err.Error())
	}

	scanner.chars = append(chars, '\000')
	scanner.current = 0
	scanner.start = 0
	scanner.line = 1
	return
}

func match(expected byte) bool {
	if is_at_end() {
		return false
	}
	if scanner.chars[scanner.current] != expected {
		return false
	}
	scanner.current++
	return true
}

func peek() byte {
	if is_at_end() {
		return '\000'
	}

	return scanner.chars[scanner.current]
}

func peek_next() byte {
	if is_at_end() {
		return '\000'
	}
	return scanner.chars[scanner.current+1]
}

func skip_whitespace() {
	for {
		var c = peek()

		switch c {
		case ' ', '\r', '\t':
			advance()

		case '\n':
			scanner.line++
			advance()

		case '/':
			if peek_next() == '/' {
				for peek() != '\n' && !is_at_end() {
					advance()
				}
			} else {
				return
			}

		default:
			return
		}
	}
}

func is_digit(c byte) bool {
	return c <= '9' && c >= '0'
}

func is_alpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func string_Token() Token {
	for peek() != '"' && !is_at_end() {
		if peek() == '\n' {
			scanner.line++
		}
		advance()
	}

	if is_at_end() {
		return error_token("Unterminated string.")
	}

	advance()

	return make_Token_len(TOKEN_STRING, scanner.start+1, scanner.current-1)
}

func number_Token() Token {
	for is_digit(peek()) {
		advance()
	}

	if peek() == '.' && is_digit(peek_next()) {
		advance()

		for is_digit(peek()) {
			advance()
		}

		return make_Token(TOKEN_DECIMAL)
	} else if peek() == 'u' {
		advance()
		return make_Token_len(TOKEN_UINT, scanner.start, scanner.current-1)
	}

	return make_Token(TOKEN_INT)
}

func identifer_Token() Token {
	for is_alpha(peek()) || is_digit(peek()) {
		advance()
	}

	switch string(scanner.chars[scanner.start:scanner.current]) {
	case "print":
		return make_Token(TOKEN_PRINT)
	case "println":
		return make_Token(TOKEN_PRINTLN)
	case "var":
		return make_Token(TOKEN_VAR)
	case "true":
		return make_Token(TOKEN_TRUE)
	case "false":
		return make_Token(TOKEN_FALSE)
	case "if":
		return make_Token(TOKEN_IF)
	case "else":
		return make_Token(TOKEN_ELSE)
	case "and":
		return make_Token(TOKEN_AND)
	case "or":
		return make_Token(TOKEN_OR)
	case "switch":
		return make_Token(TOKEN_SWITCH)
	case "for":
		return make_Token(TOKEN_FOR)
	case "break":
		return make_Token(TOKEN_BREAK)
	case "func":
		return make_Token(TOKEN_FUNC)
	case "return":
		return make_Token(TOKEN_RETURN)

	case "assign":
		return make_Token(TOKEN_ASSIGN)

	case "int":
		return make_Token(TOKEN_TYPE_INT)

	case "uint":
		return make_Token(TOKEN_TYPE_UINT)

	case "bool":
		return make_Token(TOKEN_TYPE_BOOL)

	case "decimal":
		return make_Token(TOKEN_TYPE_DECIMAL)

	case "string":
		return make_Token(TOKEN_TYPE_STRING)
	}

	return make_Token(TOKEN_IDENTIFER)
}

func scan_token() Token {
	skip_whitespace()
	scanner.start = scanner.current

	if is_at_end() {
		return make_Token(TOKEN_EOF)
	}

	var c = advance()

	if is_digit(c) {
		return number_Token()
	}

	if is_alpha(c) {
		return identifer_Token()
	}

	switch c {
	case '(':
		return make_Token(TOKEN_LEFT_PAREN)
	case ')':
		return make_Token(TOKEN_RIGHT_PAREN)
	case '{':
		return make_Token(TOKEN_LEFT_BRACE)
	case '}':
		return make_Token(TOKEN_RIGHT_BRACE)
	case '[':
		return make_Token(TOKEN_LEFT_BRACKET)
	case ']':
		return make_Token(TOKEN_RIGHT_BRACKET)
	case ';':
		return make_Token(TOKEN_SEMICOLON)
	case ',':
		return make_Token(TOKEN_COMMA)
	case '.':
		return make_Token(TOKEN_DOT)
	case '-':
		return make_Token(TOKEN_MINUS)
	case '+':
		return make_Token(TOKEN_PLUS)
	case '*':
		return make_Token(TOKEN_STAR)
	case '/':
		return make_Token(TOKEN_SLASH)

	case '!':
		if match('=') {
			return make_Token(TOKEN_NOT_EQUAL)
		} else {
			return make_Token(TOKEN_NOT)
		}

	case '=':
		if match('=') {
			return make_Token(TOKEN_EQUAL_EQUAL)
		} else {
			return make_Token(TOKEN_EQUAL)
		}

	case '>':
		if match('=') {
			return make_Token(TOKEN_GREATER_EQUAL)
		} else {
			return make_Token(TOKEN_GREATER)
		}

	case '<':
		if match('=') {
			return make_Token(TOKEN_LESS_EQUAL)
		} else {
			return make_Token(TOKEN_LESS)
		}

	case ':':
		if match('=') {
			return make_Token(TOKEN_COLON_EQUAL)
		} else {
			return make_Token(TOKEN_COLON)
		}

	case '"':
		return string_Token()
	}

	fmt.Printf("character that stopped at: '%c'\n", c)
	return error_token("Unexpected character.")
}
