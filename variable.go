package main

import (
	"fmt"
	"log"
)

type Function_Table struct {
	functions []Function_Entry
}

type Function_Type byte

const (
	FUNCTION_NATIVE Function_Type = iota
	FUNCTION_VIRTUAL
)

type Function_Entry struct {
	f_type      Function_Type
	native_body func(bool, []Value) (Value, ValueTypes)
	body        Chunk
	name        string
	arity       uint
	return_type ValueTypes
}

var ftable = Function_Table{}

func (table *Function_Table) check_if_already_exists(name string) bool {
	for _, v := range table.functions {
		if v.name == name {
			return true
		}
	}

	return false
}

func (table *Function_Table) add_virtual_entry(name string, body Chunk, arity uint, return_type ValueTypes) {
	if table.check_if_already_exists(name) {
		log.Panicf("Function '%s' already exists", name)
	}

	table.functions = append(table.functions, Function_Entry{
		FUNCTION_VIRTUAL,
		func(b bool, v []Value) (Value, ValueTypes) { return Value{}, NO_VALUE },
		body,
		name,
		arity,
		return_type,
	})
}

func (table *Function_Table) add_native_entry(name string, body func(bool, []Value) (Value, ValueTypes), arity uint, return_type ValueTypes) {
	if table.check_if_already_exists(name) {
		log.Panicf("Function '%s' already exists", name)
	}

	table.functions = append(table.functions, Function_Entry{
		FUNCTION_NATIVE,
		body,
		Chunk{},
		name,
		arity,
		return_type,
	})
}

func (table *Function_Table) get_entry(name string) Function_Entry {
	for _, v := range table.functions {
		if v.name == name {
			return v
		}
	}

	log.Panicf("Cannot find an entry by the name of '%s'\n", name)
	return Function_Entry{}
}

type Environment struct {
	Entries      []Entry
	currentScope uint8
}

type Entry struct {
	name  string
	vtype ValueTypes
	value Value
	scope uint8
}

func (env *Environment) add_entry(name string, vtype ValueTypes, value Value, scope_index uint8) {
	env.Entries = append(env.Entries, Entry{
		name,
		vtype,
		value,
		scope_index,
	})
}

func (env *Environment) assign_to_entry(name string, value Value) {
	for i := int16(env.currentScope); i >= 0; i-- {
		for k, v := range env.Entries {
			if v.scope == uint8(i) && v.name == name {
				if v.value.value_type == value.value_type {
					env.Entries[k].value = value
				} else {
					log.Panicf("Cannot assign value to '%s' as it's the wrong type.", name)
				}
				return
			}
		}
	}

	log.Panicf("Couldn't get a variable by the name of '%s'!", name)
}

func (env *Environment) get_variable_value(name string) Value {
	for i := int16(env.currentScope); i >= 0; i-- {
		for _, v := range env.Entries {
			if v.scope == uint8(i) && v.name == name {
				return v.value
			}
		}
	}

	log.Panicf("Couldn't get a variable by the name of '%s'!", name)
	return NO_VAL()
}

func (env *Environment) remove_scope(scope_to_remove uint8) {
	length := len(env.Entries)
	deletion_index := len(env.Entries)
	for i := 0; i < length; i++ {
		if env.Entries[i].scope == scope_to_remove {
			deletion_index = i
			break
		}
	}
	env.Entries = env.Entries[0:deletion_index]
}

func (env *Environment) print_entries() {
	fmt.Println("===      Entries      ===")
	fmt.Println(" Name  Scope  Type  Value")

	for i := 0; i < len(env.Entries); i++ {
		fmt.Print("[ '")
		fmt.Print(env.Entries[i].name)
		fmt.Print("'")

		if i > 0 && env.Entries[i].scope == env.Entries[i-1].scope {
			fmt.Print("   | ")
		} else {
			fmt.Printf("%4d ", env.Entries[i].scope)
		}

		fmt.Print("  ")
		fmt.Print(ValueTypes_to_string(env.Entries[i].vtype))
		fmt.Print("  ")
		print_Value(env.Entries[i].value)
		fmt.Println("]")
	}
}

func new_Environment() (result Environment) {
	result = Environment{}
	result.Entries = make([]Entry, 0, 0)
	result.currentScope = 0
	return
}
