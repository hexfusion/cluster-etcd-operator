package rollbackcopy

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"k8s.io/klog"
	"os"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

func getEtcdClient(endpoints []string) (*clientv3.Client, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithBlock(), // block until the underlying connection is up
	}

	tlsInfo := transport.TLSInfo{
		CertFile:      os.Getenv("ETCDCTL_CERT"),
		KeyFile:       os.Getenv("ETCDCTL_KEY"),
		TrustedCAFile: os.Getenv("ETCDCTL_CACERT"),
	}
	tlsConfig, err := tlsInfo.ClientConfig()

	cfg := &clientv3.Config{
		DialOptions: dialOptions,
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
		TLS:         tlsConfig,
	}

	cli, err := clientv3.New(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to make etcd client for endpoints %v: %w", endpoints, err)
	}
	return cli, nil
}

func getClusterVersionAndUpgradeInfo(cli *clientv3.Client) (string, bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clusterversionKey := "/kubernetes.io/config.openshift.io/clusterversions/version"
	resp, err := cli.Get(ctx, clusterversionKey)
	if err != nil {
		return "", false, err
	}
	klog.Info("### Response for client GET: ", resp)
	if len(resp.Kvs) != 1 {
		klog.Errorf("Expected to get a single key from etcd, got %d", len(resp.Kvs))
		return "", false, fmt.Errorf("Expected to get a single key from etcd, got %d", len(resp.Kvs))
	}

	klog.Info("### Value lookup: ", string(resp.Kvs[0].Value))
	var cv map[string]interface{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &cv); err != nil {
		return "", false, err
	}

	status := cv["status"].(map[string]interface{})
	desired := status["desired"].(map[string]interface{})
	klog.Info("desired version: ", desired["version"])

	history := status["history"].([]interface{})
	latestHistory := history[0].(map[string]interface{})
	klog.Info("latest history version:", latestHistory["version"], " status: ", latestHistory["state"])
	klog.Info("Return values: ", latestHistory["version"], " upgradeInProgress: ", desired["version"] != latestHistory["version"])

	return latestHistory["version"].(string), desired["version"].(string) != latestHistory["version"].(string), nil
}

func isLeader(cli *clientv3.Client, name string) (bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	membersResp, err := cli.MemberList(ctx)
	if err != nil {
		return false, err
	}
	for _, member := range membersResp.Members {
		if member.Name != name {
			continue
		}
		if len(member.ClientURLs) == 0 && member.Name == "" {
			return false, fmt.Errorf("EtcdMemberNotStarted")
		}

		resp, err := cli.Status(ctx, member.ClientURLs[0])
		if err != nil {
			klog.Errorf("error getting etcd member %s status: %#v", member.Name, err)
			return false, err
		}
		return resp.Header.MemberId == resp.Leader, nil
	}
	return false, fmt.Errorf("EtcdMemberStatusUnknown")
}

func SaveSnapshot(cli *clientv3.Client, dbPath string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	partpath := dbPath + ".part"
	defer os.RemoveAll(partpath)

	f, err := os.OpenFile(partpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("could not open %s (%v)", partpath, err)
	}

	opBegin := time.Now()
	var rd io.ReadCloser
	rd, err = cli.Snapshot(ctx)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, rd); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	klog.Info(
		"fetched snapshot. ",
		"Time taken: ", time.Since(opBegin),
	)

	if err := os.Rename(partpath, dbPath); err != nil {
		return fmt.Errorf("could not rename %s to %s (%v)", partpath, dbPath, err)
	}
	klog.Info("saved snapshot to path", dbPath)
	return nil
}
