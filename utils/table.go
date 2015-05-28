package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
)

func PrintJsonArray(display_mode string, fields []string, input interface{}) {
	if display_mode == "" {
		PrintPlain(fields, input)
	} else if display_mode == "table" {
		PrintTable(fields, input)
	} else {
		PrintJsonObject(display_mode, input)
	}
}

func PrintJsonObject(display_mode string, input interface{}) {
	var str []byte
	var err error
	if display_mode == "json" {
		str, err = json.Marshal(input)
	} else if display_mode == "" || display_mode == "pretty-json" {
		str, err = json.MarshalIndent(input, "", "  ")
	} else {
		fmt.Fprintln(os.Stderr, display_mode, "mode is not supported")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}
	fmt.Println(string(bytes.Trim(str, `"`)))
}

func PrintPlain(fields []string, input interface{}) {
	array, ok := input.([]interface{})
	if !ok {
		fmt.Fprintln(os.Stderr, "table print must be array")
		os.Exit(-1)
	}
	fields_width := map[string]int{}
	for _, field := range fields {
		fields_width[field] = getFieldMaxWidth(field, array, false) + 2
	}
	printTableBody(fields, fields_width, array, false)
}

func PrintTable(fields []string, input interface{}) {
	array, ok := input.([]interface{})
	if !ok {
		fmt.Fprintln(os.Stderr, "table print must be array, but it is", reflect.ValueOf(input).Kind())
		os.Exit(-1)
	}
	fields_width := map[string]int{}
	for _, field := range fields {
		fields_width[field] = getFieldMaxWidth(field, array, true) + 2
	}
	printTableSeparator(fields, fields_width)
	printTableHeader(fields, fields_width)
	printTableSeparator(fields, fields_width)
	printTableBody(fields, fields_width, array, true)
	printTableSeparator(fields, fields_width)
}

func getFieldMaxWidth(field string, array []interface{}, include_header bool) int {
	var max_width int
	if include_header {
		max_width = len(field)
	} else {
		max_width = 0
	}
	fields := []string{field}
	for _, item := range array {
		m, err := ConvInterface2StringMap(fields, item)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(-1)
		}
		width := len(m[field])
		if width > max_width {
			max_width = width
		}
	}
	return max_width
}

func printTableHeader(fields []string, fields_width map[string]int) {
	for _, field := range fields {
		width := fields_width[field]
		fmt.Print("| " + field)
		printChar(" ", width-len(field)-1)
	}
	fmt.Println("|")
}

func printTableSeparator(fields []string, fields_width map[string]int) {
	for _, field := range fields {
		width := fields_width[field]
		fmt.Print("+")
		printChar("-", width)
	}
	fmt.Println("+")
}

func printTableBody(fields []string, fields_width map[string]int, array []interface{}, show_delimiter bool) {
	for _, item := range array {
		if reflect.ValueOf(item).IsNil() {
			printTableSeparator(fields, fields_width)
			continue
		}
		m, err := ConvInterface2StringMap(fields, item)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(-1)
		}
		for _, field := range fields {
			width := fields_width[field]
			if show_delimiter {
				fmt.Print("| ")
			}
			fmt.Print(m[field])
			printChar(" ", width-len(m[field])-1)
		}
		if show_delimiter {
			fmt.Print("|")
		}
		fmt.Println()
	}
}

func printChar(s string, width int) {
	for i := 0; i < width; i++ {
		fmt.Print(s)
	}
}

func ospSysField(field string) string {
	return "(" + field + ")"
}

func ConvInterface2StringMap(fields []string, in interface{}) (map[string]string, error) {
	out := map[string]string{}

	bytes, _ := json.Marshal(in)
	var m map[string]interface{}
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, errors.New("convert json object to map failed")
	}

	for _, field := range fields {
		out[field] = fmt.Sprint(m[field])
	}
	return out, nil
}

func FlattenCustomData(input interface{}, columns []string, customDataKey string) (interface{}, []string) {
	outputArr := make([]interface{}, 0)
	outputCol := make([]string, 0)
	for _, column := range columns {
		outputCol = append(outputCol, ospSysField(column))
	}

	inputArr, ok := input.([]interface{})
	if !ok {
		fmt.Fprintln(os.Stderr, "input: interface to array failed")
		os.Exit(-1)
	}
	firstRow := true
	for _, item := range inputArr {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			fmt.Fprintln(os.Stderr, "custom data: interface to map failed")
			os.Exit(-1)
		}
		for _, column := range columns {
			itemMap[ospSysField(column)] = itemMap[column]
		}
		custom, exists := itemMap[customDataKey]
		if !exists {
			fmt.Fprintln(os.Stderr, "custom data: not exist")
			os.Exit(-1)
		}
		customMap, _ := custom.(map[string]interface{})
		for key, value := range customMap {
			if firstRow {
				outputCol = append(outputCol, key)
			}
			valueStr, _ := value.(string)
			itemMap[key] = valueStr
		}
		outputArr = append(outputArr, item)
		firstRow = false
	}
	sort.Strings(outputCol)
	return outputArr, outputCol
}
