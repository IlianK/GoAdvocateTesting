package ui

import "fmt"

func PrintHeader(s string) {
	fmt.Println()
	fmt.Println("==>", s)
}

func PrintOK(s string) {
	fmt.Println("[OK]", s)
}

func PrintError(s string) {
	fmt.Println("[ERR]", s)
}
