package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

	if command == "server" {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()

		r.POST("/execute", func(c *gin.Context) {
			raw_body, err := ioutil.ReadAll(c.Request.Body)

			if err != nil {
				c.JSON(500, gin.H{
					"error": err.Error(),
				})

				return
			}

			fmt.Println(string(raw_body))
			graph := ParseCode(string(raw_body))
			graph.Execute()
			c.String(200, graph.ExecutionResult())
		})
		r.POST("/colorize", func(c *gin.Context) {
			raw_body, err := ioutil.ReadAll(c.Request.Body)

			if err != nil {
				c.JSON(500, gin.H{
					"error": err.Error(),
				})

				return
			}

			graph := ExecutionGraph{SourceCode: string(raw_body)}
			graph.Tokenize(true)

			c.String(200, graph.ColorizedHTML())
		})
		r.POST("/currencies", func(c *gin.Context) {
			var conversionRates struct {
				USD float64
				GBP float64
				CNY float64
				CAD float64
			}
			if err := c.ShouldBindJSON(&conversionRates); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			usdUnit := UnitTable["usd"]
			usdUnit.ConversionFactor = 1 / conversionRates.USD
			UnitTable["usd"] = usdUnit

			gbpUnit := UnitTable["gbp"]
			gbpUnit.ConversionFactor = 1 / conversionRates.GBP
			UnitTable["gbp"] = gbpUnit

			cnyUnit := UnitTable["cny"]
			cnyUnit.ConversionFactor = 1 / conversionRates.CNY
			UnitTable["cny"] = cnyUnit

			cadUnit := UnitTable["cad"]
			cadUnit.ConversionFactor = 1 / conversionRates.CAD
			UnitTable["cad"] = cadUnit

			c.JSON(200, gin.H{"ok": true})
		})

		r.Run(":7894")
	} else {

		// if path is passed read file from path
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

			fmt.Println(graph.ExecutionResult())
		} else if command == "colorize" {
			graph := ExecutionGraph{SourceCode: sourceCode}
			graph.Tokenize(true)
			fmt.Println(graph.ColorizedHTML())
		}
	}
}
