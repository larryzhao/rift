package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/singbox"
	"github.com/spf13/cobra"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	tabStyle          = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("240"))
	activeTabStyle    = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("170")).Bold(true)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	RepoIndex int
	Group     string
	Name      string
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
	groupNames      []string
	groupItems      map[string][]list.Item
	selectedIndex   map[string]int
	groupIndex      int
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

		case "tab":
			m.switchGroup(1)
			return m, nil

		case "shift+tab":
			m.switchGroup(-1)
			return m, nil

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
	return "\n" + m.renderTabs() + "\n" + m.list.View()
}

func (m *model) switchGroup(step int) {
	if len(m.groupNames) == 0 {
		return
	}

	current := m.groupNames[m.groupIndex]
	m.selectedIndex[current] = m.list.Index()
	m.groupIndex = (m.groupIndex + step + len(m.groupNames)) % len(m.groupNames)

	next := m.groupNames[m.groupIndex]
	m.list.SetItems(m.groupItems[next])
	m.list.Select(m.selectedIndex[next])
	m.updateTitle()
}

func (m *model) updateTitle() {
	if len(m.groupNames) == 0 {
		m.list.Title = "No servers available"
		return
	}

	current := m.groupNames[m.groupIndex]
	m.list.Title = fmt.Sprintf("Choose a server to connect to... (%s)", current)
}

func (m *model) renderTabs() string {
	if len(m.groupNames) == 0 {
		return ""
	}

	parts := make([]string, 0, len(m.groupNames)+1)
	for idx, group := range m.groupNames {
		style := tabStyle
		if idx == m.groupIndex {
			style = activeTabStyle
		}
		parts = append(parts, style.Render(group))
	}
	parts = append(parts, tabStyle.Render("[tab/shift+tab to switch groups]"))
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// Comand Servers
//
// `rift servers`
//
// list all the servers
func NewServersCmd() *cobra.Command {
	return &cobra.Command{
		Use: "servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			groupNames := make([]string, 0)
			groupItems := make(map[string][]list.Item)
			selectedIndex := make(map[string]int)

			for idx, srv := range repo.Servers {
				if _, ok := groupItems[srv.Group]; !ok {
					groupNames = append(groupNames, srv.Group)
				}

				groupItems[srv.Group] = append(groupItems[srv.Group], item{
					RepoIndex: idx,
					Group:     srv.Group,
					Name:      fmt.Sprintf("%s %s", srv.Server.Protocol.Style().Render(srv.Server.Protocol.ShortName()), srv.Server.Name),
				})
			}

			const defaultWidth = 20

			var items []list.Item
			if len(groupNames) > 0 {
				items = groupItems[groupNames[0]]
			}

			l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(false)
			l.Styles.Title = titleStyle
			l.Styles.PaginationStyle = paginationStyle
			l.Styles.HelpStyle = helpStyle

			m := model{
				list:          l,
				groupNames:    groupNames,
				groupItems:    groupItems,
				selectedIndex: selectedIndex,
			}
			m.updateTitle()
			m.onSelect = func(ctx context.Context, item *item) {
				selectedServer := repo.Servers[item.RepoIndex]

				runner := singbox.NewRunner(repo.SingboxConfigFile(), repo.RunnerLogFile())
				confData, err := runner.ToConfig(selectedServer.Server)
				if err != nil {
					m.onSelectMessage = rift.SprintfError("convert to sing-box config err: %s", err.Error())
					return
				}
				if err := os.MkdirAll(filepath.Dir(repo.SingboxConfigFile()), 0755); err != nil {
					m.onSelectMessage = rift.SprintfError("create sing-box config dir err: %s", err.Error())
					return
				}
				if err := os.WriteFile(repo.SingboxConfigFile(), confData, 0644); err != nil {
					m.onSelectMessage = rift.SprintfError("write sing-box config file err: %s", err.Error())
					return
				}

				repo.Status.ServerGroup = selectedServer.Group
				repo.Status.ServerName = selectedServer.Server.Name
				repo.Status.Protocl = selectedServer.Server.Protocol
				err = repo.SaveStatus()
				if err != nil {
					m.onSelectMessage = rift.SprintfError("save status file err: %s", err.Error())
					return
				}

				ok, err := repo.Status.IsProxyRunning()
				if err != nil {
					m.onSelectMessage = rift.SprintfError("check if proxy is running err: %s", err.Error())
					return
				}
				if ok {
					pid := repo.Status.PIDByKind("proxy")

					rift.SprintfVerbose("stopping running proxy with pid: %d", pid)

					err = rift.Stop(pid)
					if err != nil {
						m.onSelectMessage = rift.SprintfError("stop proxy %d err: %s", pid, err.Error())
						return
					}

					rift.SprintfVerbose("starting proxy again")
					pid, err = runner.Run()
					if err != nil {
						m.onSelectMessage = rift.SprintfError("start proxy %s err: %s", selectedServer.Server.Protocol.String(), err.Error())
						return
					}
					rift.SprintfVerbose("proxy started with pid: %d", pid)

					repo.Status.UpdateRunningProcess("proxy", pid)
					err := repo.SaveStatus()
					if err != nil {
						m.onSelectMessage = rift.SprintfError("save status file err: %s", err.Error())
						return
					}
					return
				}
				m.onSelectMessage = rift.SprintfInfo("switch to server: %s", selectedServer.Server.Name)
			}

			if _, err := tea.NewProgram(&m).Run(); err != nil {
				return fmt.Errorf("list servers err: %w", err)
			}

			return nil
		},
	}
}
