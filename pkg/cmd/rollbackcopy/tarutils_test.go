package rollbackcopy

import (
	"archive/tar"
	"testing"
)

func Test_addFileToTarWriter(t *testing.T) {
	type args struct {
		src        string
		tarWriter  *tar.Writer
		prefixTrim string
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
			if err := addFileToTarWriter(tt.args.src, tt.args.tarWriter, tt.args.prefixTrim); (err != nil) != tt.wantErr {
				t.Errorf("addFileToTarWriter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createTarball(t *testing.T) {
	type args struct {
		tarballFilePath string
		filePaths       []string
		prefixTrim      string
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
			if err := createTarball(tt.args.tarballFilePath, tt.args.filePaths, tt.args.prefixTrim); (err != nil) != tt.wantErr {
				t.Errorf("createTarball() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
