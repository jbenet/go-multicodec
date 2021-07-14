//go:build ignore
// +build ignore

package main

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

const codeTemplate = `
// Code generated by gen.go; DO NOT EDIT.

package multicodec

const ({{ range . }}
// {{ if .IsDeprecated }}Deprecated: {{ end }}{{ .VarName }} is a {{ .Status }} code tagged "{{ .Tag }}"{{ if .Description }} and described by: {{ .Description }}{{ end }}.
{{ .VarName }} Code = {{ .Code }} // {{ .Name }}
{{ end }})
`

type tableEntry struct {
	Name        string
	Tag         string
	Code        string
	Status      string
	Description string
}

func (c tableEntry) IsDeprecated() bool {
	return strings.Contains(c.Description, "deprecated")
}

func (c tableEntry) VarName() string {
	var b strings.Builder
	var last rune
	for _, part := range strings.Split(c.Name, "-") {
		first, firstSize := utf8.DecodeRuneInString(part)
		if unicode.IsNumber(last) && unicode.IsNumber(first) {
			// 123-456 should result in 123_456 for readability.
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToUpper(first))
		b.WriteString(part[firstSize:])
		last, _ = utf8.DecodeLastRuneInString(part)
	}
	return b.String()
}

func main() {
	resp, err := http.Get("https://raw.githubusercontent.com/multiformats/multicodec/HEAD/table.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var entries []tableEntry
	csvReader := csv.NewReader(resp.Body)
	csvReader.Read() // skip the header line
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		entries = append(entries, tableEntry{
			Name:        strings.TrimSpace(record[0]),
			Tag:         strings.TrimSpace(record[1]),
			Code:        strings.TrimSpace(record[2]),
			Status:      strings.TrimSpace(record[3]),
			Description: strings.TrimSpace(record[4]),
		})
	}

	tmpl, err := template.New("").
		Funcs(template.FuncMap{"ToTitle": strings.Title}).
		Parse(codeTemplate)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create("code_table.go")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, entries); err != nil {
		log.Fatal(err)
	}
}
