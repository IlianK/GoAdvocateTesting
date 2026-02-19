package ui

import "fmt"

type SelectTestRootsParams struct {
	Roots []string
	Title string
}

// SelectTestRoots lets user choose:
// 1) ALL roots
// 2) ONE root
func SelectTestRoots(p SelectTestRootsParams) []string {
	if len(p.Roots) <= 1 {
		return p.Roots
	}
	PrintHeader(p.Title)
	fmt.Println("1. ALL")
	fmt.Println("2. ONE")
	choice := readChoice(1, 2)

	if choice == 1 {
		return p.Roots
	}

	PrintHeader("Select ONE dataset root:")
	for i, r := range p.Roots {
		fmt.Printf("%d. %s\n", i+1, r)
	}
	idx := readChoice(1, len(p.Roots))
	return []string{p.Roots[idx-1]}
}
