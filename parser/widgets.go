package parser

type QPoint struct {
	X, Y int
}

type QSize struct {
	Width, Height int
}

type QRect struct {
	X, Y int
	Width, Height int
}

type QFont struct {
	PointSize int
}

type Enum struct {
	Value string
}

type Set struct {
	Value string
}

type Property struct {
	Name string
	Value interface{}
}

type QSpacer struct {
	Name string
	Properties []*Property
}

type QWidgetItem struct {
	Prop *Property
}

type QLayoutItem struct {
	Row, Column int
	View interface{}
}

type QLayout struct {
	Class string
	Name string
	Stretch string
	Properties []*Property
	Items []*QLayoutItem
}

type QWidget struct {
	Class string
	Name string
	Properties []*Property
	Layout *QLayout
	Children []*QWidget
	items[] *QWidgetItem 		// ComboBox item
}
