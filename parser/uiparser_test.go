package parser

import "testing"

func TestParser_Parse(t *testing.T) {
	err, parser := NewParser("../test.xml")
	if err != nil {
		panic(err)
	}

	parser.Parse()
}
