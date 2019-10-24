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

		fmt.Println("-----")
		for _, v := range sheet.OldSurveyControl {
			m[v.Item]++
		}


		//if len(sheet.Accuracy.Network) > 0 {
			res, err := json.Marshal(sheet)

			if err != nil {
				panic(err)
			}

			fmt.Println(string(res))
		//}



	}




}

