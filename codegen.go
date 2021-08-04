package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
)

type CodeGen struct {
	current            Token
	previous           Token
	chunk              Chunk
	had_error          bool
	panic_mode         bool
	generate_EOF_token bool
}

func (gen *CodeGen) error_at(token *Token, msg string) {
	if gen.panic_mode {
		return
	}

	gen.panic_mode = true
	fmt.Printf("[Line: %d] Error", token.line)

	if is_Token_of_type(*token, TOKEN_EOF) {
		fmt.Printf(" at end")
	} else if is_Token_of_type(*token, TOKEN_ERROR) {

	} else {
		fmt.Printf(" at %s", token.lexeme)
	}

	log.Printf(": %s\n", msg)
	gen.had_error = true
}

func (gen *CodeGen) error_at_current(msg string) {
	gen.error_at(&gen.current, msg)
}

func (gen *CodeGen) error_at_previous(msg string) {
	gen.error_at(&gen.previous, msg)
}

func (gen *CodeGen) advance_g() {
	gen.previous = gen.current

	for {
		gen.current = scan_token()
		if gen.current.t_type != TOKEN_ERROR {
			break
		}

		gen.error_at_current(gen.current.lexeme)
	}
}

func (gen *CodeGen) consume(t_type Token_Type, err_msg string) {
	if is_Token_of_type(gen.current, t_type) {
		gen.advance_g()
		return
	}

	gen.error_at_current(err_msg)
}

func (gen *CodeGen) emit_byte(byte_ byte) {
	gen.chunk.write_chunk(byte_, uint32(gen.previous.line))
}

func (gen *CodeGen) emit_constant(value Value) {
	gen.chunk.write_constant(OP_PUSH, value, uint32(gen.previous.line))
}

func (gen *CodeGen) emit_jmp(op byte, index uint32) {
	gen.chunk.write_jmp(op, index, uint32(gen.previous.line))
}

func (gen *CodeGen) emit_store(name string, t ValueTypes) {
	gen.chunk.write_store(OP_STORE, name, t, uint32(gen.previous.line))
}

func (gen *CodeGen) emit_load(name string) {
	gen.chunk.write_load(OP_LOAD, name, uint32(gen.previous.line))
}

func (gen *CodeGen) emit_call_func(name string) {
	gen.chunk.write_call_func(OP_CALL_FUNC, name, uint32(gen.previous.line))
}

func (gen *CodeGen) literals() {
	gen.compile_literal(gen.current)
}

func (gen *CodeGen) compile_literal(literal Token) {
	switch literal.t_type {
	case TOKEN_UINT:
		value, err := strconv.ParseUint(literal.lexeme, 10, 64)
		if err != nil {
			log.Panicln("Failed to convert value to uint")
		}
		gen.emit_constant(UINT_VAL(value))

	case TOKEN_INT:
		value, err := strconv.ParseInt(literal.lexeme, 10, 64)
		if err != nil {
			log.Panicln("Failed to convert value to int")
		}
		gen.emit_constant(INT_VAL(value))

	case TOKEN_DECIMAL:
		value, err := strconv.ParseFloat(literal.lexeme, 10)
		if err != nil {
			log.Panicln("Failed to convert value to decimal")
		}
		gen.emit_constant(DECIMAL_VAL(value))
	}
}

func (gen *CodeGen) generate_patch_jmp(op byte) int {
	gen.emit_jmp(op, 0)

	return len(gen.chunk.code) - 5
}

func (gen *CodeGen) patch_jump(area_patch int, jmp_to_pos uint32) {
	var bytes []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, jmp_to_pos)

	gen.chunk.code[area_patch+1] = bytes[0]
	gen.chunk.code[area_patch+2] = bytes[1]
	gen.chunk.code[area_patch+3] = bytes[2]
	gen.chunk.code[area_patch+4] = bytes[3]
}

func (gen *CodeGen) lists() {
	gen.consume(TOKEN_LEFT_PAREN, "Expected '(' before list.")

	first_token := gen.current

	switch first_token.t_type {
	case TOKEN_VAR:
		gen.advance_g()
		name := gen.current.lexeme
		gen.consume(TOKEN_IDENTIFER, "Expected an identifer after 'var'")

		literal_type := gen.current.t_type
		amount := len(gen.chunk.code)
		gen.expression()
		if amount < len(gen.chunk.code) {
			gen.emit_byte(OP_EOF)
			vm := new_VM(&gen.chunk)

			value_type := vm.evaluate_operation()
			gen.chunk.code = gen.chunk.code[0 : len(gen.chunk.code)-1]
			gen.emit_store(name, value_type)
			break
		}

		gen.advance_g()
		gen.expression()

		var value_type ValueTypes

		switch literal_type {
		case TOKEN_TYPE_INT:
			value_type = INT

		case TOKEN_TYPE_UINT:
			value_type = UINT

		case TOKEN_TYPE_STRING:
			value_type = STRING

		case TOKEN_TYPE_BOOL:
			value_type = BOOL

		case TOKEN_TYPE_DECIMAL:
			value_type = DECIMAL

		default:
			gen.error_at_current("Expected a type to be specified")
		}

		gen.emit_store(name, value_type)

	case TOKEN_IF:
		gen.advance_g()
		if gen.current.t_type != TOKEN_LEFT_PAREN {
			gen.expression() // This should generate a boolean!
			gen.advance_g()
		} else {
			gen.expression()
		}

		patch_area := gen.generate_patch_jmp(OP_IF_FALSE_JMP)
		if gen.current.t_type != TOKEN_LEFT_PAREN {
			gen.expression() // This should generate a boolean!
			gen.advance_g()
		} else {
			gen.expression()
		}
		else_patch_area := gen.generate_patch_jmp(OP_JMP)
		gen.patch_jump(patch_area, uint32(len(gen.chunk.code)))
		if gen.current.t_type != TOKEN_LEFT_PAREN {
			gen.expression() // This should generate a boolean!
			gen.advance_g()
		} else {
			gen.expression()
		}
		gen.patch_jump(else_patch_area, uint32(len(gen.chunk.code)))

	case TOKEN_FUNC:
		gen.advance_g()
		name := gen.current.lexeme
		gen.consume(TOKEN_IDENTIFER, "Expected an identifer after 'func'.")

		gen.consume(TOKEN_LEFT_BRACKET, "Expected '[' before function arguments.")
		var function_args []string
		var function_types []ValueTypes

		for gen.current.t_type != TOKEN_RIGHT_BRACKET {
			arg_name := gen.current.lexeme
			gen.consume(TOKEN_IDENTIFER, "Expected an identifer for function argument.")
			literal_type := gen.current.t_type
			gen.advance_g()
			var value_type ValueTypes

			switch literal_type {
			case TOKEN_TYPE_INT:
				value_type = INT

			case TOKEN_TYPE_UINT:
				value_type = UINT

			case TOKEN_TYPE_STRING:
				value_type = STRING

			case TOKEN_TYPE_BOOL:
				value_type = BOOL

			case TOKEN_TYPE_DECIMAL:
				value_type = DECIMAL

			default:
				gen.error_at_current("Expected a type to be specified")
			}

			function_args = append(function_args, arg_name)
			function_types = append(function_types, value_type)

			if gen.current.t_type != TOKEN_RIGHT_BRACKET {
				gen.consume(TOKEN_COMMA, "Expected a ',' before next argument")
			}
		}

		gen.consume(TOKEN_RIGHT_BRACKET, "Expected ']' after function arguments.")
		literal_type := gen.current.t_type
		gen_chunk := gen.chunk
		gen.chunk.init_chunk()
		gen.emit_byte(OP_START_SCOPE)
		amount := len(gen.chunk.code)
		fmt.Println("AMOUNT:", amount)
		calculated_arugment_amount := 0
		for i, v := range function_args {
			calculated_arugment_amount += len(v) + 3
			gen.chunk.write_store(OP_STORE, v, function_types[i], uint32(gen.previous.line))
		}
		fmt.Println("CHUNK AMOUNT:", len(gen.chunk.code)-calculated_arugment_amount)
		gen.expression()

		/////////////////////////////
		// FIX INFERRING
		///////////////////////////
		if amount < len(gen.chunk.code)-calculated_arugment_amount {
			fmt.Println(amount < len(gen.chunk.code)-calculated_arugment_amount)
			fmt.Println("RUNNING")
			gen.emit_byte(OP_END_SCOPE)
			body_chunk := gen.chunk
			gen.chunk = gen_chunk
			gen_chunk.free_chunk()

			ftable.add_virtual_entry(name, body_chunk, uint(len(function_args)), NO_VALUE)
			break
		}

		gen.advance_g()
		gen.expression()
		gen.emit_byte(OP_END_SCOPE)

		body_chunk := gen.chunk
		gen.chunk = gen_chunk
		gen_chunk.free_chunk()

		var value_type ValueTypes

		switch literal_type {
		case TOKEN_TYPE_INT:
			value_type = INT

		case TOKEN_TYPE_UINT:
			value_type = UINT

		case TOKEN_TYPE_STRING:
			value_type = STRING

		case TOKEN_TYPE_BOOL:
			value_type = BOOL

		case TOKEN_TYPE_DECIMAL:
			value_type = DECIMAL

		default:
			gen.error_at_current("Expected a type to be specified")
		}

		ftable.add_virtual_entry(name, body_chunk, uint(len(function_args)), value_type)

	case TOKEN_LEFT_PAREN:
		for gen.current.t_type != TOKEN_RIGHT_PAREN {
			gen.expression()
		}
	}

	arguments := 0

	for gen.current.t_type != TOKEN_RIGHT_PAREN {
		gen.advance_g()
		for gen.current.t_type == TOKEN_LEFT_PAREN {
			gen.expression()
			arguments += 1
		}
		gen.expression()

		arguments += 1

		if gen.current.t_type == TOKEN_EOF {
			break
		}
	}

	//fmt.Println("ARGUMENTS:", arguments)
	//fmt.Println("TYPE:", first_token.t_type)

	switch first_token.t_type {
	case TOKEN_PLUS:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_ADD)
		}

	case TOKEN_MINUS:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_SUB)
		}

	case TOKEN_STAR:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_MUL)
		}

	case TOKEN_SLASH:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_DIV)
		}

	case TOKEN_LESS:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_CMP_LESS)
		}

	case TOKEN_LESS_EQUAL:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_CMP_LESS_EQUAL)
		}

	case TOKEN_GREATER:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_CMP_GREATER)
		}

	case TOKEN_GREATER_EQUAL:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_CMP_GREATER_EQUAL)
		}

	case TOKEN_EQUAL_EQUAL:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_CMP_EQUAL)
		}

	case TOKEN_NOT_EQUAL:
		for i := 0; i < arguments-2; i++ {
			gen.emit_byte(OP_CMP_NOT_EQUAL)
		}

	case TOKEN_IDENTIFER:
		gen.emit_call_func(first_token.lexeme)

	case TOKEN_RETURN:
		gen.emit_byte(OP_RETURN)
	}

	gen.consume(TOKEN_RIGHT_PAREN, "Expected ')' after list.")
}

func (gen *CodeGen) identifer_g() {
	gen.emit_load(gen.current.lexeme)
	//gen.consume(TOKEN_IDENTIFER, "Expected an identifer.")
}

func (gen *CodeGen) expression() {
	switch gen.current.t_type {
	case TOKEN_IDENTIFER:
		gen.identifer_g()
	case TOKEN_UINT, TOKEN_INT, TOKEN_BOOL, TOKEN_DECIMAL, TOKEN_STRING:
		gen.literals()
	case TOKEN_LEFT_PAREN:
		gen.lists()
	}
}

func (gen *CodeGen) generate_chunk(file_path string) Chunk {
	new_Scanner(file_path)

	gen.advance_g()
	for gen.current.t_type != TOKEN_EOF {
		gen.expression()
	}

	gen.consume(TOKEN_EOF, "Expected end of expression.")

	if gen.generate_EOF_token {
		gen.emit_byte(OP_EOF)
	}
	return gen.chunk
}

func new_CodeGen(generate_EOF_token bool) CodeGen {
	gen := CodeGen{}
	gen.chunk.init_chunk()
	gen.generate_EOF_token = generate_EOF_token

	return gen
}
