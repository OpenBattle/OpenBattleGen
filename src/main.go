package main

import (
	"OpenBattleGen"
	"flag"
	"fmt"
	"os"
)

var schema = flag.String("input", "", "Message definitions")
var lang = flag.String("lang", "c", "Output language")
var output = flag.String("output", "", "Output file name")

func main() {
	flag.Parse()
	if *schema == "" || *output == "" {
		flag.PrintDefaults()
		return
	}

	exp := OpenBattleGen.GetExporter(*lang)
	if exp == nil {
		fmt.Println("Supported exporters:")
		for _, e := range OpenBattleGen.GetExporters() {
			fmt.Println(e.Languages())
		}
		return
	}

	var defs = new(OpenBattleGen.Definitions)
	if err := defs.Parse(*schema); err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Create(*output)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	if err := exp.Export(defs, file); err != nil {
		fmt.Println(err)
		return
	}
}
