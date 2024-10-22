package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed in-out/default/*.txt
//go:embed in-out/avec-erreurs/*.txt
//go:embed in-out/sans-erreur/*.txt
var testFiles embed.FS

func main() {
	// Quick banner
	fmt.Print("CRISE - ")
	for i := 1; i <= 50; i++ {
		fmt.Print("😭")
	}
	fmt.Print("\n")

	// Compile the C program first
	cmd := exec.Command("gcc", "-o", "absence", "main.c")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to compile program: %v\n", err)
		return
	}
	defer os.Remove("absence") // Clean up the compiled binary after we're done

	sprint("in-out/default")
	sprint("in-out/avec-erreurs")
	sprint("in-out/sans-erreur")
}

func sprint(name string) {
	// Test each sprint
	for i := 1; i <= 4; i++ {
		inFile := filepath.Join(name, fmt.Sprintf("in-sp%d.txt", i))
		outFile := filepath.Join(name, fmt.Sprintf("out-sp%d.txt", i))

		// Read expected output
		expected, err := testFiles.ReadFile(outFile)
		if err != nil {
			fmt.Printf("Sprint %d: Could not read output file: %v\n", i, err)
			continue
		}
		// Normalize line endings
		expected = bytes.ReplaceAll(expected, []byte("\r\n"), []byte("\n"))

		// Read input
		input, err := testFiles.ReadFile(inFile)
		if err != nil {
			fmt.Printf("Sprint %d: Could not read input file: %v\n", i, err)
			continue
		}

		// Run the program with the input
		cmd := exec.Command("./absence")
		cmd.Stdin = bytes.NewReader(input)
		actual, err := cmd.Output()
		if err != nil {
			fmt.Printf("Sprint %d: Failed to run program: %v\n", i, err)
			continue
		}
		// Normalize line endings
		actual = bytes.ReplaceAll(actual, []byte("\r\n"), []byte("\n"))

		// Compare outputs
		if bytes.Equal(actual, expected) {
			fmt.Printf("Sprint %d: \033[32mPASS\033[0m\n", i)
		} else {
			fmt.Printf("Sprint %d: \033[31mFAIL\033[0m\n", i)
			// Show difference
			actualLines := strings.Split(string(actual), "\n")
			expectedLines := strings.Split(string(expected), "\n")

			fmt.Printf("Differences found:\n")
			for j := 0; j < len(actualLines) || j < len(expectedLines); j++ {
				var actualLine, expectedLine string
				if j < len(actualLines) {
					actualLine = actualLines[j]
				}
				if j < len(expectedLines) {
					expectedLine = expectedLines[j]
				}
				if actualLine != expectedLine {
					fmt.Printf("Line %d:\n", j+1)
					fmt.Printf("  Expected: %q\n", expectedLine)
					fmt.Printf("  Got:      %q\n", actualLine)
				}
			}
		}
	}
}
