package main

import (
	"fmt"
	"os"
)

func actually_fibonacci(n int) int {
	if n <= 1 {
		return n
	} else {
		return actually_fibonacci(n-1) + actually_fibonacci(n-2)
	}
}

func fibonacci(values []Value) (Value, ValueTypes) {
	// this function should take 1 arg
	var value = TO_INT_S(&values[0])

	return_value := actually_fibonacci(int(value))

	return INT_VAL(int64(return_value)), INT
}

func main() {
	args := os.Args // Add [1:] when doing the thing, or just don't.
	defer fmt.Println(args[0])

	// GOAL: Convert c project to go for more flexibility for the language.

	var chunko Chunk
	chunko.init_chunk()
	defer chunko.free_chunk()

	chunko.write_constant(OP_PUSH, DECIMAL_VAL(78.9), 0)
	chunko.write_constant(OP_PUSH, INT_VAL(65), 0)
	chunko.write_chunk(OP_ADD, 1)
	chunko.write_constant(OP_PUSH, UINT_VAL(5), 2)
	chunko.write_chunk(OP_DIV, 2)
	chunko.write_chunk(OP_NEGATE, 2)
	chunko.write_store(OP_STORE, "Hello", DECIMAL, 3)
	chunko.write_load(OP_LOAD, "Hello", 3)

	chunko.write_chunk(OP_START_SCOPE, 3)
	chunko.write_constant(OP_PUSH, DECIMAL_VAL(79.9), 3)
	chunko.write_store(OP_STORE, "H", DECIMAL, 3)
	chunko.write_load(OP_LOAD, "Hello", 3)
	chunko.write_load(OP_LOAD, "H", 3)
	chunko.write_chunk(OP_END_SCOPE, 4)

	chunko.write_constant(OP_PUSH, DECIMAL_VAL(4.6), 5)
	chunko.write_constant(OP_PUSH, UINT_VAL(65), 5)
	chunko.write_chunk(OP_CMP_LESS, 5)

	chunko.write_constant(OP_PUSH, DECIMAL_VAL(4.6), 5)
	chunko.write_constant(OP_PUSH, UINT_VAL(65), 5)
	chunko.write_chunk(OP_CMP_GREATER, 5)

	chunko.write_chunk(OP_CMP_AND, 5)
	chunko.write_jmp(OP_IF_FALSE_JMP, 76, 5)
	chunko.write_store(OP_STORE, "BOOLEAN", BOOL, 5)

	chunko.write_chunk(OP_RETURN, 5)
	//disassemble_chunk(&chunko, "TEST CHUNK")

	vm := new_VM(&chunko)
	defer free_VM(&vm)
	//interpret(&vm)

	gen := new_CodeGen(true)

	ftable.add_native_entry("fibonacci", fibonacci, 1, INT)
	chunk := gen.generate_chunk("./test.txt")
	body := ftable.get_entry("fib").body
	disassemble_chunk(&body, "ADD FUNC")
	disassemble_chunk(&chunk, "GENERATED CHUNK")

	nvm := new_VM(&chunk)
	defer free_VM(&nvm)

	interpret(&nvm)
}
