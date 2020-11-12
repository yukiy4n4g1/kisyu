package main

import (
	"log"
	"os"
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
