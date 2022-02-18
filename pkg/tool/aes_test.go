package tool

import (
	"reflect"
	"testing"
)

func TestAesEncrypt(t *testing.T) {
	type args struct {
		v   string
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "T1",
			args: args{
				v:   "zadig",
				key: []byte("admin"),
			},
			want:    []byte("123456"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AesEncrypt([]byte(tt.args.v), tt.args.key)
			t.Logf("got:%v", string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("AesDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AesDecrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAesDecrypt(t *testing.T) {
	type args struct {
		cryptedBase64 string
		key           []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "T1",
			args: args{
				cryptedBase64: "IIp5R+AzJUgHjAjRX+CSYw==",
				key:           []byte("admin"),
			},
			want:    []byte("zadig"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AesDecrypt(tt.args.cryptedBase64, tt.args.key)
			t.Logf("got:%v", string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("AesDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AesDecrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
