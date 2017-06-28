package parser

import (
	. "github.com/onsi/ginkgo"
	"io/ioutil"
	"path/filepath"
	"fmt"
)

var _ = Describe("TestParser", func() {
	It("test", func() {
		err, compiler := NewCompiler("test.ui")
		if err != nil {
			panic(err)
		}

		compiler.Parse()
		compiler.GenerateCode("parser", "sample/test_ui.go")
	})
})

var _ = Describe("TestMoreParser", func() {
	It("test", func() {
		root := "../ui"
		fileList, err := ioutil.ReadDir(root)
		if err != nil {
			panic(err)
		}

		for _, file := range fileList {
			if file.IsDir() {
				continue
			}
			if filepath.Ext(file.Name()) != ".ui" {
				continue
			}

			filePath := filepath.Join(root, file.Name())
			fmt.Println(filePath)
			err, compiler := NewCompiler(filePath)
			if err != nil {
				panic(err)
			}

			compiler.Parse()

			if len(compiler.buttonGroups) > 0 {
				fmt.Println("button group count: ", len(compiler.buttonGroups))
			}
		}
	})
})
