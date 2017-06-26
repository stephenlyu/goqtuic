package parser

type compiler struct {
	*parser
}

func NewCompiler(uiFile string) (error, *compiler) {
	err, parser := NewParser(uiFile)
	if err != nil {
		return err, nil
	}

	return nil, &compiler{parser: parser}
}

func (this *compiler) GenerateCode(goFile string) error {
	return nil
}
