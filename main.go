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

	for r.HasNext() {
		sheet := r.Next()

		//fmt.Println(sheet.BasicMetadata["DESIGNATION"])
		//fmt.Println(sheet.BasicMetadata["DESIGNATION"] + "  ____   " +sheet.NewSurveyControl[0].Value)

		nums := getNumbersFromString(sheet.NewSurveyControl[0].Value)
		lat := DegreesMinutesSeconds(nums[0], nums[1], nums[2])
		lng := DegreesMinutesSeconds(nums[3], nums[4], nums[5]) * - 1

		fmt.Println(sheet.BasicMetadata["DESIGNATION"], lat, lng)


		//res, err := json.Marshal(sheet)
		//
		//if err != nil {
		//	panic(err)
		//}
		//
		//fmt.Println(string(res))

	}
}

func DegreesMinutesSeconds (deg float64, min float64, sec float64) float64 {
	return deg + (min / 60) + (sec / 3600)
}