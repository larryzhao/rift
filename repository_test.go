package rye

import "testing"

func TestDiffServerCount(t *testing.T) {
	makeServer := func(name string, host string, port int) *Server {
		return &Server{
			Protocol: ProtoclVLess,
			Name:     name,
			Host:     host,
			Port:     port,
			User:     "user",
		}
	}

	tests := []struct {
		name       string
		oldServers []*Server
		newServers []*Server
		expected   int
	}{
		{
			name:       "no change",
			oldServers: []*Server{makeServer("a", "1.1.1.1", 443)},
			newServers: []*Server{makeServer("a", "1.1.1.1", 443)},
			expected:   0,
		},
		{
			name:       "add one server",
			oldServers: []*Server{},
			newServers: []*Server{makeServer("a", "1.1.1.1", 443)},
			expected:   1,
		},
		{
			name:       "remove one server",
			oldServers: []*Server{makeServer("a", "1.1.1.1", 443)},
			newServers: []*Server{},
			expected:   1,
		},
		{
			name:       "replace one server with another",
			oldServers: []*Server{makeServer("a", "1.1.1.1", 443)},
			newServers: []*Server{makeServer("b", "2.2.2.2", 8443)},
			expected:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := diffServerCount(tt.oldServers, tt.newServers)
			if actual != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}
