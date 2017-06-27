package parser

import (
	"fmt"
	"strings"
	"github.com/z-ray/log"
	"path/filepath"
	"github.com/huandu/xstrings"
	"strconv"
)

type compiler struct {
	*parser

	RootWidgetName string

	Imports map[string]bool

	FontDefined bool
	SizePolicyDefined bool
	PaletteDefined bool
	BrushDefined bool
	ListItemDefined bool
	TableItemDefined bool
	TreeItemDefined bool

	SortingEnabledDefined bool

	VariableCodes []string
	SetupUICodes []string
	TranslateCodes []string

	SetCurrentIndexCodes []string
}

func iifs(cond bool, a, b string) string {
	if cond {
		return a
	} else {
		return b
	}
}

func boolToString(v bool) string {
	return iifs(v, "true", "false")
}

func NewCompiler(uiFile string) (error, *compiler) {
	err, parser := NewParser(uiFile)
	if err != nil {
		return err, nil
	}

	return nil, &compiler{parser: parser, Imports: make(map[string]bool)}
}

func (this *compiler) addVariableCode(line string) {
	this.VariableCodes = append(this.VariableCodes, line)
}

func (this *compiler) addSetupUICode(line string) {
	this.SetupUICodes = append(this.SetupUICodes, line)
}

func (this *compiler) addTranslateCode(line string) {
	this.TranslateCodes = append(this.TranslateCodes, line)
}

func (this *compiler) addSetCurrentIndexCode(line string) {
	this.SetCurrentIndexCodes = append(this.SetCurrentIndexCodes, line)
}

func (this *compiler) toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (this *compiler) transVarName(s string) string {
	return strings.Replace(xstrings.ToCamelCase(s), "_", "", -1)
}

func (this *compiler) addImport(_import string) {
	if _, ok := this.Imports[_import]; ok {
		return
	}
	this.Imports[_import] = true
}

func (this *compiler) enumToString(enum string) string {
	if !strings.HasPrefix(enum, "Qt::") {
		log.Errorf("unknown enum %s", enum)
	}
	this.addImport("core")
	return fmt.Sprintf("core.%s", strings.Replace(enum, ":", "_", -1))
}

func (this *compiler) defineFont() {
	if !this.FontDefined {
		this.addImport("gui")
		this.addSetupUICode("var font *gui.QFont")
		this.FontDefined = true
	}
}

func (this *compiler) definePalette() {
	if !this.PaletteDefined {
		this.addImport("gui")
		this.addSetupUICode("var palette *gui.QPalette")
		this.PaletteDefined = true
	}
}

func (this *compiler) defineBrush() {
	if !this.BrushDefined {
		this.addImport("gui")
		this.addSetupUICode("var brush *gui.QBrush")
		this.BrushDefined = true
	}
}

func (this *compiler) defineSizePolicy() {
	if !this.SizePolicyDefined {
		this.addImport("widgets")
		this.addSetupUICode("var sizePolicy *widgets.QSizePolicy")
		this.SizePolicyDefined = true
	}
}

func (this *compiler) defineListItem() {
	if !this.ListItemDefined {
		this.addImport("widgets")
		this.addSetupUICode("var listItem *widgets.QListWidgetItem")
		this.ListItemDefined = true
	}
}

func (this *compiler) defineTableItem() {
	if !this.TableItemDefined {
		this.addImport("widgets")
		this.addSetupUICode("var tableItem *widgets.QTableWidgetItem")
		this.TableItemDefined = true
	}
}

func (this *compiler) defineTreeItem() {
	if !this.TreeItemDefined {
		this.addImport("widgets")
		this.addSetupUICode("var treeItem *widgets.QTreeWidgetItem")
		this.TableItemDefined = true
	}
}

func (this *compiler) defineSortingEnabled() {
	if !this.SortingEnabledDefined{
		this.addTranslateCode("var sortingEnabled bool")
		this.SortingEnabledDefined = true
	}
}

func (this *compiler) setProperty(name string, prop *Property) {
	var valueStr string
	switch prop.Value.(type) {
	case bool:
		v, _ := prop.Value.(bool)
		valueStr = boolToString(v)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QColor:
		color := prop.Value.(*QColor)
		this.addImport("gui")
		valueStr = fmt.Sprintf("gui.NewColor3(%d, %d, %d, %d)",
			color.Red,
			color.Green,
			color.Blue,
			color.Alpha)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case string:
		log.Errorf("cstring property %s not supported", prop.Name)
	case *Cursor:
	// TODO:
	case *CursorShape:
		cursorShape, _ := prop.Value.(*CursorShape)
		this.addImport("core")
		this.addImport("gui")
		valueStr = fmt.Sprintf("core.NewQCursor2(core.Qt__%s)", cursorShape.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *Enum:
		enum, _ := prop.Value.(*Enum)
		valueStr = this.enumToString(enum.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QFont:
		this.defineFont()
		this.addSetupUICode("font = gui.NewQFont()")
		font := prop.Value.(*QFont)
		if font.Family != "" {
			this.addSetupUICode(fmt.Sprintf("font.SetFamily(%s)", font.Family))
		}
		if font.PointSize != 0 {
			this.addSetupUICode(fmt.Sprintf("font.SetPointSize(%d)", font.PointSize))
		}
		if font.Weight != 0 {
			this.addSetupUICode(fmt.Sprintf("font.SetWeight(%d)", font.Weight))
		}
		if font.Italic {
			this.addSetupUICode(fmt.Sprintf("font.SetItalic(true)"))
		}
		if font.Bold {
			this.addSetupUICode(fmt.Sprintf("font.SetBold(true)"))
		}
		if font.Underline {
			this.addSetupUICode(fmt.Sprintf("font.SetUnderline(true)"))
		}
		if font.Strikeout {
			this.addSetupUICode(fmt.Sprintf("font.SetStrikeout(true)"))
		}
		if font.Kerning {
			this.addSetupUICode(fmt.Sprintf("font.SetKerning(true)"))
		}
		if font.StyleStrategy != "" {
			this.addSetupUICode(fmt.Sprintf("font.SetStyleStrategy(gui.%s)", strings.Replace(font.StyleStrategy, ":", "_", -1)))
		}
	case *QPalette:
		palette := prop.Value.(*QPalette)
		this.definePalette()
		this.defineBrush()
		this.addImport("core")
		this.addImport("gui")
		setPalette := func (groupName string, colorGroup *ColorGroup) {
			for _, item := range colorGroup.Items {
				if item.IsColor {
					log.Error("Color role required for palette")
					continue
				}

				colorRole := item.ColorRole
				brush := colorRole.Brush
				this.addSetupUICode(fmt.Sprintf("brush = gui.NewBrush3(gui.NewColor3(%d, %d, %d, %d), core.Qt__%s)", brush.Color.Red,
					brush.Color.Green,
					brush.Color.Blue,
					brush.Color.Alpha,
					brush.BrushStyle))
				this.addSetupUICode(fmt.Sprintf("palette.SetBrush2(gui.QPalette__%s, gui.QPalette__%s, brush)", groupName, colorRole.Role))
			}
		}
		if palette.Active != nil {
			setPalette("Active", palette.Active)
		} else if palette.InActive != nil {
			setPalette("Inactive", palette.InActive)
		} else if palette.Disabled != nil {
			setPalette("Disabled", palette.Disabled)
		}

	case *QPoint:
		point := prop.Value.(*QPoint)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewPoint2(%d, %d)", point.X, point.Y)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QRect:
		rect := prop.Value.(*QRect)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewRect4(%d, %d, %d, %d)", rect.X, rect.Y, rect.Width, rect.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *Set:
		set := prop.Value.(*Set)
		enums := strings.Split(set.Value, "|")
		enumStrings := make([]string, len(enums))
		for i, enum := range enums {
			enumStrings[i] = this.enumToString(enum)
		}
		valueStr = strings.Join(enumStrings, " | ")
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QLocale:
		locale := prop.Value.(*QLocale)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQLocale2(core.QLocale__%s, core.QLocale__%s)", locale.Language, locale.Country)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QSizePolicy:
		sizePolicy := prop.Value.(*QSizePolicy)
		this.defineSizePolicy()
		this.addSetupUICode(fmt.Sprintf("sizePolicy = widgets.NewQSizePolicy2(widgets.QSizePolicy__%s, widgets.QSizePolicy__%s)",
			sizePolicy.HSizeType, sizePolicy.VSizeType))
		this.addSetupUICode(fmt.Sprintf("sizePolicy.SetHorizontalStretch(%d)", sizePolicy.HorStretch))
		this.addSetupUICode(fmt.Sprintf("sizePolicy.SetVerticalStretch(%d)", sizePolicy.VerStretch))
		this.addSetupUICode(fmt.Sprintf("sizePolicy.SetHeightForWidth(%s.SizePolicy().HasHeightForWidth())", name))
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(sizePolicy)", name, this.toCamelCase(prop.Name)))
	case *QSize:
		size := prop.Value.(*QSize)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQSize2(%d, %d)", size.Width, size.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *String:
		str := prop.Value.(*String)
		if !str.NotR {
			this.addTranslateCode(fmt.Sprintf("%s.Set%s(_translate(\"%s\", \"%s\", \"\", -1)", name, this.toCamelCase(prop.Name), this.RootWidgetName, str.Value))
		} else {
			this.addSetupUICode(fmt.Sprintf("%s.Set%s(\"%s\")", name, this.toCamelCase(prop.Name), str.Value))
		}
	case *StringList:
		log.Errorf("string list prop %s not supported", prop.Name)
	case int:
		valueStr = fmt.Sprintf("%d", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case float32:
		valueStr = fmt.Sprintf("%f", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case float64:
		valueStr = fmt.Sprintf("%f", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *Date:
		date := prop.Value.(*Date)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQDate3(%d, %d, %d)", date.Year, date.Month, date.Day)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *Time:
		time := prop.Value.(*Time)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQTime3(%d, %d, %d, 0)", time.Hour, time.Minute, time.Second)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *DateTime:
		datetime := prop.Value.(*DateTime)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewDateTime3(core.NewQDate3(%d, %d, %d), core.NewTime3(%d, %d, %d, 0), core.Qt__LocalTime)",
			datetime.Year, datetime.Month, datetime.Day,
			datetime.Hour, datetime.Minute, datetime.Second)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QPointF:
		point := prop.Value.(*QPointF)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewPoint2(%d, %d)", point.X, point.Y)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QRectF:
		rect := prop.Value.(*QRectF)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewRect4(%f, %f, %f, %f)", rect.X, rect.Y, rect.Width, rect.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QSizeF:
		size := prop.Value.(*QSizeF)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQSize2(%d, %d)", size.Width, size.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case int64:
		valueStr = fmt.Sprintf("%d", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *Char:
		log.Errorf("char prop %s not supported", prop.Name)
	case *Url:
		log.Errorf("url prop %s not supported", prop.Name)
	case uint64:
		valueStr = fmt.Sprintf("%d", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
	case *QBrush:
		brush := prop.Value.(*QBrush)
		if brush.Color != nil {
			this.defineBrush()
			this.addImport("core")
			this.addSetupUICode(fmt.Sprintf("brush = gui.NewBrush3(gui.NewColor3(%d, %d, %d, %d), core.Qt__%s)", brush.Color.Red,
				brush.Color.Green,
				brush.Color.Blue,
				brush.Color.Alpha,
				brush.BrushStyle))
			this.addSetupUICode(fmt.Sprintf("%s.Set%s(brush)", name, this.toCamelCase(prop.Name)))
		} else if brush.Gradient != nil {
			//TODO:

		} else if brush.Texture != nil {
			//TODO:
		}
	}
}

func (this *compiler) setProperties(name string, props []*Property) {
	for _, prop := range props {
		this.setProperty(name, prop)
	}
}

func (this *compiler) getImports(indent string) string {
	imports := make([]string, len(this.Imports))
	i := 0
	for s, _ := range this.Imports {
		imports[i] = fmt.Sprintf("%s\"github.com/therecipe/qt/%s\"", indent, s)
		i++
	}
	return strings.Join(imports, "\n")
}

func (this *compiler) indentLines(lines []string, indent string) []string {
	result := make([]string, len(lines))
	for i, s := range lines {
		result[i] = fmt.Sprintf("%s%s", indent, s)
	}
	return result
}

func (this *compiler) getVariableCodes(indent string) string {
	return strings.Join(this.indentLines(this.VariableCodes, indent), "\n")
}

func (this *compiler) getSetupUICodes(indent string) string {
	return strings.Join(this.indentLines(this.SetupUICodes, indent), "\n")
}

func (this *compiler) getTranslateCodes(indent string) string {
	return strings.Join(this.indentLines(this.TranslateCodes, indent), "\n")
}

func (this *compiler) getSetCurrentIndexCodes(indent string) string {
	return strings.Join(this.indentLines(this.SetCurrentIndexCodes, indent), "\n")
}

func (this *compiler) getClassName() string {
	widgetName := this.widget.Name
	switch widgetName {
	case "Form":
		fallthrough
	case "Dialog":
		fallthrough
	case "MainWindow":
		baseName := filepath.Base(this.uiFile)
		xName := filepath.Ext(baseName)
		mainName := baseName[:len(baseName) - len(xName)]
		return strings.Replace(fmt.Sprintf("%s%s", xstrings.ToCamelCase(mainName), widgetName), "_", "", -1)
	default:
		return xstrings.ToCamelCase(widgetName)
	}
}

func (this *compiler) translateSpacer(parentName string, spacer *QSpacer) {
	spacerName := this.transVarName(spacer.Name)
	this.addImport("widgets")
	var w, h int
	var hPolicy, vPolicy string

	for _, prop := range spacer.Properties {
		if prop.Name == "orientation" {
			enum := prop.Value.(*Enum)
			if enum.Value == "Qt::Vertical" {
				hPolicy = "widgets.QSizePolicy__Minimum"
				vPolicy = "widgets.QSizePolicy__Expanding"
			} else {
				hPolicy = "widgets.QSizePolicy__Expanding"
				vPolicy = "widgets.QSizePolicy__Minimum"
			}
		} else if prop.Name == "sizeHint" {
			size := prop.Value.(*QSize)
			w = size.Width
			h = size.Height
		}
	}

	this.addVariableCode(fmt.Sprintf("var %s *widgets.QSpacerItem", spacer.Name))
	this.addSetupUICode(fmt.Sprintf("this.%s = widgets.NewQSpacerItem(%d, %d, %s, %s)", spacerName, w, h, hPolicy, vPolicy))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", spacerName, spacerName))
}

func (this *compiler) translateLayoutItem(parentName string, parentClass string, item *QLayoutItem) {
	var childType string
	var childName string
	switch item.View.(type) {
	case *QLayout:
		layout, _ := item.View.(*QLayout)
		this.translateLayout(parentName, layout)
		childType = "Layout"
		childName = this.transVarName(layout.Name)
	case *QSpacer:
		spacer, _ := item.View.(*QSpacer)
		this.translateSpacer(parentName, spacer)
		childType = "Item"
		childName = this.transVarName(spacer.Name)
	case *QWidget:
		widget, _ := item.View.(*QWidget)
		this.translateWidget(parentName, widget)
		childType = "Widget"
		childName = this.transVarName(widget.Name)
	}

	switch parentClass {
	case "QVBoxLayout":
		fallthrough
	case "QHBoxLayout":
		switch childType {
		case "Layout":
			this.addSetupUICode(fmt.Sprintf("%s.AddLayout(this.%s, 0)", parentName, childName))
		case "Item":
			this.addSetupUICode(fmt.Sprintf("%s.AddItem(this.%s)", parentName, childName))
		case "Widget":
			this.addSetupUICode(fmt.Sprintf("%s.AddWidget(this.%s, 0, 0)", parentName, childName))
		}
	case "QFormLayout":
		this.addImport("widgets")
		role := iifs(item.Column == 0, "widgets.QFormLayout__LabelRole", "widgets.QFormLayout__FieldRole")
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%d, %s, this.%s)", parentName, childType, role, childName))
	case "QGridLayout":
		this.addSetupUICode(fmt.Sprintf("%s.Add%s(this.%s, %d, %d, %d, %d, 0)", parentName, childType, childName, item.Row, item.Column, item.Rowspan, item.Colspan))
	}
}

func (this *compiler) translateLayout(parentName string, layout *QLayout) {
	layoutName := this.transVarName(layout.Name)
	this.addImport("widgets")
	this.addVariableCode(fmt.Sprintf("var %s *widgets.%s", layoutName, layout.Class))
	this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s(%s)", layoutName, layout.Class, parentName))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", layoutName, layoutName))

	leftMargin, rightMargin, topMargin, bottomMargin := 0, 0, 0, 0
	spacing := 0
	if this.layoutDefault != nil {
		leftMargin, rightMargin, topMargin, bottomMargin = this.layoutDefault.Margin, this.layoutDefault.Margin,
			this.layoutDefault.Margin, this.layoutDefault.Margin
		spacing = this.layoutDefault.Spacing
	}

	// Set Properties

	for _, prop := range layout.Properties {
		if prop.Name == "leftMargin" {
			leftMargin, _ = prop.Value.(int)
		} else if prop.Name == "rightMargin" {
			rightMargin, _ = prop.Value.(int)
		} else if prop.Name == "topMargin" {
			topMargin, _ = prop.Value.(int)
		} else if prop.Name == "bottomMargin" {
			bottomMargin, _ = prop.Value.(int)
		} else if prop.Name == "spacing" {
			spacing, _ = prop.Value.(int)
		}
	}
	this.addSetupUICode(fmt.Sprintf("this.%s.SetContentsMargins(%d, %d, %d, %d)", layoutName, leftMargin, topMargin, rightMargin, bottomMargin))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetSpacing(%d)", layoutName, spacing))
	for _, prop := range layout.Properties {
		if prop.Name == "leftMargin" {
		} else if prop.Name == "rightMargin" {
		} else if prop.Name == "topMargin" {
		} else if prop.Name == "bottomMargin" {
		} else if prop.Name == "spacing" {
		} else {
			this.setProperty("this." + layoutName, prop)
		}
	}

	// Set attributes
	// TODO:

	// Translate items
	if layout.Items != nil {
		for _, item := range layout.Items {
			this.translateLayoutItem("this." + layoutName, layout.Class, item)
		}
	}

	// Set stretch & minimum size
	setCommaSplitProp := func (propName string, stretchStr string) {
		if stretchStr != "" {
			parts := strings.Split(stretchStr, ",")
			for i, part := range parts {
				stretch, _ := strconv.ParseInt(strings.TrimSpace(part), 10, 32)
				if stretch > 0 {
					this.addSetupUICode(fmt.Sprintf("this.%s.Set%s(%d, %d)", propName, i, stretch))
				}
			}
		}
	}

	setCommaSplitProp("Stretch", layout.Stretch)
	setCommaSplitProp("ColumnStretch", layout.ColumnStretch)
	setCommaSplitProp("RowStretch", layout.RowStretch)
	setCommaSplitProp("RowMinimumHeight", layout.RowMinimumHeight)
	setCommaSplitProp("ColumnMinimumWidth", layout.ColumnMinimumWidth)
}

func (this *compiler) translateAction(action *Action) {

}

func (this *compiler) translateActionGroup(actionGroup *ActionGroup) {

}

func (this *compiler) translateActionRef(actionRef *ActionRef) {

}

func (this *compiler) translateZOrder(zorder string) {
	widgetName := this.transVarName(zorder)
	this.addSetupUICode(fmt.Sprintf("this.%s.Raise()", widgetName))
}

func (this *compiler) translateComboBox(widget *QWidget) {
	if widget.Items != nil {
		widgetName := this.transVarName(widget.Name)
		for _, item := range widget.Items {
			this.addSetupUICode(fmt.Sprintf("this.%s.AddItem(\"\")", widgetName))
			for _, prop := range item.Props {
				if prop.Name != "text" {
					log.Errorf("unknown combobox item property %s", prop.Name)
					continue
				}
				value, _ := prop.Value.(*String)
				this.addTranslateCode(fmt.Sprintf("this.%s.SetItemText(_translate(\"%s\", \"%s\", \"\", -1)", widgetName, this.RootWidgetName, value.Value))
			}
		}
	}
}

func (this *compiler) translateListWidget(widget *QWidget) {
	if widget.Items == nil {
		return
	}

	this.defineListItem()
	widgetName := this.transVarName(widget.Name)

	this.defineSortingEnabled()
	this.addTranslateCode(fmt.Sprintf("sortingEnabled = this.%s.IsSortingEnabled()", widgetName))

	for i, item := range widget.Items {
		this.addSetupUICode("listItem = widgets.NewQListWidgetItem(nil, 0)")
		this.addSetupUICode(fmt.Sprintf("this.%s.addItem(listItem)", widgetName))
		for _, prop := range item.Props {
			if prop.Name != "text" {
				log.Errorf("unknown list widget item property %s", prop.Name)
				continue
			}
			value, _ := prop.Value.(*String)

			this.addTranslateCode(fmt.Sprintf("this.%s.item(%d).SetText(_translate(\"%s\", \"%s\", \"\", -1)", widgetName, i, this.RootWidgetName, value.Value))
		}
	}
	this.addTranslateCode(fmt.Sprintf("this.%s.SetSortingEnabled(sortingEnabled)", widgetName))
}

func (this *compiler) translateTableWidget(widget *QWidget) {
	widgetName := this.transVarName(widget.Name)
	this.addTranslateCode(fmt.Sprintf("this.%s.SetColumnCount(%d)", widgetName, len(widget.Columns)))
	this.addTranslateCode(fmt.Sprintf("this.%s.SetRowCount(%d)", widgetName, len(widget.Rows)))

	this.defineTableItem()
	for i, row := range widget.Rows {
		this.addSetupUICode("tableItem = widgets.NewQTableWidgetItem(0)")
		this.addSetupUICode(fmt.Sprintf("this.%s.SetVerticalHeaderItem(%d, tableItem)", widgetName, i))
		for _, prop := range row.Props {
			if prop.Name != "text" {
				log.Errorf("unknown table widget header item property %s", prop.Name)
				continue
			}
			value, _ := prop.Value.(*String)

			this.addTranslateCode(fmt.Sprintf("this.%s.VerticalHeaderItem(%d).SetText(_translate(\"%s\", \"%s\", \"\", -1)", widgetName, i, this.RootWidgetName, value.Value))
		}
	}

	for i, column := range widget.Columns {
		this.addSetupUICode("tableItem = widgets.NewQTableWidgetItem(0)")
		this.addSetupUICode(fmt.Sprintf("this.%s.SetHorizontalHeaderItem(%d, tableItem)", widgetName, i))
		for _, prop := range column.Props {
			if prop.Name != "text" {
				log.Errorf("unknown table widget header item property %s", prop.Name)
				continue
			}
			value, _ := prop.Value.(*String)

			this.addTranslateCode(fmt.Sprintf("this.%s.HorizontalHeaderItem(%d).SetText(_translate(\"%s\", \"%s\", \"\", -1)", widgetName, i, this.RootWidgetName, value.Value))
		}
	}

	for _, item := range widget.Items {
		this.addSetupUICode("tableItem = widgets.NewQTableWidgetItem(0)")
		this.addSetupUICode(fmt.Sprintf("this.%s.SetItem(%d, %d, tableItem)", widgetName, item.Row, item.Column))
		for _, prop := range item.Props {
			if prop.Name != "text" {
				log.Errorf("unknown table widget header item property %s", prop.Name)
				continue
			}
			value, _ := prop.Value.(*String)

			this.addTranslateCode(fmt.Sprintf("this.%s.item(%d, %d).SetText(_translate(\"%s\", \"%s\", \"\", -1)", widgetName, item.Row, item.Column, this.RootWidgetName, value.Value))
		}
	}

	// Translate attributes

	for _, attr := range widget.Attributes {
		var propName string
		var funcName string
		if strings.HasPrefix(attr.Name, "verticalHeader") {
			if strings.HasSuffix(attr.Name, "ShowSortIndicator") {
				propName = "SortIndicatorShown"
			} else {
				propName = attr.Name[len("verticalHeader"):]
			}
			funcName = "VerticalHeader"
		} else {
			if strings.HasSuffix(attr.Name, "ShowSortIndicator") {
				propName = "SortIndicatorShown"
			} else {
				propName = attr.Name[len("horizontalHeader"):]
			}
			funcName = "HorizontalHeader"
		}
		switch attr.Value.(type) {
		case string:
			log.Errorf("string attribute not support by table widget ")
		case int:
			value := attr.Value.(int)
			this.addSetupUICode(fmt.Sprintf("this.%s.%s().set%s(%d)", widgetName, funcName, propName, value))
		case float64:
			log.Errorf("double attribute not support by table widget ")
		case bool:
			value := attr.Value.(bool)
			this.addSetupUICode(fmt.Sprintf("this.%s.%s().set%s(%s)", widgetName, funcName, propName, boolToString(value)))
		}
	}
}

func (this *compiler) translateTreeWidget(widget *QWidget) {
	// TODO:
}

func (this *compiler) translateWidget(parentName string, widget *QWidget) {
	widgetName := this.transVarName(widget.Name)
	this.addImport("widgets")
	this.addVariableCode(fmt.Sprintf("var %s *widgets.%s", widgetName, widget.Class))
	this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s(%s)", widgetName, widget.Class, parentName))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", widgetName, widgetName))

	// Set Properties
	for _, prop := range widget.Properties {
		if prop.Name == "currentIndex" {
			currentIndex, _ := prop.Value.(int)
			this.addSetCurrentIndexCode(fmt.Sprintf("this.%s.SetCurrentIndex(%d)", widgetName, currentIndex))
		} else {
			this.setProperty("this." + widgetName, prop)
		}
	}

	switch widget.Class {
	case "QComboBox":
		this.translateComboBox(widget)
	case "QListWidget":
		this.translateListWidget(widget)
	case "QTableWidget":
		this.translateTableWidget(widget)
	case "QTreeWidget":
		this.translateTreeWidget(widget)
	default:
		// Set Attributes
		// TODO:
	}

	if this.widget.Layout != nil {
		this.translateLayout("this." + widgetName, this.widget.Layout)
	}

	if this.widget.Widgets != nil {
		for _, childWidget := range widget.Widgets {
			this.translateWidget("this." + widgetName, childWidget)
			this.addSetupUICode(fmt.Sprintf("%s.addWidget(this.%s)", parentName, widgetName))
		}
	}

	if this.widget.Actions != nil {
		for _, action := range this.widget.Actions {
			this.translateAction(action)
		}
	}

	if this.widget.ActionsGroups != nil {
		for _, actionGroup := range this.widget.ActionsGroups {
			this.translateActionGroup(actionGroup)
		}
	}

	if this.widget.AddActions != nil {
		for _, actionRef := range this.widget.AddActions {
			this.translateActionRef(actionRef)
		}
	}

	if this.widget.ZOrders != nil {
		for _, zorder := range this.widget.ZOrders {
			this.translateZOrder(zorder)
		}
	}
}

func (this *compiler) GenerateCode(packageName string, goFile string) error {
	className := this.getClassName()
	widgetName := this.transVarName(this.widget.Name)
	this.RootWidgetName = widgetName

	this.addSetupUICode(fmt.Sprintf("%s.SetObjectName(\"%s\")", widgetName, widgetName))
	this.setProperties(widgetName, this.widget.Properties)

	if this.widget.Layout != nil {
		this.translateLayout(widgetName, this.widget.Layout)
	}

	if this.widget.Widgets != nil {
		for _, widget := range this.widget.Widgets {
			this.translateWidget(widgetName, widget)
		}
	}

	if this.widget.Actions != nil {
		for _, action := range this.widget.Actions {
			this.translateAction(action)
		}
	}

	if this.widget.ActionsGroups != nil {
		for _, actionGroup := range this.widget.ActionsGroups {
			this.translateActionGroup(actionGroup)
		}
	}

	if this.widget.AddActions != nil {
		for _, actionRef := range this.widget.AddActions {
			this.translateActionRef(actionRef)
		}
	}

	if this.widget.ZOrders != nil {
		for _, zorder := range this.widget.ZOrders {
			this.translateZOrder(zorder)
		}
	}

	indent := "    "
	code := fmt.Sprintf(`import (
%s
)

type UI%s struct {
%s
}

func (this *UI%s) SetupUI(%s *widgets.%s) {
%s

    this.RetranslateUi(%s)
    %s
}

func (this *UI%s) RetranslateUi(%s *widgets.%s) {
    _translate := core.QCoreApplication_Translate
%s
}
`, this.getImports(indent),
		className,
		this.getVariableCodes(indent),
		className,
		widgetName,
		this.widget.Class,
		this.getSetupUICodes(indent),
		widgetName,
		this.getSetCurrentIndexCodes(indent),
		className,
		widgetName,
		this.widget.Class,
		this.getTranslateCodes(indent))

	fmt.Println(code)

	return nil
}
