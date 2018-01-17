package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	flag "github.com/ogier/pflag"
	"github.com/tealeg/xlsx"
)

// flags
var (
	fileName     string
	startingLine string
	ignoring     string
)

func main() {
	flag.Parse()

	// if user does not supply flags, print usage
	// we can clean this up later by putting this into its own function
	if flag.NFlag() == 0 {
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	ignored := strings.Split(ignoring, ",")

	fmt.Printf("Searching file(s): %s\n", fileName)
	fmt.Printf("From line: %s\n", startingLine)
	fmt.Printf("Ignoring these sheets: %s\n", ignored)

	parsedStartingLine, error := strconv.Atoi(startingLine)

	if error != nil {
		fmt.Printf(error.Error())
	}

	readFile(fileName, parsedStartingLine, ignored)

}

func init() {
	flag.StringVarP(&fileName, "fileName", "f", "", "file name")
	flag.StringVarP(&startingLine, "startingLine", "l", "1", "starting line")
	flag.StringVarP(&ignoring, "ignoring", "i", "", "ignoring sheet")
}

func readFile(fileName string, startingLine int, ignored []string) {
	xlFile, error := xlsx.OpenFile(fileName)

	if error == nil {
		for _, sheet := range xlFile.Sheets {
			sheetName := sheet.Name
			fmt.Printf("Sheet name: %s", sheetName)
			fmt.Printf("\n")

			if contains(ignored, sheetName) == false {
				readSheet(sheet, startingLine)
			}
		}
	}
	if error != nil {
		fmt.Printf(error.Error())
	}
}

func readSheet(sheet *xlsx.Sheet, startingLine int) {

	keys := []string{}
	values := []string{}
	m := map[string]interface{}{}
	finalMap := []map[string]interface{}{}

	for rowIndex, row := range sheet.Rows {
		if rowIndex >= startingLine {

			if rowIndex == startingLine {
				keys = readRow(row)
			} else {
				values = readRow(row)

				for i, key := range keys {
					// fmt.Printf("key %s", key)
					// fmt.Printf("\n")
					// fmt.Printf("value %s", values[i])
					// fmt.Printf("\n")
					if key != "" && values[i] != "" {
						intValue, err := strconv.Atoi(values[i])

						if err != nil {
							m[key] = values[i]
						} else {
							m[key] = intValue
						}
					}
				}
				finalMap = append(finalMap, m)

			}
		}
	}

	jsonString, _ := json.Marshal(finalMap)

	request := "UPSERT INTO CouchbaseProject (KEY, VALUE) VALUES (\"" + sheet.Name + "\", " + string(jsonString) + ") RETURNING *"

	ioutil.WriteFile(sheet.Name+".query.txt", []byte(request), 0644)
}

func readRow(row *xlsx.Row) []string {
	values := []string{}

	for _, cell := range row.Cells {
		values = append(values, cell.String())
	}

	return values
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
