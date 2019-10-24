package main

type DataSheet struct {
	// the pid of the datasheet, located on the right side
	Id string `json:"id"`

	// the key value pairs at the top of the sheet
	BasicMetadata map[string]string `json:"metadata"`

	// surveys that are on going
	NewSurveyControl []Survey `json:"newSurveys"`

	// past surveys
	OldSurveyControl []Survey `json:"oldSurveys"`

	// the accuracy section
	Accuracy Accuracy `json:"accuracy"`

	DeterminationMethodology []string `json:"determinationMethodology"`

	StatePlaneCoordinates []StatePlaneCoordinates `json:"statePlaneCoordinates"`

	SpatialAddress string `json:"spatialAddress"`

	PrimaryAzimuthMarks []PrimaryAzimuthMark `json:"primaryAzimuthMark"`

	ReferenceObjects []ReferenceObject `json:"referenceObjects"`
}

type PrimaryAzimuthMark struct {
	Mark string `json:"mark"`
	GridAz []float64 `json:"gridAz"`

}

type ReferenceObject struct {
	Pid string `json:"pid"`

	Ref string `json:"ref"`

	Distance string `json:"distance"`

	GeodAz string `json:"geodAz"`
}

type Accuracy struct {
	// key as HORZ ORDER
	HorzOrder []string `json:"horzOrder"`

	// key as ELLP ORDER
	EllpOrder []string `json:"ellpOrder"`

	// key as VERT ORDER
	VertOrder []string `json:"vertOrder"`

	// key as NETWORK
	Network []NetworkAccuracy `json:"network"`
}

type StatePlaneCoordinates struct {
	North float64	`json:"north"`
	East float64	`json:"east"`
	Units string	`json:"units"`
	Scale float64	`json:"scale"`
	Factor float64	`json:"factor"`
	Converg []float64	`json:"converg"`
	Estimated string	`json:"estimated"`
}

type NetworkAccuracy struct {
	Horiz float64 `json:"horiz"`
	Ellip float64 `json:"ellip"`
	SDN float64 `json:"SDN"`
	SDE float64 `json:"SDE"`
	SDH float64 `json:"SDH"`
	CorrNE float64 `json:"corrNE"`
}

type Survey struct {
	// the type of survey
	Item string `json:"item"`

	// the value of the survey
	Value string	`json:"value"`

	// how the survey was collected
	By string	`json:"by"`
}

func (datasheet *DataSheet) Init () {
	datasheet.BasicMetadata = make(map[string]string)
	datasheet.NewSurveyControl = make([]Survey, 0)
	datasheet.OldSurveyControl = make([]Survey, 0)
	datasheet.DeterminationMethodology = make([]string, 0)
	datasheet.StatePlaneCoordinates = make([]StatePlaneCoordinates, 0)
	datasheet.PrimaryAzimuthMarks = make([]PrimaryAzimuthMark, 0)
	datasheet.ReferenceObjects = make([]ReferenceObject, 0)


	acc := Accuracy{}
	acc.Init()
	datasheet.Accuracy = acc
}

func (accuracy *Accuracy) Init () {
	accuracy.HorzOrder = make([]string, 0)
	accuracy.EllpOrder = make([]string, 0)
	accuracy.VertOrder = make([]string, 0)
	accuracy.Network = make([]NetworkAccuracy, 0)
}