package main

import (
	"fmt"
	"strings"
)

type Ast struct {
	Kind   string
	Value  string
	Params []Ast
}

func (ast Ast) String() string {
	repr := fmt.Sprintf("[%s] %s", ast.Kind, ast.Value)

	for _, node := range ast.Params {
		for _, line := range strings.Split(node.String(), "\n") {
			repr += "\n  " + line
		}
	}

	return repr
}
