package etcdcli

import "go.etcd.io/etcd/etcdserver/etcdserverpb"

type healthCheck struct {
	Member  *etcdserverpb.Member
	Health  bool
	Started bool
	Took    string
	Error   string
}

type memberHealth struct {
	Check []*healthCheck
}

// NameStatus returns a reporting of memberHealth results by name in three buckets healthy, unhealthy and unstarted.
func (h *memberHealth) Status() ([]string, []string, []string) {
	healthy := []string{}
	unhealthy := []string{}
	unstarted := []string{}
	for _, etcd := range h.Check {
		switch {
		case len(etcd.Member.ClientURLs) == 0:
			unstarted = append(unstarted, GetMemberNameOrHost(etcd.Member))
		case etcd.Health:
			healthy = append(healthy, etcd.Member.Name)
		default:
			unhealthy = append(unhealthy, etcd.Member.Name)
		}
	}

	return healthy, unhealthy, unstarted
}

// MemberStatus returns etcd members healthy or unhealthy
func (h *memberHealth) MemberStatus() ([]*etcdserverpb.Member, []*etcdserverpb.Member) {
	healthy := []*etcdserverpb.Member{}
	unhealthy := []*etcdserverpb.Member{}
	for _, etcd := range h.Check {
		switch {
		case etcd.Health:
			healthy = append(healthy, etcd.Member)
		default:
			unhealthy = append(unhealthy, etcd.Member)
		}
	}
	return healthy, unhealthy
}
