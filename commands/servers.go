package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	Index int
	Name  string
}

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list            list.Model
	onSelect        func(ctx context.Context, item *item)
	onSelectMessage string
	choice          string
	quitting        bool
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i.Name)
				m.onSelect(context.Background(), &i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(m.onSelectMessage)
	}
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// Comand Servers
//
// `rye servers`
//
// list all the servers
func NewServersCmd() *cobra.Command {
	return &cobra.Command{
		Use: "servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			var items []list.Item
			for idx, srv := range repo.Servers {
				items = append(items, item{
					Index: idx,
					Name:  srv.Server.Name,
				})
			}

			const defaultWidth = 20

			l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
			l.Title = "Choose a server to connect to..."
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(false)
			l.Styles.Title = titleStyle
			l.Styles.PaginationStyle = paginationStyle
			l.Styles.HelpStyle = helpStyle

			m := model{list: l}
			m.onSelect = func(ctx context.Context, item *item) {
				selectedServer := repo.Servers[item.Index]

				outbound, err := selectedServer.Server.ToOutbound()
				if err != nil {
					m.onSelectMessage = rye.SprintfError("parse server #%d: %s to outbound err: %s", item.Index, selectedServer.Server.Name, err.Error())
					return
				}

				repo.XrayConfig.SetOutbound("proxy", outbound)
				err = repo.XrayConfig.Save()
				if err != nil {
					m.onSelectMessage = rye.SprintfError("save xray/config.json err: %s", err.Error())
					return
				}

				// stop runner
				err = rye.StopRunner(repo.PID)
				if err != nil {
					m.onSelectMessage = rye.SprintfError("stop runner err: %s", err.Error())
					return
				}

				// start runner again
				pid, err := rye.StartRunner()
				if err != nil {
					m.onSelectMessage = rye.SprintfError("start runner err: %s", err.Error())
					return
				}
				if err := repo.WriteRunnerPID(pid); err != nil {
					m.onSelectMessage = rye.SprintfError("write pid err: %s", err.Error())
					return
				}

				m.onSelectMessage = rye.SprintfInfo("connected to server: %s", selectedServer.Server.Name)
			}

			if _, err := tea.NewProgram(&m).Run(); err != nil {
				return fmt.Errorf("list servers err: %w", err)
			}

			return nil
		},
	}
}
