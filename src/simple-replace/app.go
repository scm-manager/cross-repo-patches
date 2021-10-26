package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("usage simple-replace path/to/file old new")
		os.Exit(1)
	}

	path := os.Args[1]
	old := os.Args[2]
	new := os.Args[3]

	updateFileIfNecessary(path, old, new)
}

func updateFileIfNecessary(path string, old string, new string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("failed to read file", err)
	}

	content := string(data)
	if strings.Contains(content, old) {
		newContent := strings.ReplaceAll(content, old, new)
		err = os.WriteFile(path, []byte(newContent), os.ModeDevice)
		if err != nil {
			log.Fatal("failed to write file", err)
		}
	} else {
		log.Printf("file %s does not contain searched string %s", path, old)
	}
}
