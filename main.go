package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const content = `
# Slalom GenAI Bootcamp Project 5

This is a simple RAG search tool written in Go.

The Internet Engineering Task Force (IETF) is a non-profit consortium that develops standards for the Internet and related protocols. The IETF is best known for its work on the Transmission Control Protocol (TCP), which is the foundation of the Internet.

This body publishes technical documentation in the form of RFCs (Request for Comments). RFCs are documents that describe protocols, standards, or processes for the Internet. They are intended to provide a concise technical summary of the proposal, and to help the community discuss and refine the proposal.

Over the years, the IETF has published more than 10,000 RFCs across a wide range of topics, including networking, security, routing, and transport protocols. With so many RFCs to read, it can be challenging to find the information in the collection.

This tool allows you to search through the IETF RFCs and find the RFC that best matches your query.
`
const dataDir = "./data/"

type model struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
	input        textinput.Model
	docView      viewport.Model
	fileView     viewport.Model
	filePicker   filepicker.Model
	historyView  viewport.Model
	historyTxt   strings.Builder
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.filePicker.Init()
}

func InitModel() (*model, error) {
	// Get terminal WindowSize
	windowWidth, windowHeight, _ := term.GetSize(int(os.Stdout.Fd()))
	vpWidth := windowWidth - 10
	vpHeight := windowHeight - 10

	// Initialize filepicker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}
	fp.CurrentDirectory = dataDir
	fp.FileAllowed = true
	fp.DirAllowed = false

	// Initialize input box
	query := textinput.New()
	query.Placeholder = "Ask a question about an RFC..."
	query.Focus()
	query.Width = 100

	// Initialize file viewport
	fileView := viewport.New(vpWidth-(vpWidth-30), vpHeight)
	fileView.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		MarginRight(2)
	fileView.SetContent(fp.View())

	// Initialize history viewport
	historyView := viewport.New(vpWidth-23, 10)
	historyView.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		MarginRight(2)

	// Initialize the history text
	historyTxt := strings.Builder{}
	historyTxt.WriteString("Reading from " + dataDir + ". " + "\n")
	historyView.SetContent(historyTxt.String())

	// Initialize the document viewport
	docView := viewport.New(vpWidth-35, vpHeight-12)
	docView.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		MarginRight(35)

	// Create a new glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(vpWidth-35),
	)
	if err != nil {
		return nil, err
	}

	// Render the content
	str, err := renderer.Render(content)
	if err != nil {
		return nil, err
	}
	docView.SetContent(str)

	return &model{
		filepicker:  fp,
		input:       query,
		fileView:    fileView,
		filePicker:  fp,
		docView:     docView,
		historyView: historyView,
		historyTxt:  historyTxt,
	}, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			txt := strings.Builder{}
			txt.WriteString(m.historyTxt.String() + m.input.Value() + "\n")
			m.historyTxt = txt
			m.historyView.SetContent(m.historyTxt.String())
			m.input.SetValue("")
			return m, nil
		}
	case clearErrorMsg:
		m.err = nil
	}
	var cmd tea.Cmd
	// Update the input with any new input from keyboard
	m.input, cmd = m.input.Update(msg)
	// Update the file picker
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var tui strings.Builder
	tui.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.fileView.View(), lipgloss.JoinVertical(lipgloss.Left, m.docView.View(), m.historyView.View())))
	tui.WriteString("\n" + m.input.View())
	return tui.String()
}

func main() {
	data, err := InitModel()
	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}
	tm, err := tea.NewProgram(data).Run()
	if err != nil {
		fmt.Println("Could not run Bubble Tea program:", err)
		fmt.Println(tm.(model))
		os.Exit(1)
	}
}
