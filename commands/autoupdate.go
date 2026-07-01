package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/larryzhao/rift"
	"github.com/larryzhao/rift/pac"
	"github.com/spf13/cobra"
)

// Command autoupdate
//
//	rift autoupdate start
//	rift autoupdate stop
//
// Runs a background daemon that periodically refreshes proxy subscriptions and
// the gfwlist used by the PAC server. Each feature is scheduled independently
// based on its `autoupdate` block in settings.yaml (interval + skip).
func NewAutoUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autoupdate",
		Short: "Automatically refresh subscriptions and gfwlist in the background",
	}

	cmd.AddCommand(newAutoUpdateStartCmd())
	cmd.AddCommand(newAutoUpdateStopCmd())
	cmd.AddCommand(newAutoUpdateDaemonCmd())

	return cmd
}

func newAutoUpdateStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the autoupdate daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			pid := repo.Status.PIDByKind("autoupdate")
			if pid > 0 {
				running, err := repo.Status.IsAutoUpdateRunning()
				if err != nil {
					return fmt.Errorf("check running autoupdate daemon err: %w", err)
				}
				if running {
					rift.PrintlnInfo("autoupdate daemon already running with pid %d", pid)
					return nil
				}
			}

			executable, err := os.Executable()
			if err != nil {
				return fmt.Errorf("resolve executable err: %w", err)
			}
			pid, err = rift.Run(executable, []string{"autoupdate", "daemon"})
			if err != nil {
				return fmt.Errorf("start autoupdate daemon err: %w", err)
			}
			repo.Status.UpdateRunningProcess("autoupdate", pid)
			if err := repo.SaveStatus(); err != nil {
				return fmt.Errorf("save status err: %w", err)
			}

			rift.PrintlnInfo("autoupdate daemon started, pid=%d", pid)
			rift.PrintlnInfo("autoupdate log file: %s", repo.AutoUpdateLogFile())
			return nil
		},
	}
}

func newAutoUpdateStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the autoupdate daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)

			pid := repo.Status.PIDByKind("autoupdate")
			if pid <= 0 {
				rift.PrintlnInfo("autoupdate daemon is not running")
				return nil
			}

			if err := rift.Stop(pid); err != nil {
				return fmt.Errorf("stop autoupdate daemon err: %w", err)
			}
			repo.Status.ClearRunningProcess("autoupdate")
			if err := repo.SaveStatus(); err != nil {
				return fmt.Errorf("save status err: %w", err)
			}

			rift.PrintlnInfo("autoupdate daemon stopped")
			return nil
		},
	}
}

// newAutoUpdateDaemonCmd is the hidden long-running worker spawned by
// `autoupdate start`. It is not meant to be invoked directly.
func newAutoUpdateDaemonCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Context().Value(rift.CtxKeyRepo).(*rift.Repo)
			return runAutoUpdateDaemon(repo)
		},
	}
}

func runAutoUpdateDaemon(repo *rift.Repo) error {
	currentPID := os.Getpid()
	if pid := repo.Status.PIDByKind("autoupdate"); pid > 0 {
		running, err := repo.Status.IsAutoUpdateRunning()
		if err != nil {
			return fmt.Errorf("check running autoupdate daemon err: %w", err)
		}
		if running && pid != currentPID {
			return fmt.Errorf("autoupdate daemon already running with pid %d", pid)
		}
	}

	repo.Status.UpdateRunningProcess("autoupdate", currentPID)
	if err := repo.SaveStatus(); err != nil {
		return fmt.Errorf("save status err: %w", err)
	}

	defer func() {
		repo.Status.ClearRunningProcess("autoupdate")
		_ = repo.SaveStatus()
	}()

	subCfg := repo.Settings.SubscriptionsAutoUpdate()
	gfwCfg := repo.Settings.GFWListAutoUpdate()

	_ = repo.AppendAutoUpdateLog(
		"daemon started; subscriptions{skip=%v interval=%s}; gfwlist{skip=%v interval=%s}",
		subCfg.Skip, subCfg.Interval.Duration(), gfwCfg.Skip, gfwCfg.Interval.Duration(),
	)

	if subCfg.Skip && gfwCfg.Skip {
		_ = repo.AppendAutoUpdateLog("nothing to do; both subscriptions and gfwlist are skipped")
		return nil
	}

	now := time.Now()
	subNext := now
	gfwNext := now

	for {
		now := time.Now()

		if !subCfg.Skip && !now.Before(subNext) {
			updateSubscriptionsOnce(repo)
			subNext = time.Now().Add(subCfg.Interval.Duration())
		}
		if !gfwCfg.Skip && !now.Before(gfwNext) {
			updateGFWListOnce(repo)
			gfwNext = time.Now().Add(gfwCfg.Interval.Duration())
		}

		var next time.Time
		consider := func(t time.Time) {
			if next.IsZero() || t.Before(next) {
				next = t
			}
		}
		if !subCfg.Skip {
			consider(subNext)
		}
		if !gfwCfg.Skip {
			consider(gfwNext)
		}
		if d := time.Until(next); d > 0 {
			time.Sleep(d)
		}
	}
}

func updateSubscriptionsOnce(repo *rift.Repo) {
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
		_ = repo.AppendAutoUpdateLog("subscriptions run failed; err=%s", updateErr.Error())
	} else {
		_ = repo.AppendAutoUpdateLog("subscriptions run finished; updated_subscriptions=%d; changed_servers=%d", updatedSubs, totalChanged)
	}

	now := time.Now()
	for _, stat := range stats {
		if stat.Err != nil {
			repo.Status.UpdateSubscriptionStatus(stat.Name, now, nil, stat.Err)
			_ = repo.AppendAutoUpdateLog("subscription=%s; status=failed; err=%s", stat.Name, stat.Err.Error())
			continue
		}
		repo.Status.UpdateSubscriptionStatus(stat.Name, now, &now, nil)
		_ = repo.AppendAutoUpdateLog(
			"subscription=%s; status=updated; changed_servers=%d; previous_servers=%d; current_servers=%d",
			stat.Name,
			stat.ChangedServers,
			stat.PreviousServers,
			stat.CurrentServers,
		)
	}
	if err := repo.SaveStatus(); err != nil {
		_ = repo.AppendAutoUpdateLog("save status failed; err=%s", err.Error())
	}
}

func updateGFWListOnce(repo *rift.Repo) {
	prev := repo.Status.GFWList
	res, err := pac.SyncGFWList(repo.PACGFWListFile(), prev.ETag, prev.SHA256)

	now := time.Now()
	if err != nil {
		repo.Status.UpdateGFWListStatus(now, nil, prev.ETag, prev.SHA256, err)
		_ = repo.AppendAutoUpdateLog("gfwlist run failed; err=%s", err.Error())
	} else if res.Changed {
		repo.Status.UpdateGFWListStatus(now, &now, res.ETag, res.SHA256, nil)
		_ = repo.AppendAutoUpdateLog("gfwlist run finished; status=updated; bytes=%d; sha256=%s", res.Bytes, res.SHA256)
	} else {
		repo.Status.UpdateGFWListStatus(now, &now, res.ETag, res.SHA256, nil)
		_ = repo.AppendAutoUpdateLog("gfwlist run finished; status=unchanged")
	}

	if err := repo.SaveStatus(); err != nil {
		_ = repo.AppendAutoUpdateLog("save status failed; err=%s", err.Error())
	}
}
