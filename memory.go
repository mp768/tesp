package main

// Used to just increase capacity by double the amount it once was.
func capacity_new(old_capacity int) int {
	var new_capacity int

	if old_capacity < 8 {
		new_capacity = 8
	} else {
		new_capacity = old_capacity * 2
	}

	return new_capacity
}
