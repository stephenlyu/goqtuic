package parser

import (
	"testing"
	"io/ioutil"
	"path/filepath"
	"fmt"
)

func TestParser_Parse(t *testing.T) {
	err, parser := NewParser("../ui/ai_comparability_dialog.ui")
	if err != nil {
		panic(err)
	}

	parser.Parse()
}

func TestMoreFiles(t *testing.T) {
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
		err, parser := NewParser(filePath)
		if err != nil {
			panic(err)
		}

		parser.Parse()
	}
}
