package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gdamore/tcell"
)

const TabStop = 4

func main() {
	editor := InitEditor()
	err := editor.S.Init()
	errorCheck(err)
	defer editor.S.Fini()
	if len(os.Args) > 1 {
		err := editor.Buf.Open(os.Args[1])
		errorCheck(err)
	}
	editor.S.ShowCursor(editor.Cx, editor.Cy)

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

type Buffer struct {
	text   [][]rune
	render []rune
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

	runetxt := []rune(string(byteTxt))
	buf.text = make([][]rune, 0, 10)

	count := 0
	for i, r := range runetxt {
		if r == '\n' {
			buf.text = append(buf.text, append([]rune{}, runetxt[count:i]...))
			count = i + 1
		}
	}
	if runetxt[len(runetxt)-1] != '\n' {
		buf.text = append(buf.text, append([]rune{}, runetxt[count:len(runetxt)]...))
	}

	return nil
}

func (buf *Buffer) Line(i int) ([]rune, error) {
	if len(buf.text) > i {
		return buf.text[i], nil
	}
	return nil, errors.New("buffer out of range")
}

func (buf *Buffer) RowLen() int {
	return len(buf.text)
}

func (buf *Buffer) ColLen(i int) int {
	if i < len(buf.text) {
		length := len(buf.text[i])
		for index, r := range buf.text[i] {
			if r == '\t' {
				length += TabStop - 1 - (index % TabStop)
			}
		}
		return length
	}
	return 0
}

func (buf *Buffer) UpdateRow(i int) error {
	line, err := buf.Line(i)
	if err != nil {
		return err
	}
	buf.render = make([]rune, 0, len(line))

	for _, r := range line {
		if r == '\t' {
			buf.render = append(buf.render, ' ')
			for len(buf.render)%TabStop != 0 {
				buf.render = append(buf.render, ' ')
			}
		} else {
			buf.render = append(buf.render, r)
		}
	}
	return nil
}

func (buf *Buffer) Render(i int) ([]rune, error) {
	err := buf.UpdateRow(i)
	if err != nil {
		return nil, err
	}
	return buf.render, nil
}

type Editor struct {
	S      tcell.Screen
	Cx     int
	Cy     int
	Rowoff int
	Coloff int
	Buf    Buffer
}

func (editor *Editor) ProcessEvent() {
	ev := editor.S.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventKey:
		editor.KeyEvent(ev.Key())
	}
}

func (editor *Editor) KeyEvent(key tcell.Key) {
	switch key {

	case tcell.KeyCtrlQ:
		editor.S.Fini()
		os.Exit(0)

	case tcell.KeyUp:
		editor.MoveCursor(key)
	case tcell.KeyDown:
		editor.MoveCursor(key)
	case tcell.KeyLeft:
		editor.MoveCursor(key)
	case tcell.KeyRight:
		editor.MoveCursor(key)
	case tcell.KeyPgUp:
		editor.MoveCursor(key)
	case tcell.KeyPgDn:
		editor.MoveCursor(key)
	}
}

func (editor *Editor) DrowRows() {
	_, wy := editor.S.Size()
	for y := 0; y < wy-1; y++ {
		filerow := y + editor.Rowoff
		row, err := editor.Buf.Render(filerow)
		if err != nil {
			editor.S.SetContent(0, y, '~', nil, tcell.StyleDefault)
		} else {
			start := editor.Coloff
			n := editor.Buf.ColLen(y)
			if n <= start {
				start = n
			}
			for x, r := range row[start:] {
				editor.S.SetContent(x, y, r, nil, tcell.StyleDefault)
			}
		}
	}
}

func (editor *Editor) DrowStatusBar() {
	rowStatus := []rune(fmt.Sprintf("%d/%d", editor.Cy+1, editor.Buf.RowLen()))
	wx, wy := editor.S.Size()
	for x := 0; x < wx; x++ {
		if x < len(rowStatus) {
			editor.S.SetContent(x, wy-1, rowStatus[x], nil, tcell.StyleDefault.Reverse(true))
		} else {
			editor.S.SetContent(x, wy-1, ' ', nil, tcell.StyleDefault.Reverse(true))
		}
	}
}

func (editor *Editor) MoveCursor(key tcell.Key) {
	switch key {
	case tcell.KeyUp:
		if editor.Cy > 0 {
			editor.Cy--
		}
	case tcell.KeyDown:
		if editor.Cy < editor.Buf.RowLen()-1 {
			editor.Cy++
		}
	case tcell.KeyLeft:
		if editor.Cx > 0 {
			editor.Cx--
		} else if editor.Cy > 0 {
			editor.Cy--
			editor.Cx = editor.Buf.ColLen(editor.Cy)
		}
	case tcell.KeyRight:
		if editor.Cx < editor.Buf.ColLen(editor.Cy) {
			editor.Cx++
		} else if editor.Cy < editor.Buf.RowLen()-1 {
			editor.Cy++
			editor.Cx = 0
		}
	case tcell.KeyPgUp:
		editor.Cy = editor.Rowoff
	case tcell.KeyPgDn:
		_, wy := editor.S.Size()
		wy--
		editor.Cy = editor.Rowoff + wy - 1
		rowLen := editor.Buf.RowLen()
		if editor.Cy > rowLen {
			editor.Cy = rowLen
		}
	}

	colLen := editor.Buf.ColLen(editor.Cy)
	if editor.Cx > colLen {
		editor.Cx = colLen
	}

}

func (editor *Editor) Scroll() {
	wx, wy := editor.S.Size()
	wy--
	if editor.Cy < editor.Rowoff {
		editor.Rowoff = editor.Cy
	}
	if editor.Cy >= editor.Rowoff+wy {
		editor.Rowoff = editor.Cy - wy + 1
	}

	if editor.Cx < editor.Coloff {
		editor.Coloff = editor.Cx
	}
	if editor.Cx >= editor.Coloff+wx {
		editor.Coloff = editor.Cx - wx + 1
	}
}

func (editor *Editor) RefreshScreen() {
	editor.S.Clear()
	editor.Scroll()
	editor.DrowRows()
	editor.DrowStatusBar()
	editor.S.ShowCursor(editor.Cx-editor.Coloff, editor.Cy-editor.Rowoff)
	editor.S.Show()
}

func InitEditor() *Editor {
	s, err := tcell.NewScreen()
	errorCheck(err)
	cx, cy, rowoff, coloff := 0, 0, 0, 0

	return &Editor{s, cx, cy, rowoff, coloff, Buffer{[][]rune{}, []rune{}}}
}
