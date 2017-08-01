package diff2html

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	oldFileNameHeader = "--- "
	newFileNameHeader = "+++ "
	hunkHeaderPrefix  = "@@"

	inserts       = "d2h-ins"
	deletes       = "d2h-del"
	insertChanges = "d2h-ins d2h-change"
	deleteChanges = "d2h-del d2h-change"
	context       = "d2h-cntx"
	info          = "d2h-info"
)

var (
	oldMode             = regexp.MustCompile(`^old mode (\d{6})`)
	newMode             = regexp.MustCompile(`^new mode (\d{6})`)
	deletedFileMode     = regexp.MustCompile(`^deleted file mode (\d{6})`)
	newFileMode         = regexp.MustCompile(`^new file mode (\d{6})`)
	copyFrom            = regexp.MustCompile(`^copy from "?(.+)"?`)
	copyTo              = regexp.MustCompile(`^copy to "?(.+)"?`)
	renameFrom          = regexp.MustCompile(`^rename from "?(.+)"?`)
	renameTo            = regexp.MustCompile(`^rename to "?(.+)"?`)
	similarityIndex     = regexp.MustCompile(`^similarity index (\d+)%`)
	dissimilarityIndex  = regexp.MustCompile(`^dissimilarity index (\d+)%`)
	index               = regexp.MustCompile(`^index ([0-9a-z]+)\.\.([0-9a-z]+)\s*(\d{6})?`)
	binaryFiles         = regexp.MustCompile(`^Binary files (.*) and (.*) differ`)
	binaryDiff          = regexp.MustCompile(`^GIT binary patch`)
	combinedIndex       = regexp.MustCompile(`^index ([0-9a-z]+),([0-9a-z]+)\.\.([0-9a-z]+)`)
	combinedMode        = regexp.MustCompile(`^mode (\d{6}),(\d{6})\.\.(\d{6})`)
	combinedNewFile     = regexp.MustCompile(`^new file mode (\d{6})`)
	combinedDeletedFile = regexp.MustCompile(`^deleted file mode (\d{6}),(\d{6})`)
	gitDiffStart        = regexp.MustCompile(`^diff --git "?(.+)"? "?(.+)"?`)
	filenameRegexp      = regexp.MustCompile(`\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)? \+\d{4}.*$`)
	crlf                = regexp.MustCompile(`\r\n?`)
	combined1           = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@.*`)
	combined2           = regexp.MustCompile(`^@@@ -(\d+)(?:,\d+)? -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@@.*`)
)

type Config struct {
	DstPrefix string
	SrcPrefix string
}

func newDiff(conf Config) *Diff {
	return &Diff{
		conf:  conf,
		Files: []*File{},
	}
}

type Diff struct {
	conf  Config
	Files []*File

	currentFile     *File
	currentBlock    *Block
	oldLine         int
	oldLine2        int
	newLine         int
	possibleOldName string
	possibleNewName string
}

type File struct {
	IsGiftDiff          bool     `json:"isGiftDiff"`
	IsCombined          bool     `json:"isCombined"`
	IsDeleted           bool     `json:"isDeleted"`
	IsNew               bool     `json:"isNew"`
	IsCopy              bool     `json:"isCopy"`
	IsRename            bool     `json:"isRename"`
	IsBinary            bool     `json:"isBinary"`
	OldName             string   `json:"oldName"`
	NewName             string   `json:"newName"`
	Language            string   `json:"language"`
	UnchangedPercentage string   `json:"unchangedPercentage"`
	ChangedPercentage   string   `json:"changedPercentage"`
	ChecksumBefore      string   `json:"checksumBefore"`
	ChecksumAfter       string   `json:"checksumAfter"`
	Mode                string   `json:"mode"`
	NewFileMode         string   `json:"newFileMode"`
	DeletedFileMode     string   `json:"deletedFileMode"`
	OldMode             string   `json:"oldMode"`
	NewMode             string   `json:"newMode"`
	Blocks              []*Block `json:"blocks"`
	DeletedLines        int      `json:"deletedLines"`
	AddedLines          int      `json:"addedLines"`
}

type Block struct {
	Lines         []*Line `json:"lines"`
	OldStartLine  int     `json:"oldStartLine"`
	OldStartLine2 int     `json:"oldStartLine2"`
	NewStartLine  int     `json:"newStartLine"`
	Header        string  `json:"header"`
}

type Line struct {
	Content   string `json:"content"`
	Type      string `json:"type"`
	OldNumber int    `json:"oldNumber"`
	NewNumber int    `json:"newNumber"`
}

func (b *Block) addLine(l *Line) {
	b.Lines = append(b.Lines, l)
}

func (d *Diff) startFile() {
	d.saveBlock()
	d.saveFile()

	d.currentFile = &File{
		DeletedLines: 0,
		AddedLines:   0,
		Blocks:       []*Block{},
	}
}

func (d *Diff) startBlock(line string) {
	d.saveBlock()

	if values := combined1.FindStringSubmatch(line); len(values) >= 3 {
		d.currentFile.IsCombined = false
		oldLine, err := strconv.Atoi(values[1])
		if err != nil {
			panic(err) // todo...
		}
		d.oldLine = oldLine
		newLine, err := strconv.Atoi(values[2])
		if err != nil {
			panic(err) // todo...
		}
		d.newLine = newLine
	} else if values = combined2.FindStringSubmatch(line); len(values) >= 4 {
		d.currentFile.IsCombined = true
		oldLine, err := strconv.Atoi(values[1])
		if err != nil {
			panic(err)
		}
		d.oldLine = oldLine
		oldLine2, err := strconv.Atoi(values[2])
		if err != nil {
			panic(err)
		}
		d.oldLine2 = oldLine2
		newLine, err := strconv.Atoi(values[3])
		if err != nil {
			panic(err)
		}
		d.newLine = newLine
	} else {
		//if strings.HasPrefix(line, hunkHeaderPrefix) {
		//
		//}
		d.oldLine = 0
		d.newLine = 0
		d.currentFile.IsCombined = false
	}

	d.currentBlock = &Block{
		Lines:         []*Line{},
		OldStartLine:  d.oldLine,
		OldStartLine2: d.oldLine2,
		NewStartLine:  d.newLine,
		Header:        line,
	}
}

func (d *Diff) createLine(line string) {
	currentLine := &Line{}
	currentLine.Content = line

	newLinePrefixes := []string{"+", " +"}
	delLinePrefixes := []string{"-", " -"}
	if !d.currentFile.IsCombined {
		newLinePrefixes = []string{"+"}
		delLinePrefixes = []string{"-"}
	}

	if startsWith(line, newLinePrefixes) {
		d.currentFile.AddedLines++
		currentLine.Type = inserts
		currentLine.OldNumber = 0
		d.newLine++
		currentLine.NewNumber = d.newLine
		d.currentBlock.addLine(currentLine)
	} else if startsWith(line, delLinePrefixes) {
		d.currentFile.DeletedLines++
		currentLine.Type = deletes
		d.oldLine++
		currentLine.OldNumber = d.oldLine
		currentLine.NewNumber = 0
		d.currentBlock.addLine(currentLine)
	} else {
		currentLine.Type = context
		d.oldLine++
		currentLine.OldNumber = d.oldLine
		d.newLine++
		currentLine.NewNumber = d.newLine
		d.currentBlock.addLine(currentLine)
	}
}

func (d *Diff) saveBlock() {
	if d.currentBlock != nil {
		d.currentFile.Blocks = append(d.currentFile.Blocks, d.currentBlock)
		d.currentBlock = nil
	}
}

func (d *Diff) saveFile() {
	if d.currentFile != nil {
		if d.currentFile.OldName == "" {
			d.currentFile.OldName = d.possibleOldName
		}
		if d.currentFile.NewName == "" {
			d.currentFile.NewName = d.possibleNewName
		}
		if d.currentFile.NewName != "" {
			d.Files = append(d.Files, d.currentFile)
			d.currentFile = nil
		}
	}

	d.possibleOldName = ""
	d.possibleNewName = ""
}

func (d *Diff) Parser(input string) error {
	input = strings.Replace(input, "\\ No newline at end of file", "", -1)
	input = crlf.ReplaceAllString(input, "\n")
	lines := strings.Split(input, "\n")

	for idx, line := range lines {
		if line == "" || strings.HasPrefix(line, "*") {
			continue
		}

		prevLine := ""
		if idx > 0 {
			prevLine = lines[idx-1]
		}
		nxtLine := ""
		if idx+1 < len(lines) {
			nxtLine = lines[idx+1]
		}
		afterNxtLine := ""
		if idx+2 < len(lines) {
			afterNxtLine = lines[idx+2]
		}

		if strings.HasPrefix(line, "diff") {
			d.startFile()
			if values := gitDiffStart.FindStringSubmatch(line); len(values) >= 3 {
				var err error
				d.possibleOldName, err = getFilename("", values[1], d.conf.DstPrefix)
				if err != nil {
					return err
				}
				d.possibleNewName, err = getFilename("", values[2], d.conf.SrcPrefix)
				if err != nil {
					return err
				}
			}
			d.currentFile.IsGiftDiff = true
			continue
		}

		if d.currentFile == nil || (!d.currentFile.IsGiftDiff &&
			(strings.HasPrefix(line, oldFileNameHeader) && strings.HasPrefix(nxtLine, newFileNameHeader) && strings.HasPrefix(afterNxtLine, hunkHeaderPrefix))) {
			d.startFile()
		}

		if (strings.HasPrefix(line, oldFileNameHeader) && strings.HasPrefix(nxtLine, newFileNameHeader)) ||
			(strings.HasPrefix(line, newFileNameHeader) && strings.HasPrefix(prevLine, oldFileNameHeader)) {
			/*
			 * --- Date Timestamp[FractionalSeconds] TimeZone
			 * --- 2002-02-21 23:30:39.942229878 -0800
			 */
			srcFilename, err := getSrcFilename(line, d.conf)
			if err != nil {
				return err
			}
			if d.currentFile != nil && d.currentFile.OldName == "" && strings.HasPrefix(line, "--- ") && srcFilename != "" {
				d.currentFile.OldName = srcFilename
				d.currentFile.Language = getExtension(d.currentFile.OldName, d.currentFile.Language)
				continue
			}

			/*
			 * +++ Date Timestamp[FractionalSeconds] TimeZone
			 * +++ 2002-02-21 23:30:39.942229878 -0800
			 */
			dstFilename, err := getDstFilename(line, d.conf)
			if err != nil {
				return err
			}
			if d.currentFile != nil && d.currentFile.NewName == "" && strings.HasPrefix(line, "+++ ") && dstFilename != "" {
				d.currentFile.NewName = dstFilename
				d.currentFile.Language = getExtension(d.currentFile.NewName, d.currentFile.Language)
				continue
			}
		}

		if (d.currentFile != nil && strings.HasPrefix(line, hunkHeaderPrefix)) ||
			(d.currentFile != nil && d.currentFile.IsGiftDiff && d.currentFile.OldName != "" && d.currentFile.NewName != "" && d.currentBlock == nil) {
			d.startBlock(line)
			continue
		}

		/*
		 * There are three types of diff lines. These lines are defined by the way they start.
		 * 1. New line     starts with: +
		 * 2. Old line     starts with: -
		 * 3. Context line starts with: <SPACE>
		 */
		if d.currentBlock != nil &&
			(strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ")) {
			d.createLine(line)
			continue
		}

		doesNotExistHunkHeader := existHunkHeader(lines, line, idx)

		if values := oldMode.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.OldMode = values[1]
		} else if values = newMode.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.NewMode = values[1]
		} else if values = deletedFileMode.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.DeletedFileMode = values[1]
			d.currentFile.IsDeleted = true
		} else if values = newFileMode.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.NewFileMode = values[1]
			d.currentFile.IsNew = true
		} else if values = copyFrom.FindStringSubmatch(line); len(values) >= 2 {
			if doesNotExistHunkHeader {
				d.currentFile.OldName = values[1]
			}
			d.currentFile.IsCopy = true
		} else if values = copyTo.FindStringSubmatch(line); len(values) >= 2 {
			if doesNotExistHunkHeader {
				d.currentFile.NewName = values[1]
			}
			d.currentFile.IsCopy = true
		} else if values = renameFrom.FindStringSubmatch(line); len(values) >= 2 {
			if doesNotExistHunkHeader {
				d.currentFile.OldName = values[1]
			}
			d.currentFile.IsRename = true
		} else if values = renameTo.FindStringSubmatch(line); len(values) >= 2 {
			if doesNotExistHunkHeader {
				d.currentFile.NewName = values[1]
			}
			d.currentFile.IsRename = true
		} else if values = binaryFiles.FindStringSubmatch(line); len(values) >= 3 {
			d.currentFile.IsBinary = true
			oldName, err := getFilename("", values[1], d.conf.SrcPrefix)
			if err != nil {
				return err
			}
			d.currentFile.OldName = oldName
			newName, err := getFilename("", values[2], d.conf.DstPrefix)
			if err != nil {
				return err
			}
			d.currentFile.NewName = newName
			d.startBlock("Binary file")
		} else if values = binaryDiff.FindStringSubmatch(line); len(values) >= 1 {
			d.currentFile.IsBinary = true
			d.startBlock(line)
		} else if values = similarityIndex.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.UnchangedPercentage = values[1]
		} else if values = dissimilarityIndex.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.ChangedPercentage = values[1]
		} else if values = index.FindStringSubmatch(line); len(values) >= 3 {
			d.currentFile.ChecksumBefore = values[1]
			d.currentFile.ChecksumAfter = values[2]
			if len(values) >= 4 {
				d.currentFile.Mode = values[3]
			}
		} else if values = combinedIndex.FindStringSubmatch(line); len(values) >= 4 {
			d.currentFile.ChecksumBefore = strings.Join([]string{values[2], values[3]}, ",")
			d.currentFile.ChecksumAfter = values[1]
		} else if values = combinedMode.FindStringSubmatch(line); len(values) >= 4 {
			d.currentFile.OldMode = strings.Join([]string{values[2], values[3]}, ",")
			d.currentFile.NewMode = values[1]
		} else if values = combinedNewFile.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.NewFileMode = values[1]
			d.currentFile.IsNew = true
		} else if values = combinedDeletedFile.FindStringSubmatch(line); len(values) >= 2 {
			d.currentFile.DeletedFileMode = values[1]
			d.currentFile.IsDeleted = true
		}
	}

	d.saveBlock()
	d.saveFile()

	return nil
}

func getExtension(filename, language string) string {
	names := strings.Split(filename, ".")
	if len(names) > 1 {
		return names[len(names)-1]
	}
	return language
}

func getSrcFilename(line string, conf Config) (string, error) {
	return getFilename("---", line, conf.SrcPrefix)
}

func getDstFilename(line string, conf Config) (string, error) {
	return getFilename("\\+\\+\\+", line, conf.DstPrefix)
}

func getFilename(linePrefix, line, extraPrefix string) (string, error) {
	prefixes := []string{"a/", "b/", "i/", "w/", "c/", "o/"}
	if extraPrefix != "" {
		prefixes = append(prefixes, extraPrefix)
	}

	var reg *regexp.Regexp
	var err error
	if linePrefix != "" {
		reg, err = regexp.Compile("^" + linePrefix + `?(.+?)"?$`)
	} else {
		reg, err = regexp.Compile(`^"?(.+?)"?$`)
	}
	if err != nil {
		return "", err
	}

	var filename string
	values := reg.FindStringSubmatch(line)
	if len(values) >= 2 {
		filename = values[1]
		matchingPrefixes := []string{}
		for _, p := range prefixes {
			if strings.HasPrefix(filename, p) {
				matchingPrefixes = append(matchingPrefixes, p)
			}
		}

		if len(matchingPrefixes) >= 1 {
			filename = filename[len(matchingPrefixes[0]):]
		}

		filename = filenameRegexp.ReplaceAllString(filename, "")
	}

	return filename, nil
}

func startsWith(str string, prefixes []string) bool {
	for _, p := range prefixes {
		return strings.HasPrefix(str, p)
	}
	return false
}

func existHunkHeader(lines []string, line string, lineIdx int) bool {
	idx := lineIdx
	l := len(lines)
	for idx < l-3 {
		if strings.HasPrefix(line, "diff") {
			return false
		}
		if strings.HasPrefix(lines[idx], oldFileNameHeader) &&
			strings.HasPrefix(lines[idx+1], newFileNameHeader) &&
			strings.HasPrefix(lines[idx+2], hunkHeaderPrefix) {

			return true
		}
		idx++
	}
	return false
}
