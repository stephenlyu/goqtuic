package main

import (
	"github.com/therecipe/qt/widgets"
	"github.com/therecipe/qt/core"
)

type XUIDialog struct {
	Label1 *widgets.QLabel
	VerticalLayout *widgets.QVBoxLayout
}

func (this *XUIDialog) SetupUI(dialog *widgets.QDialog) {
	dialog.SetObjectName("Dialog")
	dialog.Resize2(706, 833)
	dialog.SetMinimumSize(core.NewQSize2(706, 0))
	dialog.SetMaximumSize(core.NewQSize2(706, 16777215))
	this.VerticalLayout = widgets.NewQVBoxLayout2(dialog)
	this.VerticalLayout.SetObjectName("verticalLayout")

	this.RetranslateUi(dialog)
}

func (this *XUIDialog) RetranslateUi(dialog *widgets.QDialog) {
	_translate := core.QCoreApplication_Translate
	dialog.SetWindowTitle(_translate("Dialog", "Dialog", "", -1))
}
