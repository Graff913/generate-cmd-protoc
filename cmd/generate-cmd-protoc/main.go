package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
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

	fmt.Printf("protoc --proto_path=%s\n", *rootPath)
	fmt.Printf("--go_opt=paths=source_relative --go-grpc_opt=paths=source_relative\n")
	fmt.Printf("--go_out=./%s --go-grpc_out=./%s\n", *generatePath, *generatePath)

	for _, value := range imports {
		index := strings.LastIndex(value, "/")
		fmt.Printf("--go_opt=M%s=%s/%s/%s\n", value, *nameProject, *generatePath, value[:index])
		fmt.Printf("--go-grpc_opt=M%s=%s/%s/%s\n", value, *nameProject, *generatePath, value[:index])
	}
	for _, value := range imports {
		fmt.Printf("%s\n", value)
	}
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
