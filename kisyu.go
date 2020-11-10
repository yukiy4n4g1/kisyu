package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"unicode/utf8"

	"golang.org/x/text/width"

	"github.com/gdamore/tcell"
)

// TabStop はタブのサイズの定数
const TabStop = 4

func main() {
	editor := InitEditor()
	err := editor.S.Init()
	errorCheck(err)
	defer editor.S.Fini()
	if len(os.Args) > 1 {
		err := editor.Buf.Open(os.Args[1])
		errorCheck(err)
		editor.FileName = os.Args[1]
	}
	editor.S.ShowCursor(0, 0)

	for {
		editor.RefreshScreen()
		editor.ProcessEvent()
	}
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type CursorMove int

const (
	Left CursorMove = iota
	Right
	Up
	Down
	Home
	End
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

type Editor struct {
	S        tcell.Screen
	Rowoff   int
	Coloff   int
	Buf      Buffer
	FileName string
}

func (editor *Editor) ProcessEvent() {
	event := editor.S.PollEvent()
	switch ev := event.(type) {
	case *tcell.EventKey:
		editor.KeyEvent(ev)
	}
}

func (editor *Editor) KeyEvent(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyCtrlQ:
		editor.S.Fini()
		os.Exit(0)
	case tcell.KeyCtrlS:
		editor.Save()
	case tcell.KeyUp:
		editor.Buf.MoveCursor(Up)
	case tcell.KeyDown:
		editor.Buf.MoveCursor(Down)
	case tcell.KeyLeft:
		editor.Buf.MoveCursor(Left)
	case tcell.KeyRight:
		editor.Buf.MoveCursor(Right)
	case tcell.KeyPgUp:
		editor.Buf.MoveCy(editor.Rowoff)
	case tcell.KeyPgDn:
		_, wy := editor.ScreenSize()
		editor.Buf.MoveCy(editor.Rowoff + wy - 1)
	case tcell.KeyHome:
		editor.Buf.MoveCx(Home)
	case tcell.KeyEnd:
		editor.Buf.MoveCx(End)
	case tcell.KeyBackspace2:
		editor.Buf.DeleteRune()
	case tcell.KeyEnter:
		editor.Buf.InsertNewLine()
	case tcell.KeyRune:
		editor.Buf.InsertRune(ev.Rune())
	}
}

func (editor *Editor) DrowRows() {
	_, wy := editor.ScreenSize()
	for y := 0; y < wy; y++ {
		filerow := y + editor.Rowoff
		row, err := editor.Buf.Render(filerow)
		if err != nil {
			editor.S.SetContent(0, y, '~', nil, tcell.StyleDefault)
		} else {
			start := editor.Coloff
			if start > editor.Buf.CursorEnd(filerow) {
				start = editor.Buf.CursorEnd(filerow)
			}
			for x, r := range row[start:] {
				editor.S.SetContent(x, y, r, nil, tcell.StyleDefault)
			}
		}
	}
}

func (editor *Editor) DrowStatusBar() {
	rowStatus := []rune(fmt.Sprintf("%d/%d", editor.Buf.Cy()+1, editor.Buf.RowLen()))
	wx, wy := editor.ScreenSize()
	for x := 0; x < wx; x++ {
		if x < len(rowStatus) {
			editor.S.SetContent(x, wy, rowStatus[x], nil, tcell.StyleDefault.Reverse(true))
		} else {
			editor.S.SetContent(x, wy, ' ', nil, tcell.StyleDefault.Reverse(true))
		}
	}
}

func (editor *Editor) Scroll() {
	wx, wy := editor.ScreenSize()
	cx, cy := editor.Buf.Cx(), editor.Buf.Cy()
	if cy < editor.Rowoff {
		editor.Rowoff = cy
	}
	if cy >= editor.Rowoff+wy {
		editor.Rowoff = cy - wy + 1
	}

	if cx < editor.Coloff {
		editor.Coloff = cx
	}
	if cx >= editor.Coloff+wx {
		editor.Coloff = cx - wx + 1
	}
}

func (editor *Editor) ScreenSize() (int, int) {
	x, y := editor.S.Size()
	y -= 1
	return x, y
}

func (editor *Editor) RefreshScreen() {
	editor.S.Clear()
	editor.Scroll()
	editor.DrowRows()
	editor.DrowStatusBar()
	editor.S.ShowCursor(editor.Buf.Cx()-editor.Coloff, editor.Buf.Cy()-editor.Rowoff)
	editor.S.Show()
}

func (editor *Editor) Save() {
	if editor.FileName != "" {
		s := editor.Buf.RowToString()
		err := ioutil.WriteFile(editor.FileName, ([]byte)(s), 0644)
		errorCheck(err)
	}
}

func InitEditor() *Editor {
	s, err := tcell.NewScreen()
	errorCheck(err)
	rowoff, coloff := 0, 0

	return &Editor{s, rowoff, coloff, InitBuffer(), ""}
}
