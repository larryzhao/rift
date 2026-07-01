package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/larryzhao/rift"
	"github.com/spf13/cobra"
)

type statusLine struct {
	Label  string
	Value  string
	Detail string
	OK     bool
}

var (
	statusTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170"))
	statusMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
	statusOKStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))
	statusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("160"))
	statusLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))
	statusPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("238")).
				Padding(0, 1).
				MarginBottom(1)
)

func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use: "status",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)
			fmt.Print(renderStatus(repo))
			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
}

func renderStatus(repo *rift.Repo) string {
	return renderStatusPanel("Runtime", runtimeStatusLines(repo)) + "\n"
}

func runtimeStatusLines(repo *rift.Repo) []statusLine {
	proxyOK, proxyErr := repo.Status.IsProxyRunning()
	pacOK, pacErr := repo.Status.IsPACServerRunning()
	proxySetOK, proxySetErr := repo.Status.IsProxySet()
	autoUpdateOK, autoUpdateErr := repo.Status.IsAutoUpdateRunning()

	protocol := "unknown"
	if repo.Status.Protocl != rift.ProtoclUnknown && repo.Status.Protocl != "" {
		protocol = repo.Status.Protocl.String()
	}

	currentServer := protocol
	if repo.Status.ServerName != "" {
		currentServer = strings.Join(nonEmptyStrings(repo.Status.ServerGroup, repo.Status.ServerName, protocol), ", ")
	}

	lines := []statusLine{
		{
			Label:  "Proxy",
			Value:  statusValue(proxyOK),
			Detail: statusDetail(proxyErr, currentServer),
			OK:     proxyOK && proxyErr == nil,
		},
		{
			Label:  "PAC server",
			Value:  statusValue(pacOK),
			Detail: statusDetail(pacErr, "http://127.0.0.1:60061/pac/proxy.js"),
			OK:     pacOK && pacErr == nil,
		},
		{
			Label:  "Wi-Fi autoproxy",
			Value:  enabledValue(proxySetOK),
			Detail: statusDetail(proxySetErr, ""),
			OK:     proxySetOK && proxySetErr == nil,
		},
		{
			Label:  "Autoupdate",
			Value:  statusValue(autoUpdateOK),
			Detail: statusDetail(autoUpdateErr, ""),
			OK:     autoUpdateOK && autoUpdateErr == nil,
		},
	}

	lines = append(lines, gfwlistStatusLine(repo))
	lines = append(lines, subscriptionStatusLines(repo)...)
	return lines
}

func gfwlistStatusLine(repo *rift.Repo) statusLine {
	st := repo.Status.GFWList
	line := statusLine{
		Label: "  gfwlist",
		Value: "unknown",
		OK:    false,
	}

	if st.LastSuccessAt > 0 {
		line.Value = "updated"
		line.Detail = fmt.Sprintf("last success %s", formatStatusTime(st.LastSuccessAt))
		line.OK = true
	}

	if st.LastError != "" {
		line.Value = "error"
		line.Detail = fmt.Sprintf("last attempt %s: %s", formatStatusTime(st.LastAttemptAt), st.LastError)
		line.OK = false
	}

	return line
}

func subscriptionStatusLines(repo *rift.Repo) []statusLine {
	subs := repo.Settings.SubscriptionItems()
	if len(subs) == 0 {
		return []statusLine{{
			Label:  "  Subscriptions",
			Value:  "none",
			Detail: "no subscriptions configured",
			OK:     false,
		}}
	}

	lines := make([]statusLine, 0, len(subs))
	for _, sub := range subs {
		subStatus, ok := repo.Status.SubscriptionStatuses[sub.Name]
		if !ok {
			lines = append(lines, statusLine{
				Label:  fmt.Sprintf("  %s", sub.Name),
				Value:  "unknown",
				Detail: "no update record",
				OK:     false,
			})
			continue
		}

		line := statusLine{
			Label: fmt.Sprintf("  %s", sub.Name),
			Value: "unknown",
			OK:    false,
		}

		if subStatus.LastSuccessAt > 0 {
			line.Value = "updated"
			line.Detail = fmt.Sprintf("last success %s", formatStatusTime(subStatus.LastSuccessAt))
			line.OK = true
		}

		if subStatus.LastError != "" {
			line.Value = "error"
			line.Detail = fmt.Sprintf("last attempt %s: %s", formatStatusTime(subStatus.LastAttemptAt), subStatus.LastError)
			line.OK = false
		}

		lines = append(lines, line)
	}

	return lines
}

func renderStatusPanel(title string, lines []statusLine) string {
	labelWidth := statusPanelLabelWidth(lines)
	valueWidth := statusPanelValueWidth(lines)
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		rendered = append(rendered, renderStatusLine(line, labelWidth, valueWidth))
	}
	return statusPanelStyle.Render(strings.Join(rendered, "\n"))
}

func renderStatusLine(line statusLine, labelWidth int, valueWidth int) string {
	valueStyle := statusErrorStyle
	if line.OK {
		valueStyle = statusOKStyle
	}

	label := statusLabelStyle.Width(labelWidth).Render(line.Label)
	value := valueStyle.Width(valueWidth).Render(line.Value)
	if line.Detail == "" {
		return fmt.Sprintf("%s  %s", label, value)
	}
	return fmt.Sprintf("%s  %s  %s", label, value, statusMutedStyle.Render(compactStatusDetail(line.Detail)))
}

func statusPanelLabelWidth(lines []statusLine) int {
	width := 0
	for _, line := range lines {
		if lipgloss.Width(line.Label) > width {
			width = lipgloss.Width(line.Label)
		}
	}
	return width
}

func statusPanelValueWidth(lines []statusLine) int {
	width := 0
	for _, line := range lines {
		if lipgloss.Width(line.Value) > width {
			width = lipgloss.Width(line.Value)
		}
	}
	return width
}

func statusValue(ok bool) string {
	if ok {
		return "running"
	}
	return "stopped"
}

func enabledValue(ok bool) string {
	if ok {
		return "enabled"
	}
	return "disabled"
}

func statusDetail(err error, fallback string) string {
	if err != nil {
		return err.Error()
	}
	return fallback
}

func formatStatusTime(unixTime int64) string {
	if unixTime <= 0 {
		return "unknown"
	}
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04")
}

func compactStatusDetail(detail string) string {
	const maxWidth = 56
	if lipgloss.Width(detail) <= maxWidth {
		return detail
	}

	var b strings.Builder
	for _, r := range detail {
		if lipgloss.Width(b.String()+string(r)) > maxWidth-3 {
			break
		}
		b.WriteRune(r)
	}
	return b.String() + "..."
}

func nonEmptyStrings(values ...string) []string {
	nonEmpty := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		nonEmpty = append(nonEmpty, value)
	}
	return nonEmpty
}
