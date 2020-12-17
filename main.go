package main

import (
	"flag"
	"fmt"
	"github.com/stephenlyu/goqtuic/parser"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func translateUIFile(uiFile string, destDir string, testGoFile string) error {
	base := filepath.Base(uiFile)
	destDir, _ = filepath.Abs(filepath.Clean(destDir))

	goFile := filepath.Join(destDir, strings.Replace(base, ".", "_", -1)+".go")

	// 检查文件更新时间，如果go文件时间比ui文件晚，就不重新生成
	goStat, err := os.Stat(goFile)
	if err == nil {
		uiStat, err := os.Stat(uiFile)
		if err != nil {
			return err
		}

		if goStat.ModTime().UnixNano() > uiStat.ModTime().UnixNano() {
			return nil
		}
	}

	fmt.Printf("Translating %s...\n", uiFile)

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
				if genPackage[len(genPackage)-1] == filepath.Separator {
					genPackage = genPackage[:len(genPackage)-1]
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
