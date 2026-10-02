package businesswords

import "testing"

func TestNormalizeBusinessWordAbbreviation(t *testing.T) {
	tests := []struct {
		name    string
		word    string
		want    string
		wantErr bool
	}{
		{"incorporated abbreviates to inc", "INCORPORATED", "INC", false},
		{"lowercase input is normalized", "incorporated", "INC", false},
		{"company abbreviates to co", "COMPANY", "CO", false},
		{"unrecognized word errors", "FROBNICATE", "", true},
		{"alt-only spelling is not recognized by primary lookup", "INCORP", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeBusinessWordAbbreviation(tt.word)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeBusinessWordAbbreviation(%q) = %q, want error", tt.word, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeBusinessWordAbbreviation(%q) returned error: %v", tt.word, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeBusinessWordAbbreviation(%q) = %q, want %q", tt.word, got, tt.want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	info, err := Info("INCORPORATED")
	if err != nil {
		t.Fatalf("Info(INCORPORATED) returned error: %v", err)
	}
	if info.Primary != "INCORPORATED" || info.Short != "INC" {
		t.Errorf("Info(INCORPORATED) = %+v, want Primary=INCORPORATED Short=INC", info)
	}

	if _, err := Info("NOT A REAL BUSINESS WORD"); err == nil {
		t.Errorf("Info(NOT A REAL BUSINESS WORD) = nil error, want error")
	}
}
