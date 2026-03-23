package main

import "strings"

// normalizeArgsForFlags reorders args so flags are parsed even if swapped:
// - advocate-runner <cmd> <positional> --flag value ...
// - Go's flag parser stops at the first non-flag argument.
func normalizeArgsForFlags(args []string) []string {
	var flags []string
	var pos []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			flags = append(flags, a)

			// Support --flag=value
			if strings.Contains(a, "=") {
				continue
			}

			// If next token is a value (doesn't start with '-'), attach it
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			pos = append(pos, a)
		}
	}

	return append(flags, pos...)
}
