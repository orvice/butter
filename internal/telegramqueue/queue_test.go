package telegramqueue

import "testing"

func TestValidateDurabilityConfig(t *testing.T) {
	for _, tc := range []struct {
		name        string
		policy      string
		appendonly  string
		save        string
		wantFailure bool
	}{
		{name: "aof", policy: "noeviction", appendonly: "yes"},
		{name: "rdb", policy: "noeviction", save: "3600 1"},
		{name: "evicting", policy: "allkeys-lru", appendonly: "yes", wantFailure: true},
		{name: "volatile", policy: "noeviction", wantFailure: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDurabilityConfig(
				map[string]string{"maxmemory-policy": tc.policy},
				map[string]string{"appendonly": tc.appendonly},
				map[string]string{"save": tc.save},
			)
			if (err != nil) != tc.wantFailure {
				t.Fatalf("err = %v, wantFailure = %v", err, tc.wantFailure)
			}
		})
	}
}
