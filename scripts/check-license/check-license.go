// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	licenseCopyright = "The prometheus-operator Authors"
	licenseHeader    = `// Copyright The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.`
)

var (
	copyrightWithDate    = regexp.MustCompile(`Copyright \d{4} The prometheus-operator Authors`)
	generatedFilePattern = regexp.MustCompile(`(Code generated|DO NOT EDIT|@generated)`)

	excludedDirs = map[string]struct{}{
		"vendor":  {},
		".git":    {},
		".github": {},
		"tmp":     {},
	}

	checkFlag bool
	fixFlag   bool
	rootDir   string
)

func main() {
	flag.BoolVar(&checkFlag, "check", false, "Check if all files contain the license header")
	flag.BoolVar(&fixFlag, "fix", false, "Fix files to add the license header where missing")
	flag.StringVar(&rootDir, "root", ".", "Root directory to scan")
	flag.Parse()

	if !checkFlag && !fixFlag {
		fmt.Fprintln(os.Stderr, "Usage: check-license -check | -fix [-root <dir>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if info, err := os.Stat(rootDir); err != nil || !info.IsDir() {
		fatalf("Root directory %q does not exist or is not a directory.", rootDir)
	}

	if checkFlag {
		check()
	}
	if fixFlag {
		fix()
	}
}

type fileResult int

const (
	fileOK fileResult = iota
	fileMissingLicense
	fileDatedLicense
)

// classifyFile reads the first 3 lines of a .go file and determines its license status.
func classifyFile(path string) (fileResult, error) {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return fileOK, err
	}
	defer f.Close() //nolint:errcheck

	scanner := bufio.NewScanner(f)
	for i := 0; i < 3 && scanner.Scan(); i++ {
		line := scanner.Text()
		if generatedFilePattern.MatchString(line) {
			return fileOK, nil
		}
		if copyrightWithDate.MatchString(line) {
			return fileDatedLicense, nil
		}
		if strings.Contains(line, licenseCopyright) {
			return fileOK, nil
		}
	}
	return fileMissingLicense, nil
}

// collectFiles walks the file tree and returns files missing a license header
// and files that have a license header with a year.
func collectFiles(rootPath string) (missingLicense, datedLicense []string) {
	err := filepath.WalkDir(rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if _, excluded := excludedDirs[info.Name()]; excluded {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		result, classifyErr := classifyFile(path)
		if classifyErr != nil {
			fatalf("Failed to read file %q: %s", path, classifyErr)
		}

		switch result {
		case fileMissingLicense:
			missingLicense = append(missingLicense, path)
		case fileDatedLicense:
			datedLicense = append(datedLicense, path)
		}
		return nil
	})
	if err != nil {
		fatalf("Failed to walk directory: %s", err)
	}
	return missingLicense, datedLicense
}

func check() {
	missingLicense, datedLicense := collectFiles(rootDir)
	if len(missingLicense) == 0 && len(datedLicense) == 0 {
		fmt.Println(">> All files have a valid license header.")
		return
	}
	if len(missingLicense) > 0 {
		fmt.Printf(">> Found %d file(s) without a license header:\n", len(missingLicense))
		for _, file := range missingLicense {
			fmt.Printf("  %s\n", file)
		}
	}
	if len(datedLicense) > 0 {
		fmt.Printf(">> Found %d file(s) with a year in the copyright line (should be \"Copyright The prometheus-operator Authors\"):\n", len(datedLicense))
		for _, file := range datedLicense {
			fmt.Printf("  %s\n", file)
		}
	}
	fmt.Println()
	fmt.Println("Run 'make fix-license' to fix these issues automatically.")
	os.Exit(1)
}

func fix() {
	missingLicense, datedLicense := collectFiles(rootDir)
	if len(missingLicense) == 0 && len(datedLicense) == 0 {
		fmt.Println(">> Nothing to fix. All files have a valid license header.")
		return
	}
	for _, file := range missingLicense {
		fmt.Printf("  Adding license header to %s\n", file)
	}
	fixFiles(missingLicense, licenseHeader+"\n\n", 0)
	for _, file := range datedLicense {
		fmt.Printf("  Removing year from copyright in %s\n", file)
	}
	fixFiles(datedLicense, "// Copyright The prometheus-operator Authors\n", 1)
	fmt.Printf(">> Fixed %d file(s).\n", len(missingLicense)+len(datedLicense))
}

func fixFiles(files []string, header string, numberOfLinesToSkip int) {
	for _, file := range files {
		if err := fixFile(file, header, numberOfLinesToSkip); err != nil {
			fatalf("Failed to fix file %q: %s", file, err)
		}
	}
}

func fixFile(path, header string, numberOfLinesToSkip int) error {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "license_fix_*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()    //nolint:errcheck
		os.Remove(tmpPath) //nolint:errcheck
	}()

	if _, err := tmpFile.WriteString(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	scanner := bufio.NewScanner(f)
	for i := 0; i < numberOfLinesToSkip && scanner.Scan(); i++ {
	}
	for scanner.Scan() {
		if _, err := tmpFile.WriteString(scanner.Text() + "\n"); err != nil {
			return fmt.Errorf("writing content: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading original file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replacing file: %w", err)
	}
	return nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
