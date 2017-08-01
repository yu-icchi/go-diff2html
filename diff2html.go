package diff2html

// GetPrettyHTML Generates the html diff.
func GetPrettyHTML(input string) (string, error) {
	d := newDiff(Config{})
	err := d.Parser(input)
	if err != nil {
		return "", err
	}
	diffHTML := newSideBySide()
	return diffHTML.GenerateSideBySideHTML(d.Files)
}
