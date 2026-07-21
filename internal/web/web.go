package web

import (
	"embed"
	"html/template"
	"io"

	"hearth/internal/library"
)

//go:embed index.html
var files embed.FS

// nodeTemplate defines the recursive "node" template. It is parsed together
// with index.html so that index.html's {{template "node" .}} calls resolve.
//
// Rules encoded here (must match the frontend contract exactly):
//   - a directory renders as <li class="hearth-folder"> with a clickable
//     .hearth-folder-row and a nested <ul class="hearth-children collapsed">.
//   - a file renders as <li class="hearth-file"> with a .hearth-copy-btn
//     carrying data-path="{{.RelPath}}".
//   - all folders start collapsed (the "collapsed" class is always present).
const nodeTemplate = `
{{define "node"}}
{{- if .IsDir -}}
<li class="hearth-folder">
  <div class="hearth-folder-row">
    <span class="hearth-chevron"><svg viewBox="0 0 16 16" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 4l4 4-4 4"/></svg></span>
    <span class="hearth-folder-icon"><svg viewBox="0 0 20 20" width="18" height="18" fill="currentColor"><path d="M2 5a2 2 0 0 1 2-2h3.2a2 2 0 0 1 1.4.6l.8.8a1 1 0 0 0 .7.3H16a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5z"/></svg></span>
    <span class="hearth-name">{{.Name}}</span>
  </div>
  <ul class="hearth-children collapsed">
    {{- range .Children}}{{template "node" .}}{{end -}}
  </ul>
</li>
{{- else -}}
<li class="hearth-file">
  <span class="hearth-file-icon"><svg viewBox="0 0 20 20" width="16" height="16" fill="currentColor"><path d="M4 3h9l3 3v11a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z"/></svg></span>
  <span class="hearth-name">{{.Name}}</span>
  <button class="hearth-copy-btn" data-path="{{.RelPath}}" aria-label="Copy link">
    <span class="icon-copy"><svg viewBox="0 0 20 20" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.6"><rect x="7" y="7" width="9" height="10" rx="1.5"/><path d="M4 13V4a1 1 0 0 1 1-1h8"/></svg></span>
    <span class="icon-tick"><svg viewBox="0 0 20 20" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 10l4 4 8-8"/></svg></span>
  </button>
</li>
{{- end -}}
{{end}}
`

// pageTemplate is index.html + the recursive node template, parsed together.
//
// IMPORTANT: there are TWO template.Must calls. The inner one wraps the first
// Parse (the page). The outer one wraps the second Parse (the node template).
// Both Parse calls return (*template.Template, error); each must be wrapped in
// template.Must so the result is a single *template.Template that can
// initialize this var. Omitting the outer template.Must is a compile error
// ("multiple-value ... in single-value context"). Do not remove either Must.
var pageTemplate = template.Must(
	template.Must(
		template.New("index.html").Parse(mustReadIndex()),
	).Parse(nodeTemplate),
)

func mustReadIndex() string {
	b, err := files.ReadFile("index.html")
	if err != nil {
		// This can only happen if the embed failed at build time, which is
		// a programmer error, not a runtime condition.
		panic(err)
	}
	return string(b)
}

// Render writes the full HTML page for the given tree to w.
func Render(w io.Writer, tree []library.Node) error {
	return pageTemplate.Execute(w, tree)
}
