package main

import (
	"fmt"
	"runtime"
)

func actually_fibonacci(n int) int {
	if n <= 1 {
		return n
	} else {
		return actually_fibonacci(n-1) + actually_fibonacci(n-2)
	}
}

func fibonacci(eval bool, values []Value) (Value, ValueTypes) {
	// this function should take 1 arg
	var value = TO_INT_S(&values[0])

	return_value := actually_fibonacci(int(value))

	return INT_VAL(int64(return_value)), INT
}

func clock(eval bool, values []Value) (Value, ValueTypes) {
	// this function takes 0 args

	//dt := time.Now()

	return NO_VAL(), NO_VALUE
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func main() {
	//args := os.Args // Add [1:] when doing the thing, or just don't.
	//defer fmt.Println(args[0])

	gen := new_CodeGen(true)

	ftable.add_native_entry("fibonacci", fibonacci, 1, INT)
	ftable.add_native_entry("clock", clock, 0, NO_VALUE)
	chunk := gen.generate_chunk("./test.txt")
	//body := ftable.get_entry("add").body
	//disassemble_chunk(&body, "ADD FUNC")
	//disassemble_chunk(&chunk, "GENERATED CHUNK")

	vm := new_VM(&chunk)
	defer free_VM(&vm)

	interpret(&vm)
}
