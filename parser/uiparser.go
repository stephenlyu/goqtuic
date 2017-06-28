package parser

import (
	xmlx "github.com/jteeuwen/go-pkg-xmlx"
	"log"
	"fmt"
)

type parser struct {
	doc *xmlx.Document

	uiFile string
	class string

	buttonGroups []string
	tabStops []string
	layoutDefault *LayoutDefault

	widget *QWidget
}

func NewParser(uiFile string) (error, *parser) {
	ret := &parser{uiFile: uiFile}
	ret.doc = xmlx.New()
	err := ret.doc.LoadFile(uiFile, nil)
	if err != nil {
		return err, nil
	}
	return nil, ret
}

func (this *parser) parsePoint(n *xmlx.Node) *QPoint {
	width := n.I("", "x")
	height := n.I("", "y")
	return &QPoint{width, height}
}

func (this *parser) parseSize(n *xmlx.Node) *QSize {
	width := n.I("", "width")
	height := n.I("", "height")
	return &QSize{width, height}
}

func (this *parser) parseRect(n *xmlx.Node) *QRect {
	x := n.I("", "x")
	y := n.I("", "y")
	width := n.I("", "width")
	height := n.I("", "height")
	return &QRect{x, y, width, height}
}

func (this *parser) parsePointF(n *xmlx.Node) *QPointF {
	width := n.F64("", "x")
	height := n.F64("", "y")
	return &QPointF{width, height}
}

func (this *parser) parseSizeF(n *xmlx.Node) *QSizeF {
	width := n.F64("", "width")
	height := n.F64("", "height")
	return &QSizeF{width, height}
}

func (this *parser) parseRectF(n *xmlx.Node) *QRectF {
	x := n.F64("", "x")
	y := n.F64("", "y")
	width := n.F64("", "width")
	height := n.F64("", "height")
	return &QRectF{x, y, width, height}
}

func (this *parser) parseFont(n *xmlx.Node) *QFont {
	return &QFont{
		Family: n.S("", "family"),
		PointSize: n.I("", "pointsize"),
		Weight: n.I("", "weight"),
		Italic: n.B("", "bool"),
		Bold: n.B("", "bool"),
		Underline: n.B("", "bool"),
		Strikeout: n.B("", "bool"),
		AntiAliasing: n.B("", "bool"),
		StyleStrategy: n.S("", "string"),
		Kerning: n.B("", "bool"),
	}
}

func (this *parser) parseGradientStop(n *xmlx.Node) *GradientStop {
	position := n.Af64("", "position")
	colorNodes := n.SelectNodesDirect("", "color")
	colors := make([]*QColor, len(colorNodes))
	for i, ch := range colorNodes {
		colors[i] = this.parseColor(ch)
	}
	return &GradientStop{Position: position, Colors: colors}
}

func (this *parser) parseGradient(n *xmlx.Node) *QGrdient {
	children := n.SelectNodesDirect("", "gradientStop")
	stops := make([]*GradientStop, len(children))
	for i, ch := range children {
		stops[i] = this.parseGradientStop(ch)
	}
	return &QGrdient{
		GradientStops: stops,
		StartX: n.Af64("", "startx"),
		StartY: n.Af64("", "starty"),
		EndX: n.Af64("", "endx"),
		EndY: n.Af64("", "endy"),
		CentralX: n.Af64("", "centralx"),
		CentralY: n.Af64("", "centraly"),
		FocalX: n.Af64("", "focalx"),
		FocalY: n.Af64("", "focaly"),
		Radius: n.Af64("", "radius"),
		Angle: n.Af64("", "angle"),
		Type: n.S("", "type"),
		Spread: n.S("", "spread"),
		CoordinateMode: n.S("", "coordinatemode"),
	}
}

func (this *parser) parseLocale(n *xmlx.Node) *QLocale {
	return &QLocale{n.As("", "language"), n.As("", "country")}
}

func (this *parser) parseEnum(n *xmlx.Node) *Enum {
	return &Enum{n.GetValue()}
}

func (this *parser) parseSet(n *xmlx.Node) *Set {
	return &Set{n.GetValue()}
}

func (this *parser) parseChar(n *xmlx.Node) *Char {
	return &Char{n.I("", "unicode")}
}

func (this *parser) parseUrl(n *xmlx.Node) *Url {
	return &Url{n.S("", "string")}
}

func (this *parser) parseDate(n *xmlx.Node) *Date {
	return &Date{Year: n.I("", "year"), Month: n.I("", "month"), Day: n.I("", "day")}
}

func (this *parser) parseTime(n *xmlx.Node) *Time {
	return &Time{Hour: n.I("", "hour"), Minute: n.I("", "minute"), Second: n.I("", "second")}
}

func (this *parser) parseDateTime(n *xmlx.Node) *DateTime {
	return &DateTime{Year: n.I("", "year"), Month: n.I("", "month"), Day: n.I("", "day"),
		Hour: n.I("", "hour"), Minute: n.I("", "minute"), Second: n.I("", "second")}
}

func (this *parser) parseStringList(n *xmlx.Node) *StringList {
	children := n.SelectNodesDirect("", "string")
	strings := make([]string, len(children))
	for i, ch := range children {
		strings[i] = ch.GetValue()
	}

	return &StringList{Strings: strings, NotR: n.Ab("", "notr")}
}

func (this *parser) parseString(n *xmlx.Node) *String {
	return &String{Value: n.GetValue(), NotR: n.Ab("", "notr")}
}

func (this *parser) parseColor(n *xmlx.Node) *QColor {
	return &QColor{
		Alpha: n.Ai("", "alpha"),
		Red: n.I("", "red"),
		Green: n.I("", "green"),
		Blue: n.I("", "blue"),
	}
}

func (this *parser) parseBrush(n *xmlx.Node) *QBrush {
	ch := n.SelectNode("", "color")
	return &QBrush{
		BrushStyle: n.As("", "brushstyle"),
		Color: this.parseColor(ch),
	}
}

func (this *parser) parseColorRole(n *xmlx.Node) *QColorRole {
	ch := n.SelectNode("", "brush")
	return &QColorRole{
		Role: n.As("", "role"),
		Brush: this.parseBrush(ch),
	}
}

func (this *parser) parseColorGroup(nodes []*xmlx.Node) *ColorGroup {
	items := make([]*ColorGroupItem, len(nodes))
	for i, n := range nodes {
		if n.Name.Local == "color" {
			items[i] = &ColorGroupItem{
				IsColor: true,
				Color: this.parseColor(n),
			}
		} else if n.Name.Local == "colorrole" {
			items[i] = &ColorGroupItem{
				IsColor: false,
				ColorRole: this.parseColorRole(n),
			}
		} else {
			log.Fatalf("bad color group %s", n)
		}
	}
	return &ColorGroup{Items: items}
}

func (this *parser) parsePalette(n *xmlx.Node) *QPalette {
	ret := &QPalette{}

	childCount := len(this.elementChildren(n))
	if childCount != 3 {
		log.Fatalf("Bad palette with %d children", childCount)
	}

	activeNode := n.SelectNode("", "active")
	ret.Active = this.parseColorGroup(this.elementChildren(activeNode))

	inActiveNode := n.SelectNode("", "inactive")
	ret.InActive = this.parseColorGroup(this.elementChildren(inActiveNode))

	disabledNode := n.SelectNode("", "disabled")
	ret.Disabled = this.parseColorGroup(this.elementChildren(disabledNode))

	return ret
}

func (this *parser) parseAttribute(n *xmlx.Node) *Attribute {
	var value interface{}

	ch := this.elementChildren(n)[0]

	name := n.As("", "name")

	switch ch.Name.Local {
	case "string":
		value = &String{Value: n.S("", "string"), NotR: ch.Ab("", "notr")}
	case "bool":
		value = n.B("", "bool")
	case "number":
		value = n.I("", "number")
	case "double":
		value = n.F64("", "double")
	default:
		log.Fatal("Bad attribute type of %s", name)
	}

	return &Attribute{Name: name, Value: value}
}

func (this *parser) parseSizePolicy(n *xmlx.Node) *QSizePolicy {
	return &QSizePolicy{
		HSizeType: n.As("", "hsizetype"),
		VSizeType: n.As("", "vsizetype"),
		HorStretch: n.I("", "horstretch"),
		VerStretch: n.I("", "verstretch"),
	}
}

func (this *parser) elementChildren(n *xmlx.Node) []*xmlx.Node {
	ret := []*xmlx.Node{}
	for _, ch := range n.Children {
		if ch.Type == xmlx.NT_ELEMENT {
			ret = append(ret, ch)
		}
	}
	return ret
}

func (this *parser) parseLayoutItem(n *xmlx.Node) *QLayoutItem {
	row := n.Ai("", "row")
	column := n.Ai("", "column")
	rowSpan := n.Ai("", "rowspan")
	colSpan := n.Ai("", "colspan")
	alignment := n.S("", "alignment")

	children := this.elementChildren(n)
	if len(children) != 1 {
		for _, ch := range (children) {
			fmt.Println(ch)
		}

		log.Fatalf("Bad layout item")
	}

	var view interface{}

	child := children[0]
	switch child.Name.Local {
	case "layout":
		view = this.parseLayout(child)
	case "spacer":
		view = this.parseSpacer(child)
	case "widget":
		view = this.parseWidget(child)
	default:
		log.Fatalf("Bad layout item child type %s", child.Name.Local)
	}
	return &QLayoutItem{
		Row: row,
		Column: column,
		Rowspan: rowSpan,
		Colspan: colSpan,
		Alignment: alignment,
		View: view,
	}
}

func (this *parser) parseSpacer(n *xmlx.Node) *QSpacer {
	name := n.As("", "name")
	children := this.elementChildren(n)
	properties := make([]*Property, len(children))

	for i, ch := range children {
		if ch.Name.Local != "property" {
			log.Fatalf("Bad child type %s of spacer", ch.Name.Local)
		}

		properties[i] = this.parseProperty(ch)
	}

	return &QSpacer{Name: name, Properties: properties}
}

func (this *parser) parseRow(n *xmlx.Node) *Row {
	children := n.SelectNodesDirect("", "property")
	props := make([]*Property, len(children))
	for i, ch := range children {
		props[i] = this.parseProperty(ch)
	}
	return &Row{Props: props}
}

func (this *parser) parseColumn(n *xmlx.Node) *Column {
	children := n.SelectNodesDirect("", "property")
	props := make([]*Property, len(children))
	for i, ch := range children {
		props[i] = this.parseProperty(ch)
	}
	return &Column{Props: props}
}

func (this *parser) parseLayout(n *xmlx.Node) *QLayout {
	name := n.As("", "name")
	class := n.As("", "class")
	stretch := n.As("", "stretch")
	rowStretch := n.As("", "rowstretch")
	columnStretch := n.As("", "columnstretch")
	rowMinimumHeight := n.As("", "rowminimumheight")
	columnMinimumWidth := n.As("", "columnminimumwidth")


	properties := []*Property{}
	items := []*QLayoutItem{}
	attributes := []*Property{}

	children := this.elementChildren(n)

	for _, ch := range children {
		switch ch.Name.Local {
		case "property":
			properties = append(properties, this.parseProperty(ch))
		case "item":
			items = append(items, this.parseLayoutItem(ch))
		case "attribute":
			attributes = append(attributes, this.parseProperty(ch))
		default:
			log.Fatalf("Bad child type %s of layout", ch.Name.Local)
		}
	}

	return &QLayout{
		Class: class,
		Name: name,
		Stretch: stretch,
		RowStretch: rowStretch,
		ColumnStretch: columnStretch,
		RowMinimumHeight: rowMinimumHeight,
		ColumnMinimumWidth: columnMinimumWidth,
		Properties: properties,
		Items: items,
		Attributes: attributes,
	}
}

func (this *parser) parseWidgetItem(n *xmlx.Node) *QWidgetItem {
	propNodes := n.SelectNodesDirect("", "property")
	props := make([]*Property, len(propNodes))
	for i, ch := range propNodes {
		props[i] = this.parseProperty(ch)
	}

	itemNodes := n.SelectNodesDirect("", "item")
	items := make([]*QWidgetItem, len(itemNodes))
	for i, ch := range itemNodes {
		items[i] = this.parseWidgetItem(ch)
	}

	return &QWidgetItem{
		Props: props,
		Items: items,
		Row: n.Ai("", "row"),
		Column: n.Ai("", "column"),
	}
}

func (this *parser) parseAction(n *xmlx.Node) *Action {
	propNodes := n.SelectNodesDirect("", "property")
	props := make([]*Property, len(propNodes))
	for i, ch := range propNodes {
		props[i] = this.parseProperty(ch)
	}

	attrNodes := n.SelectNodesDirect("", "attribute")
	attrs := make([]*Property, len(attrNodes))
	for i, ch := range attrNodes {
		attrs[i] = this.parseProperty(ch)
	}

	return &Action{
		Props: props,
		Attributes: attrs,
		Name: n.As("", "name"),
		Menu: n.As("", "menu"),
	}
}

func (this *parser) parseActionGroup(n *xmlx.Node) *ActionGroup {
	propNodes := n.SelectNodesDirect("", "property")
	props := make([]*Property, len(propNodes))
	for i, ch := range propNodes {
		props[i] = this.parseProperty(ch)
	}

	attrNodes := n.SelectNodesDirect("", "attribute")
	attrs := make([]*Property, len(attrNodes))
	for i, ch := range attrNodes {
		attrs[i] = this.parseProperty(ch)
	}

	actionNodes := n.SelectNodesDirect("", "action")
	actions := make([]*Action, len(actionNodes))
	for i, ch := range actionNodes {
		actions[i] = this.parseAction(ch)
	}

	actionGroupNodes := n.SelectNodesDirect("", "actiongroup")
	actionGroups := make([]*ActionGroup, len(actionGroupNodes))
	for i, ch := range actionGroupNodes {
		actionGroups[i] = this.parseActionGroup(ch)
	}
	return &ActionGroup{
		Props: props,
		Attributes: attrs,
		Actions: actions,
		ActionGroups: actionGroups,

		Name: n.As("", "name"),
	}
}

func (this *parser) parseActionRef(n *xmlx.Node) *ActionRef {
	return &ActionRef{Name: n.As("", "name")}
}

func (this *parser) parseWidget(n *xmlx.Node) *QWidget {
	name := n.As("", "name")
	class := n.As("", "class")

	properties := []*Property{}
	attributes := []*Attribute{}

	rows := []*Row{}
	columns := []*Column{}
	items := []*QWidgetItem{}

	var layout *QLayout
	widgets := []*QWidget{}
	actions := []*Action{}
	actionGroups := []*ActionGroup{}
	addActions := []*ActionRef{}
	zorders := []string{}

	children := this.elementChildren(n)
	for _, ch := range children {
		switch ch.Name.Local {
		case "property":
			properties = append(properties, this.parseProperty(ch))
		case "attribute":
			attributes = append(attributes, this.parseAttribute(ch))

		case "row":
			rows = append(rows, this.parseRow(ch))
		case "column":
			columns = append(columns, this.parseColumn(ch))
		case "item":
			items = append(items, this.parseWidgetItem(ch))

		case "layout":
			layout = this.parseLayout(ch)
		case "widget":
			widgets = append(widgets, this.parseWidget(ch))
		case "action":
			actions = append(actions, this.parseAction(ch))
		case "actiongroup":
			actionGroups = append(actionGroups, this.parseActionGroup(ch))
		case "addaction":
			addActions = append(addActions, this.parseActionRef(ch))
		case "zorder":
			zorders = append(zorders, ch.GetValue())
		default:
			log.Fatalf("Bad child type %s of layout, parent name: %s", ch.Name.Local, name)
		}
	}

	if len(attributes) > 0 {
		log.Printf("widget name: %s has %d attributes", name, len(attributes))
	}

	if layout != nil {
		if len(widgets) > 0 {
			log.Fatalf("MUST no child if layout set. widget name: %s", name)
		}
	}

	return &QWidget{
		Class: class,
		Name: name,
		Properties: properties,
		Attributes: attributes,

		Rows: rows,
		Columns: columns,
		Items: items,

		Widgets: widgets,
		Layout: layout,
		Actions: actions,
		ActionsGroups: actionGroups,
		AddActions: addActions,
		ZOrders: zorders,
	}
}

func (this *parser) parseProperty(n *xmlx.Node) *Property {
	name := n.As("", "name")
	children := this.elementChildren(n)

	if len(children) != 1 {
		log.Fatalf("Bad property %s", name)
	}
	var value interface{}
	child := children[0]
	switch child.Name.Local {
	case "bool":
		value = n.B("", "bool")
	case "color":
		value = this.parseColor(child)
	case "cstring":
		value = n.S("", "cstring")
	case "cursor":
		value = &Cursor{ Value:n.I("", "cursor")}
	case "cursorShape":
		value = &CursorShape{Value: n.S("", "cursorShape")}
	case "cursorshape":
		value = &CursorShape{Value: n.S("", "cursorshape")}
	case "enum":
		value = this.parseEnum(child)
	case "font":
		value = this.parseFont(child)
	case "iconset":
		log.Fatalf("iconset not support now. %s", child)
	case "pixmap":
		log.Fatalf("pixmap not support now. %s", child)
	case "palette":
		value = this.parsePalette(child)
	case "point":
		value = this.parsePoint(child)
	case "rect":
		value = this.parseRect(child)
	case "set":
		value = this.parseSet(child)
	case "locale":
		value = this.parseLocale(child)
	case "sizepolicy":
		value = this.parseSizePolicy(child)
	case "size":
		value = this.parseSize(child)
	case "string":
		value = this.parseString(child)
	case "stringlist":
		value = this.parseStringList(child)
	case "number":
		value = n.I("", "number")
	case "float":
		value = n.F32("", "double")
	case "double":
		value = n.F64("", "double")
	case "date":
		fmt.Println(child)
		value = this.parseDate(child)
	case "time":
		value = this.parseTime(child)
	case "datetime":
		value = this.parseDateTime(child)
	case "pointf":
		value = this.parsePointF(child)
	case "rectf":
		value = this.parseRectF(child)
	case "sizef":
		value = this.parseSizeF(child)
	case "longlong":
		value = n.I64("", "longlong")
	case "char":
		value = this.parseChar(child)
	case "url":
		value = this.parseUrl(child)
	case "ulonglong":
		value = n.U64("", "ulonglong")
	case "brush":
		value = this.parseBrush(child)
	default:
		log.Fatalf("Bad property type %s of %v", child.Name.Local, child)
	}
	return &Property{Name: name, Value: value, StdSet: n.Ab("", "stdset")}
}

func (this *parser) Parse() error {
	rootNode := this.doc.Root
	this.class = rootNode.S("", "class")

	widgetRoot := rootNode.SelectNode("", "widget")
	this.widget = this.parseWidget(widgetRoot)

	layoutDefault := rootNode.SelectNode("", "layoutDefault")
	if layoutDefault != nil {
		this.layoutDefault = &LayoutDefault{Margin: layoutDefault.Ai("", "margin"), Spacing: layoutDefault.Ai("", "spacing")}
	}

	// Parse tabstops
	tabStopsRoot := rootNode.SelectNode("", "tabstops")
	if tabStopsRoot != nil {
		for _, ch := range tabStopsRoot.SelectNodesDirect("", "tabstop") {
			this.tabStops = append(this.tabStops, ch.GetValue())
		}
	}

	// Parse button groups
	buttonGroupsRoot := rootNode.SelectNode("", "buttongroups")
	if buttonGroupsRoot != nil {
		for _, ch := range buttonGroupsRoot.SelectNodesDirect("", "buttongroup") {
			this.buttonGroups = append(this.buttonGroups, ch.As("", "name"))
		}
	}

	return nil
}

