package main

import (
	"errors"
	"io/ioutil"
	"os"
)

type CursorMove int

const (
	Left CursorMove = iota
	Right
	Up
	Down
	Home
	End
)

type Buffer struct {
	rows []Row
	cx   int
	cy   int
}

func (buf *Buffer) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	byteTxt, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	buf.rows = make([]Row, 0, 10)
	runetxt := []rune(string(byteTxt))

	count := 0
	for i, r := range runetxt {
		if r == '\n' {
			buf.rows = append(buf.rows, InitRow(runetxt[count:i]))
			count = i + 1
		}
	}
	if runetxt[len(runetxt)-1] != '\n' {
		buf.rows = append(buf.rows, InitRow(runetxt[count:len(runetxt)]))
	}

	return nil
}

func (buf *Buffer) Render(rowNum int) ([]rune, error) {
	if rowNum < len(buf.rows) {
		buf.rows[rowNum].UpdateRender()
		return buf.rows[rowNum].Render, nil
	}
	return []rune{}, errors.New("index out of range")
}

func (buf *Buffer) InsertRune(r rune) {
	if buf.cy < len(buf.rows) && buf.cy >= 0 {
		buf.rows[buf.cy].InsertRune(buf.cx, r)
	}
	buf.cx++
}

func (buf *Buffer) DeleteRune() {
	if buf.cx > 0 {
		buf.rows[buf.cy].DeleteRune(buf.cx)
		buf.cx--
	} else if buf.cx == 0 && buf.cy > 0 {
		for _, r := range buf.rows[buf.cy].Text {
			buf.rows[buf.cy-1].InsertRune(len(buf.rows[buf.cy-1].Text), r)
		}
		if buf.cy < buf.RowLen() {
			buf.rows = append(buf.rows[:buf.cy], buf.rows[buf.cy+1:]...)
		} else {
			buf.rows = buf.rows[:buf.cy-1]
		}
		buf.cy--
		buf.cx = len(buf.rows[buf.cy].Text)
	}
}

func (buf *Buffer) RowToString() string {
	s := ""
	for _, row := range buf.rows {
		s += string(row.Text)
		s += "\n"
	}
	return s
}

func (buf *Buffer) InsertNewLine() {
	if buf.cy < len(buf.rows) {
		buf.rows = append(buf.rows[:buf.cy+1], buf.rows[buf.cy:]...)
	} else if buf.cy == len(buf.rows) {
		buf.rows = append(buf.rows, InitRow(make([]rune, 0, 10)))
	}
	buf.rows[buf.cy+1].Text = buf.rows[buf.cy+1].Text[buf.cx:]
	buf.rows[buf.cy].Text = buf.rows[buf.cy].Text[:buf.cx]
	buf.rows[buf.cy+1].Dirty = true
	buf.rows[buf.cy].Dirty = true
	buf.cy++
	buf.cx = 0
}

func (buf *Buffer) Cy() int {
	return buf.cy
}

func (buf *Buffer) Cx() int {
	buf.rows[buf.cy].UpdateRender()
	return buf.rows[buf.cy].CursorPositions[buf.cx]
}

func (buf *Buffer) ColLen(colNum int) int {
	if colNum < len(buf.rows) {
		return len(buf.rows[colNum].Text)
	}
	return -1
}

func (buf *Buffer) CursorEnd(colNum int) int {
	if colNum < len(buf.rows) {
		buf.rows[colNum].UpdateRender()
		return buf.rows[colNum].CursorPositions[len(buf.rows[colNum].CursorPositions)-1]
	}
	return -1
}

func (buf *Buffer) RowLen() int {
	return len(buf.rows)
}

func (buf *Buffer) MoveCursor(c CursorMove) {
	switch c {
	case Up:
		if buf.cy > 0 {
			buf.cy--
		}
	case Down:
		if buf.cy < buf.RowLen()-1 {
			buf.cy++
		}
	case Left:
		if buf.cx > 0 {
			buf.cx--
		} else if buf.cy > 0 {
			buf.cy--
			buf.cx = buf.ColLen(buf.cy)
		}
	case Right:
		if buf.cx < buf.ColLen(buf.cy) {
			buf.cx++
		} else if buf.cy < buf.RowLen()-1 {
			buf.cy++
			buf.cx = 0
		}
	}
	if buf.cx > buf.ColLen(buf.cy) {
		buf.cx = buf.ColLen(buf.cy)
	}
}

func (buf *Buffer) MoveCy(y int) {
	buf.cy = y
	if buf.cy > buf.RowLen() {
		buf.cy = buf.RowLen()
	}
	if buf.cx > buf.ColLen(buf.cy) {
		buf.cx = buf.ColLen(buf.cy)
	}
}

func (buf *Buffer) MoveCx(c CursorMove) {
	switch c {
	case Home:
		buf.cx = 0
	case End:
		buf.cx = buf.ColLen(buf.cy)
	}
}

func InitBuffer() Buffer {
	return Buffer{[]Row{InitRow([]rune{})}, 0, 0}
}
