package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type emoji struct {
	server string
	name   string
	url    string
}

func readCurrentDir() []string {
	here, err := os.Open(".")
	if err != nil {
		log.Fatal(err)
	}
	defer here.Close()

	filenames, err := here.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}

	return filenames
}

func parseFileForEmoji(file string) map[string]string {
	// emojis := []emoji{}
	emojis := make(map[string]string)

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	doc, err := html.Parse(r)
	if err != nil {
		log.Fatal(err)
	}

	var parseFunc func(*html.Node)
	imgClass := "image-1CmAz0"
	parseFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			var src, alt string
			var isEmoji bool
			for _, a := range n.Attr {
				switch a.Key {
				case "src":
					src = a.Val
				case "alt":
					alt = a.Val
				case "class":
					if a.Val == imgClass {
						isEmoji = true
					}
				}
			}
			if isEmoji {
				alt = strings.Trim(alt, ":")
				emojis[alt] = src
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseFunc(c)
		}
	}

	parseFunc(doc)

	return emojis
}

func downloadEmoji(emoji emoji) {
	response, err := http.Get(emoji.url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	f, err := os.Create(fmt.Sprintf("downloaded_emoji/%s/%s", emoji.server, emoji.name))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err = io.Copy(f, response.Body); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var emojis []emoji

	filenames := readCurrentDir()
	for _, filename := range filenames {
		emojisFromFile := parseFileForEmoji(filename)
		for k, v := range emojisFromFile {
			emojis = append(emojis, emoji{name: k, url: v, server: filename})
		}
	}

	if err := os.Mkdir("downloaded_emoji", 0755); err != nil {
		log.Fatal(err)
	}
	for _, filename := range filenames {
		if err := os.Mkdir(fmt.Sprintf("downloaded_emoji/%s", filename), 0755); err != nil {
			log.Fatal(err)
		}
	}

	for _, emoji := range emojis {
		downloadEmoji(emoji)
	}
}
