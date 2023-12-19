// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for ACMEProxy.
package acmeproxy

import (
	"context"
	"reflect"
	"testing"

	"github.com/libdns/libdns"
)

func TestProvider_AppendRecords(t *testing.T) {
	type fields struct {
		Credentials Credentials
		Endpoint    string
		HTTPClient  HTTPClient
	}
	type args struct {
		ctx     context.Context
		zone    string
		records []libdns.Record
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []libdns.Record
		wantErr bool
	}{
		{
			name:   "Test AppendRecords",
			fields: fields{},
			args: args{
				ctx:     context.Background(),
				zone:    "example.com",
				records: []libdns.Record{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				Credentials: tt.fields.Credentials,
				Endpoint:    tt.fields.Endpoint,
				HTTPClient:  tt.fields.HTTPClient,
			}
			got, err := p.AppendRecords(tt.args.ctx, tt.args.zone, tt.args.records)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.AppendRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.AppendRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetRecords(t *testing.T) {
	type fields struct {
		Credentials Credentials
		Endpoint    string
		HTTPClient  HTTPClient
	}
	type args struct {
		ctx  context.Context
		zone string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []libdns.Record
		wantErr bool
	}{
		{
			name:   "Test AppendRecords",
			fields: fields{},
			args: args{
				ctx:  context.Background(),
				zone: "example.com",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				Credentials: tt.fields.Credentials,
				Endpoint:    tt.fields.Endpoint,
				HTTPClient:  tt.fields.HTTPClient,
			}
			got, err := p.GetRecords(tt.args.ctx, tt.args.zone)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.GetRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.GetRecords() = %v, want %v", got, tt.want)
			}
		})
	}
}
