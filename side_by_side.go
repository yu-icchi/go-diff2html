package diff2html

import (
	"bytes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"html/template"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	genericColumnLineNumberTemplate = template.Must(template.New("generic-column-line-number").Parse(genericColumnLineNumber))
	genericEmptyDiffTemplate        = template.Must(template.New("generic-empty-diff").Parse(genericEmptyDiff))
	genericFilePathTemplate         = template.Must(template.New("generic-file-path").Parse(genericFilePath))
	genericLineTemplate             = template.Must(template.New("generic-line").Parse(genericLine))
	genericWrapperTemplate          = template.Must(template.New("generic-wrapper").Parse(genericWrapper))
	iconFileTemplate                = template.Must(template.New("icon-file").Parse(iconFile))
	sideBySideFileDiffTemplate      = template.Must(template.New("side-by-side-file-diff").Parse(sideBySideFileDiff))
	tagFileAddedTemplate            = template.Must(template.New("tag-file-added").Parse(tagFileAdded))
	tagFileChangedTemplate          = template.Must(template.New("tag-file-changed").Parse(tagFileChanged))
	tagFileDeletedTemplate          = template.Must(template.New("tag-file-deleted").Parse(tagFileDeleted))
	tagFileRenamedTemplate          = template.Must(template.New("tag-file-renamed").Parse(tagFileRenamed))
)

type fileHTML struct {
	Left  string
	Right string
}

func newSideBySide() *sideBySidePrinter {
	return &sideBySidePrinter{}
}

type sideBySidePrinter struct {
}

func (p *sideBySidePrinter) GenerateSideBySideHTML(files []*File) (string, error) {
	content := ""
	for _, file := range files {
		var fileHTML *fileHTML
		var err error
		if len(file.Blocks) > 0 {
			fileHTML, err = p.genSideBySideFileHTML(file)
			if err != nil {
				return "", err
			}
		} else {
			fileHTML, err = p.genEmptyDiff()
			if err != nil {
				return "", err
			}
		}

		dh, err := p.makeDiffHTML(file, fileHTML)
		if err != nil {
			return "", err
		}
		content += dh
		content += "\n"
	}

	buf := &bytes.Buffer{}
	err := genericWrapperTemplate.Execute(buf, struct {
		Content template.HTML
	}{
		Content: template.HTML(content),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *sideBySidePrinter) makeDiffHTML(file *File, diffs *fileHTML) (string, error) {

	pathHTML, err := p.makePathHTML(file)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = sideBySideFileDiffTemplate.Execute(buf, struct {
		FileHTMLID string
		FilePath   template.HTML
		Language   string
		Left       template.HTML
		Right      template.HTML
	}{
		FileHTMLID: getHTMLID(file),
		FilePath:   template.HTML(pathHTML),
		Language:   file.Language,
		Left:       template.HTML(diffs.Left),
		Right:      template.HTML(diffs.Right),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *sideBySidePrinter) makePathHTML(file *File) (string, error) {
	iconHTML, err := p.makeIconHTML()
	if err != nil {
		return "", err
	}
	tagHTML, err := p.makeTagHTML(file)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = genericFilePathTemplate.Execute(buf, struct {
		FileDiffName string
		FileIcon     template.HTML
		FileTag      template.HTML
	}{
		FileDiffName: getDiffName(file),
		FileIcon:     template.HTML(iconHTML),
		FileTag:      template.HTML(tagHTML),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *sideBySidePrinter) makeIconHTML() (string, error) {
	buf := &bytes.Buffer{}
	err := iconFileTemplate.Execute(buf, struct{}{})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *sideBySidePrinter) makeTagHTML(file *File) (string, error) {
	tagTemplate := tagFileChangedTemplate
	if file.IsRename {
		tagTemplate = tagFileRenamedTemplate
	} else if file.IsCopy {
		tagTemplate = tagFileRenamedTemplate
	} else if file.IsNew {
		tagTemplate = tagFileAddedTemplate
	} else if file.IsDeleted {
		tagTemplate = tagFileDeletedTemplate
	} else if file.NewName != file.OldName {
		tagTemplate = tagFileRenamedTemplate
	}
	buf := &bytes.Buffer{}
	err := tagTemplate.Execute(buf, struct{}{})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *sideBySidePrinter) makeSideHTML(blockHeader string) (string, error) {
	buf := &bytes.Buffer{}
	err := genericColumnLineNumberTemplate.Execute(buf, struct {
		BlockHeader  string
		Type         string
		LineClass    string
		ContentClass string
	}{
		BlockHeader:  blockHeader,
		Type:         info,
		LineClass:    "d2h-code-side-linenumber",
		ContentClass: "d2h-code-side-line",
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *sideBySidePrinter) genSideBySideFileHTML(file *File) (*fileHTML, error) {
	var err error

	fileHTML := &fileHTML{
		Left:  "",
		Right: "",
	}
	for _, block := range file.Blocks {
		if fileHTML.Left, err = p.makeSideHTML(block.Header); err != nil {
			return nil, err
		}
		if fileHTML.Right, err = p.makeSideHTML(""); err != nil {
			return nil, err
		}

		oldLines := make([]*Line, 0)
		newLines := make([]*Line, 0)

		processChangeBlock := func() error {
			// conf.matching == "none" only
			oldLen := len(oldLines)
			newLen := len(newLines)
			common := int(math.Min(float64(oldLen), float64(newLen)))
			max := int(math.Max(float64(oldLen), float64(newLen)))

			for i := 0; i < common; i++ {
				oldLine := &Line{Content: ""}
				if oldLen > i {
					oldLine = oldLines[i]
				}
				newLine := &Line{Content: ""}
				if newLen > i {
					newLine = newLines[i]
				}
				highlight := diffHighlight(oldLine.Content, newLine.Content, file.IsCombined)
				left, err := p.genSingleLineHTML(file.IsCombined, deletes, oldLine.OldNumber, highlight.First.Line, highlight.First.Prefix)
				if err != nil {
					return err
				}
				fileHTML.Left += left
				right, err := p.genSingleLineHTML(file.IsCombined, inserts, newLine.NewNumber, highlight.Second.Line, highlight.Second.Prefix)
				if err != nil {
					return err
				}
				fileHTML.Right += right
			}

			if max > common {
				oldSlice := oldLines[common:]
				newSlice := newLines[common:]
				tmpHtml, err := p.processLines(file.IsCombined, oldSlice, newSlice)
				if err != nil {
					return err
				}
				fileHTML.Left += tmpHtml.Left
				fileHTML.Right += tmpHtml.Right
			}

			oldLines = make([]*Line, 0)
			newLines = make([]*Line, 0)
			return nil
		}

		for _, line := range block.Lines {
			prefix := string(line.Content[0])
			escapedLine := line.Content[1:]

			if line.Type != inserts && (len(newLines) > 0 || (line.Type != deletes && len(oldLines) > 0)) {
				if err := processChangeBlock(); err != nil {
					return nil, err
				}
			}

			if line.Type == context {
				left, err := p.genSingleLineHTML(file.IsCombined, line.Type, line.OldNumber, escapedLine, prefix)
				if err != nil {
					return nil, err
				}
				fileHTML.Left += left
				right, err := p.genSingleLineHTML(file.IsCombined, line.Type, line.NewNumber, escapedLine, prefix)
				if err != nil {
					return nil, err
				}
				fileHTML.Right += right
			} else if line.Type == inserts && len(oldLines) == 0 {
				left, err := p.genSingleLineHTML(file.IsCombined, context, 0, "", "")
				if err != nil {
					return nil, err
				}
				fileHTML.Left += left
				right, err := p.genSingleLineHTML(file.IsCombined, line.Type, line.NewNumber, escapedLine, prefix)
				if err != nil {
					return nil, err
				}
				fileHTML.Right += right
			} else if line.Type == deletes {
				oldLines = append(oldLines, line)
			} else if line.Type == inserts && len(oldLines) > 0 {
				newLines = append(newLines, line)
			} else {
				// console.error('unknown state in html side-by-side generator');
				if err := processChangeBlock(); err != nil {
					return nil, err
				}
			}
		}

		if err := processChangeBlock(); err != nil {
			return nil, err
		}
	}

	return fileHTML, nil
}

func (p *sideBySidePrinter) processLines(isCombined bool, oldLines, newLines []*Line) (*fileHTML, error) {
	fileHTML := &fileHTML{Left: "", Right: ""}

	oldLinesLen := len(oldLines)
	newLinesLen := len(newLines)
	maxLinesNumber := int(math.Max(float64(oldLinesLen), float64(newLinesLen)))
	for i := 0; i < maxLinesNumber; i++ {
		var oldLine *Line
		if oldLinesLen > i {
			oldLine = oldLines[i]
		}
		var newLine *Line
		if newLinesLen > i {
			newLine = newLines[i]
		}

		var oldContent string
		var newContent string
		var oldPrefix string
		var newPrefix string

		if oldLine != nil {
			oldContent = oldLine.Content[1:]
			oldPrefix = oldLine.Content[0:1]
		}
		if newLine != nil {
			newContent = newLine.Content[1:]
			newPrefix = newLine.Content[0:1]
		}

		if oldLine != nil && newLine != nil {
			left, err := p.genSingleLineHTML(isCombined, oldLine.Type, oldLine.OldNumber, oldContent, oldPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Left += left
			right, err := p.genSingleLineHTML(isCombined, newLine.Type, newLine.NewNumber, newContent, newPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Right += right
		} else if oldLine != nil {
			left, err := p.genSingleLineHTML(isCombined, oldLine.Type, oldLine.OldNumber, oldContent, oldPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Left += left
			right, err := p.genSingleLineHTML(isCombined, context, 0, "", "")
			if err != nil {
				return nil, err
			}
			fileHTML.Right += right
		} else if newLine != nil {
			left, err := p.genSingleLineHTML(isCombined, context, 0, "", "")
			if err != nil {
				return nil, err
			}
			fileHTML.Left += left
			right, err := p.genSingleLineHTML(isCombined, newLine.Type, newLine.NewNumber, newContent, newPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Right += right
		} else {
			// console.error('How did it get here?');
		}
	}

	return fileHTML, nil
}

func (p *sideBySidePrinter) genSingleLineHTML(isCombined bool, lineType string, num int, content string, possiblePrefix string) (string, error) {
	lineWithoutPrefix := content
	prefix := possiblePrefix

	if prefix == "" && content != "" {
		prefix, lineWithoutPrefix = separatePrefix(isCombined, content)
	}

	lineNumberStr := ""
	if num > 0 {
		lineNumberStr = strconv.Itoa(num)
	}

	buf := &bytes.Buffer{}
	err := genericLineTemplate.Execute(buf, struct {
		Type          string
		Prefix        string
		Content       template.HTML
		LineNumberStr string
		LineClass     string
		ContentClass  string
	}{
		Type:          lineType,
		Prefix:        prefix,
		Content:       template.HTML(lineWithoutPrefix),
		LineNumberStr: lineNumberStr,
		LineClass:     "d2h-code-side-linenumber",
		ContentClass:  "d2h-code-side-line",
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (p *sideBySidePrinter) genEmptyDiff() (*fileHTML, error) {
	fileHTML := &fileHTML{}
	fileHTML.Right = ""

	buf := &bytes.Buffer{}
	err := genericEmptyDiffTemplate.Execute(buf, struct {
		Type         string
		ContentClass string
	}{
		Type:         info,
		ContentClass: "d2h-code-side-line",
	})
	if err != nil {
		return nil, err
	}

	fileHTML.Left = buf.String()
	return fileHTML, nil
}

func separatePrefix(isCombined bool, line string) (string, string) {
	prefix := ""
	lineWithoutPrefix := ""

	if isCombined {
		prefix = line[0:2]
		lineWithoutPrefix = line[2:]
	} else {
		prefix = line[0:1]
		lineWithoutPrefix = line[1:]
	}

	return prefix, lineWithoutPrefix
}

func getHTMLID(file *File) string {
	name := getDiffName(file)
	hash := 0
	for i := 0; i < len(name); i++ {
		hash = ((hash << 5) - hash) + int(name[i])
		hash |= 0
	}
	name = strconv.Itoa(hash)
	l := int(math.Min(float64(len(name)), float64(6)))
	return "d2h-" + name[:l]
}

func getDiffName(file *File) string {
	oldFilename := unifyPath(file.OldName)
	newFilename := unifyPath(file.NewName)

	if oldFilename != "" && newFilename != "" && oldFilename != newFilename && !isDevNullName(oldFilename) && !isDevNullName(newFilename) {
		return oldFilename + " → " + newFilename
	} else if newFilename != "" && !isDevNullName(newFilename) {
		return newFilename
	} else if oldFilename != "" {
		return oldFilename
	}
	return "unknown/file/path"
}

func unifyPath(path string) string {
	if path != "" {
		return strings.Replace(path, "\\", "/", 1)
	}
	return path
}

func isDevNullName(str string) bool {
	return strings.HasPrefix(str, "dev/null")
}

type Highlight struct {
	First  HighlightPart
	Second HighlightPart
}

type HighlightPart struct {
	Prefix string
	Line   string
}

func diffHighlight(diffLine1, diffLine2 string, isCombined bool) Highlight {
	prefixSize := 1
	if isCombined {
		prefixSize = 2
	}

	linePrefix1 := diffLine1[0:prefixSize]
	linePrefix2 := diffLine2[0:prefixSize]
	unprefixedLine1 := diffLine1[prefixSize:]
	unprefixedLine2 := diffLine2[prefixSize:]

	differ := diffmatchpatch.New()
	diffs := differ.DiffMain(unprefixedLine1, unprefixedLine2, true)
	diffs = differ.DiffCleanupSemantic(diffs)

	highlightedLine := ""
	for _, part := range diffs {
		elemType := ""
		if part.Type == diffmatchpatch.DiffInsert {
			elemType = "ins"
		} else if part.Type == diffmatchpatch.DiffDelete {
			elemType = "del"
		}
		if elemType != "" {
			highlightedLine += "<" + elemType + ">" + part.Text + "</" + elemType + ">"
		} else {
			highlightedLine += part.Text
		}
	}

	return Highlight{
		First: HighlightPart{
			Prefix: linePrefix1,
			Line:   removeIns(highlightedLine),
		},
		Second: HighlightPart{
			Prefix: linePrefix2,
			Line:   removeDel(highlightedLine),
		},
	}
}

var (
	removeInsRegexp = regexp.MustCompile(`(<ins[^>]*>((.|\n)*?)<\/ins>)`)
	removeDelRegexp = regexp.MustCompile(`(<del[^>]*>((.|\n)*?)<\/del>)`)
)

func removeIns(str string) string {
	return removeInsRegexp.ReplaceAllString(str, "")
}

func removeDel(str string) string {
	return removeDelRegexp.ReplaceAllString(str, "")
}
