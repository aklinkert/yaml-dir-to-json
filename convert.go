package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ghodss/yaml"
)

const distDir = "dist"
const srcDir = "src"

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: convert ./source-dir ./target-dir")
	}

	src := os.Args[1]
	target := os.Args[2]
	if _, err := os.Stat(target); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error checking target directory: %v", err)
	} else if err == nil {
		log.Print("Cleaning target directory ...")
		if err = os.RemoveAll(target); err != nil {
			log.Fatalf("Error deleting target directory: %v", err)
		}
	}

	log.Print("Creating target directory ...")
	if err := os.Mkdir(target, 0755); err != nil {
		log.Fatalf("Error creating target directory: %v", err)
	}

	log.Printf("Reading source directory %s ...", src)
	dirFiles, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatalf("Error reading source directory: %v", err)
	}

	var files []string
	for _, f := range dirFiles {
		if f.IsDir() {
			continue
		}

		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		if !strings.HasSuffix(f.Name(), ".yml") &&
			!strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}

		files = append(files, f.Name())
	}

	log.Printf("Converting %d files ... \n", len(files))

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go handleFile(src, target, f, &wg)
	}

	wg.Wait()
	log.Println("Done.")
}

func handleFile(dir, targetDir, file string, wg *sync.WaitGroup) {
	sourceFile := filepath.Join(dir, file)
	defer wg.Done()

	b, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		log.Fatalln(err)
	}

	j, err := yaml.YAMLToJSON(b)
	if err != nil {
		log.Fatalf("Error reading file %q: %v \n", sourceFile, err)
	}

	var out bytes.Buffer
	json.Indent(&out, j, "", "  ")

	targetFile := file
	targetFile = strings.Replace(targetFile, "yml", "json", -1)
	targetFile = strings.Replace(targetFile, "yaml", "json", -1)
	err = ioutil.WriteFile(filepath.Join(targetDir, targetFile), out.Bytes(), 0755)
	if err != nil {
		log.Fatalf("Error writing file %q: %v \n", sourceFile, err)
	}

	log.Printf("Converted file %q to %q \n", sourceFile, filepath.Join(targetDir, targetFile))
}
