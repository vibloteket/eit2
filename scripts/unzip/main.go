// Command unzip extracts a trusted build archive for package verification.
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: unzip ARCHIVE.zip DESTINATION")
		os.Exit(2)
	}
	archive, err := zip.OpenReader(os.Args[1])
	check(err)
	defer archive.Close()
	destination, err := filepath.Abs(os.Args[2])
	check(err)
	for _, file := range archive.File {
		path := filepath.Join(destination, filepath.FromSlash(file.Name))
		if path != destination && !strings.HasPrefix(path, destination+string(os.PathSeparator)) {
			panic("unsafe zip path: " + file.Name)
		}
		if file.FileInfo().IsDir() {
			check(os.MkdirAll(path, file.Mode()))
			continue
		}
		check(os.MkdirAll(filepath.Dir(path), 0o755))
		input, err := file.Open()
		check(err)
		output, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode())
		check(err)
		_, copyErr := io.Copy(output, input)
		closeOutputErr := output.Close()
		closeInputErr := input.Close()
		check(copyErr)
		check(closeOutputErr)
		check(closeInputErr)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
