// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for ACMEProxy.
package acmeproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/libdns/libdns"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type acmeProxy struct {
	FQDN  string `json:"FQDN"`
	Value string `json:"Value"`
}

// Credentials represents the username and password required for authentication.
// The fields are optional and can be omitted.
type Credentials struct {
	Username string `json:"username,omitempty"` // Username represents the user's name for authentication.
	Password string `json:"password,omitempty"` // Password represents the user's password for authentication.
}

// Provider facilitates DNS record manipulation with ACMEProxy.
type Provider struct {
	// Credentials are the username and password required for authentication.
	// The fields are optional and can be omitted.
	Credentials
	// Endpoint is the URL of the ACMEProxy server.
	Endpoint string `json:"endpoint"`
	// HTTPClient is the client used to communicate with the ACMEProxy server.
	// If nil, a default client will be used.
	HTTPClient HTTPClient
}

func (p *Provider) getClient() HTTPClient {
	if p.HTTPClient == nil {
		return http.DefaultClient
	}
	return p.HTTPClient
}

func (p *Provider) doAction(ctx context.Context, endpoint *url.URL, action string, zone string, record libdns.Record) (libdns.Record, error) {
	// We only support TXT records
	if record.Type != "TXT" {
		return libdns.Record{}, fmt.Errorf("ACMEProxy provider only supports TXT records")
	}

	// Create Request Body
	reqBody := new(bytes.Buffer)
	{
		msg := acmeProxy{
			FQDN:  libdns.AbsoluteName(record.Name, zone),
			Value: record.Value,
		}
		if err := json.NewEncoder(reqBody).Encode(msg); err != nil {
			return libdns.Record{}, fmt.Errorf("ACMEProxy provider could not marshal JSON: %w", err)
		}
	}

	// Create Request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return libdns.Record{}, fmt.Errorf("ACMEProxy provider could not create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Add Basic Auth
	if p.Username != "" || p.Password != "" {
		req.SetBasicAuth(p.Username, p.Password)
	}

	// Send Request
	resp, err := p.getClient().Do(req)
	if err != nil {
		return libdns.Record{}, fmt.Errorf("ACMEProxy provider could not send request: %w", err)
	}
	defer resp.Body.Close()

	// Validate Response
	{
		if resp.StatusCode != http.StatusOK {
			return libdns.Record{}, fmt.Errorf("ACMEProxy provider received status code %d", resp.StatusCode)
		}

		var respMsg acmeProxy
		if err := json.NewDecoder(resp.Body).Decode(&respMsg); err != nil {
			return libdns.Record{}, fmt.Errorf("ACMEProxy provider could not unmarshal JSON: %w", err)
		}

		if respMsg.FQDN != libdns.AbsoluteName(record.Name, zone) {
			return libdns.Record{}, fmt.Errorf("ACMEProxy provider received unexpected FQDN %q", respMsg.FQDN)
		}

		if respMsg.Value != record.Value {
			return libdns.Record{}, fmt.Errorf("ACMEProxy provider received unexpected Value %q", respMsg.Value)
		}
	}

	return libdns.Record{
		ID:    record.ID,
		Type:  "TXT",
		Name:  record.Name,
		Value: record.Value,
		TTL:   record.TTL,
	}, nil
}

func (p *Provider) doActions(ctx context.Context, action string, zone string, records []libdns.Record) ([]libdns.Record, error) {
	// Validate Endpoint
	uri, err := url.ParseRequestURI(p.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("ACMEProxy provider invalid endpoint [%s]: %w", p.Endpoint, err)
	}
	endpoint := uri.JoinPath(action)

	// Loop through records
	// This is not atomic, but we want to try to complete as many records as possible
	// So when an error occurs, we return the completed records along with the error
	completedRecords := []libdns.Record{}
	for _, record := range records {
		completedRecord, err := p.doAction(ctx, endpoint, action, zone, record)
		if err != nil {
			return completedRecords, fmt.Errorf("ACMEProxy provider could not %s record %q: %w", action, record.Name, err)
		}
		// Add Record to completed set
		completedRecords = append(completedRecords, completedRecord)
	}
	return completedRecords, nil
}

// GetRecords lists all the records in the zone.
// This is not supported by the ACMEProxy provider.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	return nil, fmt.Errorf("ACMEProxy provider does not support listing records")
}

// AppendRecords adds records to the zone. It returns the records that were added.
// It does the same as SetRecords.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.SetRecords(ctx, zone, records)
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.doActions(ctx, "present", zone, records)
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.doActions(ctx, "cleanup", zone, records)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
