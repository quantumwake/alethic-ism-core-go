package nats

import (
	"testing"
)

func TestFindRouteBySelectorWildcard(t *testing.T) {
	// Create a test config with wildcard routes
	config := &Config{
		MessageConfig: MessageConfig{
			Routes: []NatConfig{
				{
					Selector: "processor/usage",
					Subject:  "processor.usage",
					URL:      "nats://127.0.0.1:4222",
				},
				{
					Selector: "language/models/openai/*",
					Subject:  "processor.models.openai",
					URL:      "nats://127.0.0.1:4222",
				},
				{
					Selector: "language/models/llama/*",
					Subject:  "processor.models.llama",
					URL:      "nats://127.0.0.1:4222",
				},
				{
					Selector: "language/models/anthropic/*",
					Subject:  "processor.models.anthropic",
					URL:      "nats://127.0.0.1:4222",
				},
			},
		},
	}

	// Build route maps
	config.BuildRouteMaps()

	tests := []struct {
		name        string
		selector    string
		wantSubject string
		wantErr     bool
	}{
		{
			name:        "exact match",
			selector:    "processor/usage",
			wantSubject: "processor.usage",
			wantErr:     false,
		},
		{
			name:        "wildcard match - openai gpt-4",
			selector:    "language/models/openai/gpt-4",
			wantSubject: "processor.models.openai",
			wantErr:     false,
		},
		{
			name:        "wildcard match - openai gpt-3.5-turbo",
			selector:    "language/models/openai/gpt-3.5-turbo",
			wantSubject: "processor.models.openai",
			wantErr:     false,
		},
		{
			name:        "wildcard match - llama",
			selector:    "language/models/llama/llama-3-70b",
			wantSubject: "processor.models.llama",
			wantErr:     false,
		},
		{
			name:        "wildcard match - anthropic",
			selector:    "language/models/anthropic/claude-3-opus",
			wantSubject: "processor.models.anthropic",
			wantErr:     false,
		},
		{
			name:     "no match",
			selector: "language/models/cohere/command",
			wantErr:  true,
		},
		{
			name:     "partial match should fail",
			selector: "language/models",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route, err := config.FindRouteBySelectorWildcard(tt.selector)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindRouteBySelectorWildcard() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("FindRouteBySelectorWildcard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if route.Subject != tt.wantSubject {
				t.Errorf("FindRouteBySelectorWildcard() subject = %v, want %v", route.Subject, tt.wantSubject)
			}
		})
	}
}

func TestFindRouteBySelectorWildcardMultipleMatches(t *testing.T) {
	// Create a config with overlapping wildcard routes
	config := &Config{
		MessageConfig: MessageConfig{
			Routes: []NatConfig{
				{
					Selector: "language/models/*",
					Subject:  "processor.models.all",
					URL:      "nats://127.0.0.1:4222",
				},
				{
					Selector: "language/models/openai/*",
					Subject:  "processor.models.openai",
					URL:      "nats://127.0.0.1:4222",
				},
			},
		},
	}

	config.BuildRouteMaps()

	// This should return error due to multiple matches
	_, err := config.FindRouteBySelectorWildcard("language/models/openai/gpt-4")
	if err == nil {
		t.Error("FindRouteBySelectorWildcard() expected error for multiple matches, got nil")
	}
}

func TestFindRouteBySelector(t *testing.T) {
	config := &Config{
		MessageConfig: MessageConfig{
			Routes: []NatConfig{
				{
					Selector: "processor/usage",
					Subject:  "processor.usage",
					URL:      "nats://127.0.0.1:4222",
				},
			},
		},
	}

	config.BuildRouteMaps()

	route, err := config.FindRouteBySelector("processor/usage")
	if err != nil {
		t.Errorf("FindRouteBySelector() error = %v", err)
	}

	if route.Subject != "processor.usage" {
		t.Errorf("FindRouteBySelector() subject = %v, want processor.usage", route.Subject)
	}

	// Test not found
	_, err = config.FindRouteBySelector("nonexistent")
	if err == nil {
		t.Error("FindRouteBySelector() expected error for nonexistent selector, got nil")
	}
}
