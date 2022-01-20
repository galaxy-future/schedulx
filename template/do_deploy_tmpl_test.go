package template

import "testing"

func TestParseDeployCmd(t *testing.T) {
	type args struct {
		params DeployParams
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"singe env",
			args{DeployParams{
				EnvVariables: []string{"xx=xx"},
				RawCmd:       "echo $path",
			}},
			`
#!/bin/bash
export xx=xx
echo $path
`,
			false,
		},
		{
			"multi env",
			args{DeployParams{
				EnvVariables: []string{"xx=xx", "aa=bb"},
				RawCmd:       "echo $path",
			}},
			`
#!/bin/bash
export xx=xx
export aa=bb
echo $path
`,
			false,
		},
		{
			"empty env",
			args{DeployParams{
				EnvVariables: []string{},
				RawCmd:       "echo $path",
			}},
			`
#!/bin/bash

echo $path
`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDeployCmd(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDeployCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDeployCmd() got = %v, want %v", got, tt.want)
			}
		})
	}
}
