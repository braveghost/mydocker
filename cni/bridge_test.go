package cni

import "testing"

func Test_createBridgeInterface(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "testbridge1",
			args:    args{"bridgxxxxxxxexx"},
			wantErr: false,
		},
		{
			name:    "testbridge2",
			args:    args{"bridgxxxxxxxexxxx"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createBridgeInterface(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("createBridgeInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
