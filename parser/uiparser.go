package parser

import (
	xmlx "github.com/jteeuwen/go-pkg-xmlx"
	"log"
	"fmt"
)

type parser struct {
	doc *xmlx.Document

	class string

	buttonGroups []string
	tabStops []string

	widget *QWidget
}

func NewParser(uiFile string) (error, *parser) {
	ret := &parser{}
	ret.doc = xmlx.New()
	err := ret.doc.LoadFile(uiFile, nil)
	if err != nil {
		return err, nil
	}
	return nil, ret
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

func (this *parser) parseFont(n *xmlx.Node) *QFont {
	return &QFont{n.I("", "pointsize")}
}

func (this *parser) parseEnum(n *xmlx.Node) *Enum {
	return &Enum{n.GetValue()}
}

func (this *parser) parseSet(n *xmlx.Node) *Set {
	return &Set{n.GetValue()}
}

func (this *parser) parseDate(n *xmlx.Node) *Date {
	return &Date{Year: n.I("", "year"), Month: n.I("", "month"), Day: n.I("", "day")}
}

func (this *parser) parseAttribute(n *xmlx.Node) *Attribute {
	var value interface{}

	ch := this.elementChildren(n)[0]

	name := n.As("", "name")

	switch ch.Name.Local {
	case "string":
		value = n.S("", "string")
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
	return &QLayoutItem{Row: row, Column: column, View: view}
}

func (this *parser) parseSpacer(n *xmlx.Node) *QSpacer {
	name := n.As("", "name")
	properties := make([]*Property, len(n.Children))
	children := this.elementChildren(n)

	for i, ch := range children {
		if ch.Name.Local != "property" {
			log.Fatalf("Bad child type %s of spacer", ch.Name.Local)
		}

		properties[i] = this.parseProperty(ch)
	}

	return &QSpacer{Name: name, Properties: properties}
}

func (this *parser) parseLayout(n *xmlx.Node) *QLayout {
	name := n.As("", "name")
	class := n.As("", "class")
	stretch := n.As("", "stretch")
	properties := []*Property{}
	items := []*QLayoutItem{}
	children := this.elementChildren(n)

	for _, ch := range children {
		switch ch.Name.Local {
		case "property":
			properties = append(properties, this.parseProperty(ch))
		case "item":
			items = append(items, this.parseLayoutItem(ch))
		default:
			fmt.Println(n)
			fmt.Println(ch)
			log.Fatalf("Bad child type %s of layout", ch.Name.Local)
		}
	}

	return &QLayout{Class: class, Name: name, Stretch: stretch, Properties: properties, Items: items}
}

func (this *parser) parseWidgetItem(n *xmlx.Node) *QWidgetItem {
	ch := n.SelectNode("", "property")
	return &QWidgetItem{Prop: this.parseProperty(ch)}
}

func (this *parser) parseWidget(n *xmlx.Node) *QWidget {
	name := n.As("", "name")
	class := n.As("", "class")
	properties := []*Property{}
	widgets := []*QWidget{}
	var layout *QLayout
	items := []*QWidgetItem{}
	attributes := []*Attribute{}
	zorders := []string{}

	children := this.elementChildren(n)
	for _, ch := range children {
		switch ch.Name.Local {
		case "property":
			properties = append(properties, this.parseProperty(ch))
		case "widget":
			widgets = append(widgets, this.parseWidget(ch))
		case "layout":
			layout = this.parseLayout(ch)
		case "item":
			items = append(items, this.parseWidgetItem(ch))
		case "attribute":
			attributes = append(attributes, this.parseAttribute(ch))
		case "zorder":
			zorders = append(zorders, ch.GetValue())
		default:
			fmt.Println(n)
			fmt.Println(ch)
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

	return &QWidget{Class: class, Name: name, Properties: properties, Attributes: attributes,
		Children: widgets, Layout: layout, ZOrders: zorders}
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
	case "number":
		value = n.I("", "number")
	case "double":
		value = n.F64("", "double")
	case "string":
		value = n.S("", "string")
	case "bool":
		value = n.B("", "bool")
	case "enum":
		value = this.parseEnum(child)
	case "set":
		value = this.parseSet(child)
	case "size":
		value = this.parseSize(child)
	case "rect":
		value = this.parseRect(child)
	case "font":
		value = this.parseFont(child)
	case "sizepolicy":
		value = this.parseSizePolicy(child)
	case "date":
		value = this.parseDate(child)
	default:
		log.Fatalf("Bad property type %s of %v", child.Name.Local, child)
	}
	return &Property{Name: name, Value: value}
}

func (this *parser) Parse() error {
	rootNode := this.doc.Root
	this.class = rootNode.S("", "class")

	widgetRoot := rootNode.SelectNode("", "widget")
	this.widget = this.parseWidget(widgetRoot)

	// Parse tabstops
	tabStopsRoot := rootNode.SelectNode("", "tabstops")
	if tabStopsRoot != nil {
		for _, ch := range tabStopsRoot.SelectNodes("", "tabstop") {
			this.tabStops = append(this.tabStops, ch.GetValue())
		}
	}

	// Parse button groups
	buttonGroupsRoot := rootNode.SelectNode("", "buttongroups")
	if buttonGroupsRoot != nil {
		for _, ch := range buttonGroupsRoot.SelectNodes("", "buttongroup") {
			this.buttonGroups = append(this.buttonGroups, ch.As("", "name"))
		}
	}

	return nil
}

func (this *parser) GenerateCode(goFile string) error {
	return nil
}
