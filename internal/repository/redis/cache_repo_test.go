package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractUserIDFromKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "Valid task by ID key",
			key:      "user:c0646df4-f341-47a6-a295-f141bcaac615:task:98ce3320",
			expected: "c0646df4-f341-47a6-a295-f141bcaac615",
		},
		{
			name:     "Valid task list key",
			key:      "user:c0646df4-f341-47a6-a295-f141bcaac615:tasks:p1:l10:s:q",
			expected: "c0646df4-f341-47a6-a295-f141bcaac615",
		},
		{
			name:     "Valid pattern key",
			key:      "user:c0646df4-f341-47a6-a295-f141bcaac615:*",
			expected: "c0646df4-f341-47a6-a295-f141bcaac615",
		},
		{
			name:     "Invalid prefix",
			key:      "other:c0646df4-f341-47a6-a295-f141bcaac615:task",
			expected: "",
		},
		{
			name:     "Empty userID segment",
			key:      "user::task",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUserIDFromKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
