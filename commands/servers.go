package commands

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/larryzhao/rye"
	"github.com/larryzhao/rye/repo"
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
	list     list.Model
	logger   *slog.Logger
	onSelect func(ctx context.Context, item *item)
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
	if m.choice != "" {
		return ""
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
			r, err := repo.LoadRepo()
			if err != nil {
				return err
			}

			cliLog, err := os.OpenFile(path.Join(r.Dir, "cli.log"), os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				return err
			}

			logger := slog.New(slog.NewJSONHandler(cliLog, nil))
			logger.Info("servers command entered")

			var items []list.Item
			for idx, srv := range r.Servers {
				items = append(items, item{
					Index: idx,
					Name:  srv.Server.Name,
				})
			}

			const defaultWidth = 20

			l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
			l.Title = "What do you want for dinner?"
			l.SetShowStatusBar(false)
			l.SetFilteringEnabled(false)
			l.Styles.Title = titleStyle
			l.Styles.PaginationStyle = paginationStyle
			l.Styles.HelpStyle = helpStyle

			m := model{list: l, logger: logger}
			m.onSelect = func(ctx context.Context, item *item) {
				selectedServer := r.Servers[item.Index]

				outbound, err := selectedServer.Server.ToOutbound()
				if err != nil {
					logger.Info(fmt.Sprintf("parse server #%d: %s to outbound err: %s", item.Index, selectedServer.Server.Name, err.Error()))
					rye.PrintError("parse server #%d: %s to outbound err: %s", item.Index, selectedServer.Server.Name, err.Error())
					return
				}

				r.XrayConfig.SetOutbound("proxy", outbound)
				err = r.XrayConfig.Save()
				if err != nil {
					logger.Info(fmt.Sprintf("save xray/config.json err: %s", err.Error()))
					rye.PrintError("save xray/config.json err: %s", err.Error())
					return
				}

				// stop runner
				err = rye.StopRunner(r.PID)
				if err != nil {
					logger.Info(fmt.Sprintf("stop runner err: %s", err.Error()), slog.Int("runner_pid", r.PID))
					rye.PrintError("stop runner err: %s", err.Error())
					return
				}
				logger.Info("runner stopped")

				// start runner again
				pid, err := rye.StartRunner()
				if err != nil {
					logger.Info(fmt.Sprintf("start runner err: %s", err.Error()))
					return
				}
				logger.Info(fmt.Sprintf("runner started: %d", pid))
				if err := r.WritePID(pid); err != nil {
					logger.Info(fmt.Sprintf("write pid err: %s", err.Error()))
					return
				}
				logger.Info("runner pid written")
			}

			if _, err := tea.NewProgram(m).Run(); err != nil {
				fmt.Println("Error running program:", err)
				return err
			}

			return nil
		},
	}
}
