package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/gdamore/tcell"
)

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
	editor.DrowRows()

	for {
		editor.ProcessEvent()
		editor.DrowRows()
	}
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Buffer struct {
	text [][]rune
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

type Editor struct {
	S   tcell.Screen
	Cx  int
	Cy  int
	Buf Buffer
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
	}
}

func (editor *Editor) DrowRows() {
	_, wy := editor.S.Size()
	for y := 0; y < wy; y++ {
		row, err := editor.Buf.Line(y)
		if err != nil {
			editor.S.SetContent(0, y, '~', nil, tcell.StyleDefault)
		} else {
			for x, r := range row {
				editor.S.SetContent(x, y, r, nil, tcell.StyleDefault)
			}
		}
	}
	editor.S.Show()
}

func (editor *Editor) MoveCursor(key tcell.Key) {
	wx, wy := editor.S.Size()
	switch key {
	case tcell.KeyUp:
		if editor.Cy > 0 {
			editor.Cy--
		}
	case tcell.KeyDown:
		if editor.Cy < wy-1 {
			editor.Cy++
		}
	case tcell.KeyLeft:
		if editor.Cx > 0 {
			editor.Cx--
		}
	case tcell.KeyRight:
		if editor.Cx < wx-1 {
			editor.Cx++
		}
	}
	editor.S.ShowCursor(editor.Cx, editor.Cy)
}

func InitEditor() *Editor {
	s, err := tcell.NewScreen()
	errorCheck(err)
	cx, cy := 0, 0

	return &Editor{s, cx, cy, Buffer{[][]rune{}}}
}
