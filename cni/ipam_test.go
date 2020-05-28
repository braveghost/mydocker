package cni

import (
	"net"
	"testing"
)

func TestIpam_Allocator(t *testing.T) {

	type args struct {
		subnet *net.IPNet
	}
	_, ipnet, _ := net.ParseCIDR("192.168.0.0/24")
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test-allocator",
			args: args{ipnet},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewIpam()
			got, err := i.Allocator(tt.args.subnet)
			t.Log(got)
			t.Log(err)
		})
	}
}

func TestIpam_Release(t *testing.T) {
	type args struct {
		subnet *net.IPNet
		ipaddr *net.IP
	}
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-replease",
			args: args{
				subnet: ipnet,
				ipaddr: &ip,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewIpam()
			if err := i.Release(tt.args.subnet, tt.args.ipaddr); (err != nil) != tt.wantErr {
				t.Errorf("Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
