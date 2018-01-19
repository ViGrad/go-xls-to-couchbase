package main

import (
	"bytes"
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

func cleanString(str string) string {
	str = strings.Replace(str, "'", "", -1)
	str = strings.Replace(str, "é", "e", -1)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "-", "", -1)

	return str
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func readRow(row *xlsx.Row) []string {
	values := []string{}

	for _, cell := range row.Cells {
		values = append(values, cell.String())
	}

	return values
}

func createRequestFile(bucketName string, queries string, num int) {
	str := "UPSERT INTO " + bucketName + " (KEY, VALUE) VALUES \n" + strings.TrimRight(queries, ", \n") + " Returning *;"
	bytes := []byte(str)
	fileName := bucketName + strconv.Itoa(num) + ".query.txt"

	fmt.Printf(fileName)
	fmt.Printf("\n")
	ioutil.WriteFile("outputs/"+fileName, bytes, 0644)
}

func readSheet(sheet *xlsx.Sheet, startingLine int) {
	MAX_QUERIES_PER_FILE := 3000

	keys := []string{}
	values := []string{}
	m := map[string]interface{}{}
	sheetName := cleanString(sheet.Name)
	nbRequests := 1
	var buffer bytes.Buffer

	for rowIndex, row := range sheet.Rows {
		if rowIndex >= startingLine {

			if rowIndex == startingLine {
				keys = readRow(row)
			} else {
				values = readRow(row)

				for i, key := range keys {
					if key != "" && values[i] != "" {
						intValue, err := strconv.Atoi(values[i])

						if err != nil {
							m[key] = values[i]
						} else {
							m[key] = intValue
						}
					}
				}

				jsonString, _ := json.Marshal(m)
				buffer.WriteString("(\"" + sheetName + strconv.Itoa(rowIndex) + "\", " + string(jsonString) + "), \n")

				if rowIndex%MAX_QUERIES_PER_FILE == 0 {
					createRequestFile(sheetName, buffer.String(), nbRequests)
					buffer.Reset()
					nbRequests++
				}
			}
		}
	}

	createRequestFile(sheetName, buffer.String(), nbRequests)
}

func readFile(fileName string, startingLine int, ignored []string) {
	xlFile, error := xlsx.OpenFile(fileName)

	if error == nil {
		for _, sheet := range xlFile.Sheets {
			sheetName := sheet.Name

			if contains(ignored, sheetName) == false {
				readSheet(sheet, startingLine)
			}
		}
	}
	if error != nil {
		fmt.Printf(error.Error())
	}
}

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
