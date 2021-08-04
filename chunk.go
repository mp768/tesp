package main

const (
	OP_EOF byte = iota
	OP_RETURN

	OP_PUSH
	OP_POP
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV

	OP_CALL_FUNC

	OP_CMP_LESS
	OP_CMP_GREATER
	OP_CMP_EQUAL
	OP_CMP_NOT_EQUAL
	OP_CMP_LESS_EQUAL
	OP_CMP_GREATER_EQUAL

	OP_CMP_AND
	OP_CMP_OR

	OP_JMP
	OP_IF_FALSE_JMP

	OP_NEGATE

	OP_STORE
	OP_END_QUOTE

	OP_LOAD
	OP_START_SCOPE
	OP_END_SCOPE
)

type Chunk struct {
	code      []byte
	lines     []uint32
	constants ValueArray
}

func (chunk *Chunk) init_chunk() {
	chunk.code = make([]byte, 0, 0)
	init_ValueArray(&chunk.constants)
}

func (chunk *Chunk) write_load(byte_ byte, name string, line uint32) {
	chunk.write_chunk(byte_, line)
	for i := 0; i < len(name); i++ {
		chunk.write_chunk(name[i], line)
	}
	chunk.write_chunk(OP_END_QUOTE, line)
}

func (chunk *Chunk) write_call_func(byte_ byte, name string, line uint32) {
	chunk.write_load(byte_, name, line) // LOL
}

func (chunk *Chunk) write_chunk(byte_ byte, line uint32) {
	chunk.code = append(chunk.code, byte_)
	chunk.lines = append(chunk.lines, line)
}

func (chunk *Chunk) write_store(byte_ byte, name string, t ValueTypes, line uint32) {
	chunk.write_chunk(byte_, line)
	name_bytes := []byte(name)
	for i := 0; i < len(name_bytes); i++ {
		chunk.write_chunk(name_bytes[i], line)
	}
	chunk.write_chunk(OP_END_QUOTE, line)
	chunk.write_chunk(byte(t), line)
}

func (chunk *Chunk) write_jmp(byte_ byte, jmp_value uint32, line uint32) {
	chunk.write_chunk(byte_, line)
	chunk.write_chunk(byte(jmp_value>>24), line)
	chunk.write_chunk(byte(jmp_value>>16), line)
	chunk.write_chunk(byte(jmp_value>>8), line)
	chunk.write_chunk(byte(jmp_value), line)
}

func (chunk *Chunk) write_constant(byte_ byte, constant Value, line uint32) {
	chunk.write_chunk(byte_, line)
	write_ValueArray(&chunk.constants, constant)
	index := len(chunk.constants.values) - 1
	chunk.write_chunk(byte(index>>8), line)
	chunk.write_chunk(byte(index&0xff), line)
}

func (chunk *Chunk) free_chunk() {
	chunk.init_chunk()
	free_ValueArray(&chunk.constants)
}
