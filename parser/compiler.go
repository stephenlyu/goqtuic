package parser

import (
	"bytes"
	"fmt"
	"github.com/z-ray/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type compiler struct {
	*parser

	RootWidgetName string

	Imports map[string]bool

	FontDefined       bool
	SizePolicyDefined bool
	PaletteDefined    bool
	BrushDefined      bool
	ListItemDefined   bool
	TableItemDefined  bool
	IconDefined       bool

	SortingEnabledDefined bool

	VariableCodes  []string
	SetupUICodes   []string
	TranslateCodes []string

	AddActionCodes []string
	BuddyCodes     []string

	SetCurrentIndexCodes []string

	DefinedButtonGroups map[string]bool
	DefinedTreeItems    map[string]bool
}

// ToCamelCase can convert all lower case characters behind underscores
// to upper case character.
// Underscore character will be removed in result except following cases.
//     * More than 1 underscore.
//           "a__b" => "A_B"
//     * At the beginning of string.
//           "_a" => "_A"
//     * At the end of string.
//           "ab_" => "Ab_"
func ToCamelCase(str string) string {
	if len(str) == 0 {
		return ""
	}

	buf := &bytes.Buffer{}
	var r0, r1 rune
	var size int

	// leading '_' will appear in output.
	for len(str) > 0 {
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if r0 != '_' {
			break
		}

		buf.WriteRune(r0)
	}

	if len(str) == 0 {
		return buf.String()
	}

	buf.WriteRune(unicode.ToUpper(r0))
	r0, size = utf8.DecodeRuneInString(str)
	str = str[size:]

	for len(str) > 0 {
		r1 = r0
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if r1 == '_' && r0 != '_' {
			r0 = unicode.ToUpper(r0)
		} else {
			buf.WriteRune(r1)
		}
	}

	buf.WriteRune(r0)
	return buf.String()
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

	return nil, &compiler{
		parser:              parser,
		Imports:             make(map[string]bool),
		DefinedButtonGroups: make(map[string]bool),
		DefinedTreeItems:    make(map[string]bool),
	}
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

func (this *compiler) addAddActionCode(line string) {
	this.AddActionCodes = append(this.AddActionCodes, line)
}

func (this *compiler) addBuddyCode(line string) {
	this.BuddyCodes = append(this.BuddyCodes, line)
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
	return strings.Replace(ToCamelCase(s), "_", "", -1)
}

func (this *compiler) addImport(_import string) {
	if _, ok := this.Imports[_import]; ok {
		return
	}
	this.Imports[_import] = true
}

func (this *compiler) enumToString(enum string) string {
	parts := strings.Split(enum, "::")

	ns := parts[0]
	switch ns {
	case "Qt":
		this.addImport("core")
		return fmt.Sprintf("core.%s", strings.Replace(enum, ":", "_", -1))
	case "QDialogButtonBox":
		fallthrough
	case "QFrame":
		fallthrough
	case "QLineEdit":
		fallthrough
	case "QLayout", "QFormLayout":
		fallthrough
	case "QAbstractItemView", "QProgressBar":
		this.addImport("widgets")
		return fmt.Sprintf("widgets.%s", strings.Replace(enum, ":", "_", -1))
	}

	log.Errorf("unknown enum %s", enum)
	return ""
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

func (this *compiler) defineIcon() {
	if !this.IconDefined {
		this.addImport("gui")
		this.addSetupUICode("var icon *gui.QIcon")
		this.IconDefined = true
	}
}

func (this *compiler) defineButtonGroup(buttonGroupName string) string {
	varName := this.transVarName(buttonGroupName)
	if _, ok := this.DefinedButtonGroups[varName]; ok {
		return varName
	}

	this.addImport("widgets")
	this.addVariableCode(fmt.Sprintf("%s *widgets.QButtonGroup", varName))
	this.addSetupUICode(fmt.Sprintf("this.%s = widgets.NewQButtonGroup(%s)", varName, this.RootWidgetName))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", varName, buttonGroupName))
	this.DefinedButtonGroups[varName] = true
	return varName
}

func (this *compiler) defineTreeItem() string {
	for k, v := range this.DefinedTreeItems {
		if !v {
			this.DefinedTreeItems[k] = true
			return k
		}
	}

	this.addImport("widgets")
	varName := fmt.Sprintf("treeItem%d", len(this.DefinedTreeItems)+1)
	this.addSetupUICode(fmt.Sprintf("var %s *widgets.QTreeWidgetItem", varName))
	this.DefinedTreeItems[varName] = true
	return varName
}

func (this *compiler) undefineTreeItem(varName string) {
	v, ok := this.DefinedTreeItems[varName]
	if !ok {
		log.Fatalf("undefined tree item var %s", varName)
	}
	if !v {
		log.Fatalf("unused tree item var %s", varName)
	}
	this.DefinedTreeItems[varName] = false
}

func (this *compiler) defineSortingEnabled() {
	if !this.SortingEnabledDefined {
		this.addTranslateCode("var sortingEnabled bool")
		this.SortingEnabledDefined = true
	}
}

func (this *compiler) setProperty(name string, prop *Property) {
	this.setPropertyEx(name, "", prop)
}

func (this *compiler) translateIcon(icon *QIcon) {
	this.defineIcon()
	this.addImport("gui")
	if icon.Theme != "" {
		this.addSetupUICode(fmt.Sprintf("icon = gui.QIcon_FromTheme(\"%s\")", icon.Theme))
	} else {
		this.addImport("core")
		this.addSetupUICode("icon = gui.NewQIcon()")

		if icon.NormalOff != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Normal, gui.QIcon__Off)", icon.NormalOff))
		}

		if icon.NormalOn != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Normal, gui.QIcon__On)", icon.NormalOn))
		}

		if icon.DisabledOff != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Disabled, gui.QIcon__Off)", icon.DisabledOff))
		}

		if icon.DisabledOn != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Disabled, gui.QIcon__On)", icon.DisabledOn))
		}

		if icon.ActiveOff != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Active, gui.QIcon__Off)", icon.ActiveOff))
		}

		if icon.ActiveOff != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Active, gui.QIcon__On)", icon.ActiveOn))
		}

		if icon.SelectedOff != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Selected, gui.QIcon__Off)", icon.SelectedOff))
		}

		if icon.SelectedOn != "" {
			this.addSetupUICode(fmt.Sprintf("icon.AddPixmap(gui.NewQPixmap5(\"%s\", \"\", core.Qt__AutoColor), gui.QIcon__Selected, gui.QIcon__On)", icon.SelectedOn))
		}
	}
}

func (this *compiler) setPropertyEx(name string, paramPrefix string, prop *Property) {
	var valueStr string
	switch prop.Value.(type) {
	case bool:
		v, _ := prop.Value.(bool)
		valueStr = boolToString(v)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QColor:
		color := prop.Value.(*QColor)
		this.addImport("gui")
		valueStr = fmt.Sprintf("gui.NewQColor3(%d, %d, %d, %d)",
			color.Red,
			color.Green,
			color.Blue,
			color.Alpha)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case string:
		switch prop.Name {
		case "buddy":
			valueStr, _ := prop.Value.(string)
			this.addBuddyCode(fmt.Sprintf("%s.Set%s(%sthis.%s)", name, this.toCamelCase(prop.Name), paramPrefix, this.transVarName(valueStr)))
		default:
			log.Errorf("cstring property %s not supported", prop.Name)
		}
	case *Cursor:
	// TODO:
	case *CursorShape:
		cursorShape, _ := prop.Value.(*CursorShape)
		this.addImport("core")
		this.addImport("gui")
		valueStr = fmt.Sprintf("gui.NewQCursor2(core.Qt__%s)", cursorShape.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *Enum:
		enum, _ := prop.Value.(*Enum)
		valueStr = this.enumToString(enum.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QFont:
		this.defineFont()
		this.addSetupUICode("font = gui.NewQFont()")
		font := prop.Value.(*QFont)
		if font.Family != "" {
			this.addSetupUICode(fmt.Sprintf("font.SetFamily(\"%s\")", font.Family))
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
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%sfont)", name, this.toCamelCase(prop.Name), paramPrefix))
	case *QPixmap:
		this.addImport("gui")
		this.addImport("core")
		pixmap := prop.Value.(*QPixmap)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(gui.NewQPixmap5(%s\"%s\", \"\", core.Qt__AutoColor))", name, this.toCamelCase(prop.Name), paramPrefix, pixmap.Value))
	case *QIcon:
		icon := prop.Value.(*QIcon)
		this.translateIcon(icon)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%sicon)", name, this.toCamelCase(prop.Name), paramPrefix))
	case *QPalette:
		palette := prop.Value.(*QPalette)
		this.definePalette()
		this.defineBrush()
		this.addImport("core")
		this.addImport("gui")
		this.addSetupUICode(fmt.Sprintf("palette = gui.NewQPalette()"))
		setPalette := func(groupName string, colorGroup *ColorGroup) {
			for _, item := range colorGroup.Items {
				if item.IsColor {
					log.Error("Color role required for palette")
					continue
				}

				colorRole := item.ColorRole
				brush := colorRole.Brush
				this.addSetupUICode(fmt.Sprintf("brush = gui.NewQBrush3(gui.NewQColor3(%d, %d, %d, %d), core.Qt__%s)", brush.Color.Red,
					brush.Color.Green,
					brush.Color.Blue,
					brush.Color.Alpha,
					brush.BrushStyle))
				this.addSetupUICode(fmt.Sprintf("palette.SetBrush2(gui.QPalette__%s, gui.QPalette__%s, brush)", groupName, colorRole.Role))
			}
		}
		if palette.Active != nil {
			setPalette("Active", palette.Active)
		}

		if palette.InActive != nil {
			setPalette("Inactive", palette.InActive)
		}

		if palette.Disabled != nil {
			setPalette("Disabled", palette.Disabled)
		}
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%spalette)", name, this.toCamelCase(prop.Name), paramPrefix))

	case *QPoint:
		point := prop.Value.(*QPoint)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQPoint2(%d, %d)", point.X, point.Y)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QRect:
		rect := prop.Value.(*QRect)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQRect4(%d, %d, %d, %d)", rect.X, rect.Y, rect.Width, rect.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *Set:
		set := prop.Value.(*Set)
		enums := strings.Split(set.Value, "|")
		enumStrings := make([]string, len(enums))
		for i, enum := range enums {
			enumStrings[i] = this.enumToString(enum)
		}
		valueStr = strings.Join(enumStrings, " | ")
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QLocale:
		locale := prop.Value.(*QLocale)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQLocale3(core.QLocale__%s, core.QLocale__%s)", locale.Language, locale.Country)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QSizePolicy:
		sizePolicy := prop.Value.(*QSizePolicy)
		this.defineSizePolicy()
		this.addSetupUICode(fmt.Sprintf("sizePolicy = widgets.NewQSizePolicy2(widgets.QSizePolicy__%s, widgets.QSizePolicy__%s, widgets.QSizePolicy__DefaultType)",
			sizePolicy.HSizeType, sizePolicy.VSizeType))
		this.addSetupUICode(fmt.Sprintf("sizePolicy.SetHorizontalStretch(%d)", sizePolicy.HorStretch))
		this.addSetupUICode(fmt.Sprintf("sizePolicy.SetVerticalStretch(%d)", sizePolicy.VerStretch))
		this.addSetupUICode(fmt.Sprintf("sizePolicy.SetHeightForWidth(%s.SizePolicy().HasHeightForWidth())", name))
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%ssizePolicy)", name, this.toCamelCase(prop.Name), paramPrefix))
	case *QSize:
		size := prop.Value.(*QSize)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQSize2(%d, %d)", size.Width, size.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *String:
		str := prop.Value.(*String)
		if !str.NotR {
			this.addTranslateCode(fmt.Sprintf("%s.Set%s(_translate(\"%s\", %s, \"\", -1))", name, this.toCamelCase(prop.Name), this.RootWidgetName, strconv.Quote(str.Value)))
		} else {
			this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, strconv.Quote(str.Value)))
		}
	case *StringList:
		log.Errorf("string list prop %s not supported", prop.Name)
	case int:
		valueStr = fmt.Sprintf("%d", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case float32:
		valueStr = fmt.Sprintf("%f", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case float64:
		valueStr = fmt.Sprintf("%f", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *Date:
		date := prop.Value.(*Date)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQDate3(%d, %d, %d)", date.Year, date.Month, date.Day)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *Time:
		time := prop.Value.(*Time)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQTime3(%d, %d, %d, 0)", time.Hour, time.Minute, time.Second)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *DateTime:
		datetime := prop.Value.(*DateTime)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQDateTime3(core.NewQDate3(%d, %d, %d), core.NewQTime3(%d, %d, %d, 0), core.Qt__LocalTime)",
			datetime.Year, datetime.Month, datetime.Day,
			datetime.Hour, datetime.Minute, datetime.Second)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QPointF:
		point := prop.Value.(*QPointF)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQPointF2(%d, %d)", point.X, point.Y)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QRectF:
		rect := prop.Value.(*QRectF)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQRect4(%f, %f, %f, %f)", rect.X, rect.Y, rect.Width, rect.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QSizeF:
		size := prop.Value.(*QSizeF)
		this.addImport("core")
		valueStr = fmt.Sprintf("core.NewQSize2(%f, %f)", size.Width, size.Height)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case int64:
		valueStr = fmt.Sprintf("%d", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *Char:
		log.Errorf("char prop %s not supported", prop.Name)
	case *Url:
		log.Errorf("url prop %s not supported", prop.Name)
	case uint64:
		valueStr = fmt.Sprintf("%d", prop.Value)
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s%s)", name, this.toCamelCase(prop.Name), paramPrefix, valueStr))
	case *QBrush:
		brush := prop.Value.(*QBrush)
		if brush.Color != nil {
			this.defineBrush()
			this.addImport("core")
			this.addSetupUICode(fmt.Sprintf("brush = gui.NewQBrush3(gui.NewQColor3(%d, %d, %d, %d), core.Qt__%s)", brush.Color.Red,
				brush.Color.Green,
				brush.Color.Blue,
				brush.Color.Alpha,
				brush.BrushStyle))
			this.addSetupUICode(fmt.Sprintf("%s.Set%s(%sbrush)", name, this.toCamelCase(prop.Name), paramPrefix))
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
	if lines == nil || len(lines) == 0 {
		return lines
	}

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

func (this *compiler) getBuddyCodes(indent string) string {
	return "\n" + strings.Join(this.indentLines(this.BuddyCodes, indent), "\n")
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
		mainName := baseName[:len(baseName)-len(xName)]
		return strings.Replace(fmt.Sprintf("%s%s", ToCamelCase(mainName), widgetName), "_", "", -1)
	default:
		return ToCamelCase(widgetName)
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

	this.addVariableCode(fmt.Sprintf("%s *widgets.QSpacerItem", spacerName))
	this.addSetupUICode(fmt.Sprintf("this.%s = widgets.NewQSpacerItem(%d, %d, %s, %s)", spacerName, w, h, hPolicy, vPolicy))
}

func (this *compiler) translateLayoutItem(parentWidgetName, parentName string, parentClass string, item *QLayoutItem) {
	var childType string
	var childName string
	switch item.View.(type) {
	case *QLayout:
		layout, _ := item.View.(*QLayout)
		this.translateLayout(parentWidgetName, layout)
		childType = "Layout"
		childName = this.transVarName(layout.Name)
	case *QSpacer:
		spacer, _ := item.View.(*QSpacer)
		this.translateSpacer(parentWidgetName, spacer)
		childType = "Item"
		childName = this.transVarName(spacer.Name)
	case *QWidget:
		widget, _ := item.View.(*QWidget)
		this.translateWidget(parentWidgetName, widget)
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
		this.addSetupUICode(fmt.Sprintf("%s.Set%s(%d, %s, this.%s)", parentName, childType, item.Row, role, childName))
	case "QGridLayout":
		rowSpan := item.Rowspan
		colSpan := item.Colspan
		if rowSpan == 0 {
			rowSpan = 1
		}
		if colSpan == 0 {
			colSpan = 1
		}
		switch childType {
		case "Widget":
			this.addSetupUICode(fmt.Sprintf("%s.Add%s3(this.%s, %d, %d, %d, %d, 0)", parentName, childType, childName, item.Row, item.Column, rowSpan, colSpan))
		case "Layout":
			this.addSetupUICode(fmt.Sprintf("%s.Add%s2(this.%s, %d, %d, %d, %d, 0)", parentName, childType, childName, item.Row, item.Column, rowSpan, colSpan))
		default:
			log.Errorf("QGridLayout.AddItem not support now, QLayout.AddItem used")
			this.addSetupUICode(fmt.Sprintf("%s.Add%s(this.%s)", parentName, childType, childName))
		}
	}
}

func (this *compiler) translateLayout(parentName string, layout *QLayout) {
	layoutName := this.transVarName(layout.Name)
	this.addImport("widgets")
	this.addVariableCode(fmt.Sprintf("%s *widgets.%s", layoutName, layout.Class))
	switch layout.Class {
	case "QVBoxLayout":
		fallthrough
	case "QHBoxLayout":
		this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s2(%s)", layoutName, layout.Class, parentName))
	default:
		this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s(%s)", layoutName, layout.Class, parentName))
	}
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", layoutName, layout.Name))

	leftMargin, rightMargin, topMargin, bottomMargin := 0, 0, 0, 0
	spacing := 0
	if this.layoutDefault != nil {
		leftMargin, rightMargin, topMargin, bottomMargin = this.layoutDefault.Margin, this.layoutDefault.Margin,
			this.layoutDefault.Margin, this.layoutDefault.Margin
		spacing = this.layoutDefault.Spacing
	}

	// Set Properties

	for _, prop := range layout.Properties {
		switch prop.Value.(type) {
		case int:
			value := prop.Value.(int)
			switch prop.Name {
			case "margin":
				leftMargin = value
				rightMargin = value
				topMargin = value
				bottomMargin = value
			case "leftMargin":
				leftMargin = value
			case "rightMargin":
				rightMargin = value
			case "topMargin":
				topMargin = value
			case "bottomMargin":
				bottomMargin = value
			case "spacing":
				spacing = value
			}
		case *Enum:
		}
	}
	this.addSetupUICode(fmt.Sprintf("this.%s.SetContentsMargins(%d, %d, %d, %d)", layoutName, leftMargin, topMargin, rightMargin, bottomMargin))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetSpacing(%d)", layoutName, spacing))
	for _, prop := range layout.Properties {
		switch prop.Name {
		case "margin":
		case "leftMargin":
		case "rightMargin":
		case "topMargin":
		case "bottomMargin":
		case "spacing":
		default:
			this.setProperty("this."+layoutName, prop)
		}
	}

	// Set attributes
	// TODO:

	// Translate items
	if layout.Items != nil {
		for _, item := range layout.Items {
			this.translateLayoutItem(parentName, "this."+layoutName, layout.Class, item)
		}
	}

	// Set stretch & minimum size
	setCommaSplitProp := func(propName string, stretchStr string) {
		if stretchStr != "" {
			parts := strings.Split(stretchStr, ",")
			for i, part := range parts {
				stretch, _ := strconv.ParseInt(strings.TrimSpace(part), 10, 32)
				if stretch > 0 {
					this.addSetupUICode(fmt.Sprintf("this.%s.Set%s(%d, %d)", layoutName, propName, i, stretch))
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
	varName := this.transVarName(action.Name)
	this.addImport("widgets")
	this.addVariableCode(fmt.Sprintf("%s *widgets.QAction", varName))
	this.addSetupUICode(fmt.Sprintf("this.%s = widgets.NewQAction(%s)", varName, this.RootWidgetName))
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", varName, action.Name))

	for _, prop := range action.Props {
		if prop.Name == "shortcut" {
			value, _ := prop.Value.(*String)
			this.addImport("gui")
			this.addTranslateCode(fmt.Sprintf("%s.Set%s(gui.QKeySequence_FromString(_translate(\"%s\", %s, \"\", -1), gui.QKeySequence__NativeText))", "this."+varName,
				this.toCamelCase(prop.Name), this.RootWidgetName, strconv.Quote(value.Value)))
		} else {
			this.setProperty("this."+varName, prop)
		}
	}
}

func (this *compiler) translateActionGroup(actionGroup *ActionGroup) {
	// TODO:
}

func (this *compiler) translateActionRef(parentName string, parentClass string, actionRef *ActionRef) {
	if actionRef.Name == "separator" {
		this.addAddActionCode(fmt.Sprintf("this.%s.AddSeparator()", parentName))
	} else {
		switch parentClass {
		case "QToolBar":
			fallthrough
		case "QMenu":
			this.addAddActionCode(fmt.Sprintf("this.%s.QWidget.AddAction(this.%s)", parentName, this.transVarName(actionRef.Name)))
		case "QMenuBar":
			this.addAddActionCode(fmt.Sprintf("this.%s.QWidget.AddAction(this.%s.MenuAction())", parentName, this.transVarName(actionRef.Name)))
		default:
			log.Errorf("%s action not supported")
		}
	}
}

func (this *compiler) translateZOrder(zorder string) {
	widgetName := this.transVarName(zorder)
	this.addSetupUICode(fmt.Sprintf("this.%s.Raise()", widgetName))
}

func (this *compiler) translateComboBox(widget *QWidget) {
	if widget.Items != nil {
		widgetName := this.transVarName(widget.Name)
		for i, item := range widget.Items {
			var hasIcon bool
			for _, prop := range item.Props {
				if prop.Name == "icon" {
					icon := prop.Value.(*QIcon)
					this.translateIcon(icon)
					hasIcon = true
					break
				}
			}
			if hasIcon {
				this.addSetupUICode(fmt.Sprintf("this.%s.AddItem2(icon, \"\", core.NewQVariant())", widgetName))
			} else {
				this.addSetupUICode(fmt.Sprintf("this.%s.AddItem(\"\", core.NewQVariant())", widgetName))
			}

			for _, prop := range item.Props {
				if prop.Name != "text" {
					if prop.Name != "icon" {
						log.Errorf("unknown combobox item property %s", prop.Name)
					}
					continue
				}
				value, _ := prop.Value.(*String)
				this.addTranslateCode(fmt.Sprintf("this.%s.SetItemText(%d, _translate(\"%s\", %s, \"\", -1))", widgetName, i, this.RootWidgetName, strconv.Quote(value.Value)))
			}
		}
	}
}

func (this *compiler) translateListWidget(widget *QWidget) {
	if widget.Items == nil {
		return
	}

	widgetName := this.transVarName(widget.Name)

	if widget.Items != nil && len(widget.Items) > 0 {
		this.defineListItem()

		this.defineSortingEnabled()
		this.addTranslateCode(fmt.Sprintf("sortingEnabled = this.%s.IsSortingEnabled()", widgetName))
		for i, item := range widget.Items {
			this.addSetupUICode("listItem = widgets.NewQListWidgetItem(nil, 0)")
			this.addSetupUICode(fmt.Sprintf("this.%s.AddItem2(listItem)", widgetName))
			for _, prop := range item.Props {
				if prop.Name != "text" {
					log.Errorf("unknown list widget item property %s", prop.Name)
					continue
				}
				value, _ := prop.Value.(*String)

				this.addTranslateCode(fmt.Sprintf("this.%s.Item(%d).SetText(_translate(\"%s\", %s, \"\", -1))", widgetName, i, this.RootWidgetName, strconv.Quote(value.Value)))
			}
		}
		this.addTranslateCode(fmt.Sprintf("this.%s.SetSortingEnabled(sortingEnabled)", widgetName))
	}

}

func (this *compiler) translateTableWidget(widget *QWidget) {
	widgetName := this.transVarName(widget.Name)
	this.addTranslateCode(fmt.Sprintf("this.%s.SetColumnCount(%d)", widgetName, len(widget.Columns)))
	this.addTranslateCode(fmt.Sprintf("this.%s.SetRowCount(%d)", widgetName, len(widget.Rows)))

	if len(widget.Rows) > 0 {
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

				this.addTranslateCode(fmt.Sprintf("this.%s.VerticalHeaderItem(%d).SetText(_translate(\"%s\", %s, \"\", -1))", widgetName, i, this.RootWidgetName, strconv.Quote(value.Value)))
			}
		}
	}

	if len(widget.Columns) > 0 {
		this.defineTableItem()
		for i, column := range widget.Columns {
			this.addSetupUICode("tableItem = widgets.NewQTableWidgetItem(0)")
			this.addSetupUICode(fmt.Sprintf("this.%s.SetHorizontalHeaderItem(%d, tableItem)", widgetName, i))
			for _, prop := range column.Props {
				if prop.Name != "text" {
					log.Errorf("unknown table widget header item property %s", prop.Name)
					continue
				}
				value, _ := prop.Value.(*String)

				this.addTranslateCode(fmt.Sprintf("this.%s.HorizontalHeaderItem(%d).SetText(_translate(\"%s\", %s, \"\", -1))", widgetName, i, this.RootWidgetName, strconv.Quote(value.Value)))
			}
		}
	}

	if len(widget.Items) > 0 {
		this.defineTableItem()
		this.defineSortingEnabled()
		this.addTranslateCode(fmt.Sprintf("sortingEnabled = this.%s.IsSortingEnabled()", widgetName))

		for _, item := range widget.Items {
			this.addSetupUICode("tableItem = widgets.NewQTableWidgetItem(0)")
			this.addSetupUICode(fmt.Sprintf("this.%s.SetItem(%d, %d, tableItem)", widgetName, item.Row, item.Column))
			for _, prop := range item.Props {
				if prop.Name != "text" {
					log.Errorf("unknown table widget header item property %s", prop.Name)
					continue
				}
				value, _ := prop.Value.(*String)

				this.addTranslateCode(fmt.Sprintf("this.%s.Item(%d, %d).SetText(_translate(\"%s\", %s, \"\", -1))", widgetName, item.Row, item.Column, this.RootWidgetName, strconv.Quote(value.Value)))
			}
		}
		this.addTranslateCode(fmt.Sprintf("this.%s.SetSortingEnabled(sortingEnabled)", widgetName))
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
			this.addSetupUICode(fmt.Sprintf("this.%s.%s().Set%s(%d)", widgetName, funcName, propName, value))
		case float64:
			log.Errorf("double attribute not support by table widget ")
		case bool:
			value := attr.Value.(bool)
			this.addSetupUICode(fmt.Sprintf("this.%s.%s().Set%s(%s)", widgetName, funcName, propName, boolToString(value)))
		}
	}
}

func (this *compiler) translateTreeItemProps(callObject string, varName string, item *QWidgetItem) {
	column := -1
	for _, prop := range item.Props {
		switch prop.Name {
		case "text":
			column++
			value := prop.Value.(*String)
			if value.Value != "" {
				this.addTranslateCode(fmt.Sprintf("%s.SetText(%d, _translate(\"%s\", %s, \"\", -1))", callObject, column, this.RootWidgetName, strconv.Quote(value.Value)))
			}
			continue
		}

		this.setPropertyEx(varName, fmt.Sprintf("%d, ", column), prop)
	}
}

func (this *compiler) needDefineTreeItemVar(item *QWidgetItem) bool {
	if len(item.Items) > 0 {
		return true
	}

	for _, prop := range item.Props {
		if prop.Name != "text" {
			return true
		}
	}
	return false
}

func (this *compiler) translateTreeWidgetItem(callObject string, parentName string, item *QWidgetItem) {
	if item.Items == nil || len(item.Items) == 0 {
		return
	}
	for i, childItem := range item.Items {
		if this.needDefineTreeItemVar(childItem) {
			varName := this.defineTreeItem()

			this.addSetupUICode(fmt.Sprintf("%s = widgets.NewQTreeWidgetItem6(%s, 0)", varName, parentName))
			childCallObject := fmt.Sprintf("%s.Child(%d)", callObject, i)
			this.translateTreeItemProps(childCallObject, varName, childItem)
			this.translateTreeWidgetItem(childCallObject, varName, childItem)

			this.undefineTreeItem(varName)
		} else {
			this.addSetupUICode(fmt.Sprintf("widgets.NewQTreeWidgetItem6(%s, 0)", parentName))
		}
	}
}

func (this *compiler) translateTreeWidget(widget *QWidget) {
	widgetName := this.transVarName(widget.Name)

	this.defineSortingEnabled()
	this.addTranslateCode(fmt.Sprintf("sortingEnabled = this.%s.IsSortingEnabled()", widgetName))

	if widget.Columns != nil && len(widget.Columns) > 0 {
		varName := this.defineTreeItem()
		this.addSetupUICode(fmt.Sprintf("%s = widgets.NewQTreeWidgetItem3(this.%s, 0)", varName, widgetName))
		this.addSetupUICode(fmt.Sprintf("this.%s.SetHeaderItem(%s)", widgetName, varName))
		for i, column := range widget.Columns {
			for _, prop := range column.Props {
				if prop.Name != "text" {
					log.Errorf("unknown tree widget header item property %s", prop.Name)
					continue
				}
				value, _ := prop.Value.(*String)

				this.addTranslateCode(fmt.Sprintf("this.%s.HeaderItem().SetText(%d, _translate(\"%s\", %s, \"\", -1))", widgetName, i, this.RootWidgetName, strconv.Quote(value.Value)))
			}
		}
		this.undefineTreeItem(varName)
	}
	this.addTranslateCode(fmt.Sprintf("this.%s.SetSortingEnabled(sortingEnabled)", widgetName))

	if widget.Items != nil && len(widget.Items) > 0 {
		for i, item := range widget.Items {
			if this.needDefineTreeItemVar(item) {
				varName := this.defineTreeItem()
				this.addSetupUICode(fmt.Sprintf("%s = widgets.NewQTreeWidgetItem3(this.%s, 0)", varName, widgetName))
				callObject := fmt.Sprintf("this.%s.TopLevelItem(%d)", widgetName, i)
				this.translateTreeItemProps(callObject, varName, item)
				this.translateTreeWidgetItem(callObject, varName, item)

				this.undefineTreeItem(varName)
			} else {
				this.addSetupUICode(fmt.Sprintf("widgets.NewQTreeWidgetItem3(this.%s, 0)", widgetName))
			}
		}
	}

	// Translate attributes

	for _, attr := range widget.Attributes {
		var propName string
		if strings.HasSuffix(attr.Name, "ShowSortIndicator") {
			propName = "SortIndicatorShown"
		} else {
			propName = attr.Name[len("Header"):]
		}
		switch attr.Value.(type) {
		case string:
			log.Errorf("string attribute not support by tree widget ")
		case int:
			value := attr.Value.(int)
			this.addSetupUICode(fmt.Sprintf("this.%s.Header().Set%s(%d)", widgetName, propName, value))
		case float64:
			log.Errorf("double attribute not support by table widget ")
		case bool:
			value := attr.Value.(bool)
			this.addSetupUICode(fmt.Sprintf("this.%s.Header().Set%s(%s)", widgetName, propName, boolToString(value)))
		}
	}
}

func (this *compiler) convertLineWidget(widget *QWidget) *QWidget {
	if widget.Class != "Line" {
		return widget
	}

	props := []*Property{}
	for _, prop := range widget.Properties {
		if prop.Name == "orientation" {
			props = append(props, &Property{Name: "frameShadow", Value: &Enum{Value: "QFrame::Sunken"}})
			orientation := prop.Value.(*Enum)
			if orientation.Value == "Qt::Horizontal" {
				props = append(props, &Property{Name: "frameShape", Value: &Enum{Value: "QFrame::HLine"}})
			} else {
				props = append(props, &Property{Name: "frameShape", Value: &Enum{Value: "QFrame::VLine"}})
			}
		} else {
			props = append(props, prop)
		}
	}
	return &QWidget{Name: widget.Name, Class: "QFrame", Properties: props}
}

func (this *compiler) translateWidget(parentName string, widget *QWidget) {
	widget = this.convertLineWidget(widget)

	widgetName := this.transVarName(widget.Name)
	this.addImport("widgets")
	this.addVariableCode(fmt.Sprintf("%s *widgets.%s", widgetName, widget.Class))
	switch widget.Class {
	case "QWidget":
		fallthrough
	case "QFrame":
		fallthrough
	case "QLabel":
		this.addImport("core")
		this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s(%s, core.Qt__Widget)", widgetName, widget.Class, parentName))
	case "QToolBar":
		this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s2(%s)", widgetName, widget.Class, parentName))
	default:
		this.addSetupUICode(fmt.Sprintf("this.%s = widgets.New%s(%s)", widgetName, widget.Class, parentName))
	}
	this.addSetupUICode(fmt.Sprintf("this.%s.SetObjectName(\"%s\")", widgetName, widgetName))

	// Set Properties
	for _, prop := range widget.Properties {
		if prop.Name == "currentIndex" {
			currentIndex, _ := prop.Value.(int)
			this.addSetCurrentIndexCode(fmt.Sprintf("this.%s.SetCurrentIndex(%d)", widgetName, currentIndex))
		} else {
			this.setProperty("this."+widgetName, prop)
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
		for _, attr := range widget.Attributes {
			switch attr.Name {
			case "buttonGroup":
				buttonGroupName := attr.Value.(*String)
				varName := this.defineButtonGroup(buttonGroupName.Value)
				this.addSetupUICode(fmt.Sprintf("this.%s.AddButton(this.%s, -1)", varName, widgetName))
			default:
				// TODO:
			}
		}
	}

	if widget.Layout != nil {
		this.translateLayout("this."+widgetName, widget.Layout)
	}

	if widget.Widgets != nil {
		for _, childWidget := range widget.Widgets {
			this.translateWidget("this."+widgetName, childWidget)
			childWidgetName := this.transVarName(childWidget.Name)
			switch widget.Class {
			case "QTabWidget":
				this.addSetupUICode(fmt.Sprintf("this.%s.AddTab(this.%s, \"\")", widgetName, childWidgetName))

				for _, attr := range childWidget.Attributes {
					if attr.Name == "title" {
						value := attr.Value.(*String)
						this.addTranslateCode(fmt.Sprintf("this.%s.SetTabText(this.%s.IndexOf(this.%s), _translate(\"%s\", %s, \"\", -1))",
							widgetName,
							widgetName,
							childWidgetName,
							this.RootWidgetName,
							strconv.Quote(value.Value)))
					}
				}
			case "QStackedWidget":
				this.addSetupUICode(fmt.Sprintf("this.%s.AddWidget(this.%s)", widgetName, childWidgetName))
			case "QWidget":
			case "QFrame":
			case "QSplitter":
			case "QMenuBar":
			case "QScrollArea":
				this.addSetupUICode(fmt.Sprintf("this.%s.SetWidget(this.%s)", widgetName, childWidgetName))
			default:
				log.Warnf("Should add code for %s inner widget?", widgetName)
			}
		}
	}

	if widget.Actions != nil {
		for _, action := range widget.Actions {
			this.translateAction(action)
		}
	}

	if widget.ActionsGroups != nil {
		for _, actionGroup := range widget.ActionsGroups {
			this.translateActionGroup(actionGroup)
		}
	}

	if widget.AddActions != nil {
		for _, actionRef := range widget.AddActions {
			this.translateActionRef(widgetName, widget.Class, actionRef)
		}
	}

	if widget.ZOrders != nil {
		for _, zorder := range widget.ZOrders {
			this.translateZOrder(zorder)
		}
	}
}

func (this *compiler) getTabStopCodes(indent string) string {
	if len(this.tabStops) == 0 {
		return ""
	}

	lines := make([]string, len(this.tabStops)-1)

	for i, n := range this.tabStops {
		if i == len(this.tabStops)-1 {
			break
		}

		next := this.tabStops[i+1]
		lines[i] = fmt.Sprintf("%s%s.SetTabOrder(this.%s, this.%s)", indent, this.RootWidgetName, this.transVarName(n), this.transVarName(next))
	}
	return "\n" + strings.Join(lines, "\n")
}

func (this *compiler) parseSignature(signature string) (name string, params []string) {
	i := strings.Index(signature, "(")
	j := strings.Index(signature, ")")

	name = strings.TrimSpace(signature[:i])
	argString := strings.TrimSpace(signature[i+1 : j])
	if argString != "" {
		params = strings.Split(argString, ",")
	}
	return
}

func (this *compiler) getConnectionCodes(indent string) string {
	if len(this.connections) == 0 {
		return ""
	}

	var lines []string

outer:
	for _, n := range this.connections {
		var sender, receiver string

		sender = this.transVarName(n.Sender)
		receiver = this.transVarName(n.Receiver)

		if sender != this.RootWidgetName {
			sender = "this." + sender
		}

		if receiver != this.RootWidgetName {
			receiver = "this." + receiver
		} else {
			receiver = "this"
		}

		signal, signalParams := this.parseSignature(n.Signal)
		slot, slotParams := this.parseSignature(n.Slot)

		// Check params

		if len(slotParams) > len(signalParams) {
			log.Errorf("%s.%s and %s.%s argument mismatched!!!", n.Sender, n.Signal, n.Receiver, n.Slot)
			continue
		}
		for i, paramType := range slotParams {
			signalParamType := signalParams[i]
			println(signalParamType, paramType)
			if paramType != signalParamType {
				log.Errorf("%s.%s and %s.%s argument type mismatched!!!", n.Sender, n.Signal, n.Receiver, n.Slot)
				continue outer
			}
		}

		if len(signalParams) == len(slotParams) {
			lines = append(lines, fmt.Sprintf("%s%s.Connect%s(%s.%s)", indent, sender, ToCamelCase(signal), receiver, ToCamelCase(slot)))
		} else {
			// Wrap slot to fit signal prototype
			wrapperCodes := []string{}
			var addCode = func(line string) {
				wrapperCodes = append(wrapperCodes, line)
			}

			signalArgs := make([]string, len(signalParams))
			for i, paramType := range signalParams {
				signalArgs[i] = fmt.Sprintf("arg%d %s", i, paramType)
			}

			slotArgs := make([]string, len(slotParams))
			for i, _ := range slotParams {
				slotArgs[i] = fmt.Sprintf("arg%d", i)
			}

			addCode(fmt.Sprintf("func (%s) {", strings.Join(signalArgs, ", ")))
			addCode(fmt.Sprintf("%s%s%s.%s(%s)", indent, indent, receiver, ToCamelCase(slot), strings.Join(slotArgs, ", ")))
			addCode(fmt.Sprintf("%s}", indent))
			lines = append(lines, fmt.Sprintf("%s%s.Connect%s(%s)", indent, sender, ToCamelCase(signal), strings.Join(wrapperCodes, "\n")))
		}
	}
	return "\n" + strings.Join(lines, "\n")
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
			switch widget.Class {
			case "QMenuBar":
				this.addSetupUICode(fmt.Sprintf("%s.SetMenuBar(this.%s)", widgetName, this.transVarName(widget.Name)))
			case "QStatusBar":
				this.addSetupUICode(fmt.Sprintf("%s.SetStatusBar(this.%s)", widgetName, this.transVarName(widget.Name)))
			case "QToolBar":
				for _, attr := range widget.Attributes {
					if attr.Name == "toolBarArea" {
						value := attr.Value.(*Enum)
						this.addImport("core")
						this.addSetupUICode(fmt.Sprintf("%s.AddToolBar(core.Qt__%s, this.%s)", widgetName, value.Value, this.transVarName(widget.Name)))
						break
					}
				}
			case "QWidget":
				if this.widget.Class == "QMainWindow" {
					this.addSetupUICode(fmt.Sprintf("%s.SetCentralWidget(this.%s)", widgetName, this.transVarName(widget.Name)))
				}
			}
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
			this.translateActionRef(widgetName, this.widget.Class, actionRef)
		}
	}

	if this.widget.ZOrders != nil {
		for _, zorder := range this.widget.ZOrders {
			this.translateZOrder(zorder)
		}
	}

	this.SetupUICodes = append(this.SetupUICodes, this.AddActionCodes...)

	indent := "	"
	code := fmt.Sprintf(`// WARNING! All changes made in this file will be lost!
package %s

import (
%s
)

type UI%s struct {
%s
}

func (this *UI%s) SetupUI(%s *widgets.%s) {
%s%s

    this.RetranslateUi(%s)
%s%s%s
}

func (this *UI%s) RetranslateUi(%s *widgets.%s) {
    _translate := core.QCoreApplication_Translate
%s
}
`, packageName,
		this.getImports(indent),
		className,
		this.getVariableCodes(indent),
		className,
		widgetName,
		this.widget.Class,
		this.getSetupUICodes(indent),
		this.getBuddyCodes(indent),
		widgetName,
		this.getSetCurrentIndexCodes(indent),
		this.getTabStopCodes(indent),
		this.getConnectionCodes(indent),
		className,
		widgetName,
		this.widget.Class,
		this.getTranslateCodes(indent))

	dir := filepath.Dir(goFile)
	os.MkdirAll(dir, 0755)
	return ioutil.WriteFile(goFile, []byte(code), 0644)
}

func (this *compiler) needSubclassing() bool {
	if len(this.connections) == 0 {
		return false
	}

	for _, conn := range this.connections {
		if conn.Receiver == this.widget.Name {
			return true
		}
	}
	return false
}

func (this *compiler) generateSlotOverrideFunctionCode() string {
	codes := []string{}

	var addCode = func(code string) {
		codes = append(codes, code)
	}

	for _, conn := range this.connections {
		if conn.Receiver != this.widget.Name {
			continue
		}

		slot, slotParams := this.parseSignature(conn.Slot)

		delcaredArgs := make([]string, len(slotParams))
		for i, paramType := range slotParams {
			delcaredArgs[i] = fmt.Sprintf("arg%d %s", i, paramType)
		}

		callArgs := make([]string, len(slotParams))
		for i, _ := range slotParams {
			callArgs[i] = fmt.Sprintf("arg%d", i)
		}

		goSlot := ToCamelCase(slot)
		code := fmt.Sprintf(`func (this *Window) %s(%s) {
	this.%s.%s(%s)
	// TODO: Add code here
}`, goSlot,
			strings.Join(delcaredArgs, ", "),
			this.widget.Class,
			goSlot,
			strings.Join(callArgs, ", "))
		addCode(code)
	}
	return strings.Join(codes, "\n\n")
}

func (this *compiler) GenerateTestCode(goFile string, genPackage string) error {
	var uiPackage string

	if genPackage != "" {
		baseName := filepath.Base(genPackage)
		if baseName != "" {
			uiPackage = baseName + "."
		}

		genPackage = `"` + genPackage + `"`
	}

	var widgetType string = this.widget.Class[1:]
	if this.widget.Class == "QMainWindow" {
		widgetType = "Window"
	}

	var code string
	if this.needSubclassing() {
		code = fmt.Sprintf(`package main

import (
	"github.com/therecipe/qt/widgets"
	"github.com/therecipe/qt/core"
	"os"
	%s
)

//go:generate qtmoc
type Window struct {
	widgets.%s
	%sUI%s
}

func NewWidget(parent widgets.QWidget_ITF) *Window {
	window := NewWindow(parent, core.Qt__%s)

	window.SetupUI(&window.%s)
	return window
}

// Generated Override Slots

%s

// TODO: Add UI logic here
func (this *Window) TestFunction() {
	// I am a test function.
}

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	w := NewWidget(nil)
	w.Show()

	os.Exit(app.Exec())
}
`, genPackage,
			this.widget.Class,
			uiPackage,
			this.getClassName(),
			widgetType,
			this.widget.Class,
			this.generateSlotOverrideFunctionCode(),
		)
	} else {
		code = fmt.Sprintf(`package main

import (
	"github.com/therecipe/qt/widgets"
	"github.com/therecipe/qt/core"
	"os"
	%s
)

type Window struct {
	%sUI%s
	Widget *widgets.%s
}

func NewWidget(parent widgets.QWidget_ITF) *Window {
	window := &Window{
		Widget: widgets.New%s(parent, core.Qt__%s),
	}

	window.SetupUI(window.Widget)
	return window
}

// TODO: Add UI logic here
func (this *Window) TestFunction() {
	// I am a test function.
}

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	w := NewWidget(nil)
	w.Widget.Show()

	os.Exit(app.Exec())
}
`, genPackage,
			uiPackage,
			this.getClassName(),
			this.widget.Class,
			this.widget.Class,
			widgetType)
	}

	dir := filepath.Dir(goFile)
	os.MkdirAll(dir, 0755)
	return ioutil.WriteFile(goFile, []byte(code), 0644)
}
