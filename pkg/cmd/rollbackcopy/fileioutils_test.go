package rollbackcopy

import (
	"reflect"
	"testing"
)

func Test_checkAndCreateDir(t *testing.T) {
	type args struct {
		dirName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkAndCreateDir(tt.args.dirName); (err != nil) != tt.wantErr {
				t.Errorf("checkAndCreateDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dirExists(t *testing.T) {
	type args struct {
		dirname string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dirExists(tt.args.dirname); got != tt.want {
				t.Errorf("dirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fileExists(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileExists(tt.args.filename); got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findTheLatestRevision(t *testing.T) {
	type args struct {
		dir     string
		podname string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findTheLatestRevision(tt.args.dir, tt.args.podname)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTheLatestRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findTheLatestRevision() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVersion(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    *backupVersion
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVersion(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_putVersion(t *testing.T) {
	type args struct {
		c   *backupVersion
		dir string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := putVersion(tt.args.c, tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("putVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_safeDirRename(t *testing.T) {
	type args struct {
		src            string
		dest           string
		srcMayNotExist bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := safeDirRename(tt.args.src, tt.args.dest, tt.args.srcMayNotExist); (err != nil) != tt.wantErr {
				t.Errorf("safeDirRename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_versionChanged(t *testing.T) {
	type args struct {
		dir1 string
		dir2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionChanged(tt.args.dir1, tt.args.dir2); got != tt.want {
				t.Errorf("versionChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}
