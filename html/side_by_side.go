package html

import (
	"bytes"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/yu-ichiko/go-diff2html/diff"
	"html"
	"html/template"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
}

type sideBySide struct {
	FileHTMLID string
	FilePath   string
	Diff       sideDiff
	File       *diff.File
}

type sideDiff struct {
	Left  string
	Right string
}

func NewSideBySide(conf Config) *SideBySidePrinter {
	return &SideBySidePrinter{
		conf: conf,
	}
}

type SideBySidePrinter struct {
	conf Config
}

func (p *SideBySidePrinter) GenerateSideBySideHTML(files []*diff.File) (string, error) {
	content := ""
	for _, file := range files {
		var fileHTML *sideDiff
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

	tmpl, err := template.ParseFiles("generic-wrapper.tmpl")
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct {
		Content template.HTML
	}{
		Content: template.HTML(content),
	})

	return buf.String(), nil
}

func (p *SideBySidePrinter) makeDiffHTML(file *diff.File, diffs *sideDiff) (string, error) {

	pathHTML, err := p.makePathHTML(file)
	if err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles("side-by-side-file-diff.tmpl")
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct {
		FileHTMLID string
		FilePath   template.HTML
		Language   string
		Left       template.HTML
		Right      template.HTML
	}{
		FileHTMLID: "d2h-sample",
		FilePath:   template.HTML(pathHTML),
		Language:   file.Language,
		Left:       template.HTML(diffs.Left),
		Right:      template.HTML(diffs.Right),
	})
	return buf.String(), nil
}

func (p *SideBySidePrinter) makePathHTML(file *diff.File) (string, error) {
	iconHTML, err := p.makeIconHTML()
	if err != nil {
		return "", err
	}
	tagHTML, err := p.makeTagHTML(file)
	if err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles("generic-file-path.tmpl")
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct {
		FileDiffName string
		FileIcon     template.HTML
		FileTag      template.HTML
	}{
		FileDiffName: getDiffName(file),
		FileIcon:     template.HTML(iconHTML),
		FileTag:      template.HTML(tagHTML),
	})
	return buf.String(), nil
}

func (p *SideBySidePrinter) makeIconHTML() (string, error) {
	tmpl, err := template.ParseFiles("icon-file.tmpl")
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct{}{})
	return buf.String(), nil
}

func (p *SideBySidePrinter) makeTagHTML(file *diff.File) (string, error) {
	tag := getFileTypeIcon(file)
	tmpl, err := template.ParseFiles("tag-" + tag + ".tmpl")
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, struct {
	}{})
	return buf.String(), nil
}

func (p *SideBySidePrinter) makeSideHTML(blockHeader string) (string, error) {
	buf := &bytes.Buffer{}
	tmpl, err := template.ParseFiles("generic-column-line-number.tmpl")
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(buf, struct {
		BlockHeader  string
		Type         string
		LineClass    string
		ContentClass string
	}{
		BlockHeader:  blockHeader,
		Type:         diff.Info,
		LineClass:    "d2h-code-side-linenumber",
		ContentClass: "d2h-code-side-line",
	})
	return buf.String(), nil
}

func (p *SideBySidePrinter) genSideBySideFileHTML(file *diff.File) (*sideDiff, error) {
	var err error

	fileHTML := &sideDiff{
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

		oldLines := make([]*diff.Line, 0)
		newLines := make([]*diff.Line, 0)

		processChangeBlock := func() error {
			// conf.matching == "none" only
			oldLen := len(oldLines)
			newLen := len(newLines)
			common := int(math.Min(float64(oldLen), float64(newLen)))
			max := int(math.Max(float64(oldLen), float64(newLen)))

			for i := 0; i < common; i++ {
				oldLine := &diff.Line{Content: ""}
				if oldLen > i {
					oldLine = oldLines[i]
				}
				newLine := &diff.Line{Content: ""}
				if newLen > i {
					newLine = newLines[i]
				}
				highlight := diffHighlight(oldLine.Content, newLine.Content, file.IsCombined)
				left, err := p.genSingleLineHTML(file.IsCombined, diff.Deletes, oldLine.OldNumber, highlight.First.Line, highlight.First.Prefix)
				if err != nil {
					return err
				}
				fileHTML.Left += left
				right, err := p.genSingleLineHTML(file.IsCombined, diff.Inserts, newLine.NewNumber, highlight.Second.Line, highlight.Second.Prefix)
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

			oldLines = make([]*diff.Line, 0)
			newLines = make([]*diff.Line, 0)
			return nil
		}

		for _, line := range block.Lines {
			prefix := string(line.Content[0])
			escapedLine := html.EscapeString(line.Content[1:])

			if line.Type != diff.Inserts &&
				(len(newLines) > 0) || (line.Type != diff.Deletes && len(oldLines) > 0) {
				if err := processChangeBlock(); err != nil {
					return nil, err
				}
			}

			if line.Type == diff.Ctx {
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
			} else if line.Type == diff.Inserts && len(oldLines) == 0 {
				left, err := p.genSingleLineHTML(file.IsCombined, diff.Ctx, 0, "", "")
				if err != nil {
					return nil, err
				}
				fileHTML.Left += left
				right, err := p.genSingleLineHTML(file.IsCombined, line.Type, line.NewNumber, escapedLine, prefix)
				if err != nil {
					return nil, err
				}
				fileHTML.Right += right
			} else if line.Type == diff.Deletes {
				oldLines = append(oldLines, line)
			} else if line.Type == diff.Inserts && len(oldLines) > 0 {
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

func (p *SideBySidePrinter) processLines(isCombined bool, oldLines, newLines []*diff.Line) (*sideDiff, error) {
	fileHTML := &sideDiff{}
	fileHTML.Left = ""
	fileHTML.Right = ""

	oldLinesLen := len(oldLines)
	newLinesLen := len(newLines)
	maxLinesNumber := int(math.Max(float64(oldLinesLen), float64(newLinesLen)))
	for i := 0; i < maxLinesNumber; i++ {
		var oldLine *diff.Line
		if oldLinesLen > i {
			oldLine = oldLines[i]
		}
		var newLine *diff.Line
		if newLinesLen > i {
			newLine = newLines[i]
		}

		var oldContent string
		var newContent string
		var oldPrefix string
		var newPrefix string

		if oldLine != nil {
			oldContent = html.EscapeString(oldLine.Content[1:])
			oldPrefix = oldLine.Content[0:1]
		}
		if newLine != nil {
			newContent = html.EscapeString(newLine.Content[1:])
			newPrefix = newLine.Content[0:1]
		}

		if oldLine != nil && newLine != nil {
			left, err := p.genSingleLineHTML(isCombined, oldLine.Type, oldLine.OldNumber, oldContent, oldPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Left = left
			right, err := p.genSingleLineHTML(isCombined, newLine.Type, newLine.NewNumber, newContent, newPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Right = right
		} else if oldLine != nil {
			left, err := p.genSingleLineHTML(isCombined, oldLine.Type, oldLine.OldNumber, oldContent, oldPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Left = left
			right, err := p.genSingleLineHTML(isCombined, diff.Ctx, 0, "", "")
			if err != nil {
				return nil, err
			}
			fileHTML.Right = right
		} else if newLine != nil {
			left, err := p.genSingleLineHTML(isCombined, diff.Ctx, 0, "", "")
			if err != nil {
				return nil, err
			}
			fileHTML.Left = left
			right, err := p.genSingleLineHTML(isCombined, newLine.Type, newLine.NewNumber, newContent, newPrefix)
			if err != nil {
				return nil, err
			}
			fileHTML.Right = right
		} else {
			// console.error('How did it get here?');
		}
	}

	return fileHTML, nil
}

func (p *SideBySidePrinter) genSingleLineHTML(isCombined bool, lineType string, num int, content string, possiblePrefix string) (string, error) {
	lineWithoutPrefix := content
	prefix := possiblePrefix

	if prefix == "" && content != "" {
		prefix, lineWithoutPrefix = separatePrefix(isCombined, content)
	}

	buf := &bytes.Buffer{}
	tmpl, err := template.ParseFiles("generic-line.tmpl")
	if err != nil {
		return "", err
	}

	lineNumberStr := ""
	if num > 0 {
		lineNumberStr = strconv.Itoa(num)
	}

	err = tmpl.Execute(buf, struct {
		Type          string
		Prefix        string
		Content       string
		LineNumberStr string
		LineClass     string
		ContentClass  string
	}{
		Type:          lineType,
		Prefix:        prefix,
		Content:       lineWithoutPrefix,
		LineNumberStr: lineNumberStr,
		LineClass:     "d2h-code-side-linenumber",
		ContentClass:  "d2h-code-side-line",
	})
	str := buf.String()
	return str, nil
}

func (p *SideBySidePrinter) genEmptyDiff() (*sideDiff, error) {
	fileHTML := &sideDiff{}
	fileHTML.Right = ""

	buf := &bytes.Buffer{}
	tmpl, err := template.ParseFiles("generic-empty-diff.tmpl")
	if err != nil {
		return nil, err
	}
	err = tmpl.Execute(buf, struct {
		Type         string
		ContentClass string
	}{
		Type:         diff.Info,
		ContentClass: "d2h-code-side-line",
	})
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

func getHTMLID(file *diff.File) {

}

func getFileTypeIcon(file *diff.File) string {
	templateName := "file-changed"
	if file.IsRename {
		templateName = "file-renamed"
	} else if file.IsCopy {
		templateName = "file-renamed"
	} else if file.IsNew {
		templateName = "file-added"
	} else if file.IsDeleted {
		templateName = "file-deleted"
	} else if file.NewName != file.OldName {
		templateName = "file-renamed"
	}
	return templateName
}

func getDiffName(file *diff.File) string {
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

	highlightedLine := ""
	for _, part := range diffs {
		elemType := ""
		if part.Type == diffmatchpatch.DiffInsert {
			elemType = "ins"
		} else if part.Type == diffmatchpatch.DiffDelete {
			elemType = "del"
		}
		escapedValue := html.EscapeString(part.Text)
		if elemType != "" {
			highlightedLine += "<" + elemType + ">" + escapedValue + "</" + elemType + ">"
		} else {
			highlightedLine += escapedValue
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
