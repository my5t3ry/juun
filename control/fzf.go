package main

import (
	"encoding/json"
	"fmt"
	. "github.com/jackdoe/juun/common"
	"os"
	"strings"
)

func main() {
	result := []*HistoryLine{}

	if len(os.Args) == 3 {
		qry := os.Args[2]
		encoded := QueryService("search", os.Args[1], qry)
		err := json.Unmarshal([]byte(encoded), &result)
		if err == nil && len(result) > 0 {
			var r []string
			maxLineLength := 80
			for _, curResult := range result {
				if len(curResult.Line) >= maxLineLength {
					r = append(r, curResult.Line[0:maxLineLength])
				} else {
					r = append(r, curResult.Line[0:len(curResult.Line)])
				}
			}
			fmt.Printf("%s", strings.Join(r, "\n"))
		}
	}
	os.Exit(0)
}
