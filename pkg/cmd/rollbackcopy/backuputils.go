package rollbackcopy

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"time"
)

//This backup mimics the functionality of cluster-backup.sh

var backupResourcePodList = []string{
	"kube-apiserver-pod",
	"kube-controller-manager-pod",
	"kube-scheduler-pod",
	"etcd-pod",
}

func archiveLatestResources(configDir, backupFile string) error {
	klog.Info("In backup, backupFile is", backupFile)

	paths := []string{}
	for _, podName := range backupResourcePodList {
		latestPod, err := findTheLatestRevision(filepath.Join(configDir, "static-pod-resources"), podName)
		if err != nil {
			return err
		}
		paths = append(paths, latestPod)
		klog.Info("Adding the latest revision for podName ", podName, ": ", latestPod)
	}

	err := createTarball(backupFile, paths, configDir)
	if err != nil {
		klog.Error("Got error creating tar", err)
		return err
	}
	return nil
}

func backup(cli *clientv3.Client, configDir string) error {

	currentClusterVersion, upgradeInProgress, err := getClusterVersionAndUpgradeInfo(cli)
	if err != nil {
		return err
	}

	if upgradeInProgress {
		klog.Error("The cluster is being upgraded. Skipping backup!")
		return fmt.Errorf("The cluster is being upgraded. Skipping backup!")
	}

	tmpBackupDir := filepath.Join(configDir, "rollbackcopy", "tmp")
	defer os.RemoveAll(tmpBackupDir)

	if err := checkAndCreateDir(tmpBackupDir); err != nil {
		return err
	}

	// Trying to match the output file formats with the formats of the current cluster-backup.sh script
	dateString := time.Now().Format("2006-01-02_150405")
	outputArchive := "static_kuberesources_" + dateString + ".tar.gz"
	snapshotOutFile := "snapshot_" + dateString + ".db"

	// Save snapshot
	if err := SaveSnapshot(cli, filepath.Join(tmpBackupDir, snapshotOutFile)); err != nil {
		return err
	}

	// Save the corresponding static pod resources
	if err := archiveLatestResources(configDir, filepath.Join(tmpBackupDir, outputArchive)); err != nil {
		return err
	}

	// Write the version
	version := backupVersion{currentClusterVersion, dateString}
	putVersion(&version, tmpBackupDir)

	return checkVersionsAndMoveDirs(configDir, tmpBackupDir, upgradeInProgress)
}

func checkVersionsAndMoveDirs(configDir, newBackupDir string, beingUpgraded bool) error {
	if beingUpgraded {
		return nil
	}
	currentVersionlatestDir := filepath.Join(configDir, "rollbackcopy", "currentVersion.latest")
	currentVersionPrevDir := filepath.Join(configDir, "rollbackcopy", "currentVersion.prev")
	if versionChanged(currentVersionlatestDir, newBackupDir) {
		olderVersionlatestDir := filepath.Join(configDir, "rollbackcopy", "olderVersion.latest")
		olderVersionPrevDir := filepath.Join(configDir, "rollbackcopy", "olderVersion.prev")
		klog.Info("Version changed")
		if err := safeDirRename(currentVersionPrevDir, olderVersionPrevDir, true); err != nil {
			return err
		}
		if err := safeDirRename(currentVersionlatestDir, olderVersionlatestDir, true); err != nil {
			return err
		}
	} else {
		if err := safeDirRename(currentVersionlatestDir, currentVersionPrevDir, true); err != nil {
			return err
		}
	}
	if err := safeDirRename(newBackupDir, currentVersionlatestDir, false); err != nil {
		return err
	}
	klog.Info("Backed up resources and snapshot to ", currentVersionlatestDir)
	return nil
}

func checkLeadership(cli *clientv3.Client, name string) bool {
	//return rand.Float32() < 0.5
	flag, err := isLeader(cli, name)
	if err != nil {
		klog.Info("Failed to check leadership: ", err)
		return false
	}
	return flag
}
