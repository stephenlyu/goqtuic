package main

import (
	"github.com/therecipe/qt/widgets"
	"github.com/therecipe/qt/core"
	"os"
)

type Dialog struct {
	XUIDialog
	Dialog *widgets.QDialog
}

func NewDialog(parent widgets.QWidget_ITF) *Dialog {
	dialog := &Dialog{
		Dialog: widgets.NewQDialog(parent, core.Qt__Dialog),
	}

	dialog.SetupUI(dialog.Dialog)
	return dialog
}

// TODO: Add more functions here
func (this *Dialog) TestFunction() {
	// I am a test function.
}

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	w := NewDialog(nil)
	w.Dialog.Show()

	os.Exit(app.Exec())
}
