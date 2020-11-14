package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gdamore/tcell"
)

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
