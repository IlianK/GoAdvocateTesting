package ui

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func readChoice(min, max int) int {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Choice: ")
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(line)
		if err != nil || n < min || n > max {
			fmt.Printf("Enter a number between %d and %d\n", min, max)
			continue
		}
		return n
	}
}

func chooseOne(prompt string, options []string) int {
	PrintHeader(prompt)
	for i, opt := range options {
		fmt.Printf("%d. %s\n", i+1, opt)
	}
	return readChoice(1, len(options))
}

func chooseOneFromList(prompt string, options []string) string {
	options = uniqueSorted(options)
	if len(options) == 0 {
		return ""
	}
	PrintHeader(prompt)
	for i, opt := range options {
		fmt.Printf("%d. %s\n", i+1, opt)
	}
	idx := readChoice(1, len(options))
	return options[idx-1]
}

func uniqueSorted(xs []string) []string {
	m := map[string]struct{}{}
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		m[x] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for x := range m {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}
