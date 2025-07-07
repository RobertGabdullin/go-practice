package main

import (
	"fmt"
	"io"
	"os"
	"sort"
)

func writePrefix(active []bool, out io.Writer) error {
	for _, flag := range active {
		if flag {
			out.Write([]byte("│"))
		}
		out.Write([]byte{'\t'})
	}
	return nil
}

func Out(flag bool) []byte {
	if flag {
		return []byte("├───")
	}
	return []byte("└───")
}

func printSize(flag bool, entry os.DirEntry) []byte {
	if !flag || entry.IsDir() {
		return []byte{}
	}
	fileInfo, _ := entry.Info()
	sz := fileInfo.Size()

	add := fmt.Sprint(sz) + "b"
	if sz == 0 {
		add = "empty"
	}

	return []byte(" (" + add + ")")
}

func dfsDirTree(active []bool, out io.Writer, path string, printFiles bool) error {
	list, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var newList []os.DirEntry
	for _, entry := range list {
		if entry.IsDir() {
			newList = append(newList, entry)
		}
	}

	if !printFiles {
		list = newList
	}

	if len(list) == 0 {
		return nil
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})

	flag := true

	for idx, entry := range list {

		writePrefix(active, out)

		if idx == len(list)-1 {
			flag = false
		}

		out.Write(Out(flag))

		out.Write([]byte(entry.Name()))

		out.Write(printSize(printFiles, entry))

		out.Write([]byte("\n"))

		if entry.IsDir() {
			err = dfsDirTree(append(active, flag), out, path+string(os.PathSeparator)+entry.Name(), printFiles)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dfsDirTree([]bool{}, out, path, printFiles)
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
