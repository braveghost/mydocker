package cni

import "testing"

func TestCreateNetwork(t *testing.T) {
	type args struct {
		driver string
		subnet string
		name   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "TestCreateNetwork",
			args:    args{
				driver: "bridge",
				subnet: "172.16.0.0/24",
				name:   "testnetwork",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateNetwork(tt.args.driver, tt.args.subnet, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("CreateNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListNetwork(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "list-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ListNetwork()
		})
	}
}