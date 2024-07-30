package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var (
	o            = flag.String("o", "", "The output file for the cmd protoc.")
	nameProject  = flag.String("n", "<NAME_PROJECT>", "The name project: module name from go.mod")
	rootPath     = flag.String("r", "<ROOT_PATH>", "The root dictionary path with proto-files")
	generatePath = flag.String("g", "<GENERATE_PATH>", "The output dictionary for files generate.")
)

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Printf("  paths")
		fmt.Printf("\tThe input JSON Schema files.")
	}

	flag.Parse()
	rootFiles := flag.Args()

	mapFiles := make(map[string]struct{})
	for _, rootFile := range rootFiles {
		readImports(*rootPath, rootFile, mapFiles)
	}
	imports := make([]string, 0, len(mapFiles))
	for key, _ := range mapFiles {
		imports = append(imports, key)
	}
	sort.Strings(imports)

	var w io.Writer = os.Stdout
	var postfix = "\n"
	var err error

	if *o != "" {
		w, err = os.Create(*o)
		postfix = " "
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error opening output file: ", err)
			return
		}
	}

	output(w, imports, postfix)
}

func readImports(rootPath, pathFile string, mapFiles map[string]struct{}) {
	file, err := os.Open(fmt.Sprintf("%s/%s", rootPath, pathFile))
	if err != nil {
		return
	}
	defer file.Close()

	if _, ok := mapFiles[pathFile]; !ok {
		mapFiles[pathFile] = struct{}{}
	}

	reader := bufio.NewReader(file)

	var line []byte
	for err == nil {
		line, _, err = reader.ReadLine()

		if strings.Contains(string(line), "import \"") {
			importPath := strings.Split(string(line), "\"")[1]
			readImports(rootPath, importPath, mapFiles)
		}
		if strings.Contains(string(line), "package") {
			break
		}
	}
}

func output(w io.Writer, imports []string, postfix string) {
	_, _ = fmt.Fprintf(w, "protoc --proto_path=%s%s", *rootPath, postfix)
	_, _ = fmt.Fprintf(w, "--go_opt=paths=source_relative --go-grpc_opt=paths=source_relative%s", postfix)
	_, _ = fmt.Fprintf(w, "--go_out=./%s --go-grpc_out=./%s%s", *generatePath, *generatePath, postfix)

	for _, value := range imports {
		index := strings.LastIndex(value, "/")
		_, _ = fmt.Fprintf(w, "--go_opt=M%s=%s/%s/%s%s", value, *nameProject, *generatePath, value[:index], postfix)
		_, _ = fmt.Fprintf(w, "--go-grpc_opt=M%s=%s/%s/%s%s", value, *nameProject, *generatePath, value[:index], postfix)
	}
	for _, value := range imports {
		_, _ = fmt.Fprintf(w, "%s%s", value, postfix)
	}
}
