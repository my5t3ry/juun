package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	. "../common"
	. "../config"
)

func main() {
	result := []*HistoryLine{}
	cfg := GetConfig()
	encoded := ""

	if len(os.Args) == 3 {
		qry := os.Args[2]
		if len(qry) == 0 {
			encoded = QueryService("list", os.Args[1], qry)
		} else {
			encoded = QueryService("search", os.Args[1], qry)
		}
	} else if len(os.Args) == 2 {
		encoded = QueryService("list", os.Args[1], "")
	}

	err := json.Unmarshal([]byte(encoded), &result)
	if err == nil && len(result) > 0 {
		var r []string
		for _, curResult := range result {
			if len(curResult.Line) >= cfg.ResultLineLength {
				r = append(r, curResult.Line[0:cfg.ResultLineLength])
			} else {
				r = append(r, curResult.Line[0:len(curResult.Line)])
			}
		}
		fmt.Printf("%s", strings.Join(r, "\n"))
	}
	os.Exit(0)
}
