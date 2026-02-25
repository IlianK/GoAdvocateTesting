package ui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"GoAdvocateTesting/internal/app"
)

type Kind string

const (
	KindAnalysis Kind = "analysis"
	KindFuzzing  Kind = "fuzzing"
)

type Scope int

const (
	ScopeAllTests Scope = iota
	ScopeOneTest
)

type ModeType int

const (
	AllOnAll ModeType = iota
	AllOnOne
	OneOnAll
	OneOnOne
)

type RunMenuParams struct {
	Modes []string
	Tests []app.TestCase

	AnalysisProfile map[string]app.AnalysisProfile
	FuzzProfiles    map[string]app.FuzzProfile
}

type Selection struct {
	Kind    Kind
	Profile string

	// analysis selection uses Scope
	Scope Scope
	Test  app.TestCase

	// fuzzing selection uses ModeType/Mode and optionally Test
	ModeType ModeType
	Mode     string
}

func RunMenu(p RunMenuParams) Selection {
	reader := bufio.NewReader(os.Stdin)

	kind := chooseKind(reader)

	if kind == KindAnalysis {
		profile := chooseProfile(reader, "Select analysis profile:", keys(p.AnalysisProfile))
		scope := chooseScope(reader, "Analysis scope:")
		if scope == ScopeOneTest {
			tc := chooseTest(reader, "Select test:", p.Tests)
			return Selection{Kind: kind, Profile: profile, Scope: scope, Test: tc}
		}
		return Selection{Kind: kind, Profile: profile, Scope: scope}
	}

	// fuzzing
	profile := chooseProfile(reader, "Select fuzz profile:", keys(p.FuzzProfiles))
	mt, mode, tc := chooseFuzzPlan(reader, p.Modes, p.Tests)
	return Selection{Kind: kind, Profile: profile, ModeType: mt, Mode: mode, Test: tc}
}

// DisplayTest shows a unique label in recursive discovery
// If File contains a subdir, display "subdir :: TestName"
func DisplayTest(tc app.TestCase) string {
	// tc.File is either "md_test.go" or "subdir/md_test.go" in recursive mode
	if strings.Contains(tc.File, string(os.PathSeparator)) {
		dir := filepath.Dir(tc.File)
		if dir != "." && dir != "" {
			return fmt.Sprintf("%s :: %s", dir, tc.Name)
		}
	}
	return tc.Name
}

func chooseKind(reader *bufio.Reader) Kind {
	fmt.Println("Select run kind:")
	fmt.Println("1. Analysis")
	fmt.Println("2. Fuzzing")
	for {
		fmt.Print("Choice: ")
		in, _ := reader.ReadString('\n')
		switch strings.TrimSpace(in) {
		case "1":
			return KindAnalysis
		case "2":
			return KindFuzzing
		}
	}
}

func chooseScope(reader *bufio.Reader, title string) Scope {
	fmt.Println(title)
	fmt.Println("1. Run on ALL tests")
	fmt.Println("2. Run on ONE test")
	for {
		fmt.Print("Choice: ")
		in, _ := reader.ReadString('\n')
		switch strings.TrimSpace(in) {
		case "1":
			return ScopeAllTests
		case "2":
			return ScopeOneTest
		}
	}
}

func chooseFuzzPlan(reader *bufio.Reader, modes []string, tests []app.TestCase) (ModeType, string, app.TestCase) {
	fmt.Println("1. Run ALL modes on ALL test cases")
	fmt.Println("2. Run ALL modes on ONE test")
	fmt.Println("3. Run ONE mode on ALL tests")
	fmt.Println("4. Run ONE mode on ONE test")
	for {
		fmt.Print("Choice: ")
		c, _ := reader.ReadString('\n')
		c = strings.TrimSpace(c)

		switch c {
		case "1":
			return AllOnAll, "", app.TestCase{}
		case "2":
			return AllOnOne, "", chooseTest(reader, "Select test:", tests)
		case "3":
			return OneOnAll, chooseString(reader, "Select mode:", modes), app.TestCase{}
		case "4":
			return OneOnOne, chooseString(reader, "Select mode:", modes), chooseTest(reader, "Select test:", tests)
		}
	}
}

func chooseProfile(reader *bufio.Reader, title string, options []string) string {
	if len(options) == 0 {
		panic("no profiles available")
	}
	return chooseString(reader, title, options)
}

func chooseString(reader *bufio.Reader, title string, options []string) string {
	fmt.Println(title)
	for i, o := range options {
		fmt.Printf("%d. %s\n", i+1, o)
	}
	for {
		fmt.Print("#? ")
		input, _ := reader.ReadString('\n')
		idx, err := strconv.Atoi(strings.TrimSpace(input))
		if err == nil && idx > 0 && idx <= len(options) {
			return options[idx-1]
		}
	}
}

func chooseTest(reader *bufio.Reader, title string, tests []app.TestCase) app.TestCase {
	fmt.Println(title)
	for i, t := range tests {
		fmt.Printf("%d. %s\n", i+1, DisplayTest(t))
	}
	for {
		fmt.Print("#? ")
		input, _ := reader.ReadString('\n')
		idx, err := strconv.Atoi(strings.TrimSpace(input))
		if err == nil && idx > 0 && idx <= len(tests) {
			return tests[idx-1]
		}
	}
}

func keys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
