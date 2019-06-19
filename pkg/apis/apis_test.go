package apis

import (
	"testing"
)

func TestGroupName(t *testing.T) {
	if got, want := GroupName, "etcd.openshift.io"; got != want {
		t.Fatalf("mismatch group name, got: %s want: %s", got, want)
	}
}
