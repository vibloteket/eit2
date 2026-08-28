// Command zipdir creates a ZIP archive containing the named source directory.
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: zipdir SOURCE_DIR OUTPUT.zip")
		os.Exit(2)
	}
	source, err := filepath.Abs(os.Args[1])
	check(err)
	output, err := os.Create(os.Args[2])
	check(err)
	archive := zip.NewWriter(output)

	var paths []string
	check(filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	}))
	sort.Strings(paths)
	base := filepath.Dir(source)
	for _, path := range paths {
		info, err := os.Stat(path)
		check(err)
		header, err := zip.FileInfoHeader(info)
		check(err)
		relative, err := filepath.Rel(base, path)
		check(err)
		header.Name = filepath.ToSlash(relative)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		header.Modified = info.ModTime()
		writer, err := archive.CreateHeader(header)
		check(err)
		if info.IsDir() {
			continue
		}
		file, err := os.Open(path)
		check(err)
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		check(copyErr)
		check(closeErr)
	}
	check(archive.Close())
	check(output.Close())
	if !strings.HasSuffix(os.Args[2], ".zip") {
		fmt.Fprintln(os.Stderr, "warning: output does not end in .zip")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
