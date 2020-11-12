package main

import (
	"unicode/utf8"

	"golang.org/x/text/width"
)

type Row struct {
	Text            []rune
	Render          []rune
	CursorPositions []int
	Dirty           bool
}

func (row *Row) UpdateRender() {
	if row.Dirty {
		row.Render = make([]rune, 0, len(row.Text))
		row.CursorPositions = make([]int, len(row.Text)+1)
		cursorPosition := 0

		for i, r := range row.Text {
			if r == '\t' {
				row.Render = append(row.Render, ' ')
				row.CursorPositions[i] = cursorPosition
				cursorPosition++
				for len(row.Render)%TabStop != 0 {
					row.Render = append(row.Render, ' ')
					cursorPosition++
				}
			} else if utf8.RuneLen(r) > 1 && (width.LookupRune(r).Kind() != width.EastAsianHalfwidth || width.LookupRune(r).Kind() != width.EastAsianNarrow) {
				row.CursorPositions[i] = cursorPosition
				row.Render = append(row.Render, r)
				row.Render = append(row.Render, ' ')
				cursorPosition += 2
			} else {
				row.CursorPositions[i] = cursorPosition
				row.Render = append(row.Render, r)
				cursorPosition++
			}
		}
		row.CursorPositions[len(row.CursorPositions)-1] = cursorPosition
		row.Dirty = false
	}
}

func (row *Row) InsertRune(colNum int, r rune) {
	if colNum < len(row.Text) {
		row.Text = append(row.Text[:colNum+1], row.Text[colNum:]...)
		row.Text[colNum] = r
		row.Dirty = true
	} else if colNum == len(row.Text) {
		row.Text = append(row.Text, r)
		row.Dirty = true
	}
}

func (row *Row) DeleteRune(colNum int) {
	if colNum < len(row.Text) {
		row.Text = append(row.Text[:colNum-1], row.Text[colNum:]...)
	} else {
		row.Text = row.Text[:len(row.Text)-1]
	}
	row.Dirty = true
}

func InitRow(text []rune) Row {
	row := Row{text, []rune{}, []int{0}, true}
	return row
}
