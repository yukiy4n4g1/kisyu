package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell"
)

func main() {
	editor := InitEditor()
	err := editor.S.Init()
	errorCheck(err)
	defer editor.S.Fini()
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

type Editor struct {
	S  tcell.Screen
	Cx int
	Cy int
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
		editor.S.SetContent(0, y, '~', nil, tcell.StyleDefault)
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

	return &Editor{s, cx, cy}
}
