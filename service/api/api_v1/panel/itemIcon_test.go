package panel

import (
	"testing"

	"x-panel/models"
)

func TestCandidateIcon(t *testing.T) {
	for _, tt := range []struct {
		json string
		want bool
	}{
		{`{"itemType":1,"src":"manual"}`, true},
		{`{"itemType":2,"src":""}`, true},
		{`{"itemType":2,"src":"   "}`, true},
		{`{"itemType":2,"src":"/uploads/icon.png"}`, false},
		{`{"itemType":3,"src":""}`, false},
		{`bad json`, false},
	} {
		_, got := candidateIcon(models.ItemIcon{IconJson: tt.json})
		if got != tt.want {
			t.Errorf("candidateIcon(%q) = %v, want %v", tt.json, got, tt.want)
		}
	}
}
