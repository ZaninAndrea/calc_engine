package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) < 1 {
		fmt.Print("You need to pass the command and the path of the file to execute")
		os.Exit(1)
	}

	command := argsWithoutProg[0]

	sourceCode := ""
	LoadUnitAliases()

	// if path is passed read file from path, otherwise
	if len(argsWithoutProg) > 1 {
		rawSource, err := ioutil.ReadFile(argsWithoutProg[1])

		if err != nil {
			panic(err)
		}
		sourceCode = string(rawSource)
	} else {
		reader := bufio.NewReader(os.Stdin)

		for {
			data := make([]byte, 1<<16) // read in blocks of 2^16 Bytes
			count, err := reader.Read(data)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatalf("Problems reading from input: %s", err)
			}

			sourceCode += string(data[:count])
		}
	}

	if command == "execute" {
		graph := ParseCode(sourceCode)
		graph.Execute()

		for i := range graph.Lines {
			if graph.Lines[i].HasError() {
				fmt.Println("!")
			} else if graph.Lines[i].IsEmpty() {
				fmt.Println("X")
			} else {
				unitString := graph.Lines[i].Unit.String()

				if unitString != "" {
					unitString = " " + unitString
				}

				fmt.Printf("%f%s\n", roundToDecimal(graph.Lines[i].Value, 13), unitString)
			}
		}
	} else if command == "colorize" {
		graph := ExecutionGraph{SourceCode: sourceCode}
		graph.Tokenize(true)
		fmt.Println(graph.ColorizedHTML())
	}
}
