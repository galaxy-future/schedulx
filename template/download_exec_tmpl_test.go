package template

import "testing"

func TestParseDownloadExecCmd(t *testing.T) {
	type args struct {
		params DownloadParams
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"1",
			args{params: DownloadParams{
				DownloadFileUrl: "xx",
				DeployFilePath:  "/aa/bb/cc",
				DeployFileName:  "dd",
			}},
			`
#!/bin/bash
mkdir -p /aa/bb/cc
wget xx -O /aa/bb/cc/dd
`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDownloadExecCmd(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDownloadExecCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDownloadExecCmd() got = %v, want %v", got, tt.want)
			}
		})
	}
}
