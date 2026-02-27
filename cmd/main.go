package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// Check command-line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <path-to-csv>")
	}
	csvPath := os.Args[1]

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Read CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Find the column indices
	header := records[0]
	testIdx := -1
	bugIdx := -1
	for i, h := range header {
		if h == "Test" {
			testIdx = i
		} else if h == "Bug_Types" {
			bugIdx = i
		}
	}
	if testIdx == -1 || bugIdx == -1 {
		log.Fatal("Required columns not found")
	}

	// Collect tests with P06
	var testsWithP06 []string
	for _, row := range records[1:] { // skip header
		if bugIdx >= len(row) {
			continue
		}
		if strings.Contains(row[bugIdx], "P06") {
			testsWithP06 = append(testsWithP06, row[testIdx])
		}
	}

	// Print results
	fmt.Println("Tests with P06 bug type:")
	for _, test := range testsWithP06 {
		fmt.Println(test)
	}
	fmt.Printf("\nTotal count: %d\n", len(testsWithP06))
}
