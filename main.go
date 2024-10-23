package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	greenIn  = "\033[32m"
	greenOut = "\033[0m"
	redIn    = "\033[31m"
	redOut   = "\033[0m"
)

func RemoveSpacesAndNewlines(s string) string {
	// Use strings.Map to remove non-printable characters
	cleanText := strings.Map(func(r rune) rune {
		if r <= unicode.MaxASCII {
			return r
		}
		return -1
	}, s)
	return strings.Join(strings.Fields(cleanText), "")
}

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

	nSprintPtr := flag.Int("dir", 1, "1=default, 2=sans-erreur, 3=avec-erreurs")
	nSprintTestPtr := flag.Int("sprint", 1, "1-4=a specific one")
	flag.Parse()

	if nSprintPtr == nil || *nSprintPtr < 1 || *nSprintPtr > 3 {
		fmt.Println("Invalid sprint number")
		return
	}
	if nSprintTestPtr == nil || *nSprintTestPtr < 1 || *nSprintTestPtr > 4 {
		fmt.Println("Invalid sprint test number")
		return
	}

	var err error
	switch *nSprintPtr {
	case 1:
		err = sprint("in-out/default", *nSprintTestPtr)
	case 2:
		err = sprint("in-out/sans-erreur", *nSprintTestPtr)
	case 3:
		err = sprint("in-out/avec-erreurs", *nSprintTestPtr)
	}
	if err != nil {
		log.Fatalln("je sais pas comment tu fait mais toi c'est vraiment la crise: ", err)
	}
}

func sprint(name string, n int) error {
	// Test each sprint
	inFile := filepath.Join(name, fmt.Sprintf("in-sp%d.txt", n))
	outFile := filepath.Join(name, fmt.Sprintf("out-sp%d.txt", n))

	// Read expected output
	expected, err := testFiles.ReadFile(outFile)
	if err != nil {
		return fmt.Errorf("Sprint %d: Could not read output file: %v\n", n, err)
	}
	// Normalize line endings
	expected = bytes.ReplaceAll(expected, []byte("\r\n"), []byte("\n"))

	// Read input
	input, err := testFiles.ReadFile(inFile)
	if err != nil {
		return fmt.Errorf("Sprint %d: Could not read input file: %v\n", n, err)
	}

	// Run the program with the input
	cmd := exec.Command("./absence")
	cmd.Stdin = bytes.NewReader(input)
	actual, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Sprint %d: Failed to run program: %v\n", n, err)
	}
	// Normalize line endings
	actual = bytes.ReplaceAll(actual, []byte("\r\n"), []byte("\n"))

	// Compare outputs
	if bytes.Equal(actual, expected) {
		fmt.Printf("Sprint %d: %sPASS%s\n", n, greenIn, greenOut)
	} else {
		if RemoveSpacesAndNewlines(string(actual)) == RemoveSpacesAndNewlines(string(expected)) {
			fmt.Printf("Sprint %d: %sFAIL (SPACES or NEWLINES)%s\n", n, redIn, redOut)
		} else {
			fmt.Printf("Sprint %d: %sFAIL%s\n", n, redIn, redOut)
		}
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
				var in string
				var out string
				if RemoveSpacesAndNewlines(actualLine) == RemoveSpacesAndNewlines(expectedLine) {
					in = greenIn
					out = greenOut
				} else {
					in = redIn
					out = redOut
				}
				fmt.Printf("%sLine %d:%s\n", in, j+1, out)
				fmt.Printf("  %sExpected%s: %q\n", in, out, expectedLine)
				fmt.Printf("  %sGot%s:      %q\n", in, out, actualLine)
			}
		}
	}
	return nil
}
