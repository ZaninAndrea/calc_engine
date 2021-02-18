package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParseExecute(t *testing.T) {
	path := "./test.cal"
	rawSource, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}
	sourceCode := string(rawSource)

	graph := ParseCode(sourceCode)
	graph.Execute()

	fmt.Println(graph.Lines[0].Value)
	if graph.Lines[0].Value != 71 {
		t.Errorf("Output should be 71")
	}
}
