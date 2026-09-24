package example

import "testing"

func TestExample(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{name: "empty", give: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Example(tt.give)
			if got != tt.want {
				t.Errorf("Example(%q) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
