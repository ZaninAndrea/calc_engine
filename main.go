package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	path := "./test.cal"
	rawSource, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}
	sourceCode := string(rawSource)

	graph := ParseCode(sourceCode)
	graph.Execute()

	fmt.Println(graph.Lines[0].Value)
}
