package memberagent

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"time"

	cmcfgv1 "github.com/openshift/cluster-etcd-operator/pkg/apis/etcd.openshift.io/v1"
	mapi "github.com/openshift/cluster-etcd-operator/pkg/apis/etcd.openshift.io/v1"
	clustermemberclient "github.com/openshift/cluster-etcd-operator/pkg/generated/clientset/versioned/typed/etcd.openshift.io/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"

	"k8s.io/client-go/tools/clientcmd"
)

// memberAgentOpts holds values to drive the member-agent command.
type memberOpts struct {
	errOut io.Writer

	memberName    string
	peerURLs      []string
	etcdConfigDir string
	kubeconfig    string
}

// NewRenderCommand creates a render command.
func NewMemberCommand(errOut io.Writer) *cobra.Command {
	memberOpts := memberOpts{
		errOut:        errOut,
		etcdConfigDir: "/etc/etcd",
	}
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Add new member to cluster",
		Run: func(cmd *cobra.Command, args []string) {
			must := func(fn func() error) {
				if err := fn(); err != nil {
					if cmd.HasParent() {
						klog.Fatal(err)
					}
					fmt.Fprint(memberOpts.errOut, err.Error())
				}
			}

			must(memberOpts.Validate)
			must(memberOpts.Run)
		},
	}

	memberOpts.AddFlags(cmd.Flags())

	return cmd
}

func (m *memberOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&m.memberName, "name", m.memberName, "Name of etcd member to be added to clustem.")
	fs.StringSliceVar(&m.peerURLs, "peer-urls", m.peerURLs, "Comma serperated list of this member's peer URLs to advertise to the rest of the clustem.")
	fs.StringVar(&m.etcdConfigDir, "etcd-config-dir", m.etcdConfigDir, "Path to etcd config directory")
	fs.StringVar(&m.kubeconfig, "kubeconfig", m.kubeconfig, "Path to the kubeconfig file to connect to apiservem. If \"\", InClusterConfig is used which uses the service account kubernetes gives to pods.")
}

func (m *memberOpts) Validate() error {
	if m.memberName == "" {
		return errors.New("missing required flag: --name")
	}
	if len(m.peerURLs) == 0 {
		return errors.New("missing required flag: --peer-urls")
	}
	if m.etcdConfigDir == "" {
		return errors.New("missing required flag: --etcd-config-dir")
	}
	if m.kubeconfig == "" {
		return errors.New("missing required flag: --kubeconfig")
	}
	return nil

}

func (m *memberOpts) Run() error {
	config := ClusterMemberConfig{
		Name:     m.memberName,
		PeerURLs: m.peerURLs,
	}
	a, err := NewAgent(config, m.kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating agent: %s", err)
	}
	if err := a.RequestClusterMemberConfig(); err != nil {
		return fmt.Errorf("error requesting etcd config: %s", err)
	}
	return nil
}

type ClusterMemberConfig struct {
	Name          string
	PeerURLs      []string
	EtcdConfigDir string
}

// ClusterMemberAgent
type ClusterMemberAgent struct {
	// client implements the ClusterMemberInterface
	client clustermemberclient.ClusterMemberInterface
	// config holds all the new member values.
	config ClusterMemberConfig
}

// NewAgent returns an initialized ClusterClusterMemberAgent instance or an error is unsuccessful
func NewAgent(config ClusterMemberConfig, kubeconfigFile string) (*ClusterMemberAgent, error) {
	content, err := ioutil.ReadFile(kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", kubeconfigFile, err)
	}
	kubeconfig, err := clientcmd.Load(content)
	if err != nil {
		return nil, fmt.Errorf("error reading config from bytes: %v", err)
	}
	clientConfig, err := clientcmd.NewDefaultClientConfig(*kubeconfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating client config: %v", err)
	}

	client, err := clustermemberclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %v", err)
	}

	return &ClusterMemberAgent{
		client: client.ClusterMembers(),
		config: config,
	}, nil
}

// RequestClusterMemberConfig
func (c *ClusterMemberAgent) RequestClusterMemberConfig() error {
	cm := &cmcfgv1.ClusterMember{
		metav1.TypeMeta{
			Kind:       "ClusterMember",
			APIVersion: "etcd.openshift.io/v1",
		},
		metav1.ObjectMeta{},
		cmcfgv1.ClusterMemberSpec{
			Name:     c.config.Name,
			PeerURLs: c.config.PeerURLs,
			// EtcdConfigDir: c.config.etcdConfigDir,
		},
		cmcfgv1.ClusterMemberStatus{},
	}

	duration := 10 * time.Second
	// wait forever for success and retry every duration interval
	wait.PollInfinite(duration, func() (bool, error) {
		_, err := c.client.Create(cm)
		if err != nil {
			klog.Errorf("error sending ClusterMember request: %v", err)
			return false, nil
		}
		return true, nil
	})

	rcvdConfig, err := c.WaitForConfig()
	if err != nil {
		return fmt.Errorf("error obtaining etcd configuration: %v", err)
	}

	// write out etcd configuration to disk.
	configFile := path.Join(c.config.EtcdConfigDir, "etcd.conf")
	if err := ioutil.WriteFile(configFile, rcvdConfig.Status.Config, 0644); err != nil {
		return fmt.Errorf("unable to write to %s: %v", configFile, err)
	}
	return nil
}

// ClusterMemberAgent
func (c *ClusterMemberAgent) WaitForConfig() (req *mapi.ClusterMember, err error) {
	interval := 3 * time.Second
	timeout := 10 * time.Second

	// implement the client GET request to the signer in a poll loop.
	if err = wait.PollImmediate(interval, timeout, func() (bool, error) {
		req, err = c.client.Get(c.config.Name, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("unable to retrieve approved ClusterMember: %v. Retrying.", err)
			return false, nil
		}
		// if a CSR is returned without explicitly being `approved` or `denied` we want to retry
		if approved, denied := GetClusterMemberCondition(&req.Status); !approved && !denied {
			klog.Error("status on ClusterMember not set. Retrying.")
			return false, nil
		}
		// if a CSR is returned with `approved` status set and no signed certificate we want to retry
		if IsClusterMemberRequestApproved(req) && len(req.Status.Config) == 0 {
			klog.Error("status on ClusterMember set to `approved` but config is empty. Retrying.")
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}
	return
}

func IsClusterMemberRequestApproved(cm *mapi.ClusterMember) bool {
	approved, denied := GetClusterMemberCondition(&cm.Status)
	return approved && !denied
}

func GetClusterMemberCondition(status *mapi.ClusterMemberStatus) (approved bool, denied bool) {
	for _, c := range status.Conditions {
		if c.Type == mapi.ClusterMemberApproved {
			approved = true
		}
		if c.Type == mapi.ClusterMemberDenied {
			denied = true
		}
	}
	return
}
