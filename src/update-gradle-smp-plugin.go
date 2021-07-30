package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// id 'org.scm-manager.smp' version '0.8.3'
var regex = regexp.MustCompile("^\\s*id\\s+'org\\.scm-manager\\.smp'\\s+version\\s+'([0-9]\\.[0-9]\\.[0-9])'\\s*$")

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage update-gradle-smp-plugin path/to/source/code newVersion")
		os.Exit(1)
	}

	path := os.Args[1]
	newVersion := os.Args[2]

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if info.Name() == "build.gradle" {
			updateFileIfNecessary(p, newVersion)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func updateFileIfNecessary(path string, newVersion string) {
	needsUpdate, lines := checkFile(path, newVersion)
	if needsUpdate {
		writeUpdatedFile(path, lines)
	} else {
		fmt.Printf("file %s does not contain plugin or is up to date\n", path)
	}
}

func checkFile(path string, newVersion string) (bool, []string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var lines []string
	updated := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		groups := regex.FindStringSubmatch(line)
		if len(groups) > 0 {
			oldVersion := groups[1]
			if oldVersion != newVersion {
				newLine := strings.Replace(groups[0], oldVersion, newVersion, 1)
				lines = append(lines, newLine)
				updated = true
			} else {
				lines = append(lines, line)
			}

		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return updated, lines
}

func writeUpdatedFile(path string, lines []string) {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range lines {
		fmt.Fprintln(f, v)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("updated file", path)
}
