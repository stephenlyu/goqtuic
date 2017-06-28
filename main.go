package main

import (
	"flag"
	"os"
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/stephenlyu/goqtuic/parser"
	"fmt"
)

func translateUIFile(uiFile string, destDir string, testGoFile string) error {
	fmt.Printf("Translating %s...\n", uiFile)
	base := filepath.Base(uiFile)
	destDir, _ = filepath.Abs(filepath.Clean(destDir))

	goFile := filepath.Join(destDir, strings.Replace(base, ".", "_", -1) + ".go")
	packageName := filepath.Base(destDir)

	err, compiler := parser.NewCompiler(uiFile)
	if err != nil {
		return err
	}

	compiler.Parse()
	err = compiler.GenerateCode(packageName, goFile)
	if err != nil {
		return err
	}

	if testGoFile != "" {
		var genPackage string
		goPath := filepath.Clean(os.Getenv("GOPATH"))
		if goPath != "" {
			sourcePath := filepath.Join(goPath, "src")
			if strings.HasPrefix(destDir, sourcePath) {
				genPackage = destDir[len(sourcePath):]
				if genPackage[0] == filepath.Separator {
					genPackage = genPackage[1:]
				}
				if genPackage[len(genPackage) - 1] == filepath.Separator {
					genPackage = genPackage[:len(genPackage) - 1]
				}
			}
		}

		return compiler.GenerateTestCode(testGoFile, genPackage)
	}

	return nil
}

func main() {
	uiFile := flag.String("ui-file", "ui", "QT Designer ui file or directory")
	uiGoDir := flag.String("go-ui-dir", "uigen", "Generated ui go files directory")
	testGoFile := flag.String("go-test-file", "", "Test go file path")

	flag.Parse()


	stat, err := os.Stat(*uiFile)
	if err != nil {
		panic(err)
	}

	if stat.IsDir() {
		files, err := ioutil.ReadDir(*uiFile)
		if err != nil {
			panic(err)
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			if filepath.Ext(f.Name()) != ".ui" {
				continue
			}

			filePath := filepath.Join(*uiFile, f.Name())

			translateUIFile(filePath, *uiGoDir, "")
		}
	} else {
		translateUIFile(*uiFile, *uiGoDir, *testGoFile)
	}
}
