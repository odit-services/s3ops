# Changelog

All notable changes to this project will be documented in this file.
Versions are based on [Semantic Versioning](http://semver.org/), and the changelog is generated with [Chglog](https://github.com/git-chglog/git-chglog).

## Version History

{{ range .Versions -}}
* [{{ .Tag.Name }}](#{{ .Tag.Name }})
{{ end }}
## Changes
{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
### {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }}

> {{ datetime "2006-01-02" .Tag.Date }}

{{ range .CommitGroups -}}
{{ if eq .Title "Feat" }}#### 🚀 Enhancements{{ else if eq .Title "Perf" }}#### 🔥 Performance{{ else if eq .Title "Fix" }}#### 🩹 Fixes{{ else if eq .Title "Refactor" }}#### 💅 Refactors{{ else if eq .Title "Docs" }}#### 📖 Documentation{{ else if eq .Title "Build" }}#### 📦 Build{{ else if eq .Title "Chore" }}#### 🏡 Chore{{ else if eq .Title "Test" }}#### ✅ Tests{{ else if eq .Title "Style" }}#### 🎨 Styles{{ end }}

{{ range .Commits -}}
* {{ if .Scope }}**{{ .Scope }}:** {{ end }}{{ .Subject }}
{{ end }}
{{ end -}}

{{- if .MergeCommits -}}
#### Pull Requests

{{ range .MergeCommits -}}
* {{ .Header }}
{{ end }}
{{ end -}}

{{- if .NoteGroups -}}
{{ range .NoteGroups -}}
#### {{ .Title }}

{{ range .Notes }}
{{ .Body }}
{{ end }}
{{ end -}}
{{ end -}}
{{ end -}}