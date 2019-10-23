package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main () {
	file, err := os.Open("/Users/carterharrison/Downloads/ca.txt")

	if err != nil {
		panic(err)
	}

	r := NewReader(file)


	m := make(map[string]int)
	i := 0
	for r.HasNext() {
		i++
		//fmt.Println("---------------------------")
		sheet := r.Next()

		for _, v := range sheet.OldSurveyControl {
			m[v.Item]++
		}
	}

	res, err := json.Marshal(m)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(res))



}

