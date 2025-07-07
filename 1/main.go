package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
)

func removeFiles(files []fs.FileInfo) (ans []fs.FileInfo) {
	for _, elem := range files {
		if elem.IsDir() {
			ans = append(ans, elem)
		}
	}
	return
}

func writePref(out io.Writer, pref []bool) {
	for _, elem := range pref {
		if elem {
			fmt.Fprint(out, "│\t")
		} else {
			fmt.Fprint(out, " \t")
		}
	}
}

func writeBranch(out io.Writer, last bool) {
	start := "├"
	if last {
		start = "└"
	}
	fmt.Fprint(out, start+"───")
}

func dfsDirTree(out io.Writer, path string, printFiles bool, prefix []bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	files, err := f.Readdir(0)
	if err != nil {
		return err
	}

	if !printFiles {
		files = removeFiles(files)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for i, elem := range files {
		last := i == len(files)-1
		writePref(out, prefix)
		writeBranch(out, last)
		fmt.Fprint(out, elem.Name())
		next := "\n"
		if last && len(prefix) == 0 {
			next = ""
		}
		if elem.IsDir() {
			fmt.Fprint(out, "\n")
			suf := true
			if last {
				suf = false
			}
			dfsDirTree(out, path+string(os.PathSeparator)+elem.Name(), printFiles, append(prefix, suf))
		} else {
			if elem.Size() != 0 {
				fmt.Fprintf(out, " (%db)"+next, elem.Size())
			} else {
				fmt.Fprintf(out, " (empty)"+next)
			}
		}
	}

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dfsDirTree(out, path, printFiles, make([]bool, 0))
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
