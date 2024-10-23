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

type Status int

const (
	Fail Status = iota
	PassWithoutSpaces
	Pass
	Other // Other status, for errors, not handled
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
	spaces := flag.Bool("spaces", false, "show spaces issues")
	quick := flag.Bool("quick", false, "don't show differences, just pass/fail")
	flag.Parse()

	if nSprintPtr == nil || *nSprintPtr < 1 || *nSprintPtr > 3 {
		fmt.Println("Invalid sprint number")
		return
	}
	if nSprintTestPtr == nil || *nSprintTestPtr < 1 || *nSprintTestPtr > 4 {
		fmt.Println("Invalid sprint test number")
		return
	}

	if *quick {
		var statuses []Status
		for _, dir := range []string{"in-out/default", "in-out/sans-erreur", "in-out/avec-erreurs"} {
			status, err := sprint(dir, *nSprintTestPtr, *spaces, *quick)
			if err != nil {
				log.Fatalln("error: ", err)
			}
			statuses = append(statuses, status)
		}
		for i, status := range statuses {
			switch status {
			case Pass:
				fmt.Printf("Sprint %d: PASS\n", i+1)
			case PassWithoutSpaces:
				fmt.Printf("Sprint %d: PASS (without spaces)\n", i+1)
			case Fail:
				fmt.Printf("Sprint %d: FAIL\n", i+1)
			}
		}
		return
	}

	switch *nSprintPtr {
	case 1:
		_, err := sprint("in-out/default", *nSprintTestPtr, *spaces, *quick)
		if err != nil {
			log.Fatalln("error: ", err)
		}
	case 2:
		_, err := sprint("in-out/sans-erreur", *nSprintTestPtr, *spaces, *quick)
		if err != nil {
			log.Fatalln("error: ", err)
		}
	case 3:
		_, err := sprint("in-out/avec-erreurs", *nSprintTestPtr, *spaces, *quick)
		if err != nil {
			log.Fatalln("error: ", err)
		}
	}
}

func sprint(name string, n int, showSpace bool, quick bool) (Status, error) {
	// Test each sprint
	inFile := filepath.Join(name, fmt.Sprintf("in-sp%d.txt", n))
	outFile := filepath.Join(name, fmt.Sprintf("out-sp%d.txt", n))

	// Read expected output
	expected, err := testFiles.ReadFile(outFile)
	if err != nil {
		return Other, fmt.Errorf("Sprint %d: Could not read output file: %v\n", n, err)
	}
	// Normalize line endings
	expected = bytes.ReplaceAll(expected, []byte("\r\n"), []byte("\n"))

	// Read input
	input, err := testFiles.ReadFile(inFile)
	if err != nil {
		return Other, fmt.Errorf("Sprint %d: Could not read input file: %v\n", n, err)
	}

	// Run the program with the input
	cmd := exec.Command("./absence")
	cmd.Stdin = bytes.NewReader(input)
	actual, err := cmd.Output()
	if err != nil {
		return Other, fmt.Errorf("Sprint %d: Failed to run program: %v\n", n, err)
	}
	// Normalize line endings
	actual = bytes.ReplaceAll(actual, []byte("\r\n"), []byte("\n"))

	// Compare outputs
	if bytes.Equal(actual, expected) {
		fmt.Printf("Sprint %d: %sPASS%s\n", n, greenIn, greenOut)
		return Pass, nil
	} else {
		// Show difference
		actualLines := strings.Split(string(actual), "\n")
		expectedLines := strings.Split(string(expected), "\n")

		if !quick {
			fmt.Printf("Differences found:\n")
		}
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
					if !showSpace {
						// we don't need to show spaces here.
						continue
					}
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

		spaces := false
		if RemoveSpacesAndNewlines(string(actual)) == RemoveSpacesAndNewlines(string(expected)) {
			spaces = true
		}
		if spaces {
			if !quick {
				fmt.Printf("Sprint %d: %sPASS without (SPACES or NEWLINES)%s\n", n, greenIn, greenOut)
			}
			return PassWithoutSpaces, nil
		}
		if !quick {
			fmt.Printf("Sprint %d: %sFAIL%s\n", n, redIn, redOut)
		}
		return Fail, nil
	}
}
