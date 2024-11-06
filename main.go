package main

import (
  "fmt"
  "os"
  "strings"
  "golang.org/x/term"
  "github.com/charmbracelet/bubbles/viewport"
  "github.com/charmbracelet/bubbles/textinput"
  "github.com/charmbracelet/lipgloss"
  "github.com/charmbracelet/glamour"
  tea "github.com/charmbracelet/bubbletea"
)

const content = `
# Slalom GenAI Bootcamp Project 5

This is a simple RAG search tool written in Go.

The Internet Engineering Task Force (IETF) is a non-profit consortium that develops standards for the Internet and related protocols. The IETF is best known for its work on the Transmission Control Protocol (TCP), which is the foundation of the Internet.

This body publishes technical documentation in the form of RFCs (Request for Comments). RFCs are documents that describe protocols, standards, or processes for the Internet. They are intended to provide a concise technical summary of the proposal, and to help the community discuss and refine the proposal.

Over the years, the IETF has published more than 10,000 RFCs across a wide range of topics, including networking, security, routing, and transport protocols. With so many RFCs to read, it can be challenging to find the information in the collection.

This tool allows you to search through the IETF RFCs and find the RFC that best matches your query.
`

// Define the application state model
type model struct {
  input textinput.Model
  docView  viewport.Model
  fileView viewport.Model
  historyView viewport.Model
}

// Define the application state components
func appState() (*model, error) {
  windowWidth, windowHeight, err := term.GetSize(int(os.Stdout.Fd()))
  vpWidth := windowWidth - 10
  vpHeight := windowHeight - 10

  docVp := viewport.New(vpWidth - 35, vpHeight - 12)
  docVp.Style = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("205")).
    MarginRight(35)

  renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(vpWidth - 35),
  )
  if err != nil {
    return nil, err
  }

  str, err := renderer.Render(content)
  if err != nil {
    return nil, err
  }

  docVp.SetContent(str)

  query := textinput.New()
  query.Placeholder = "Ask a question about an RFC..."
  query.Focus()
  query.Width = 100

  fileView := viewport.New(vpWidth - (vpWidth - 30), vpHeight)
  fileView.Style = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("205")).
    MarginRight(2)

  historyView := viewport.New(vpWidth - 23, 10)
  historyView.Style = lipgloss.NewStyle().
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("205")).
    MarginRight(2)

  return &model{
    docView: docVp,
    fileView: fileView,
    historyView: historyView,
    input: query,
  }, nil
}

func (m model) Init() tea.Cmd {
  return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var cmd tea.Cmd

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.Type {
      case tea.KeyEnter:
        m.historyView.SetContent(m.input.Value())
        m.input.SetValue("")
        return m, nil
      case tea.KeyCtrlC:
        return m, tea.Quit
    }
  case tea.WindowSizeMsg:
    m.docView.Width = msg.Width
  }

  m.input, cmd = m.input.Update(msg)

  return m, cmd
}

func (m model) View() string {
  doc := strings.Builder{}
  doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.fileView.View(), lipgloss.JoinVertical(lipgloss.Left, m.docView.View(), m.historyView.View())))
  doc.WriteString("\n" + m.input.View())
  return doc.String()
}

func main() {
  model, err := appState()
  if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Println("Could not run Bubble Tea program:", err)
		os.Exit(1)
	}
}
