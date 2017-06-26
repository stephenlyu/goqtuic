package parser

type LayoutDefault struct {
	Spacing, Margin int
}

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

type QPointF struct {
	X, Y float64
}

type QSizeF struct {
	Width, Height float64
}

type QRectF struct {
	X, Y float64
	Width, Height float64
}

type Char struct {
	Unicode int
}

type Url struct {
	String string
}

type QFont struct {
	Family string
	PointSize int
	Weight int
	Italic bool
	Bold bool
	Underline bool
	Strikeout bool
	AntiAliasing bool
	StyleStrategy string
	Kerning bool
}

type QLocale struct {
	Language string
	Country string
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

type Time struct {
	Hour, Minute, Second int
}

type DateTime struct {
	Year, Month, Day int
	Hour, Minute, Second int
}

type StringList struct {
	Strings []string
}

type ResourcePixmap struct {
	Resource string
	Alias string
}

type ResourceIcon struct {
	//TODO:
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

type GradientStop struct {
	Position float64
	Colors []*QColor
}

type QGrdient struct {
	GradientStops []*GradientStop

	StartX, StartY float64
	EndX, EndY float64
	CentralX, CentralY float64
	FocalX, FocalY float64
	Radius float64
	Angle float64
	Type string
	Spread string
	CoordinateMode string
}

type QBrush struct {
	BrushStyle string
	Color *QColor
	Texture *Property
	Gradient *QGrdient
}

type QColorRole struct {
	Role string
	Brush *QBrush
}

type ColorGroupItem struct {
	IsColor bool
	Color *QColor
	ColorRole *QColorRole
}

type ColorGroup struct {
	Items []*ColorGroupItem
}

type QPalette struct {
	Active *ColorGroup
	InActive *ColorGroup
	Disabled *ColorGroup
}

type Property struct {
	Name string
	StdSet bool
	Value interface{}
}

type QSpacer struct {
	Name string
	Properties []*Property
}

type QWidgetItem struct {
	Props []*Property
	Items []*QWidgetItem 		// TODO: what's this?
	Row, Column int
}

type QLayoutItem struct {
	Row, Column int
	Rowspan, Colspan int
	Alignment string
	View interface{}
}

type QLayout struct {
	Class string
	Name string
	Stretch string
	RowStretch string
	ColumnStretch string
	RowMinimumHeight string
	ColumnMinimumWidth string

	Properties []*Property
	Items []*QLayoutItem
	Attributes []*Property
}

type Row struct {
	Props []*Property
}

type Column struct {
	Props []*Property
}

type ActionGroup struct {
	Name string

	Actions []*Action
	ActionGroups []*ActionGroup
	Props []*Property
	Attributes []*Property
}

type ActionRef struct {
	Name string
}

type Action struct {
	Name string
	Menu string
	Props []*Property
	Attributes []*Property
}

type QWidget struct {
	// No element: class, script, widgetdata

	Class string
	Name string
	Native bool

	Properties []*Property
	Attributes []*Attribute

	Rows []*Row
	Columns []*Column
	Items[] *QWidgetItem 		// ComboBox item

	Layout *QLayout
	Widgets []*QWidget
	Actions []*Action
	ActionsGroups []*ActionGroup
	AddActions []*ActionRef
	ZOrders []string
}
