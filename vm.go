package main

import (
	"encoding/binary"
	"fmt"
	"log"
)

const (
	VM_TYPE_SCRIPT byte = iota
	VM_TYPE_FUNCTION
)

var env Environment

// The main purpose of this is to test out features of the compiler to see if they are implemented correctly
// This will essentially emulate what I plan for the bytecode to be compiled to
// It will also be used as the base for the passes in the compiler, like the infer pass or optimization pass

type VM struct {
	chunk         *Chunk
	index         uint32
	stack         ValueArray
	type_to_check byte
}

type InterpreterResult byte

const (
	INTERPRETER_RESULT_OK InterpreterResult = iota
	INTERPETER_RESULT_COMPILE_ERROR
	INTERPRETER_RESULT_INTERPET_ERROR
)

func new_VM(chunk *Chunk) (result VM) {
	var valueStack ValueArray
	init_ValueArray(&valueStack)
	result = VM{
		chunk,
		0,
		valueStack,
		VM_TYPE_SCRIPT,
	}
	return
}

func (vm *VM) evaluate_operation() (result ValueTypes) {
	interpret(vm)
	env.remove_scope(env.currentScope)
	env.currentScope--

	vm.index = 0
	result = pop_ValueArray(&vm.stack).value_type
	free_ValueArray(&vm.stack)
	return
}

func evaluate_function(name string, values []Value) (result Value, returned_type ValueTypes) {
	function := ftable.get_entry(name)

	if function.f_type == FUNCTION_VIRTUAL {

		vm := new_VM(&function.body)
		vm.type_to_check = VM_TYPE_FUNCTION
		for _, v := range values {
			write_ValueArray(&vm.stack, v)
		}
		interpret(&vm)
		if function.return_type != NO_VALUE {
			result = pop_ValueArray(&vm.stack)
			returned_type = result.value_type
		} else {
			result = Value{}
			returned_type = NO_VALUE
		}

		free_VM(&vm)
	} else if function.f_type == FUNCTION_NATIVE {
		result, returned_type = function.native_body(values)
	}

	return
}

func interpret(vm *VM) InterpreterResult {
	READ_BYTE := func() (result byte) {

		result = vm.chunk.code[vm.index]
		vm.index++
		return
	}

	GENERATE_VALUE_FOR_BINARY_OP := func() (a Value, b Value, types ValueTypes) {
		b = pop_ValueArray(&vm.stack)
		a = pop_ValueArray(&vm.stack)

		types = 0

		if types < a.value_type {
			types = a.value_type
		}

		if types < b.value_type {
			types = b.value_type
		}

		return
	}

	var READ_CONSTANT func() Value = func() Value {
		index := binary.BigEndian.Uint16([]byte{READ_BYTE(), READ_BYTE()})
		return vm.chunk.constants.values[index]
	}

	// I guess this is my only solution for debugging.
	// ¯\_(ツ)_/¯
	debugging := false
	debug_entries := true
	debug_values := true

	for {
		if len(vm.chunk.code) == int(vm.index) {
			return INTERPRETER_RESULT_OK
		}

		if debugging {
			if debug_entries {
				env.print_entries()
				fmt.Println()
			}

			if debug_values {
				fmt.Println("===       Stack       ===")
				for _, v := range vm.stack.values {
					print_Value(v)
					fmt.Println(" ", ValueTypes_to_string(v.value_type), " ")
				}
				fmt.Println()
			}
		}

		instruction := READ_BYTE()

		switch instruction {
		case OP_START_SCOPE:
			env.currentScope++
		case OP_END_SCOPE:
			env.remove_scope(env.currentScope)
			env.currentScope--

		case OP_ASSIGN:
			var bytes_of_name []byte
			current_byte := READ_BYTE()
			for i := 0; current_byte != OP_END_QUOTE; i++ {
				bytes_of_name = append(bytes_of_name, current_byte)
				current_byte = READ_BYTE()
			}
			name := string(bytes_of_name)

			env.assign_to_entry(name, pop_ValueArray(&vm.stack))

		case OP_CALL_FUNC:
			var bytes_of_name []byte
			current_byte := READ_BYTE()
			for i := 0; current_byte != OP_END_QUOTE; i++ {
				bytes_of_name = append(bytes_of_name, current_byte)
				current_byte = READ_BYTE()
			}
			name := string(bytes_of_name)

			function := ftable.get_entry(name)
			var values []Value

			for i := 0; uint(i) < function.arity; i++ {
				values = append(values, pop_ValueArray(&vm.stack))
			}

			value, rtype := evaluate_function(name, values)

			if rtype != NO_VALUE {
				write_ValueArray(&vm.stack, value)
			}

		case OP_JMP:
			value := binary.BigEndian.Uint32([]byte{READ_BYTE(), READ_BYTE(), READ_BYTE(), READ_BYTE()})
			vm.index = value

		case OP_IF_FALSE_JMP:
			var result = pop_ValueArray(&vm.stack)

			value := binary.BigEndian.Uint32([]byte{READ_BYTE(), READ_BYTE(), READ_BYTE(), READ_BYTE()})

			if !IS_OF_TYPE(&result, BOOL) {
				log.Panic("Boolean value is required for an If False Jump instruction.")
			}

			if !TO_BOOL_S(&result) {
				vm.index = value
			}

		case OP_NEGATE:
			value := pop_ValueArray(&vm.stack)

			switch value.value_type {
			case INT:
			case UINT:
				write_ValueArray(&vm.stack, INT_VAL(-TO_INT_S(&value)))

			case DECIMAL:
				write_ValueArray(&vm.stack, DECIMAL_VAL(-TO_DECIMAL_S(&value)))

			default:
				log.Panic("Value cannot be negated!")
			}

		case OP_CMP_LESS:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_INT_S(&a) < TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_UINT_S(&a) < TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_DECIMAL_S(&a) < TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot compare (CMP_LESS) these two values!")
			}

		case OP_CMP_GREATER:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_INT_S(&a) > TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_UINT_S(&a) > TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_DECIMAL_S(&a) > TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot compare (CMP_GREATER) these two values!")
			}

		case OP_CMP_EQUAL:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_INT_S(&a) == TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_UINT_S(&a) == TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_DECIMAL_S(&a) == TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot compare (CMP_EQUAL) these two values!")
			}

		case OP_CMP_NOT_EQUAL:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_INT_S(&a) != TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_UINT_S(&a) != TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_DECIMAL_S(&a) != TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot compare (CMP_NOT_EQUAL) these two values!")
			}

		case OP_CMP_LESS_EQUAL:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_INT_S(&a) <= TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_UINT_S(&a) <= TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_DECIMAL_S(&a) <= TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot compare (CMP_LESS_EQUAL) these two values!")
			}

		case OP_CMP_GREATER_EQUAL:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_INT_S(&a) >= TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_UINT_S(&a) >= TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, BOOL_VAL(TO_DECIMAL_S(&a) >= TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot compare (CMP_GREATER_EQUAL) these two values!")
			}

		case OP_CMP_AND:
			a, b, _ := GENERATE_VALUE_FOR_BINARY_OP()

			if a.value_type != BOOL && b.value_type != BOOL {
				log.Panic("The two values are not booleans, for an 'and' operation it is required to have two booleans")
			} else {
				write_ValueArray(&vm.stack, BOOL_VAL(TO_BOOL_S(&a) && TO_BOOL_S(&b)))
			}

		case OP_CMP_OR:
			a, b, _ := GENERATE_VALUE_FOR_BINARY_OP()

			if a.value_type != BOOL && b.value_type != BOOL {
				log.Panic("The two values are not booleans, for an 'or' operation it is required to have two booleans")
			} else {
				write_ValueArray(&vm.stack, BOOL_VAL(TO_BOOL_S(&a) || TO_BOOL_S(&b)))
			}

		// Don't know how to debloat this
		case OP_ADD:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, INT_VAL(TO_INT_S(&a)+TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, UINT_VAL(TO_UINT_S(&a)+TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, DECIMAL_VAL(TO_DECIMAL_S(&a)+TO_DECIMAL_S(&b)))
			case STRING:
				write_ValueArray(&vm.stack, STRING_VAL(TO_STRING_S(&a)+TO_STRING_S(&b)))

			default:
				log.Panic("Cannot add these two binary operations!")
			}

		case OP_SUB:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, INT_VAL(TO_INT_S(&a)-TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, UINT_VAL(TO_UINT_S(&a)-TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, DECIMAL_VAL(TO_DECIMAL_S(&a)-TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot subtract these two binary operations!")
			}

		case OP_MUL:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, INT_VAL(TO_INT_S(&a)*TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, UINT_VAL(TO_UINT_S(&a)*TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, DECIMAL_VAL(TO_DECIMAL_S(&a)*TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot mutliple these two binary operations!")
			}

		case OP_DIV:
			a, b, value_type := GENERATE_VALUE_FOR_BINARY_OP()

			switch value_type {
			case INT:
				write_ValueArray(&vm.stack, INT_VAL(TO_INT_S(&a)/TO_INT_S(&b)))
			case UINT:
				write_ValueArray(&vm.stack, UINT_VAL(TO_UINT_S(&a)/TO_UINT_S(&b)))
			case DECIMAL:
				write_ValueArray(&vm.stack, DECIMAL_VAL(TO_DECIMAL_S(&a)/TO_DECIMAL_S(&b)))

			default:
				log.Panic("Cannot divide these two binary operations!")
			}

		case OP_LOAD:
			var bytes_of_name []byte
			current_byte := READ_BYTE()
			for i := 0; current_byte != OP_END_QUOTE; i++ {
				bytes_of_name = append(bytes_of_name, current_byte)
				current_byte = READ_BYTE()
			}
			current_name := string(bytes_of_name)

			write_ValueArray(&vm.stack, env.get_variable_value(current_name))

		// This is a long instruction, Idk how to compact this
		case OP_STORE:
			var bytes_of_name []byte
			current_byte := READ_BYTE()
			for i := 0; current_byte != OP_END_QUOTE; i++ {
				bytes_of_name = append(bytes_of_name, current_byte)
				current_byte = READ_BYTE()
			}

			current_name := string(bytes_of_name)
			current_type := ValueTypes(READ_BYTE())
			current_value := pop_ValueArray(&vm.stack)

			if current_type != current_value.value_type {
				switch current_type {
				case INT:
					current_value = INT_VAL(TO_INT_S(&current_value))
				case UINT:
					current_value = UINT_VAL(TO_UINT_S(&current_value))
				case DECIMAL:
					current_value = DECIMAL_VAL(TO_DECIMAL_S(&current_value))
				case BOOL:

				default:
					log.Panic("Unknown type used for variable!")
				}
			}

			env.add_entry(current_name, current_type, current_value, env.currentScope)

		case OP_PUSH:
			write_ValueArray(&vm.stack, READ_CONSTANT())

		case OP_PRINT:
			print_Value(pop_ValueArray(&vm.stack))

		case OP_PRINTLN:
			print_Value(pop_ValueArray(&vm.stack))
			fmt.Println()

		case OP_RETURN:
			if vm.type_to_check == VM_TYPE_SCRIPT {
				fmt.Println("Cannot return in a script!")
				return INTERPRETER_RESULT_INTERPET_ERROR
			} else if vm.type_to_check == VM_TYPE_FUNCTION {
				env.remove_scope(env.currentScope)
				env.currentScope--
				return INTERPRETER_RESULT_OK
			}

		case OP_EOF:
			return INTERPRETER_RESULT_OK

		default:
			log.Printf("Opcode used at %d and of type %d in bytecode doesn't have implementation or isn't correct.\n", vm.index, instruction)
			return INTERPRETER_RESULT_INTERPET_ERROR
		}
	}
}

func free_VM(vm *VM) {
	vm.chunk = nil
	vm.index = 0
	free_ValueArray(&vm.stack)
}
