package main

import (
	"fmt"
	"strconv"
	"strings"
)

// spec at https://www.ngs.noaa.gov/DATASHEET/dsdata.pdf

var (
	basicMetadataSection = 0
	currentSurveyControlSection = 1
	accuracySection = 2
	dataDeterminationMethodologySection = 3
	projectionsSection = 4
	azimuthMarksSection = 5
	supersededSurveyControlSection = 6
	monumentationSection = 7
	historySection = 8
	descriptionAndRecoverySection = 9
	headers = make(map[string]string)
)

type Page struct {
	CurrentSheet DataSheet
	CurrentSection int
	LineNum int
	CurrentBuffer string
}

func NewPage () Page {
	page := Page{}
	page.Reset()
	return page
}

func (page *Page) Reset () {
	page.CurrentSheet = DataSheet{}
	page.CurrentSheet.Init()
	page.CurrentSection = basicMetadataSection
	page.CurrentBuffer = ""
	page.LineNum = 0
}

func (page *Page) AddLine (line string) {
	if page.LineNum == 0 {
		page.ReadId(line)
	}

	defer func () {
		page.LineNum++
	}()


	// we give the current line to the correct parser

	if page.CurrentSection == basicMetadataSection {
		page.BasicMetadataSection(line)
		return
	}

	if page.CurrentSection == currentSurveyControlSection {
		page.CurrentSurveyControlSection(line)
		return
	}

	if page.CurrentSection == accuracySection {
		page.AccuracySection(line)
		return
	}

	if page.CurrentSection == dataDeterminationMethodologySection {
		page.DataDeterminationMethodologySection(line)
		return
	}

	if page.CurrentSection == projectionsSection {
		page.ProjectionsSection(line)
		return
	}

	if page.CurrentSection == azimuthMarksSection {
		page.AzimuthMarksSection(line)
		return
	}

	if page.CurrentSection == supersededSurveyControlSection {
		page.SupersededSurveyControlSection(line)
		return
	}

	if page.CurrentSection == monumentationSection {
		page.MonumentationSection(line)
		return
	}
}

func (page *Page) Make () DataSheet {
	defer page.Reset()

	return page.CurrentSheet
}


// gets the first 1:7 chars of the line
func (page *Page) ReadId (line string) {
	if len(line) <= 8 {
		return
	}

	page.CurrentSheet.Id = line[1:7]
}

func (page *Page) BasicMetadataSection (line string) {

	// if there is a new blank line, move to the next section
	if len(line) == 7 {
		page.CurrentSection = currentSurveyControlSection
		return
	}

	// the line is not long enough to hold basic metadata
	if len(line) <= 10 {
		return
	}

	// we separate the key and value by -, and remove whitespace
	line = line[9:]
	onKey := true
	whiteSpace := ""
	key := ""
	value := ""

	for _, l := range line {
		if onKey {
			if string(l) == "-" {
				whiteSpace = ""
				onKey = false
			} else if string(l) != " " {
				key = key + whiteSpace + string(l)
				whiteSpace = ""
			} else {
				whiteSpace = whiteSpace + string(l)
			}
		} else {
			if string(l) != " " || value != "" {
				value = value + string(l)
			}
		}
	}

	page.CurrentSheet.BasicMetadata[key] = value
}

func (page *Page) CurrentSurveyControlSection (line string) {

	// if there is a line break, we have moved on to the next section
	if len(line) == 7 {
		page.CurrentSection = accuracySection
		return
	}

	// the line is not long enough to hold survey data, lets see what else it could be
	if len(line) <= 70 {

		// there could be an accuracy line here, sometimes they are merged
		if len(line) > 26 {
			// this is the accuracy section
			if keyIsAccuracy(line) {
				page.CurrentSection = accuracySection
				page.AccuracySection(line)
				return
			}
		}

		return
	}

	// if the start of the line has a space or _, it is the header
	if string(line[10]) == " " || string(line[10]) == "_" {
		return
	}

	// if there is a star, it is a current survey
	isNew := string(line[7]) == "*"

	if isNew {
		survey := Survey{
			Item:  trimWhiteSpace(line[9:30]),
			Value: trimWhiteSpace(line[31:70]),
			By:    trimWhiteSpace(line[71:]),
		}

		page.CurrentSheet.NewSurveyControl = append(page.CurrentSheet.NewSurveyControl, survey)
		return
	}

	survey := Survey{
		Item:  trimWhiteSpace(line[9:25]),
		Value: trimWhiteSpace(line[26:70]),
		By:    trimWhiteSpace(line[71:]),
	}

	page.CurrentSheet.OldSurveyControl = append(page.CurrentSheet.OldSurveyControl, survey)

}

func (page *Page) AccuracySection (line string) {

	// make sure the line is long enough to hold the data
	if len(line) < 26 {
		return
	}

	if string(line[7]) == "." {
		page.CurrentSection = dataDeterminationMethodologySection
		page.DataDeterminationMethodologySection(line)
		return
	}

	// check for the network key
	isNetwork := line[9:16] == "NETWORK"

	// this is a network line
	if isNetwork {
		// make and add the network line to the array
		data := networkLine(line[17:])
		page.CurrentSheet.Accuracy.Network = append(page.CurrentSheet.Accuracy.Network, data)
		return
	}

	if !keyIsAccuracy(line) {
		return
	}

	// we are on a line with accuracy data to be found
	key := trimWhiteSpace(line[9:25])
	valParts := strings.Split(line, "-")

	if len(valParts) != 2 {
		return
	}

	if len(valParts[1]) < 2 {
		return
	}

	val := trimWhiteSpace(valParts[1][1:])

	switch key {
	case "HORZ ORDER": page.CurrentSheet.Accuracy.HorzOrder = append(page.CurrentSheet.Accuracy.HorzOrder, val)
	case "ELLP ORDER": page.CurrentSheet.Accuracy.EllpOrder = append(page.CurrentSheet.Accuracy.EllpOrder, val)
	case "VERT ORDER": page.CurrentSheet.Accuracy.VertOrder = append(page.CurrentSheet.Accuracy.VertOrder, val)
	}
}

func (page *Page) DataDeterminationMethodologySection (line string) {
	if len(line) > 10 {
		if  string(line[7]) == ";" || line[7:9] == ". " {
			page.CurrentSection = projectionsSection
			page.ProjectionsSection(line)
			return
		}
	}


	// end of paragraph
	if len(line) == 7 {
		page.CurrentSheet.DeterminationMethodology = append(page.CurrentSheet.DeterminationMethodology, page.CurrentBuffer)
		page.CurrentBuffer = ""
	} else if len(line) > 9 {
		// regular sentence

		if page.CurrentBuffer == "" {
			page.CurrentBuffer = page.CurrentBuffer + line[8:]
		} else {
			page.CurrentBuffer = page.CurrentBuffer + " " + line[8:]
		}
	}

}

func (page *Page) ProjectionsSection (line string) {
	// make sure line has content for this section
	if len(line) < 20 {
		return
	}

	if len(line) == 58 {
		// if header survey line is present, move to next section
		if line[33:] == "SUPERSEDED SURVEY CONTROL" {
			page.CurrentSection = azimuthMarksSection
			page.AzimuthMarksSection(line)
			return
		}
	}

	// also move  to next section if data starts right away
	if string(line[7]) == ":" || string(line[7]) == "|" {
		page.CurrentSection = azimuthMarksSection
		page.AzimuthMarksSection(line)
		return
	}

	// check for State Plane Coordinates
	if string(line[7]) == ";" {
		page.statePlaneCoordinates(line)
	}

	page.checkSpatialAddress(line)
}



func (page *Page) AzimuthMarksSection (line string) {

	// check for header of next section
	if len(line) > 40 {
		if line[33:] == "SUPERSEDED SURVEY CONTROL" {
			page.CurrentSection = supersededSurveyControlSection
			page.SupersededSurveyControlSection(line)
			return
		}
	}

	// we have to check for spatial address here because it could be out of order
	if page.checkSpatialAddress(line) {
		return
	}

	// make sure line has content
	if len(line) < 68 {
		return
	}
	// check for Primary Azimuth Mark
	if string(line[7]) == ":" && string(line[8]) != " " {
		name := trimWhiteSpace(line[25:65])
		nums := getNumbersFromString(line[66:])

		if len(nums) != 3 {
			return
		}

		mark := PrimaryAzimuthMark{
			Mark:  name,
			GridAz: nums,
		}

		page.CurrentSheet.PrimaryAzimuthMarks = append(page.CurrentSheet.PrimaryAzimuthMarks, mark)
	}

	// check for reference object table row
	if string(line[7]) == "|" && string(line[8]) == " " && line[9:12] != "PID" && string(line[9]) != " " {
		pid := line[9:15]
		name := trimWhiteSpace(line[16:52])
		distance := trimWhiteSpace(line[52:67])
		geodAz := trimWhiteSpace(line[67:76])

		reference := ReferenceObject{
			Pid: pid,
			Ref: name,
			Distance: distance,
			GeodAz: geodAz,
		}

		page.CurrentSheet.ReferenceObjects = append(page.CurrentSheet.ReferenceObjects, reference)
	}

}

func (page *Page) SupersededSurveyControlSection (line string) {
	if len(line) < 10 {
		return
	}

	// check for end of section to move onto next section
	if line[7:] == ".See file dsdata.pdf to determine how the superseded data were derived." ||
		line[7:] == ".No superseded survey control is available for this station." {
		page.CurrentSection = monumentationSection
		page.MonumentationSection(line)
		return
	}

	if !(line[7:9] == "  " && line[10:11] != " ") {
		return
	}

	// check for Latitude and Longitude

	if string(line[21]) == "-" {
		//name := line[9:21]
		//pos := line[24:41]
		//order := line[76]

		fmt.Println(line)
	}

	// check for Ellipsoid Height

	// check for Orthometric Height


	//fmt.Println(line)

}

func (page *Page) MonumentationSection (line string) {

}

// see if line is grid spatial address
func (page *Page) checkSpatialAddress (line string) bool {
	if len(line) > 45 {
		if line[8:42] == "U.S. NATIONAL GRID SPATIAL ADDRESS" {
			address := line[44:]
			page.CurrentSheet.SpatialAddress = address
			return true
		}
	}

	return false
}

func (page *Page) statePlaneCoordinates (line string) {
	// check for header
	if string(line[8]) == " " {
		page.CurrentBuffer = line[28:]
	} else if page.CurrentBuffer == "North         East     Units Scale Factor Converg." {
		nums := getNumbersFromString(line[19:])

		// todo handle * number
		if len(nums) != 6 {
			return
		}

		unit := getProjectionUnit(line)

		coords := StatePlaneCoordinates{
			North: nums[0],
			East: nums[1],
			Units: unit,
			Scale: nums[2],
			Factor: nums[3],
			Converg: []float64{nums[4], nums[5]},
			Estimated: "",
		}

		page.CurrentSheet.StatePlaneCoordinates = append(page.CurrentSheet.StatePlaneCoordinates, coords)
	} else if page.CurrentBuffer == "North         East    Units  Estimated Accuracy" {
		// then parse "EW5045;SPC CA 5     -   615,560.    1,886,710.      MT  (+/- 180 meters Scaled)"

		unit := getProjectionUnit(line)
		nums := getNumbersFromString(line[19:50])

		if len(nums) != 2 {
			return
		}

		coords := StatePlaneCoordinates{
			North: nums[0],
			East: nums[1],
			Units: unit,
			Scale: 0,
			Factor: 0,
			Converg: []float64{},
			Estimated: line[57:],
		}

		page.CurrentSheet.StatePlaneCoordinates = append(page.CurrentSheet.StatePlaneCoordinates, coords)
	}
}


// gets safe and removes whitespace
func getProjectionUnit (line string) string {
	if len(line) < 56 {
		return ""
	}

	unit := line[52:55]

	if string(unit[0]) == " " {
		unit = unit[1:]
	}

	return unit
}

// network line data without prefix, strip " EW4726  NETWORK"
func networkLine (line string) NetworkAccuracy {
	ntw := NetworkAccuracy{}
	nums := getNumbersFromString(line)

	// add numbers based on order
	if len(nums) == 6 {
		ntw.Horiz = nums[0]
		ntw.Ellip = nums[1]
		ntw.SDN = nums[2]
		ntw.SDE = nums[3]
		ntw.SDH = nums[4]
		ntw.CorrNE = nums[5]
	}

	return ntw
}

func getNumbersFromString (s string) []float64 {
	s = s + " "
	nums := make([]float64, 0)
	currentNum := ""

	for _, c := range s {
		cs := string(c)
		isWhite := cs == " " || cs == "\t"

		// if we have reached the end of a number, make it
		if isWhite && currentNum != "" {
			n, err := strconv.ParseFloat(currentNum, 64)

			if err == nil {
				nums = append(nums, n)
			}

			// reset num for next
			currentNum = ""
		} else if isNumChar(cs) {
			currentNum = currentNum + cs
		}
	}

	return nums
}

func isNumChar (c string) bool {
	return strings.Contains("1234567890.-", c)
}

func keyIsAccuracy (line string) bool {
	// we make sure line is long enough
	if len(line) < 25 {
		return false
	}

	// check to see if key is one the the accuracy keys
	key := trimWhiteSpace(line[9:25])
	return key == "HORZ ORDER" || key == "ELLP ORDER" || key == "VERT ORDER" || key == "NETWORK"
}

// gets rid of the whitespace on the end
func trimWhiteSpace (s string) string {
	n := ""
	white := ""

	for _, l := range s {
		if string(l) != " " {
			n = n + white + string(l)
			white = ""
		} else if n != "" {
			white = white + " "
		}
	}

	return n
}