package main


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
)

type Page struct {
	CurrentSheet DataSheet
	CurrentSection int
	LineNum int
}

func NewPage () Page {
	page := Page{}
	page.Reset()
	return page
}

func (page *Page) Reset () {
	page.CurrentSheet = DataSheet{}
	page.CurrentSheet.BasicMetadata = make(map[string]string)
	page.CurrentSheet.NewSurveyControl = make([]Survey, 0)
	page.CurrentSheet.OldSurveyControl = make([]Survey, 0)
	page.CurrentSection = basicMetadataSection
	page.LineNum = 0
}

func (page *Page) AddLine (line string) {
	if page.LineNum == 0 {
		page.ReadId(line)
	}

	if page.CurrentSection == basicMetadataSection {
		page.BasicMetadataSection(line)
		return
	}

	if page.CurrentSection == currentSurveyControlSection {
		page.CurrentSurveyControlSection(line)
		return
	}

	page.LineNum++
}

func (page *Page) Make () DataSheet {
	defer page.Reset()

	return page.CurrentSheet
}

func (page *Page) ReadId (line string) {
	if len(line) <= 8 {
		return
	}

	page.CurrentSheet.Id = line[1:7]
}

func (page *Page) BasicMetadataSection (line string) {
	if len(line) == 7 {
		page.CurrentSection = currentSurveyControlSection
		return
	}

	if len(line) <= 10 {
		return
	}

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
	if len(line) == 7 {
		page.CurrentSection = accuracySection
		return
	}

	if len(line) <= 70 {
		return
	}

	if string(line[10]) == " " || string(line[10]) == "_" {
		return
	}

	isNew := string(line[7]) == "*"

	if isNew {
		page.CurrentSheet.NewSurveyControl = append(page.CurrentSheet.NewSurveyControl, Survey{
			Item:  trimWhiteSpace(line[9:30]),
			Value: trimWhiteSpace(line[31:70]),
			By:    trimWhiteSpace(line[71:]),
		})
	} else {
		page.CurrentSheet.OldSurveyControl = append(page.CurrentSheet.OldSurveyControl, Survey{
			Item:  trimWhiteSpace(line[9:25]),
			Value: trimWhiteSpace(line[26:70]),
			By:    trimWhiteSpace(line[71:]),
		})
	}
}

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