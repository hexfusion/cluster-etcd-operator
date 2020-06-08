package rollbackcopy

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"k8s.io/klog"
	"math/rand"
	"os"
	"time"
)

type rollbackCopyOpts struct {
	checkLeadershipInterval time.Duration
	rollbackCopyInterval    time.Duration
	configDir               string
	errOut                  io.Writer
}

const (
	leaderShipCheckInterval time.Duration = 5 * time.Minute
	rollbackSaveInterval    time.Duration = 60 * time.Minute
)

func NewRollbackCopy(errOut io.Writer) *cobra.Command {
	rollbackCopyOpts := &rollbackCopyOpts{
		checkLeadershipInterval: leaderShipCheckInterval,
		rollbackCopyInterval:    rollbackSaveInterval,
		errOut:                  errOut,
	}
	cmd := &cobra.Command{
		Use:   "rollbackcopy",
		Short: "Periodically save snapshot and resources every hour, useful for a rollback",
		Run: func(cmd *cobra.Command, args []string) {
			must := func(fn func() error) {
				if err := fn(); err != nil {
					if cmd.HasParent() {
						klog.Fatal(err)
					}
					fmt.Fprint(rollbackCopyOpts.errOut, err.Error())
				}
			}

			must(rollbackCopyOpts.Run)
		},
	}
	rollbackCopyOpts.AddFlags(cmd.Flags())
	return cmd
}

func (r *rollbackCopyOpts) AddFlags(fs *pflag.FlagSet) {
	fs.Set("logtostderr", "true")
	fs.StringVar(&r.configDir, "config-dir", "/etc/kubernetes", "Dir containing kubernetes resources")
}

func checkAndScheduleBackup(memberName string, delay time.Duration, configDir string, done <-chan struct{}) {
	backupTicker := time.NewTicker(delay)
	for {
		cli, err := getEtcdClient([]string{"localhost:2379"})
		if err != nil {
			klog.Error("Failed to get etcd client")
			return
		}
		klog.Info(" checking amLeader in schedule...")
		if !checkLeadership(cli, memberName) {
			klog.Info("Member is NOT the leader. Returning")
			cli.Close()
			return
		}
		klog.Info("Member IS the leader. Backing up")
		backup(cli, configDir)
		cli.Close()

		select {
		case <-done:
			backupTicker.Stop()
			return
		case <-backupTicker.C:
			continue
		}
	}
}

func (r *rollbackCopyOpts) Run() error {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	localEtcdName := os.Getenv("ETCD_NAME")
	ticker := time.NewTicker(r.checkLeadershipInterval)
	go func() {
		for {
			checkAndScheduleBackup(localEtcdName, r.rollbackCopyInterval, r.configDir, ctx.Done())
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				continue
			}
		}
	}()

	<-ctx.Done()
	klog.Info("Done !!")
	return nil
}
