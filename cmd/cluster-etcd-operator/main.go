package main

import (
	goflag "flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/openshift/cluster-etcd-operator/pkg/cmd/certsigner"
	"github.com/openshift/cluster-etcd-operator/pkg/cmd/render"
	"github.com/openshift/cluster-etcd-operator/pkg/cmd/setupetcd"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	command := NewSSCSCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func NewSSCSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster-etcd-operator",
		Short: "OpenShift cluster etcd operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	cmd.AddCommand(certsigner.NewCertSignerCommand(os.Stderr))
	cmd.AddCommand(render.NewRenderCommand(os.Stderr))
	cmd.AddCommand(setupetcd.NewSetupEtcdCommand(os.Stderr))

	return cmd
}
