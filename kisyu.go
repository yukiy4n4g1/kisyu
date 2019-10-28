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

func (buf *Buffer) ColLen(i int) int {
	if i < len(buf.text) {
		return len(buf.text[i])
	}
	return 0
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
	}
}

func (editor *Editor) DrowRows() {
	_, wy := editor.S.Size()
	for y := 0; y < wy; y++ {
		filerow := y + editor.Rowoff
		row, err := editor.Buf.Line(filerow)
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
	}

	colLen := editor.Buf.ColLen(editor.Cy)
	if editor.Cx > colLen {
		editor.Cx = colLen
	}

}

func (editor *Editor) Scroll() {
	wx, wy := editor.S.Size()
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
	editor.S.ShowCursor(editor.Cx-editor.Coloff, editor.Cy-editor.Rowoff)
	editor.S.Show()
}

func InitEditor() *Editor {
	s, err := tcell.NewScreen()
	errorCheck(err)
	cx, cy, rowoff, coloff := 0, 0, 0, 0

	return &Editor{s, cx, cy, rowoff, coloff, Buffer{[][]rune{}}}
}
