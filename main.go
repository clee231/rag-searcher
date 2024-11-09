package main

import (
  "fmt"
  "os"
  "strings"
  "bytes"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "golang.org/x/term"
  "github.com/charmbracelet/bubbles/viewport"
  "github.com/charmbracelet/bubbles/textinput"
  "github.com/charmbracelet/bubbles/filepicker"
  "github.com/charmbracelet/lipgloss"
  "github.com/charmbracelet/glamour"
  tea "github.com/charmbracelet/bubbletea"
)


func getEmbedding(text string) ([]float32, error) {
    // Prepare the request payload
    jsonData := map[string]string{"text": text}
    jsonValue, _ := json.Marshal(jsonData)

    // Send POST request to local server
    resp, err := http.Post("http://127.0.0.1:8000/embed", "application/json", bytes.NewBuffer(jsonValue))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse response
    body, _ := ioutil.ReadAll(resp.Body)
    var result map[string][]float32
    json.Unmarshal(body, &result)

    return result["embedding"], nil
}

const content = `
# Slalom GenAI Bootcamp Project 5

This is a simple RAG search tool written in Go.

The Internet Engineering Task Force (IETF) is a non-profit consortium that develops standards for the Internet and related protocols. The IETF is best known for its work on the Transmission Control Protocol (TCP), which is the foundation of the Internet.

This body publishes technical documentation in the form of RFCs (Request for Comments). RFCs are documents that describe protocols, standards, or processes for the Internet. They are intended to provide a concise technical summary of the proposal, and to help the community discuss and refine the proposal.

Over the years, the IETF has published more than 10,000 RFCs across a wide range of topics, including networking, security, routing, and transport protocols. With so many RFCs to read, it can be challenging to find the information in the collection.

This tool allows you to search through the IETF RFCs and find the RFC that best matches your query.
`
const dataDir = "./data/"

// Define the application state model
type model struct {
  input textinput.Model
  docView  viewport.Model
  fileView viewport.Model
  filePicker filepicker.Model
  historyView viewport.Model
  historyTxt strings.Builder
}

// Define a custom item type that implements list.Item
type item struct {
    title string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.title }

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

  // Initialize the file picker
  fp := filepicker.New()
  fp.AllowedTypes = []string{".txt"}
  fp.CurrentDirectory = dataDir
  fp.Init()


  //// Get files from data directory
  //files, err := ioutil.ReadDir("./data/")
  //if err != nil {
  //  fmt.Println("Could not read data directory:", err)
  //}
  //// Log a message about files read
  historyTxt := strings.Builder{}
  //historyTxt.WriteString("Read in " + strconv.Itoa(len(files)) + " files." + "\n")
  //historyView.SetContent(historyTxt.String())

  //// Convert the filenames into list items
  //items := make([]list.Item, 0)
  //for _, file := range files {
  //    if !file.IsDir() { // Only add files, not subdirectories
  //        items = append(items, item{title: file.Name()})
  //    }
  //}

  // Initialize the list
  const defaultWidth = 20
  //l := list.New(items, list.NewDefaultDelegate(), windowWidth - (windowWidth - 40) , windowHeight)
  //l.Title = "Files in Directory"
  

  fileView.SetContent(fp.View())


  return &model{
    docView: docVp,
    fileView: fileView,
    filePicker: fp,
    historyView: historyView,
    historyTxt: historyTxt,
    input: query,
  }, nil
}

// Define message types to handle the result of the command
type dataMsg struct{ data string }
type errMsg struct{ err error }

// Command to process files
func (m model) processDataCmd() tea.Cmd {
    return func() tea.Msg {
        embedData := strings.Builder{}
        // Get files from data directory
        files, err := ioutil.ReadDir(dataDir)
        if err != nil {
          fmt.Println("Could not read data directory:", err)
        }

        // Convert file contents into embeddings
        for _, file := range files {
            if !file.IsDir() { // Only process files
              filePtr, err := os.ReadFile(dataDir + file.Name())
              if err != nil {
                fmt.Println("Could not read file:", err)
              }
              // Prepare the request payload
              jsonData := map[string]string{"text": string(filePtr)}
              jsonValue, _ := json.Marshal(jsonData)

              // Send POST request to local server
              resp, err := http.Post("http://127.0.0.1:8000/embed", "application/json", bytes.NewBuffer(jsonValue))
              if err != nil {
                  fmt.Println("Could not get embedding", err)
              }
              defer resp.Body.Close()

              // Parse response
              body, _ := ioutil.ReadAll(resp.Body)
              embedData.WriteString(string(body))

              // Log message in HistoryTxt
            }
        }

        return dataMsg{string(embedData.String())}
    }
}

func (m model) Init() tea.Cmd {
  return m.processDataCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var cmd tea.Cmd

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.Type {
      case tea.KeyEnter:
        txt := strings.Builder{}
        txt.WriteString(m.historyTxt.String() + m.input.Value() + "\n")
        m.historyTxt = txt
        m.historyView.SetContent(m.historyTxt.String())
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
