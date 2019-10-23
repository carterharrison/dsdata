package main

type DataSheet struct {
	Id string `json:"id"`
	BasicMetadata map[string]string `json:"metadata"`
	NewSurveyControl []Survey `json:"newSurveys"`
	OldSurveyControl []Survey `json:"oldSurveys"`
}

type Survey struct {
	Item string `json:"item"`
	Value string	`json:"value"`
	By string	`json:"by"`
}