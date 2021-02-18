package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Line contains the compiled data for one line of code
type Line struct {
	Name         string  // it's name as variable, if assigned
	Tokens       []Token // it's tokenization
	RawTokens    []Token // tokenization including whitespace and comment
	Dependencies []int
	Value        float64
	Ast          Ast
}

// IsEmpty returns whether the Line contains an empty expression
func (line *Line) IsEmpty() bool {
	return len(line.Tokens) == 0
}

// ExecutionGraph contains the interpreted code
type ExecutionGraph struct {
	Lines          []Line
	Variables      map[string]int // map from variable to the corresponding line
	ExecutionOrder []int
	SourceCode     string
}

// ParseCode parses a sourcecode into an ExecutionGraph
func ParseCode(sourceCode string) ExecutionGraph {
	graph := ExecutionGraph{SourceCode: sourceCode}

	for _, line := range strings.Split(sourceCode, "\n") {
		tokens := tokenizer(line)
		graph.Lines = append(graph.Lines, Line{Tokens: removeNonSemanticTokens(tokens), RawTokens: tokens})
	}

	graph.parseVariableDeclarations()
	graph.parseLineDependencies()

	if graph.hasCyclicalDependencies() {
		panic("Cyclical definitions detected")
	}

	graph.findExecutionOrder()

	for i := range graph.Lines {
		graph.Lines[i].Ast = parser(graph.Lines[i].Tokens, graph.Variables)
	}

	return graph
}

// Parse a line of code into a list of tokens
func tokenizer(source string) []Token {
	current := 0
	tokens := []Token{}

	digits := []byte("0123456789")
	numberChars := []byte("0123456789.,%")

	literalStartChars := []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_")
	literalChars := []byte("qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_0123456789")

	operators := []byte("+-*/^")

	for current < len(source) {
		char := source[current]

		// Everything after the comment marker is ignored
		if char == '#' {
			tokens = append(tokens, Token{"comment", source[current:]})

			break
		}

		// skip whitespace
		if char == ' ' || char == '\t' {
			val := ""

			for (source[current] == ' ' || source[current] == '\t') && current < len(source) {
				val += string(source[current])
				current++
			}

			tokens = append(tokens, Token{"whitespace", val})
			continue
		}

		// match open and close parenthesis and definitions
		if char == '(' {
			tokens = append(tokens, Token{"paren", "("})

			current++
			continue
		}
		if char == ')' {
			tokens = append(tokens, Token{"paren", ")"})

			current++
			continue
		}
		if char == ':' {
			tokens = append(tokens, Token{"definition", ":"})

			current++
			continue
		}

		if containsByte(operators, char) {
			tokens = append(tokens, Token{"operator", string(char)})

			current++
			continue
		}

		// match a number
		if containsByte(digits, char) {
			value := ""

			for containsByte(numberChars, char) {
				value += string(char)
				current++

				if current >= len(source) {
					break
				}

				char = source[current]
			}

			tokens = append(tokens, Token{"number", value})

			continue
		}

		// match a variable
		if containsByte(literalStartChars, char) {
			value := ""

			for containsByte(literalChars, char) {
				value += string(char)
				current++

				if current >= len(source) {
					break
				}

				char = source[current]
			}

			tokens = append(tokens, Token{"literal", value})

			continue
		}

		// handling unknown characters
		if char == '\n' {
			panic("Tokenizer should parse single lines, \\n found")
		}
		panic("Unknown character " + string(char))
	}

	return tokens
}

func removeNonSemanticTokens(tokens []Token) []Token {
	filteredSlice := []Token{}

	for i := range tokens {
		if tokens[i].Kind != "comment" && tokens[i].Kind != "whitespace" {
			filteredSlice = append(filteredSlice, tokens[i])
		}
	}

	return filteredSlice
}

// Check which lines are declaring a variable
func (graph *ExecutionGraph) parseVariableDeclarations() {
	graph.Variables = map[string]int{}
	for i := range graph.Lines {
		line := &graph.Lines[i]
		if len(line.Tokens) > 1 && line.Tokens[0].Kind == "literal" && line.Tokens[1].Kind == "definition" {
			graph.Variables[line.Tokens[0].Value] = i
			line.Name = line.Tokens[0].Value

			line.Tokens = line.Tokens[2:]
		}
	}
}

// For every line find which variables it references
func (graph *ExecutionGraph) parseLineDependencies() {
	for i := range graph.Lines {
		line := &graph.Lines[i]

		// loop over all tokens and check if they are variable literals
		for _, token := range line.Tokens {
			if token.Kind == "literal" {
				val, ok := graph.Variables[token.Value]

				if ok {
					line.Dependencies = append(line.Dependencies, val)
				}
			}
		}
	}
}

// Checks if there are cycles in the dependency graph
func (graph *ExecutionGraph) hasCyclicalDependencies() bool {
	dt := make([]int, len(graph.Lines))
	ft := make([]int, len(graph.Lines))
	step := 1

	for i := range graph.Lines {
		if dt[i] == 0 {
			if recHasCycles(graph, &dt, &ft, i, &step) {
				return true
			}
		}
	}

	return false
}

// Uses a Depth First Search to look for cycles in the directed graph given by the dependencies
func recHasCycles(graph *ExecutionGraph, dt *[]int, ft *[]int, node int, step *int) bool {
	(*dt)[node] = *step
	(*step)++

	for _, n := range graph.Lines[node].Dependencies {
		// If the node hasn't been visited call recursively, if it has check if this edge closes a loop
		if (*dt)[n] == 0 {
			if recHasCycles(graph, dt, ft, n, step) {
				return true
			}
		} else if (*ft)[n] == 0 {
			return true
		}
	}

	(*ft)[node] = *step
	(*step)++

	return false
}

// Computes a topological order in the dependencies graph
func (graph *ExecutionGraph) findExecutionOrder() {
	visited := make([]bool, len(graph.Lines))
	order := []int{}

	for i := range graph.Lines {
		if !visited[i] {
			recTopologicalOrder(graph, i, &order, &visited)
		}
	}

	graph.ExecutionOrder = order
}

// Computes a topological order using a Depth First Visit of the graph, appending values in post-order
func recTopologicalOrder(graph *ExecutionGraph, line int, order *[]int, visited *[]bool) {
	// Add to the order first all the depended lines (if not already added)
	for _, l := range graph.Lines[line].Dependencies {
		if !(*visited)[l] {
			recTopologicalOrder(graph, l, order, visited)
		}
	}

	// Then add to the order this line
	(*order) = append(*order, line)
	(*visited)[line] = true
}

func parser(tokens []Token, variables map[string]int) Ast {
	functions := []string{"sqrt", "log", "ln", "sin", "cos", "tan", "abs", "ln", "round", "ceil", "floor"}
	constants := []string{"pi", "e"}

	current := 0

	var walk func() Ast
	walk = func() Ast {
		token := tokens[current]

		if token.Kind == "number" {
			current++

			return Ast{Kind: "NumberLiteral", Value: token.Value}
		}

		// Match all the tokens inside the parenthesis
		if token.Kind == "paren" && token.Value == "(" {
			current++
			token = tokens[current]

			ast := Ast{Kind: "Expression", Params: []Ast{}}

			for token.Kind != "paren" || token.Value != ")" {
				ast.Params = append(ast.Params, walk())
				token = tokens[current]
			}

			current++

			return ast
		}

		// literals can be constants, variables or functions
		if token.Kind == "literal" {
			if containsString(constants, token.Value) {
				current++

				return Ast{Kind: "Constant", Value: token.Value}
			}

			if _, ok := variables[token.Value]; ok {
				current++

				return Ast{Kind: "Variable", Value: token.Value}
			}

			if containsString(functions, token.Value) {
				ast := Ast{Kind: "Function", Value: token.Value}

				current++
				token = tokens[current]
				ast.Params = []Ast{walk()}

				return ast
			}
		}

		if token.Kind == "operator" {
			current++

			return Ast{Kind: "RawOperator", Value: token.Value}
		}

		panic("Unrecognized syntax")
	}

	ast := &Ast{Kind: "Expression", Params: []Ast{}}

	for current < len(tokens) {
		ast.Params = append(ast.Params, walk())
	}

	for _, operator := range []string{"^", "*", "/", "-", "+"} {
		ast = parseOperator(ast, operator)
	}

	return *ast
}

func parseOperator(ast *Ast, operator string) *Ast {
	if ast.Kind == "NumberLiteral" || ast.Kind == "Constant" || ast.Kind == "Variable" {
		return ast
	}

	if ast.Kind == "Function" {
		if len(ast.Params) == 0 {
			panic("Function called without argument")
		}

		if ast.Params[0].Kind == "RawOperator" {
			panic("Cannot pass operation as argument to function")
		}

		ast.Params = []Ast{*parseOperator(&ast.Params[0], operator)}

		return ast
	}
	if ast.Kind == "Operator" {
		firstParam := *parseOperator(&ast.Params[0], operator)
		secondParam := *parseOperator(&ast.Params[1], operator)
		ast.Params = []Ast{firstParam, secondParam}

		return ast
	}

	if ast.Kind == "Expression" {
		parsedParams := []Ast{}

		for i := 0; i < len(ast.Params); i++ {
			token := ast.Params[i]

			if token.Kind != "RawOperator" {
				parsedParams = append(parsedParams, *parseOperator(&token, operator))
				continue
			} else {
				if token.Value != operator {
					parsedParams = append(parsedParams, token)
					continue
				}

				// operators cannot end an expression
				if i >= len(ast.Params)-1 {
					panic("Cannot end expression with operation")
				}

				// only - operator can start an expression
				if len(parsedParams) == 0 && operator != "-" {
					panic("Cannot start expression with operation")
				}

				newAst := Ast{Kind: "Operator", Value: token.Value}

				var firstToken Ast
				// -expression is parsed as 0-expression
				if len(parsedParams) == 0 {
					firstToken = Ast{Kind: "NumberLiteral", Value: "0"}
				} else {
					firstToken = parsedParams[len(parsedParams)-1]
					parsedParams = parsedParams[:len(parsedParams)-1]
				}

				i++
				token = ast.Params[i]
				if token.Kind == "RawOperator" {
					panic("Cannot have 2 operations consecutively")
				}
				secondToken := *parseOperator(&token, operator)

				newAst.Params = []Ast{firstToken, secondToken}
				parsedParams = append(parsedParams, newAst)
			}
		}

		ast.Params = parsedParams
		return ast
	}

	panic("Unrecognized AST")
}

// Execute computes the value of each line in the file
func (graph *ExecutionGraph) Execute() {
	for _, line := range graph.ExecutionOrder {
		if !graph.Lines[line].IsEmpty() {
			graph.Lines[line].Value = executeAst(&graph.Lines[line].Ast, graph)
		}
	}
}

func executeAst(ast *Ast, graph *ExecutionGraph) float64 {
	if ast.Kind == "NumberLiteral" {
		raw := ast.Value
		raw = strings.ReplaceAll(raw, ".", "")
		raw = strings.ReplaceAll(raw, ",", ".")

		isPercentage := false
		if raw[len(raw)-1] == '%' {
			raw = raw[:len(raw)-1]
			isPercentage = true
		}

		val, err := strconv.ParseFloat(raw, 64)

		if err != nil {
			panic("Invalid number literal")
		}

		if isPercentage {
			val /= 100
		}

		return val
	}

	if ast.Kind == "Variable" {
		line, _ := graph.Variables[ast.Value]

		if graph.Lines[line].IsEmpty() {
			panic("Referring to a variable defined by empty expression")
		}

		return graph.Lines[line].Value
	}

	if ast.Kind == "Expression" {
		if len(ast.Params) == 0 {
			panic("Cannot evaluate empty expression")
		}

		return executeAst(&ast.Params[0], graph)
	}

	if ast.Kind == "Operator" {
		firstValue := executeAst(&ast.Params[0], graph)
		secondValue := executeAst(&ast.Params[1], graph)

		switch ast.Value {
		case "+":
			return firstValue + secondValue
		case "-":
			return firstValue - secondValue
		case "*":
			return firstValue * secondValue
		case "/":
			return firstValue / secondValue
		case "^":
			return math.Pow(firstValue, secondValue)
		default:
			panic("Unknown operation")
		}
	}

	if ast.Kind == "Function" {
		value := executeAst(&ast.Params[0], graph)

		switch ast.Value {
		case "sqrt":
			return math.Sqrt(value)
		case "log":
			return math.Log10(value)
		case "ln":
			return math.Log(value)
		case "sin":
			return math.Sin(value)
		case "cos":
			return math.Cos(value)
		case "tan":
			return math.Tan(value)
		case "abs":
			return math.Abs(value)
		case "round":
			return math.Round(value)
		case "ceil":
			return math.Ceil(value)
		case "floor":
			return math.Floor(value)
		default:
			panic("Unknown function")
		}
	}

	if ast.Kind == "Constant" {
		switch ast.Value {
		case "pi":
			return math.Pi
		case "e":
			return math.E
		default:
			panic("Unknown constant")
		}
	}

	panic("Unrecognized AST")
}

// ColorizedHTML returns the source code as HTML with CSS classes to tag the parsed semantic
func (graph *ExecutionGraph) ColorizedHTML() string {
	colorizedLines := []string{}

	for _, line := range graph.Lines {
		colorizedLine := ""

		for _, token := range line.RawTokens {
			colorizedLine += fmt.Sprintf(`<span class="calc-token-%s">%s</span>`, token.Kind, token.Value)
		}

		colorizedLines = append(colorizedLines, colorizedLine)
	}

	return strings.Join(colorizedLines, "<br/>")
}
