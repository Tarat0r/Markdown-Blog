package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// ConvertMarkdown processes Markdown and applies AST modifications
func MarkdownToHTML(mdText string) string {
	md := []byte(mdText)

	// Create a new Markdown parser
	gm := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)

	// Parse the Markdown into an AST (Abstract Syntax Tree)
	reader := text.NewReader(md)
	doc := gm.Parser().Parse(reader)

	// ✅ Call the separate function to modify AST
	// modifyAST(doc)
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {

		// ✅ Modify image links (add CDN prefix)
		case *ast.Image:
			if entering {
				oldSrc := string(node.Destination)
				newSrc := "https://cdn.example.com" + oldSrc
				node.Destination = []byte(newSrc)
			}

		// ✅ Modify Markdown links (.md → .html)
		case *ast.Link:
			if entering {
				oldHref := string(node.Destination)
				if strings.HasSuffix(oldHref, ".md") {
					newHref := strings.TrimSuffix(oldHref, ".md") + ".html"
					node.Destination = []byte(newHref)
				}
			}
		}
		return ast.WalkContinue, nil
	})

	// Render the modified AST back into HTML
	var buf bytes.Buffer
	if err := gm.Renderer().Render(&buf, md, doc); err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

// Test Function
func main() {
	markdown := `<script></script># My Notes
See [Note 1](note1.md) and [Note 2](note2.md) for details.

Here is an image:
![Alt text](/images/example.jpg)
<script></script>`

	html := MarkdownToHTML(markdown)
	fmt.Println(html)
}

// func modifyAST(doc ast.Node) {
// 	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
// 		switch node := n.(type) {

// 		// ✅ Modify image links (add CDN prefix)
// 		case *ast.Image:
// 			if entering {
// 				oldSrc := string(node.Destination)
// 				newSrc := "https://cdn.example.com" + oldSrc
// 				node.Destination = []byte(newSrc)
// 			}

// 		// ✅ Modify Markdown links (.md → .html)
// 		case *ast.Link:
// 			if entering {
// 				oldHref := string(node.Destination)
// 				if strings.HasSuffix(oldHref, ".md") {
// 					newHref := strings.TrimSuffix(oldHref, ".md") + ".html"
// 					node.Destination = []byte(newHref)
// 				}
// 			}
// 		}
// 		return ast.WalkContinue, nil
// 	})
// }
