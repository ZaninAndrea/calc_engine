package main

import "fmt"

// Token stores the information about a single syntactical token, e.g. a constant or a function name
type Token struct {
	Kind  string
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("[%s] %s", t.Kind, t.Value)
}
