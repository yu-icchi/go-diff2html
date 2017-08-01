package diff2html

const (
	genericColumnLineNumber = `<tr>
    <td class="{{.LineClass}} {{.Type}}"></td>
    <td class="{{.Type}}">
        <div class="{{.ContentClass}} {{.Type}}">{{.BlockHeader}}</div>
    </td>
</tr>`

	genericEmptyDiff = `<tr>
  <td class="{{.Type}}">
    <div class="{{.ContentClass}} {{.Type}}">
      File without changes
    </div>
  </td>
</tr>`

	genericFilePath = `<span class="d2h-file-name-wrapper">
    <span class="d2h-icon-wrapper">{{.FileIcon}}</span>
    <span class="d2h-file-name">{{.FileDiffName}}</span>
    {{.FileTag}}
</span>`

	genericLine = `<tr>
    <td class="{{.LineClass}} {{.Type}}">
        {{.LineNumberStr}}
    </td>
    <td class="{{.Type}}">
        <div class="{{.ContentClass}} {{.Type}}">
            {{if .Prefix}}<span class="d2h-code-line-prefix">{{.Prefix}}</span>{{end}}
            {{if .Content}}<span class="d2h-code-line-ctn">{{.Content}}</span>{{end}}
        </div>
    </td>
</tr>`

	genericWrapper = `<div class="d2h-wrapper">
    {{.Content}}
</div>`

	iconFile = `<svg aria-hidden="true" class="d2h-icon" height="16" version="1.1" viewBox="0 0 12 16" width="12">
    <path d="M6 5H2v-1h4v1zM2 8h7v-1H2v1z m0 2h7v-1H2v1z m0 2h7v-1H2v1z m10-7.5v9.5c0 0.55-0.45 1-1 1H1c-0.55 0-1-0.45-1-1V2c0-0.55 0.45-1 1-1h7.5l3.5 3.5z m-1 0.5L8 2H1v12h10V5z"></path>
</svg>`

	sideBySideFileDiff = `<div id="{{.FileHTMLID}}" class="d2h-file-wrapper" data-lang="{{.Language}}">
    <div class="d2h-file-header">
        {{.FilePath}}
    </div>
    <div class="d2h-file-diff">
        <div class="d2h-file-side-diff">
            <div class="d2h-code-wrapper">
                <table class="d2h-diff-table">
                    <tbody class="d2h-diff-tbody">
                    {{.Left}}
                    </tbody>
                </table>
            </div>
        </div>
        <div class="d2h-file-side-diff">
            <div class="d2h-code-wrapper">
                <table class="d2h-diff-table">
                    <tbody class="d2h-diff-tbody">
                    {{.Right}}
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</div>`

	tagFileAdded = `<span class="d2h-tag d2h-added d2h-added-tag">ADDED</span>`

	tagFileChanged = `<span class="d2h-tag d2h-changed d2h-changed-tag">CHANGED</span>`

	tagFileDeleted = `<span class="d2h-tag d2h-deleted d2h-deleted-tag">DELETED</span>`

	tagFileRenamed = `<span class="d2h-tag d2h-moved d2h-moved-tag">RENAMED</span>`
)
