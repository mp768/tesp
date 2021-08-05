package main

import (
	"encoding/binary"
	"fmt"
)

func simple_instruction(name string, offset uint64) uint64 {
	fmt.Println(name)
	return offset + 1
}

func constant_instruction(name string, chunk *Chunk, offset uint64) uint64 {
	fmt.Print(name, "   '")

	var index uint16 = binary.BigEndian.Uint16([]byte{chunk.code[offset+1], chunk.code[offset+2]})

	value := chunk.constants.values[index]

	print_Value(value)
	fmt.Print("'  ")
	fmt.Println(ValueTypes_to_string(*&value.value_type))
	return offset + 3
}

func jmp_instruction(name string, chunk *Chunk, offset uint64) uint64 {
	fmt.Print(name, "   '")

	index := binary.BigEndian.Uint32([]byte{chunk.code[offset+1], chunk.code[offset+2], chunk.code[offset+3], chunk.code[offset+4]})

	fmt.Print(index)
	fmt.Println("'")
	return offset + 5
}

func store_instruction(name string, load_op bool, chunk *Chunk, offset uint64) uint64 {
	fmt.Print(name, "   '")

	offset++
	original_offset := offset
	var count uint32

	for chunk.code[offset] != OP_END_QUOTE {
		offset++
		count++
	}

	offset = original_offset

	name_bytes := make([]byte, count, count)
	for chunk.code[offset] != OP_END_QUOTE {
		name_bytes[offset-original_offset] = chunk.code[offset]
		offset++
	}
	fmt.Print(string(name_bytes))
	offset++
	fmt.Print("'  ")
	if !load_op {
		fmt.Println(ValueTypes_to_string(ValueTypes(chunk.code[offset])))
		return offset + 1
	} else {
		fmt.Println()
		return offset
	}
}

func disassemble_chunk(chunk *Chunk, name string) {
	fmt.Println("== ", name, " ==")

	for offset := uint64(0); offset < uint64(len(chunk.code)); {
		offset = disassemble_instruction(chunk, offset)
	}
}

func disassemble_instruction(chunk *Chunk, offset uint64) uint64 {
	fmt.Printf("%04d ", offset)

	if offset > 0 && chunk.lines[offset] == chunk.lines[offset-1] {
		fmt.Print("   | ")
	} else {
		fmt.Printf("%4d ", chunk.lines[offset])
	}

	instruction := chunk.code[offset]
	switch instruction {
	case OP_START_SCOPE:
		return simple_instruction("OP_START_SCOPE", offset)

	case OP_END_SCOPE:
		return simple_instruction("OP_END_SCOPE", offset)

	case OP_LOAD:
		return store_instruction("OP_LOAD", true, chunk, offset)

	case OP_ASSIGN:
		return store_instruction("OP_ASSIGN", true, chunk, offset)

	case OP_CALL_FUNC:
		return store_instruction("OP_CALL_FUNC", true, chunk, offset)

	case OP_PRINT:
		return simple_instruction("OP_PRINT", offset)

	case OP_PRINTLN:
		return simple_instruction("OP_PRINTLN", offset)

	case OP_NEGATE:
		return simple_instruction("OP_NEGATE", offset)

	case OP_POP:
		return simple_instruction("OP_POP", offset)

	case OP_JMP:
		return jmp_instruction("OP_JMP", chunk, offset)

	case OP_IF_FALSE_JMP:
		return jmp_instruction("OP_IF_FALSE_JMP", chunk, offset)

	case OP_CMP_GREATER:
		return simple_instruction("OP_CMP_GREATER", offset)

	case OP_CMP_LESS:
		return simple_instruction("OP_CMP_LESS", offset)

	case OP_CMP_EQUAL:
		return simple_instruction("OP_CMP_EQUAL", offset)

	case OP_CMP_NOT_EQUAL:
		return simple_instruction("OP_CMP_NOT_EQUAL", offset)

	case OP_CMP_GREATER_EQUAL:
		return simple_instruction("OP_CMP_GREATER_EQUAL", offset)

	case OP_CMP_LESS_EQUAL:
		return simple_instruction("OP_CMP_LESS_EQUAL", offset)

	case OP_CMP_AND:
		return simple_instruction("OP_CMP_AND", offset)

	case OP_CMP_OR:
		return simple_instruction("OP_CMP_OR", offset)

	case OP_ADD:
		return simple_instruction("OP_ADD", offset)

	case OP_DIV:
		return simple_instruction("OP_DIV", offset)

	case OP_SUB:
		return simple_instruction("OP_SUB", offset)

	case OP_MUL:
		return simple_instruction("OP_MUL", offset)

	case OP_STORE:
		return store_instruction("OP_STORE", false, chunk, offset)

	case OP_PUSH:
		return constant_instruction("OP_PUSH", chunk, offset)

	case OP_RETURN:
		return simple_instruction("OP_RETURN", offset)

	case OP_EOF:
		return simple_instruction("OP_EOF", offset)

	default:
		fmt.Println("Unknown opcode ", instruction)
		return offset + 1
	}
}
