package main

import (
	"fmt"
	"os"
)

func main () {
	file, err := os.Open(os.Args[1])

	if err != nil {
		panic(err)
	}

	r := NewReader(file)
	//markers := make(map[string]int)

	for r.HasNext() {
		sheet := r.Next()

		nums := getNumbersFromString(sheet.NewSurveyControl[0].Value)
		lat := DegreesMinutesSeconds(nums[0], nums[1], nums[2])
		lng := DegreesMinutesSeconds(nums[3], nums[4], nums[5]) * - 1
		name := sheet.BasicMetadata["DESIGNATION"]
		//fmt.Println(sheet.Monumentation["_MARKER"] )
		//markers[sheet.Monumentation["_MARKER"]]++
		//
		if sheet.Monumentation["_MARKER"] == "" {
			fmt.Println(name, lat, lng, sheet.Monumentation["_MARKER"])

		}

	}

	//v, _ := json.MarshalIndent(markers, "", "      ")
	//fmt.Println(string((v)))
}

func DegreesMinutesSeconds (deg float64, min float64, sec float64) float64 {
	return deg + (min / 60) + (sec / 3600)
}
