package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/hysteria2"
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
					Name:  fmt.Sprintf("%s %s", srv.Server.Protocol.Style().Render(srv.Server.Protocol.ShortName()), srv.Server.Name),
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

				var runner rye.Runnable

				switch selectedServer.Server.Protocol {
				case rye.ProtoclHysteria2:
					confData, err := hysteria2.ToConfig(selectedServer.Server)
					if err != nil {
						m.onSelectMessage = rye.SprintfError("convert to hysteria2 config err: %s", err.Error())
						return
					}

					err = os.WriteFile(repo.HysteriaConfigFile(), confData, 0644)
					if err != nil {
						m.onSelectMessage = rye.SprintfError("write hysteria config file err: %s", err.Error())
						return
					}

					runner = hysteria2.NewRunner("/opt/homebrew/bin/hysteria", repo.HysteriaConfigFile())
				}

				ok, err := repo.Status.IsProxyRunning()
				if err != nil {
					m.onSelectMessage = rye.SprintfError("check if proxy is running err: %s", err.Error())
					return
				}
				if ok {
					pid := repo.Status.PIDByKind("proxy")
					err = rye.Stop(pid)
					if err != nil {
						m.onSelectMessage = rye.SprintfError("stop proxy %d err: %s", pid, err.Error())
						return
					}

					pid, err = runner.Run()
					if err != nil {
						m.onSelectMessage = rye.SprintfError("start proxy %s err: %s", selectedServer.Server.Protocol.String(), err.Error())
						return
					}

					repo.Status.ServerGroup = selectedServer.Group
					repo.Status.ServerName = selectedServer.Server.Name
					repo.Status.Protocl = selectedServer.Server.Protocol
					repo.Status.UpdateRunningProcess("proxy", pid)
					err = repo.SaveStatus()
					if err != nil {
						m.onSelectMessage = rye.SprintfError("save status file err: %s", err.Error())
						return
					}
					return
				}

				// outbound, err := selectedServer.Server.ToOutbound()
				// if err != nil {
				// 	m.onSelectMessage = rye.SprintfError("parse server #%d: %s to outbound err: %s", item.Index, selectedServer.Server.Name, err.Error())
				// 	return
				// }

				// repo.XrayConfig.SetOutbound("proxy", outbound)
				// err = repo.XrayConfig.Save()
				// if err != nil {
				// 	m.onSelectMessage = rye.SprintfError("save xray/config.json err: %s", err.Error())
				// 	return
				// }

				// stop runner
				// err = rye.StopRunner(repo.Status.PID)
				// if err != nil {
				// 	m.onSelectMessage = rye.SprintfError("stop runner err: %s", err.Error())
				// 	return
				// }

				// // start runner again
				// pid, err := rye.StartRunner()
				// if err != nil {
				// 	m.onSelectMessage = rye.SprintfError("start runner err: %s", err.Error())
				// 	return
				// }
				// repo.Status.PID = pid
				// repo.Status.ServerGroup = selectedServer.Group
				// repo.Status.ServerName = selectedServer.Server.Name
				// repo.Status.Protocl = selectedServer.Server.Protocol
				// if err := repo.SaveStatus(); err != nil {
				// 	m.onSelectMessage = rye.SprintfError("update runner status err: %s", err.Error())
				// }
				m.onSelectMessage = rye.SprintfInfo("switch to server: %s", selectedServer.Server.Name)
			}

			if _, err := tea.NewProgram(&m).Run(); err != nil {
				return fmt.Errorf("list servers err: %w", err)
			}

			return nil
		},
	}
}
