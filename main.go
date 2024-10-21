package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/net/html"
)

// Browser represents the main application structure
type Browser struct {
	app        *tview.Application // The main application
	pageView   *tview.TextView    // Displays the webpage content
	urlInput   *tview.InputField  // Input field for entering URLs
	statusBar  *tview.TextView    // Displays status messages
	history    []string           // Stores browsing history
	currentURL string             // Current URL being displayed
}

// NewBrowser creates and initializes a new Browser instance
func NewBrowser() *Browser {
	// Initialize a new Browser struct with its components
	b := &Browser{
		app:       tview.NewApplication(),
		pageView:  tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(true).SetScrollable(true),
		urlInput:  tview.NewInputField().SetLabel("URL: "),
		statusBar: tview.NewTextView().SetTextAlign(tview.AlignCenter),
		history:   make([]string, 0),
	}

	// Set up the URL input field to handle Enter key press
	b.urlInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			url := b.urlInput.GetText()
			b.loadURL(url)
		}
	})

	// Set up scrolling for the page view
	b.pageView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			// Scroll up one row
			row, _ := b.pageView.GetScrollOffset()
			b.pageView.ScrollTo(row-1, 0)
		case tcell.KeyDown:
			// Scroll down one row
			row, _ := b.pageView.GetScrollOffset()
			b.pageView.ScrollTo(row+1, 0)
		case tcell.KeyPgUp:
			// Scroll up one page
			_, _, _, height := b.pageView.GetInnerRect()
			row, _ := b.pageView.GetScrollOffset()
			b.pageView.ScrollTo(row-height, 0)
		case tcell.KeyPgDn:
			// Scroll down one page
			_, _, _, height := b.pageView.GetInnerRect()
			row, _ := b.pageView.GetScrollOffset()
			b.pageView.ScrollTo(row+height, 0)
		}
		return event
	})

	// Set color scheme for the browser components
	b.urlInput.SetFieldBackgroundColor(tcell.ColorWhite)
	b.urlInput.SetFieldTextColor(tcell.ColorBlack)
	b.pageView.SetBackgroundColor(tcell.ColorBlack)
	b.pageView.SetTextColor(tcell.ColorWhite)
	b.statusBar.SetBackgroundColor(tcell.ColorDarkGray)
	b.statusBar.SetTextColor(tcell.ColorWhite)

	return b
}

// loadURL fetches and displays the content of the given URL
func (b *Browser) loadURL(url string) {
	b.statusBar.SetText("Loading...")
	content, err := fetchContent(processURL(url))
	if err != nil {
		b.statusBar.SetText(fmt.Sprintf("Error: %v", err))
		return
	}
	b.currentURL = url
	b.history = append(b.history, url)
	b.displayContent(content)
	b.statusBar.SetText("Loaded: " + url)
}

// processURL ensures the URL has a proper scheme (http:// or https://)
func processURL(url string) string {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return "https://" + url
	}
	return url
}

// fetchContent retrieves the content of a webpage
func fetchContent(url string) (string, error) {
	// Send an HTTP GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// displayContent parses and displays the HTML content
func (b *Browser) displayContent(content string) {
	// Parse the HTML content
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		b.pageView.SetText(fmt.Sprintf("Error parsing HTML: %v", err))
		return
	}

	var displayText strings.Builder
	var title string
	var inBody bool

	// Define a recursive function to traverse the HTML tree
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				title = extractText(n)
			case "body":
				inBody = true
			case "h1", "h2", "h3", "h4", "h5", "h6":
				if inBody {
					displayText.WriteString(fmt.Sprintf("\n[yellow::b]%s[-::-]\n", extractText(n)))
				}
			case "p":
				if inBody {
					displayText.WriteString(fmt.Sprintf("\n%s\n", extractText(n)))
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	// Traverse the HTML tree
	traverse(doc)

	// Format and display the final text
	finalText := fmt.Sprintf("[green::b]Title: %s[-::-]\n\n%s", title, displayText.String())
	b.pageView.SetText(tview.TranslateANSI(finalText))
	b.pageView.ScrollToBeginning()
}

// extractText retrieves the text content from an HTML node
func extractText(n *html.Node) string {
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else if c.Type == html.ElementNode {
			text += extractText(c)
		}
	}
	return strings.TrimSpace(text)
}

// Run starts the browser application
func (b *Browser) Run() error {
	// Create a grid layout for the browser UI
	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(0).
		SetBorders(true).
		AddItem(b.urlInput, 0, 0, 1, 1, 0, 0, true).
		AddItem(b.pageView, 1, 0, 1, 1, 0, 0, false).
		AddItem(b.statusBar, 2, 0, 1, 1, 0, 0, false)

	// Set the root of the application and run it
	return b.app.SetRoot(grid, true).Run()
}

func main() {
	// Create a new Browser instance and run it
	browser := NewBrowser()
	if err := browser.Run(); err != nil {
		fmt.Printf("Error running browser: %v\n", err)
	}
}
