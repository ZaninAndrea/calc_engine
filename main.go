package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		fmt.Print("You need to pass the path of the file to execute")
		os.Exit(1)
	}

	path := argsWithoutProg[0]
	rawSource, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}
	sourceCode := string(rawSource)

	graph := ParseCode(sourceCode)
	graph.Execute()

	for i := range graph.Lines {
		if graph.Lines[i].IsEmpty() {
			fmt.Println("X")
		} else {
			fmt.Println(graph.Lines[i].Value)
		}
	}
}
