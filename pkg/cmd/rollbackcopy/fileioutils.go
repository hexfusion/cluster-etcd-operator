package rollbackcopy

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func safeDirRename(src, dest string, srcMayNotExist bool) error {
	if !dirExists(src) {
		if !srcMayNotExist {
			return fmt.Errorf("SafeDirRename: src dir %s does not exist", src)
		} else {
			return nil
		}
	}
	destExists := false
	destToBeRemoved := dest + ".to_be_removed"
	if dirExists(dest) {
		destExists = true
		if dirExists(destToBeRemoved) {
			os.RemoveAll(destToBeRemoved)
		}
		if err := os.Rename(dest, destToBeRemoved); err != nil {
			klog.Error("Got error ", err)
			return err
		}
	}

	if err := os.Rename(src, dest); err != nil {
		if destExists {
			os.Rename(destToBeRemoved, dest)
		}
		klog.Error("Got error ", err)
		return err
	}
	klog.Info("Successfully moved ", src, " to ", dest)
	os.RemoveAll(destToBeRemoved)
	return nil
}

// copyDir copies a whole directory recursively

func findTheLatestRevision(dir, podname string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var modTime time.Time
	var latest string
	found := false
	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), podname) {
			if f.ModTime().After(modTime) {
				modTime = f.ModTime()
				latest = f.Name()
				found = true
			}
		}
	}
	if !found {
		return "", fmt.Errorf("Not found")
	}
	return filepath.Join(dir, latest), nil
}

type backupVersion struct {
	ClusterVersion string `yaml:"ClusterVersion"`
	TimeStamp      string `yaml:"TimeStamp"`
}

func versionChanged(dir1, dir2 string) bool {
	dir1Version, err := getVersion(dir1)
	if err != nil {
		return false
	}
	dir2Version, err := getVersion(dir2)
	if err != nil {
		return false
	}
	return (dir1Version.ClusterVersion != dir2Version.ClusterVersion)
}

func getVersion(dir string) (*backupVersion, error) {
	version := backupVersion{}
	yamlFile, err := ioutil.ReadFile(filepath.Join(dir, "backupenv.yaml"))
	if err != nil {
		klog.Warningf("getVersion: ReadFile #%v ", err)
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &version)
	if err != nil {
		klog.Warningf("getVersion: Unmarshal %v", err)
		return nil, err
	}

	return &version, nil
}

func putVersion(c *backupVersion, dir string) error {
	confBytes, err := yaml.Marshal(c)
	if err != nil {
		klog.Warningf("putVersion: Marshal err #%v ", err)
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dir, "backupenv.yaml"), confBytes, 0644)
	if err != nil {
		klog.Warningf("putVersion: WriteFile #%v ", err)
		return err
	}

	return nil
}

func checkAndCreateDir(dirName string) error {
	_, err := os.Stat(dirName)
	// If dirName already exists, remove it
	if err == nil || !os.IsNotExist(err) {
		os.RemoveAll(dirName)
	}
	errDir := os.MkdirAll(dirName, os.ModePerm)
	if errDir != nil {
		return errDir
	}
	return nil
}
