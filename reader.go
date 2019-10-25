package main

import (
	"bufio"
	"io"
)

type Reader struct {
	Scanner *bufio.Scanner
	BottomHeader string
	Page Page
}

func NewReader (r io.Reader) Reader {
	return Reader{
		Scanner: bufio.NewScanner(r),
		BottomHeader: "",
		Page: NewPage(),
	}
}

func (reader *Reader) HasNext () bool {
	next := false

	if len(reader.BottomHeader) > 0 {
		next = lineIsNewData(reader.BottomHeader)
		reader.BottomHeader = ""
	}

	for reader.Scanner.Scan() && !next {
		line := reader.Scanner.Text()
		next = lineIsNewData(line)
	}

	return next
}

func (reader *Reader) Next () DataSheet {
	for reader.Scanner.Scan() {
		line := reader.Scanner.Text()
		next := lineIsNewData(line)

		if next {
			reader.BottomHeader = line
			return reader.Page.Make()
		}

		if !next {
			reader.Page.AddLine(line)
		}
	}

	return reader.Page.Make()
}

func lineIsNewData (s string) bool {
	if len(s) > 0 {
		return string(s[0]) == "1"
	}

	return false
}