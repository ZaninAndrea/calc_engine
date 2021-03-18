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
	Unit         CompositeUnit
	Error        error
}

// IsEmpty returns whether the Line contains an empty expression
func (line *Line) IsEmpty() bool {
	return len(line.Tokens) == 0
}

// HasError returns whether the Line is invalid
func (line *Line) HasError() bool {
	return line.Error != nil
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

	graph.Tokenize(false)
	graph.parseVariableDeclarations()
	graph.parseLineDependencies()

	if graph.hasCyclicalDependencies() {
		panic("Cyclical definitions detected")
	}

	graph.findExecutionOrder()

	for i := range graph.Lines {
		ast, err := parser(graph.Lines[i].Tokens, graph.Variables)

		if err != nil {
			graph.Lines[i].Error = err
		} else {
			graph.Lines[i].Ast = ast
		}
	}

	return graph
}

// Tokenize computes the token representation of each line
func (graph *ExecutionGraph) Tokenize(allowUnknown bool) *ExecutionGraph {
	for _, line := range strings.Split(graph.SourceCode, "\n") {
		tokens, err := tokenizer(line, allowUnknown)

		if err != nil {
			graph.Lines = append(graph.Lines, Line{Error: err})
		} else {
			graph.Lines = append(graph.Lines, Line{Tokens: removeNonSemanticTokens(tokens), RawTokens: tokens})
		}
	}

	return graph
}

// Parse a line of code into a list of tokens
func tokenizer(source string, allowUnknown bool) ([]Token, error) {
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

			for current < len(source) && (source[current] == ' ' || source[current] == '\t') {
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
		if char == '[' {
			tokens = append(tokens, Token{"bracket", "["})

			current++
			continue
		}
		if char == ']' {
			tokens = append(tokens, Token{"bracket", "]"})

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

		if allowUnknown {
			tokens = append(tokens, Token{"unknown", string(char)})
			current++

			continue
		}
		return nil, fmt.Errorf("Unknown character " + string(char))
	}

	return tokens, nil
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

func parser(tokens []Token, variables map[string]int) (Ast, error) {
	functions := []string{"sqrt", "log", "ln", "sin", "cos", "tan", "abs", "ln", "round", "ceil", "floor"}
	constants := []string{"pi", "e"}

	current := 0

	walkUnit := func() (Ast, error) {
		if current >= len(tokens) {
			return Ast{}, fmt.Errorf("Line ends unexpectedly")
		}

		token := tokens[current]

		if token.Kind == "number" {
			current++

			return Ast{Kind: "UnitNumberLiteral", Value: token.Value}, nil
		}

		// literals can be known units or unknown units
		if token.Kind == "literal" {
			if val, ok := UnitAliasesMap[token.Value]; ok {
				current++

				return Ast{Kind: "FundamentalUnit", Value: val}, nil
			} else {
				current++

				return Ast{Kind: "CustomUnit", Value: token.Value}, nil
			}
		}

		if token.Kind == "operator" && token.Value == "^" {
			current++

			return Ast{Kind: "UnitExponent", Value: token.Value}, nil
		}
		if token.Kind == "operator" && token.Value == "/" {
			current++

			return Ast{Kind: "UnitDivision", Value: token.Value}, nil
		}

		return Ast{}, fmt.Errorf("Unrecognized unit syntax")
	}

	var walk func() (Ast, error)
	walk = func() (Ast, error) {
		if current >= len(tokens) {
			return Ast{}, fmt.Errorf("Line ends unexpectedly")
		}

		token := tokens[current]

		if token.Kind == "number" {
			current++

			return Ast{Kind: "NumberLiteral", Value: token.Value}, nil
		}

		// Match all the tokens inside the parenthesis
		if token.Kind == "paren" && token.Value == "(" {
			current++

			if current >= len(tokens) {
				return Ast{}, fmt.Errorf("Line ends unexpectedly")
			}

			token = tokens[current]

			ast := Ast{Kind: "Expression", Params: []Ast{}}

			for token.Kind != "paren" || token.Value != ")" {
				content, err := walk()

				if err != nil {
					return Ast{}, err
				}

				if content.Kind != "UnitExpression" {
					ast.Params = append(ast.Params, content)
				} else {
					ast.Unit = content.Unit
				}

				if current >= len(tokens) {
					return Ast{}, fmt.Errorf("Line ends unexpectedly")
				}
				token = tokens[current]
			}

			current++

			return ast, nil
		}

		// Match all the tokens inside the brackets
		if token.Kind == "bracket" && token.Value == "[" {
			current++

			if current >= len(tokens) {
				return Ast{}, fmt.Errorf("Line ends unexpectedly")
			}

			token = tokens[current]

			ast := Ast{Kind: "UnitExpression", Params: []Ast{}}

			for token.Kind != "bracket" || token.Value != "]" {
				content, err := walkUnit()

				if err != nil {
					return Ast{}, err
				}
				ast.Params = append(ast.Params, content)

				if current >= len(tokens) {
					return Ast{}, fmt.Errorf("Line ends unexpectedly")
				}
				token = tokens[current]
			}

			unit, err := parseUnitAst(ast)

			if err != nil {
				return Ast{}, err
			}
			ast.Unit = unit

			current++
			return ast, nil
		}

		// literals can be constants, variables or functions
		if token.Kind == "literal" {
			if containsString(constants, token.Value) {
				current++

				return Ast{Kind: "Constant", Value: token.Value}, nil
			}

			if _, ok := variables[token.Value]; ok {
				current++

				return Ast{Kind: "Variable", Value: token.Value}, nil
			}

			if containsString(functions, token.Value) {
				ast := Ast{Kind: "Function", Value: token.Value}

				current++

				if current >= len(tokens) {
					return Ast{}, fmt.Errorf("Line ends unexpectedly")
				}

				token = tokens[current]

				content, err := walk()

				if err != nil {
					return Ast{}, err
				}

				ast.Params = []Ast{content}

				return ast, nil
			}
		}

		if token.Kind == "operator" {
			current++

			return Ast{Kind: "RawOperator", Value: token.Value}, nil
		}

		return Ast{}, fmt.Errorf("Unrecognized syntax")
	}

	ast := &Ast{Kind: "Expression", Params: []Ast{}}

	for current < len(tokens) {
		content, err := walk()

		if err != nil {
			return Ast{}, err
		}

		if content.Kind != "UnitExpression" {
			ast.Params = append(ast.Params, content)
		} else {
			ast.Unit = content.Unit
		}
	}

	for _, operator := range []string{"^", "*", "/", "-", "+"} {
		newAst, err := parseOperator(ast, operator)

		if err != nil {
			return Ast{}, err
		}

		ast = newAst
	}

	return *ast, nil
}

func parseUnitAst(ast Ast) (CompositeUnit, error) {
	cu := CompositeUnit{}

	// TODO: give error on m^2^3

	exponentSign := float64(1)

	curr := 0

	for curr < len(ast.Params) {
		token := ast.Params[curr]

		if token.Kind == "FundamentalUnit" {
			cu.UnitsList = append(cu.UnitsList, UnitExponent{UnitTable[token.Value], exponentSign})
			curr++
			continue
		}

		if token.Kind == "CustomUnit" {
			cu.UnitsList = append(cu.UnitsList, UnitExponent{FundamentalUnit{
				ID:               token.Value,
				DisplayValue:     token.Value,
				Aliases:          []string{token.Value},
				BaseUnit:         token.Value,
				ConversionFactor: 1,
				ConversionShift:  0,
			}, exponentSign})
			curr++
			continue
		}

		if token.Kind == "UnitDivision" && exponentSign == 1 {
			exponentSign = -1

			curr++
			continue
		}

		if token.Kind == "UnitExponent" && len(cu.UnitsList) > 0 {
			curr++

			if curr < len(ast.Params) && ast.Params[curr].Kind == "UnitNumberLiteral" {
				exp, err := strconv.ParseFloat(ast.Params[curr].Value, 64)

				if err != nil {
					return CompositeUnit{}, err
				}

				cu.UnitsList[len(cu.UnitsList)-1].Exponent = exp * exponentSign
				curr++
				continue
			} else {
				return CompositeUnit{}, fmt.Errorf("Failed to parse unit expression")
			}
		}

		if token.Kind == "UnitNumberLiteral" && len(cu.UnitsList) == 0 {
			curr++
			continue
		}

		return CompositeUnit{}, fmt.Errorf("Failed to parse unit expression")
	}

	return cu, nil
}

func parseOperator(ast *Ast, operator string) (*Ast, error) {
	if ast.Kind == "NumberLiteral" || ast.Kind == "Constant" || ast.Kind == "Variable" {
		return ast, nil
	}

	if ast.Kind == "Function" {
		if len(ast.Params) == 0 {
			return nil, fmt.Errorf("Function called without argument")
		}

		if ast.Params[0].Kind == "RawOperator" {
			return nil, fmt.Errorf("Cannot pass operation as argument to function")
		}

		content, err := parseOperator(&ast.Params[0], operator)
		if err != nil {
			return nil, err
		}

		ast.Params = []Ast{*content}

		return ast, nil
	}
	if ast.Kind == "Operator" {
		firstParam, err1 := parseOperator(&ast.Params[0], operator)
		secondParam, err2 := parseOperator(&ast.Params[1], operator)

		if err1 != nil {
			return nil, err1
		} else if err2 != nil {
			return nil, err2
		}

		ast.Params = []Ast{*firstParam, *secondParam}

		return ast, nil
	}

	if ast.Kind == "Expression" {
		parsedParams := []Ast{}

		for i := 0; i < len(ast.Params); i++ {
			token := ast.Params[i]

			if token.Kind != "RawOperator" {
				parsed, err := parseOperator(&token, operator)
				if err != nil {
					return nil, err
				}

				parsedParams = append(parsedParams, *parsed)
				continue
			} else {
				if token.Value != operator {
					parsedParams = append(parsedParams, token)
					continue
				}

				// operators cannot end an expression
				if i >= len(ast.Params)-1 {
					return nil, fmt.Errorf("Cannot end expression with operation")
				}

				// only - operator can start an expression
				if len(parsedParams) == 0 && operator != "-" {
					return nil, fmt.Errorf("Cannot start expression with operation")
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
					return nil, fmt.Errorf("Cannot have 2 operations consecutively")
				}

				secondToken, err := parseOperator(&token, operator)
				if err != nil {
					return nil, err
				}

				newAst.Params = []Ast{firstToken, *secondToken}
				parsedParams = append(parsedParams, newAst)
			}
		}

		ast.Params = parsedParams
		return ast, nil
	}

	panic("Unrecognized AST")
}

// Execute computes the value of each line in the file
func (graph *ExecutionGraph) Execute() {
	for _, line := range graph.ExecutionOrder {
		if !graph.Lines[line].IsEmpty() && !graph.Lines[line].HasError() {
			val, unit, err := executeAst(&graph.Lines[line].Ast, graph)

			if err != nil {
				graph.Lines[line].Error = err
			} else {
				graph.Lines[line].Value = val
				graph.Lines[line].Unit = unit
			}
		}
	}
}

func executeAst(ast *Ast, graph *ExecutionGraph) (float64, CompositeUnit, error) {
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
			return 0, CompositeUnit{}, fmt.Errorf("Invalid number literal")
		}

		if isPercentage {
			val /= 100
		}

		return val, CompositeUnit{}, nil
	}

	if ast.Kind == "Variable" {
		line, _ := graph.Variables[ast.Value]

		if graph.Lines[line].IsEmpty() {
			return 0, CompositeUnit{}, fmt.Errorf("Referring to a variable defined by empty expression")
		} else if graph.Lines[line].HasError() {
			return 0, CompositeUnit{}, fmt.Errorf("Referring to a variable whose definition has an error")
		}

		return graph.Lines[line].Value, graph.Lines[line].Unit, nil
	}

	if ast.Kind == "Expression" {
		if len(ast.Params) == 0 {
			panic("Cannot evaluate empty expression")
		}

		val, unit, err := executeAst(&ast.Params[0], graph)

		if err != nil {
			return val, unit, err
		}

		if ast.Unit.IsEmpty() {
			return val, unit, nil
		}

		if unit.IsEmpty() {
			return val, ast.Unit, nil
		}

		val, err = ConvertCompositeUnits(val, unit, ast.Unit)
		return val, ast.Unit, err
	}

	if ast.Kind == "Operator" {
		firstValue, unit1, err1 := executeAst(&ast.Params[0], graph)
		secondValue, unit2, err2 := executeAst(&ast.Params[1], graph)

		if err1 != nil {
			return 0, CompositeUnit{}, err1
		} else if err2 != nil {
			return 0, CompositeUnit{}, err2
		}

		switch ast.Value {
		case "+":
			secondValueConverted, err := ConvertCompositeUnits(secondValue, unit2, unit1)
			if err != nil {
				return 0, CompositeUnit{}, err
			}

			return firstValue + secondValueConverted, unit1, nil
		case "-":
			secondValueConverted, err := ConvertCompositeUnits(secondValue, unit2, unit1)
			if err != nil {
				return 0, CompositeUnit{}, err
			}

			return firstValue - secondValueConverted, CompositeUnit{}, nil
		case "*":
			val, unit := CompositeUnitProduct(firstValue, secondValue, unit1, unit2)
			return val, unit, nil
		case "/":
			val, unit := CompositeUnitDivision(firstValue, secondValue, unit1, unit2)
			return val, unit, nil
		case "^":
			if !unit2.IsEmpty() {
				return 0, CompositeUnit{}, fmt.Errorf("Exponent must be a number with no unit")
			}

			return math.Pow(firstValue, secondValue), CompositeUnitExponentiation(unit1, secondValue), nil
		default:
			panic("Unknown operation")
		}
	}

	if ast.Kind == "Function" {
		value, unit, err := executeAst(&ast.Params[0], graph)

		if err != nil {
			return 0, CompositeUnit{}, err
		}

		switch ast.Value {
		case "sqrt":
			return math.Sqrt(value), CompositeUnitExponentiation(unit, 0.5), nil
		case "log":
			return math.Log10(value), unit, nil
		case "ln":
			return math.Log(value), unit, nil
		case "sin":
			if unit.String() == "rad" {
				return math.Sin(value), CompositeUnit{}, nil
			}
			if unit.String() == "deg" {
				value = ConvertFundamentalUnits(value, UnitTable["degrees"], UnitTable["radians"], 1)

				return math.Sin(value), CompositeUnit{}, nil
			}

			return math.Sin(value), unit, nil
		case "cos":
			if unit.String() == "rad" {
				return math.Cos(value), CompositeUnit{}, nil
			}
			if unit.String() == "deg" {
				value = ConvertFundamentalUnits(value, UnitTable["degrees"], UnitTable["radians"], 1)
				return math.Cos(value), CompositeUnit{}, nil
			}

			return math.Cos(value), unit, nil
		case "tan":
			if unit.String() == "rad" {
				return math.Tan(value), CompositeUnit{}, nil
			}
			if unit.String() == "deg" {
				value = ConvertFundamentalUnits(value, UnitTable["degrees"], UnitTable["radians"], 1)
				return math.Tan(value), CompositeUnit{}, nil
			}

			return math.Tan(value), unit, nil
		case "abs":
			return math.Abs(value), unit, nil
		case "round":
			return math.Round(value), unit, nil
		case "ceil":
			return math.Ceil(value), unit, nil
		case "floor":
			return math.Floor(value), unit, nil
		default:
			panic("Unknown function")
		}
	}

	if ast.Kind == "Constant" {
		switch ast.Value {
		case "pi":
			return math.Pi, CompositeUnit{}, nil
		case "e":
			return math.E, CompositeUnit{}, nil
		default:
			panic("Unknown constant")
		}
	}

	panic("Unrecognized AST")
}

// ColorizedHTML returns the source code as HTML with CSS classes to tag the parsed semantic
func (graph *ExecutionGraph) ColorizedHTML() string {
	colorizedLines := []string{}
	functions := []string{"sqrt", "log", "ln", "sin", "cos", "tan", "abs", "ln", "round", "ceil", "floor"}
	constants := []string{"pi", "e"}

	for _, line := range graph.Lines {
		colorizedLine := ""
		insideUnitTag := ""

		for _, token := range line.RawTokens {
			if token.Kind == "bracket" && token.Value == "[" {
				insideUnitTag = "-unit"
			}

			if token.Kind != "literal" {
				colorizedLine += fmt.Sprintf(`<span class="calc-token-%s">%s</span>`, token.Kind+insideUnitTag, token.Value)
			} else {
				switch {
				case containsString(functions, token.Value):
					colorizedLine += fmt.Sprintf(`<span class="calc-token-%s">%s</span>`, "function"+insideUnitTag, token.Value)
				case containsString(constants, token.Value):
					colorizedLine += fmt.Sprintf(`<span class="calc-token-%s">%s</span>`, "constant"+insideUnitTag, token.Value)
				default:
					colorizedLine += fmt.Sprintf(`<span class="calc-token-%s">%s</span>`, "literal"+insideUnitTag, token.Value)
				}
			}

			if token.Kind == "bracket" && token.Value == "]" {
				insideUnitTag = ""
			}
		}

		colorizedLines = append(colorizedLines, colorizedLine)
	}

	return strings.Join(colorizedLines, "<br/>")
}
