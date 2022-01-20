package dataConversion

import (
	"encoding/json"
	"fmt"
	"strings"
)

var xmlBuilder strings.Builder

func json2Xml(jsonString string) string {
	log.Infof("Trying to convert json to xml now...")
	xmlBuilder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" ?>")
	xmlBuilder.WriteString("\n<root>")

	// Creating the maps for JSON
	m := map[string]interface{}{}

	// Parsing/Unmarshalling JSON encoding/json
	err := json.Unmarshal([]byte(jsonString), &m)
	if err != nil {
		panic(err)
	}

	parseMap(m)

	xmlBuilder.WriteString("\n</root>")

	return xmlBuilder.String()
}

func parseMap(aMap map[string]interface{}) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			fmt.Println(key)
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			fmt.Println(key)
			var value = fmt.Sprintf("\n  <%v>", key)
			xmlBuilder.WriteString(value)

			parseArray(val.([]interface{}))

			value = fmt.Sprintf("</%v>", key)
			xmlBuilder.WriteString(value)
		default:
			var value = fmt.Sprintf("\n  <%v>%v</%v>", key, concreteVal, key)
			xmlBuilder.WriteString(value)
		}
	}
}

func parseArray(anArray []interface{}) {
	for _, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			xmlBuilder.WriteString("  ")
			parseMap(val.(map[string]interface{}))
			xmlBuilder.WriteString("\n  ")
		case []interface{}:
			parseArray(val.([]interface{}))
			xmlBuilder.WriteString("\n")
		default:
			var value = fmt.Sprintf("%v", concreteVal)
			xmlBuilder.WriteString(value)
		}
	}
}
