package main

import (
	"fmt"
	"log"
	"strconv"
)

type ValueTypes byte

const (
	UINT ValueTypes = iota
	INT
	DECIMAL
	BOOL
	STRING
	NO_VALUE
)

func ValueTypes_to_string(type_ ValueTypes) string {
	switch type_ {
	case INT:
		return "int"
	case UINT:
		return "uint"
	case DECIMAL:
		return "decimal"
	case BOOL:
		return "boolean"
	case STRING:
		return "string"
	case NO_VALUE:
		return "no value"

	default:
		return "Unknown"
	}
}

type Value struct {
	value_type ValueTypes
	as         struct {
		STR string
		U64 uint64
		I64 int64
		F64 float64
		B1  bool
	}
}

func print_Value(value Value) {
	switch value.value_type {
	case DECIMAL:
		fmt.Printf("%g", TO_DECIMAL_S(&value))
	case INT:
		fmt.Print(TO_INT_S(&value))
	case STRING:
		fmt.Print(TO_STRING_S(&value))
	case UINT:
		fmt.Print(TO_UINT_S(&value))
	case BOOL:
		fmt.Print(TO_BOOL_S(&value))
	}
}

type ValueArray struct {
	values []Value
}

func init_ValueArray(array *ValueArray) {
	array.values = make([]Value, 0, 0)
}

func write_ValueArray(array *ValueArray, value Value) {
	array.values = append(array.values, value)
}

func pop_ValueArray(array *ValueArray) (result Value) {
	length := len(array.values)
	result = array.values[length-1]
	array.values = array.values[0 : length-1]
	return
}

func free_ValueArray(array *ValueArray) {
	init_ValueArray(array)
}

///////////////////////////////////////////////////////
//               VALUE FUNCTIONS HERE                //
///////////////////////////////////////////////////////

func IS_OF_TYPE(value *Value, type_to_check ValueTypes) bool {
	return value.value_type == type_to_check
}

func TO_DECIMAL_S(value *Value) float64 {
	if IS_OF_TYPE(value, UINT) {
		return float64(value.as.U64)
	} else if IS_OF_TYPE(value, INT) {
		return float64(value.as.I64)
	} else if IS_OF_TYPE(value, DECIMAL) {
		return float64(value.as.F64)
	} else {
		log.Panic("Cannot convert value to a decimal!")
	}

	return 0
}

func TO_INT_S(value *Value) int64 {
	if IS_OF_TYPE(value, UINT) {
		return int64(value.as.U64)
	} else if IS_OF_TYPE(value, INT) {
		return int64(value.as.I64)
	} else if IS_OF_TYPE(value, DECIMAL) {
		return int64(value.as.F64)
	} else {
		log.Panic("Cannot convert value to a int!")
	}

	return 0
}

func TO_UINT_S(value *Value) uint64 {
	if IS_OF_TYPE(value, UINT) {
		return uint64(value.as.U64)
	} else if IS_OF_TYPE(value, INT) {
		return uint64(value.as.I64)
	} else if IS_OF_TYPE(value, DECIMAL) {
		return uint64(value.as.F64)
	} else {
		log.Panic("Cannot convert value to a uint!")
	}

	return 0
}

func TO_BOOL_S(value *Value) bool {
	switch value.value_type {
	case UINT, INT, DECIMAL:
		return 0 < TO_INT_S(value)

	case BOOL:
		return value.as.B1

	default:
		log.Panic("Cannot convert value to a bool!")
	}

	return false
}

func TO_STRING_S(value *Value) string {
	switch value.value_type {
	case UINT:
		return strconv.FormatUint(value.as.U64, 10)

	case INT:
		return strconv.FormatInt(value.as.I64, 10)

	case DECIMAL:
		return strconv.FormatFloat(value.as.F64, 'E', -1, 64)

	case STRING:
		return value.as.STR

	case BOOL:
		return strconv.FormatBool(value.as.B1)

	default:
		log.Panic("Cannot convert value to a string!")
	}

	return ""
}

func STRING_VAL(value string) Value {
	return Value{
		STRING,
		struct {
			STR string
			U64 uint64
			I64 int64
			F64 float64
			B1  bool
		}{
			value,
			0,
			0,
			0,
			false,
		},
	}
}

func NO_VAL() Value {
	return Value{
		NO_VALUE,
		struct {
			STR string
			U64 uint64
			I64 int64
			F64 float64
			B1  bool
		}{
			"",
			0,
			0,
			0,
			false,
		},
	}
}

func UINT_VAL(value uint64) Value {
	return Value{
		UINT,
		struct {
			STR string
			U64 uint64
			I64 int64
			F64 float64
			B1  bool
		}{
			"",
			value,
			0,
			0,
			false,
		},
	}
}

func INT_VAL(value int64) Value {
	return Value{
		INT,
		struct {
			STR string
			U64 uint64
			I64 int64
			F64 float64
			B1  bool
		}{
			"",
			0,
			value,
			0,
			false,
		},
	}
}

func BOOL_VAL(value bool) Value {
	return Value{
		BOOL,
		struct {
			STR string
			U64 uint64
			I64 int64
			F64 float64
			B1  bool
		}{
			"",
			0,
			0,
			0,
			value,
		},
	}
}

func DECIMAL_VAL(value float64) Value {
	return Value{
		DECIMAL,
		struct {
			STR string
			U64 uint64
			I64 int64
			F64 float64
			B1  bool
		}{
			"",
			0,
			0,
			value,
			false,
		},
	}
}
