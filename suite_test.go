package main

import (
	"fmt"
	"testing"
)

func BenchmarkChunkCreation(b *testing.B) {
	var chunko Chunk
	chunko.init_chunk()
	for i := 0; i < b.N; i++ {
		chunko.write_chunk(byte(i*2), 1)
	}

	for i := 0; i < b.N; i++ {
		fmt.Println(i, ": ", chunko.code[i])
	}
	chunko.free_chunk()
}
