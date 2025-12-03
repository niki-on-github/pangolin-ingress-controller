package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizzz/pangolin-ingress-controller/internal/util"
)

func TestSplitHost(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		wantSubdomain string
		wantDomain    string
		wantErr       bool
	}{
		{
			name:          "standard subdomain",
			host:          "app.example.com",
			wantSubdomain: "app",
			wantDomain:    "example.com",
			wantErr:       false,
		},
		{
			name:          "multi-level subdomain",
			host:          "api.staging.example.com",
			wantSubdomain: "api.staging",
			wantDomain:    "example.com",
			wantErr:       false,
		},
		{
			name:          "apex domain",
			host:          "example.com",
			wantSubdomain: "",
			wantDomain:    "example.com",
			wantErr:       false,
		},
		{
			name:          "co.uk domain",
			host:          "www.example.co.uk",
			wantSubdomain: "www",
			wantDomain:    "example.co.uk",
			wantErr:       false,
		},
		{
			name:          "deep subdomain with co.uk",
			host:          "api.staging.example.co.uk",
			wantSubdomain: "api.staging",
			wantDomain:    "example.co.uk",
			wantErr:       false,
		},
		{
			name:    "wildcard host",
			host:    "*.example.com",
			wantErr: true,
		},
		{
			name:    "IP address",
			host:    "192.168.1.1",
			wantErr: true,
		},
		{
			name:    "empty host",
			host:    "",
			wantErr: true,
		},
		{
			name:    "localhost",
			host:    "localhost",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subdomain, domain, err := util.SplitHost(tt.host)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSubdomain, subdomain)
			assert.Equal(t, tt.wantDomain, domain)
		})
	}
}
