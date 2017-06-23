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

type Date struct {
	Year, Month, Day int
}

type QSizePolicy struct {
	VSizeType, HSizeType string
	HorStretch int
	VerStretch int
}

type Attribute struct {
	Name string
	Value interface{}
}

type QColor struct {
	Alpha int
	Red int
	Green int
	Blue int
}

type QBrush struct {
	BrushStyle string
	Color *QColor
}

type QColorRole struct {
	Role string
	Brush *QBrush
}

type QPalette struct {
	Active []*QColorRole
	InActive []*QColorRole
	Disabled []*QColorRole
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
	Attributes []*Attribute
	Layout *QLayout
	Children []*QWidget
	items[] *QWidgetItem 		// ComboBox item
	ZOrders []string
}
