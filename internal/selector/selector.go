// Package selector provides interactive fuzzy selection for branches and worktrees.
package selector

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lucas-stellet/wk/internal/worktree"
)

// ErrCancelled is returned when the user cancels the selection.
var ErrCancelled = errors.New("selection cancelled")

// Options configures the branch selector behavior.
type Options struct {
	AllowCreate    bool // shows "[+] Create new branch..." option
	FilterExisting bool // filters out branches that already have worktrees
}

// branchItem represents a branch in the list.
type branchItem struct {
	name        string
	description string
	isCreate    bool
}

func (i branchItem) Title() string       { return i.name }
func (i branchItem) Description() string { return i.description }
func (i branchItem) FilterValue() string { return i.name }

// worktreeItem represents a worktree in the list.
type worktreeItem struct {
	branch string
	path   string
	commit string
}

func (i worktreeItem) Title() string       { return i.branch }
func (i worktreeItem) Description() string { return i.path }
func (i worktreeItem) FilterValue() string { return i.branch }

// Custom delegate for our list styling
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc string
		isCreate    bool
	)

	switch i := listItem.(type) {
	case branchItem:
		title = i.name
		desc = i.description
		isCreate = i.isCreate
	case worktreeItem:
		title = i.branch
		desc = i.path
	}

	// Styles
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("78"))

	isSelected := index == m.Index()

	// Indicator
	var indicator string
	if isSelected {
		indicator = selectedStyle.Render(">")
	} else {
		indicator = " "
	}

	// Circle/bullet
	var bullet string
	if isSelected {
		bullet = selectedStyle.Render("●")
	} else {
		bullet = dimStyle.Render("○")
	}

	// Title styling
	var titleStr string
	if isCreate {
		titleStr = createStyle.Render(title)
	} else if isSelected {
		titleStr = selectedStyle.Render(title)
	} else {
		titleStr = normalStyle.Render(title)
	}

	// Description styling
	descStr := dimStyle.Render(desc)

	fmt.Fprintf(w, "%s %s %s\n", indicator, bullet, titleStr)
	fmt.Fprintf(w, "    %s\n", descStr)
}

// selectorModel is the bubbletea model for our selector.
type selectorModel struct {
	list     list.Model
	choice   string
	isCreate bool
	quitting bool
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(branchItem); ok {
				m.choice = item.name
				m.isCreate = item.isCreate
			} else if item, ok := m.list.SelectedItem().(worktreeItem); ok {
				m.choice = item.branch
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// SelectBranch opens an interactive selector for branches.
func SelectBranch(opts Options) (string, error) {
	selected, _, err := SelectOrCreate(Options{
		AllowCreate:    false,
		FilterExisting: opts.FilterExisting,
	})
	return selected, err
}

// SelectOrCreate opens an interactive selector for branches with an option to create new.
func SelectOrCreate(opts Options) (string, bool, error) {
	branches, err := worktree.ListBranches()
	if err != nil {
		return "", false, fmt.Errorf("list branches: %w", err)
	}

	var existingWorktrees map[string]bool
	if opts.FilterExisting {
		existingWorktrees, err = worktree.ListWorktreeBranches()
		if err != nil {
			return "", false, fmt.Errorf("list worktree branches: %w", err)
		}
	}

	var items []list.Item

	if opts.AllowCreate {
		items = append(items, branchItem{
			name:        "[+] Create new branch...",
			description: "Enter a name to create a new branch",
			isCreate:    true,
		})
	}

	for _, b := range branches {
		if opts.FilterExisting && existingWorktrees[b.Name] {
			continue
		}

		status := formatBranchStatus(b)
		desc := fmt.Sprintf("%s · %s · %s", status, b.CommitShort, b.CommitDate)
		items = append(items, branchItem{
			name:        b.Name,
			description: desc,
		})
	}

	if len(items) == 0 {
		return "", false, errors.New("no branches available")
	}

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "Select branch"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		MarginLeft(2)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	l.SetShowHelp(true)

	m := selectorModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}

	result := finalModel.(selectorModel)
	if result.quitting && result.choice == "" {
		return "", false, ErrCancelled
	}

	return result.choice, result.isCreate, nil
}

// SelectWorktree opens an interactive selector for existing worktrees.
func SelectWorktree() (string, error) {
	worktrees, err := worktree.List()
	if err != nil {
		return "", fmt.Errorf("list worktrees: %w", err)
	}

	if len(worktrees) == 0 {
		return "", errors.New("no worktrees found")
	}

	var items []list.Item
	for _, wt := range worktrees {
		items = append(items, worktreeItem{
			branch: wt.Branch,
			path:   wt.Path,
			commit: wt.Commit,
		})
	}

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "Select worktree"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		MarginLeft(2)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	l.SetShowHelp(true)

	m := selectorModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	result := finalModel.(selectorModel)
	if result.quitting && result.choice == "" {
		return "", ErrCancelled
	}

	return result.choice, nil
}

func formatBranchStatus(b worktree.Branch) string {
	if b.IsLocal && b.IsRemote {
		return "synced"
	}
	if b.IsLocal {
		return "local"
	}
	return "remote"
}

// promptForBranchName prompts the user for a new branch name.
// This is called after selecting "Create new branch" option.
func PromptForBranchName() (string, error) {
	fmt.Print("Enter new branch name: ")
	var name string
	_, err := fmt.Scanln(&name)
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("branch name cannot be empty")
	}
	return name, nil
}
