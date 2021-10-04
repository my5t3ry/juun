package main

import (
	"encoding/json"
	"fmt"
	. "github.com/jackdoe/juun/common"
	"os"
)

func main() {
	result := []*HistoryLine{}

	if len(os.Args) == 3 {
		qry := os.Args[2]
		encoded := QueryService("search", os.Args[1], qry)
		err := json.Unmarshal([]byte(encoded), &result)
		if err == nil && len(result) > 0 {
			r := ""
			maxLineLength := 80
			if len(result[0].Line) >= maxLineLength {
				r = result[0].Line[0:maxLineLength]
			} else {
				r = result[0].Line[0:len(result[0].Line)]
			}
			fmt.Printf("%s", r)
		}
	}
	os.Exit(0)
}
