package views

import (
	"bytes"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed all:templates
var viewsFS embed.FS

var markdown = sync.OnceValue(func() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithExtensions(highlighting.NewHighlighting(
			highlighting.WithStyle("nord"),
		)),
		goldmark.WithExtensions(extension.Linkify),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
})

func ConvertMarkdown(val string) (string, error) {
	var buf bytes.Buffer
	if err := markdown().Convert([]byte(val), &buf); err != nil {
		return "", fmt.Errorf("convert markdown: %w", err)
	}
	return buf.String(), nil
}

var getBase = sync.OnceValue(func() *template.Template {
	content, err := viewsFS.ReadFile("templates/_layout.gohtml")
	if err != nil {
		panic(err)
	}
	funcs := sprig.HtmlFuncMap()
	funcs["hideContext"] = func(val string) string {
		outer, inner := getTag(val, "context")
		if inner.empty() {
			return val
		}
		left := val[:outer.start]
		right := val[outer.end:]
		content := val[inner.start:inner.end]

		var out bytes.Buffer
		if err := markdown().Convert([]byte("```xml\n<context>"+content+"</context>\n```\n"), &out); err != nil {
			slog.Error("failed to convert markdown", "error", err)
		} else {
			content = out.String()
		}

		s := left + "\n<details>\n<summary>Context</summary>\n" + content + "\n</details>\n" + right
		return s
	}
	sessionUID := uuid.New()
	sessionID := hex.EncodeToString(sessionUID[:])
	funcs["sessionID"] = func() string {
		return sessionID
	}
	funcs["getContext"] = func(val string) string {
		_, inner := getTag(val, "context")
		if inner.empty() {
			return ""
		}
		content := val[inner.start:inner.end]
		return strings.TrimSpace(content)
	}
	funcs["codeblock"] = func(lang, val string) string {
		return "```" + lang + "\n" + val + "\n```"
	}
	funcs["withoutContext"] = func(val string) string {
		outer, inner := getTag(val, "context")
		if inner.empty() {
			return val
		}
		left := val[:outer.start]
		right := val[outer.end:]

		return left + right
	}
	funcs["firstLine"] = func(val string) string {
		return strings.TrimSpace(strings.Split(strings.TrimSpace(val), "\n")[0])
	}
	funcs["toHTML"] = func(val string) template.HTML {
		return template.HTML(val)
	}
	funcs["pathquote"] = func(val string) template.URL {
		return template.URL(url.PathEscape(val))
	}
	funcs["jsonBlock"] = func(val string) template.HTML {
		var out bytes.Buffer
		if err := markdown().Convert([]byte("```json\n"+val+"\n```\n"), &out); err != nil {
			return template.HTML("<pre>" + template.HTMLEscaper(val) + "</pre>")
		}
		return template.HTML(out.String())
	}
	funcs["words"] = func(val string) int {
		words := strings.FieldsFunc(val, func(r rune) bool {
			return !unicode.In(r, unicode.Letter, unicode.Number)
		})
		return len(words)
	}
	funcs["markdown"] = func(val string) template.HTML {
		var buf bytes.Buffer
		if err := markdown().Convert([]byte(val), &buf); err != nil {
			slog.Error("failed to convert markdown", "error", err)
			return template.HTML("")
		}
		return template.HTML(buf.String())
	}
	funcs["runesTruncate"] = func(length int, val string) string {
		r := []rune(val)
		if len(r) < length {
			return val
		}
		return string(r[:length]) + "..."
	}

	funcs["linesCount"] = func(val string) int {
		return strings.Count(strings.TrimSpace(val), "\n") + 1
	}

	return template.Must(template.New("").Funcs(funcs).Parse(string(content)))
})

func Inherit[T any](base *template.Template, name string) *DynamicView[T] {
	cp, err := base.Clone()
	if err != nil {
		panic(err)
	}

	content, err := viewsFS.ReadFile("templates/" + name)
	if err != nil {
		panic(err)
	}

	return &DynamicView[T]{view: template.Must(cp.New("").Parse(string(content)))}
}

func InheritBase[T any](assets fs.ReadFileFS, name string) *DynamicView[T] {
	cp, err := getBase().Clone()
	if err != nil {
		panic(err)
	}

	content, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return &DynamicView[T]{view: template.Must(cp.New("").Parse(string(content)))}
}

type DynamicView[T any] struct {
	view *template.Template
}

func (v *DynamicView[T]) Code(code int, data T) *Builder[T] {
	return &Builder[T]{
		code: code,
		view: v.view,
		data: data,
	}
}

func (v *DynamicView[T]) OK(data T) *Builder[T] {
	return v.Code(http.StatusOK, data)
}

func (v *DynamicView[T]) HTML(w http.ResponseWriter, data T) {
	v.Code(http.StatusOK, data).HTML(w)
}

type Builder[T any] struct {
	data T
	code int
	view *template.Template
}

func (b *Builder[T]) Code(code int) *Builder[T] {
	b.code = code
	return b
}

func (b *Builder[T]) Data(data T) *Builder[T] {
	b.data = data
	return b
}

func (b *Builder[T]) HTML(out http.ResponseWriter) {
	var buf bytes.Buffer
	if err := b.view.Execute(&buf, b.data); err != nil {
		slog.Error("failed to render template", "error", err)
		b.code = http.StatusInternalServerError
		buf.Reset()
		buf.WriteString("<html><body><h1>Internal Server Error</h1><pre>" + template.HTMLEscaper(err.Error()) + "</pre></body></html>")
	}

	out.Header().Set("Content-Type", "text/html")
	out.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	out.WriteHeader(b.code)
	_, _ = out.Write(buf.Bytes())
}

type span struct {
	start int
	end   int
}

func (s span) empty() bool {
	return s.start == s.end
}

func getTag(content string, name string) (outer span, inner span) {
	openTag := "<" + name + ">"
	closeTag := "</" + name + ">"
	a := strings.Index(content, openTag)
	if a == -1 {
		return span{0, len(content)}, span{}
	}
	b := strings.LastIndex(content, closeTag)
	if b == -1 || b < a {
		return span{0, len(content)}, span{}
	}
	return span{a, b + len(closeTag)}, span{a + len(openTag), b}
}

func RenderError(w http.ResponseWriter, err error) {
	slog.Error("failed to handle request", "error", err)
	Error(ErrorParams{
		Error: err,
	}).Code(http.StatusInternalServerError).HTML(w)
}
