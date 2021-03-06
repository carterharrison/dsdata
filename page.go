package main

import (
	"strconv"
	"strings"
)

// spec at https://www.ngs.noaa.gov/DATASHEET/dsdata.pdf

// sections
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

	networkKey = "NETWORK"
	horzOrderKey = "HORZ ORDER"
	ellpOrderKey = "ELLP ORDER"
	vertOrderKey = "VERT ORDER"
	surveyControlHeader = "SUPERSEDED SURVEY CONTROL"
	pidKey = "PID"
	surveyControlEndA = ".See file dsdata.pdf to determine how the superseded data were derived."
	surveyControlEndB = ".No superseded survey control is available for this station."
	ellipHKey = "ELLIP H"
	orthometricHeightKey = "NAVD"
	historyKey = "HISTORY"
	stationDescriptionHeader = "STATION DESCRIPTION"
	stationRevoveryHeader = "STATION RECOVERY"
	historyHeader = "Date     Condition        Report By"
	accuracyHeader = "North         East    Units  Estimated Accuracy"
	statePlaneHeader = "North         East     Units Scale Factor Converg."
	spatialAddressKey = "U.S. NATIONAL GRID SPATIAL ADDRESS"
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

	// we give the current line to the correct parser
	switch page.CurrentSection {
	case basicMetadataSection: page.BasicMetadataSection(line)
	case currentSurveyControlSection: page.CurrentSurveyControlSection(line)
	case accuracySection: page.AccuracySection(line)
	case dataDeterminationMethodologySection: page.DataDeterminationMethodologySection(line)
	case projectionsSection: page.ProjectionsSection(line)
	case azimuthMarksSection: page.AzimuthMarksSection(line)
	case supersededSurveyControlSection: page.SupersededSurveyControlSection(line)
	case monumentationSection: page.MonumentationSection(line)
	case historySection: page.HistorySection(line)
	case descriptionAndRecoverySection: page.DescriptionAndRecoverySection(line)
	}

	page.LineNum++
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
	isNetwork := line[9:16] == networkKey

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
	case horzOrderKey: page.CurrentSheet.Accuracy.HorzOrder = append(page.CurrentSheet.Accuracy.HorzOrder, val)
	case ellpOrderKey: page.CurrentSheet.Accuracy.EllpOrder = append(page.CurrentSheet.Accuracy.EllpOrder, val)
	case vertOrderKey: page.CurrentSheet.Accuracy.VertOrder = append(page.CurrentSheet.Accuracy.VertOrder, val)
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
		if line[33:] == surveyControlHeader {
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
		if line[33:] == surveyControlHeader {
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
	if string(line[7]) == "|" && string(line[8]) == " " && line[9:12] != pidKey && string(line[9]) != " " {
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
	if line[7:] == surveyControlEndA || line[7:] == surveyControlEndB {
		page.CurrentSection = monumentationSection
		page.MonumentationSection(line)
		return
	}

	if !(line[7:9] == "  " && line[10:11] != " ") {
		return
	}

	// make sure line is long enough to get the correct information
	if len(line) < 77 {
		return
	}

	// check for Latitude and Longitude
	if string(line[21]) == "-" {
		name := line[9:21]
		pos := line[24:41]
		body := line[45:75]
		order := string(line[76])

		latLng := SurveyLatitudeLongitude{
			Name: name,
			Pos: pos,
			Body: body,
			Order: order,
		}

		page.CurrentSheet.SurveyLatitudeLongitudes = append(page.CurrentSheet.SurveyLatitudeLongitudes, latLng)
		return
	}


	// make sure line is log enough for ellip
	if len(line) < 79 {
		return
	}

	// check for Ellipsoid Height
	if line[9:16] == ellipHKey {
		date := line[18:26]
		height := getNumbersFromString(line[30:36])

		// make sure there is exactly one height number in the section we expect
		if len(height) != 1 {
			return
		}

		unit := getInnerParValue(line[37:45])
		method := line[64:78]
		order := line[78:79]

		ellipH := SurveyEllipsoidHeight{
			Date: date,
			Height: height[0],
			Unit: unit,
			Method: method,
			Order: order,
		}

		page.CurrentSheet.SurveyEllipsoidHeights = append(page.CurrentSheet.SurveyEllipsoidHeights, ellipH)

		return
	}

	// check for Orthometric Height
	if line[9:13] == orthometricHeightKey {
		date := trimWhiteSpace(line[18:26])
		numbers := getNumbersFromString(line[29:37])

		// we expect there to be one height measurement used
		if len(numbers) != 1 {
			return
		}

		unit := getInnerParValue(line[37:43])

		// todo figure out what the model is because sometimes it is feet
		model := line[43:63]
		if model[17:18] == "(" {

		}

		method := trimWhiteSpace(line[64:76])
		orders := getNumbersFromString(line[76:79])

		navdH := SurveyOrthometricHeight{
			Date: date,
			Height: numbers[0],
			Method: method,
			Order: orders,
			Unit: unit,
		}

		page.CurrentSheet.SurveyOrthometricHeights = append(page.CurrentSheet.SurveyOrthometricHeights, navdH)

		return
	}

}

func (page *Page) MonumentationSection (line string) {
	if len(line) < 8 {
		return
	}

	// check for history section and move on
	if len(line) > 17 {
		if line[9:16] == historyKey {
			page.CurrentSection = historySection
			page.HistorySection(line)
			return
		}
	}

	// grab everything after the pid
	content := line[7:]

	if len(content) < 1 {
		return
	}

	if content[0:1] == "." {
		return
	}

	// extract key values after the pid, separated by :
	parts := strings.Split(content, ":")

	if len(parts) != 2 {
		return
	}

	// todo looks at + for appending to it
	page.CurrentSheet.Monumentation[parts[0]] = trimWhiteSpace(parts[1])
}

func (page *Page) HistorySection (line string) {

	// if line is long enough, check for next sections header
	if len(line) > 51 {

		// this is the header of the next section
		if line[33:52] == stationDescriptionHeader {
			page.CurrentSection = descriptionAndRecoverySection
			page.DescriptionAndRecoverySection(line)
			return
		}
	}

	if len(line) < 24 {
		return
	}

	// make sure this is an history line
	if line[9:16] != historyKey {
		return
	}

	// ignore the header line
	if line[23:] == historyHeader {
		return
	}

	// we are on the history row
	date := trimWhiteSpace(line[23:31])
	condition := trimWhiteSpace(line[32:min(49, len(line) - 1)])
	by := ""

	if len(line) >= 50 {
		by = line[49:]
	}

	history := History{
		Date: date,
		Condition: condition,
		By: by,
	}

	page.CurrentSheet.History = append(page.CurrentSheet.History, history)
}

// last section, still alive
func (page *Page) DescriptionAndRecoverySection (line string)  {

	// if line is long enough for heading
	if len(line) > 33 {
		content := line[33:]
		// if header is desc
		if content == stationDescriptionHeader {
			page.CurrentBuffer = stationDescriptionHeader

			desc := StationDescription{
				Description: "",
			}

			page.CurrentSheet.StationDescription = append(page.CurrentSheet.StationDescription, desc)
			return
		}

		if len(content) > 17 {
			// if header is recovery
			if content[:16] == stationRevoveryHeader {
				page.CurrentBuffer = stationRevoveryHeader

				date := getInnerParValue(content)
				
				rec := StationRecovery{
					Date: date,
					Description: "",
				}

				page.CurrentSheet.StationRecoveries = append(page.CurrentSheet.StationRecoveries, rec)
				return
			}
		}
	}

	// make sure line can have some content
	if len(line) < 10 {
		return
	}

	// make sure line has text
	if line[7:8] != "'" {
		return
	}

	if page.CurrentBuffer == stationDescriptionHeader && len(page.CurrentSheet.StationDescription) > 0 {
		lastIndex := len(page.CurrentSheet.StationDescription) - 1
		lastDesc := page.CurrentSheet.StationDescription[lastIndex].Description
		newDesc := line[8:]
		sep := " "

		if lastDesc == "" {
			sep = ""
		}

		page.CurrentSheet.StationDescription[lastIndex].Description = lastDesc + sep + newDesc

		return
	}

	if page.CurrentBuffer == stationRevoveryHeader && len(page.CurrentSheet.StationRecoveries) > 0 {
		lastIndex := len(page.CurrentSheet.StationRecoveries) - 1
		lastDesc := page.CurrentSheet.StationRecoveries[lastIndex].Description
		newDesc := line[8:]
		sep := " "

		if lastDesc == "" {
			sep = ""
		}

		page.CurrentSheet.StationRecoveries[lastIndex].Description = lastDesc + sep + newDesc

		return
	}
}

// see if line is grid spatial address
func (page *Page) checkSpatialAddress (line string) bool {
	if len(line) > 45 {
		if line[8:42] == spatialAddressKey {
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
	} else if page.CurrentBuffer == statePlaneHeader {
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
	} else if page.CurrentBuffer == accuracyHeader {
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
	return key == horzOrderKey || key == ellpOrderKey|| key == vertOrderKey || key == networkKey
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

// (abcdef) => abcdef
func getInnerParValue (s string) string {
	isStarted := false
	val := ""

	for _, c := range s {
		char := string(c)

		if char == "(" && !isStarted {
			isStarted = true
		} else if char == ")" && isStarted {
			break
		} else if isStarted {
			val = val + char
		}
	}

	return val
}

func min (x int, y int) int {
	if x < y {
		return x
	}

	return y
}
