package parser

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"io/ioutil"
	"path/filepath"
)

var _ = Describe("TestParser", func() {
	It("test", func() {
		err, compiler := NewCompiler("../sample/ui/test.ui")
		if err != nil {
			panic(err)
		}

		compiler.Parse()
		compiler.GenerateCode("main", "test/test_ui/test_ui.go")
		compiler.GenerateTestCode("test/test_ui/main.go", "")
	})
})

var _ = Describe("TestParseMW", func() {
	It("test", func() {
		err, compiler := NewCompiler("../sample/ui/test_main_window.ui")
		if err != nil {
			panic(err)
		}

		compiler.Parse()
		compiler.GenerateCode("main", "test/test_main_window_ui/test_main_window_ui.go")
		compiler.GenerateTestCode("test/test_main_window_ui/main.go", "")
	})
})

var _ = Describe("TestConnection", func() {
	It("test", func() {
		err, compiler := NewCompiler("../sample/ui/test_signal_slot.ui")
		if err != nil {
			panic(err)
		}

		compiler.Parse()
		compiler.GenerateCode("main", "test/test_signal_slot_ui/test_signal_slot_ui.go")
		compiler.GenerateTestCode("test/test_signal_slot_ui/main.go", "")
	})
})

var _ = XDescribe("TestMoreParser", func() {
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
