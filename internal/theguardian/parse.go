package theguardian

import (
	"bytes"
	"errors"
	"strings"

	"golang.org/x/net/html"
)

// parse body of a liveblog to get the latest entry
func ParseLiveBlogBody(body string) (string, error) {
	doc, err := parseHTML(body)
	if err != nil {
		return "", err
	}
	node, found := findFirstDivWithClass(doc)
	if found {
		return extractTextContent(node), nil
	}
	return "", errors.New("couldn't find update in live blog body")
}

func parseHTML(input string) (*html.Node, error) {
	r := strings.NewReader(input)
	return html.Parse(r)
}

func findFirstDivWithClass(node *html.Node) (*html.Node, bool) {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, a := range node.Attr {
			if a.Key == "class" {
				if strings.Contains(a.Val, "block") && !strings.Contains(a.Val, "is-summary") {
					return node, true
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		foundNode, found := findFirstDivWithClass(c)
		if found {
			return foundNode, true
		}
	}
	return nil, false
}

func renderNode(node *html.Node) (string, error) {
	var buf bytes.Buffer
	err := html.Render(&buf, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func extractTextContent(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	var content string
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		content += extractTextContent(c)
	}
	return content
}
