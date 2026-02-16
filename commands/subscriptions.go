package commands

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/larryzhao/rye"
	"github.com/spf13/cobra"
)

func NewSubscriptionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(newSubscriptionsUpdateCmd())
	cmd.AddCommand(newSubscriptionsAddCmd())
	cmd.AddCommand(newSubscriptionsAutoUpdateCmd())

	return cmd
}

func newSubscriptionsUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			updatedSubs, err := repo.UpdateSubscriptions()
			if err != nil {
				return err
			}

			var updatedSubNames []string
			for _, sub := range updatedSubs {
				updatedSubNames = append(updatedSubNames, sub.Name)
			}
			rye.PrintlnInfo("%s updated", strings.Join(updatedSubNames, ","))

			return nil
		},
	}
}

func newSubscriptionsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use: "add",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			name := args[0]
			url := args[1]

			if name == "" {
				return fmt.Errorf("you need to provide a name for the subscription")
			}

			if url == "" {
				return fmt.Errorf("you need to provide a url for the subscription")
			}

			sub := &rye.Subscription{
				Name:       name,
				URL:        url,
				AddedAt:    time.Now(),
				SkipUpdate: false,
			}

			err := repo.AddSubscription(sub)
			if err != nil {
				return err
			}

			rye.PrintlnInfo("Subscription %s added.", name)
			return nil
		},
	}
}

func newSubscriptionsAutoUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "autoupdate",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rye.CtxKeyRepo).(*rye.Repo)

			interval, err := cmd.Flags().GetDuration("interval")
			if err != nil {
				return err
			}
			if interval <= 0 {
				return fmt.Errorf("interval must be greater than 0")
			}

			daemon, err := cmd.Flags().GetBool("daemon")
			if err != nil {
				return err
			}

			if daemon {
				return runSubscriptionsAutoUpdateDaemon(repo, interval)
			}

			pid, err := readAutoUpdatePID(repo)
			if err != nil {
				return err
			}
			if pid > 0 {
				running, err := rye.Running(pid)
				if err != nil {
					return fmt.Errorf("check running autoupdate daemon err: %w", err)
				}
				if running {
					rye.PrintlnInfo("autoupdate daemon already running with pid %d", pid)
					return nil
				}
			}

			executable, err := os.Executable()
			if err != nil {
				return fmt.Errorf("resolve executable err: %w", err)
			}
			pid, err = rye.Run(executable, []string{
				"subscriptions", "autoupdate", "--daemon", fmt.Sprintf("--interval=%s", interval),
			})
			if err != nil {
				return fmt.Errorf("start autoupdate daemon err: %w", err)
			}

			rye.PrintlnInfo("autoupdate daemon started, pid=%d", pid)
			rye.PrintlnInfo("autoupdate log file: %s", repo.AutoUpdateLogFile())
			return nil
		},
	}

	cmd.Flags().Duration("interval", 30*time.Minute, "auto update interval")
	cmd.Flags().Bool("daemon", false, "internal flag for daemon mode")
	err := cmd.Flags().MarkHidden("daemon")
	if err != nil {
		panic(err)
	}
	return cmd
}

func runSubscriptionsAutoUpdateDaemon(repo *rye.Repo, interval time.Duration) error {
	err := ensureAutoUpdatePID(repo)
	if err != nil {
		return err
	}
	defer os.Remove(repo.AutoUpdatePIDFile())

	err = appendAutoUpdateLog(repo, "daemon started; interval=%s", interval)
	if err != nil {
		return err
	}

	for {
		stats, updateErr := repo.UpdateSubscriptionsWithStats()

		updatedSubs := 0
		totalChanged := 0
		for _, stat := range stats {
			if stat.Updated {
				updatedSubs++
				totalChanged += stat.ChangedServers
			}
		}

		if updateErr != nil {
			_ = appendAutoUpdateLog(repo, "run failed; err=%s", updateErr.Error())
		} else {
			_ = appendAutoUpdateLog(repo, "run finished; updated_subscriptions=%d; changed_servers=%d", updatedSubs, totalChanged)
		}

		for _, stat := range stats {
			if stat.Err != nil {
				_ = appendAutoUpdateLog(repo, "subscription=%s; status=failed; err=%s", stat.Name, stat.Err.Error())
				continue
			}
			_ = appendAutoUpdateLog(
				repo,
				"subscription=%s; status=updated; changed_servers=%d; previous_servers=%d; current_servers=%d",
				stat.Name,
				stat.ChangedServers,
				stat.PreviousServers,
				stat.CurrentServers,
			)
		}

		time.Sleep(interval)
	}
}

func ensureAutoUpdatePID(repo *rye.Repo) error {
	pid, err := readAutoUpdatePID(repo)
	if err != nil {
		return err
	}

	if pid > 0 {
		running, err := rye.Running(pid)
		if err != nil {
			return fmt.Errorf("check running autoupdate daemon err: %w", err)
		}
		if running {
			return fmt.Errorf("autoupdate daemon already running with pid %d", pid)
		}
	}

	currentPID := os.Getpid()
	err = os.WriteFile(repo.AutoUpdatePIDFile(), []byte(strconv.Itoa(currentPID)), 0644)
	if err != nil {
		return fmt.Errorf("write autoupdate pid file err: %w", err)
	}
	return nil
}

func readAutoUpdatePID(repo *rye.Repo) (int, error) {
	bb, err := os.ReadFile(repo.AutoUpdatePIDFile())
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read autoupdate pid file err: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(bb)))
	if err != nil {
		return 0, fmt.Errorf("parse autoupdate pid file err: %w", err)
	}
	return pid, nil
}

func appendAutoUpdateLog(repo *rye.Repo, format string, args ...interface{}) error {
	logFile := repo.AutoUpdateLogFile()
	err := os.MkdirAll(path.Dir(logFile), 0755)
	if err != nil {
		return fmt.Errorf("create autoupdate log directory err: %w", err)
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open autoupdate log file err: %w", err)
	}
	defer file.Close()

	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	_, err = file.WriteString(line)
	if err != nil {
		return fmt.Errorf("write autoupdate log file err: %w", err)
	}
	return nil
}
