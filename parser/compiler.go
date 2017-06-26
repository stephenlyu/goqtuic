package parser

import (
	"fmt"
	"strings"
	"github.com/z-ray/log"
	"path/filepath"
	"github.com/huandu/xstrings"
)

type compiler struct {
	*parser

	Imports map[string]bool

	FontDefined bool
	SizePolicyDefined bool
	PaletteDefined bool

	VariableCodes []string
	SetupUICodes []string
	TranslateCodes []string
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

func (this *compiler) toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
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

func (this *compiler) setProperties(name string, props []*Property) {
	for _, prop := range props {
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
			// TODO:
		case *Enum:
			enum, _ := prop.Value.(*Enum)
			valueStr = this.enumToString(enum.Value)
			this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
		case *QFont:
			if !this.FontDefined {
				this.addImport("gui")
				this.addSetupUICode("var font *gui.QFont")
			}
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
			this.addSetupUICode(fmt.Sprintf("font.SetItalic(%s)", boolToString(font.Italic)))
			this.addSetupUICode(fmt.Sprintf("font.SetBold(%s)", boolToString(font.Bold)))
			this.addSetupUICode(fmt.Sprintf("font.SetUnderline(%s)", boolToString(font.Underline)))
			this.addSetupUICode(fmt.Sprintf("font.SetStrikeout(%s)", boolToString(font.Strikeout)))
			this.addSetupUICode(fmt.Sprintf("font.SetKerning(%s)", boolToString(font.Kerning)))
			this.addSetupUICode(fmt.Sprintf("font.SetStyleStrategy(gui.%s)", strings.Replace(font.StyleStrategy, ":", "_", -1)))
		case *QPalette:
			// TODO:
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
			valueStr = fmt.Sprintf("core.NewQLocale2(%s, %s)", locale.Language, locale.Country)
			this.addSetupUICode(fmt.Sprintf("%s.Set%s(%s)", name, this.toCamelCase(prop.Name), valueStr))
		case *QSizePolicy:
			sizePolicy := prop.Value.(*QSizePolicy)
			this.addImport("widgets")
			if !this.SizePolicyDefined {
				this.addSetupUICode("var sizePolicy *widgets.QSizePolicy")
			}
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
				this.addTranslateCode(fmt.Sprintf("%s.Set%s(_translate(\"%s\", \"%s\", \"\", -1)", name, this.toCamelCase(prop.Name), this.widget.Name, str.Value))
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
			//TODO:
		}
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

func (this *compiler) GenerateCode(packageName string, goFile string) error {
	this.addSetupUICode(fmt.Sprintf("%s.SetObjectName(\"%s\")", this.widget.Name, this.widget.Name))
	this.setProperties(this.widget.Name, this.widget.Properties)

	className := this.getClassName()
	widgetName := this.widget.Name

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
		className,
		widgetName,
		this.widget.Class,
		this.getTranslateCodes(indent))

	fmt.Println(code)

	return nil
}
