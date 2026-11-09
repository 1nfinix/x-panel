package systemSetting

import (
	"encoding/json"
	"testing"
)

func TestLoginCaptchaDefaultAndExplicitDisable(t *testing.T) {
	for _, tt := range []struct {
		stored string
		want   bool
	}{
		{`{}`, true},
		{`{"loginCaptcha":false}`, false},
		{`{"loginCaptcha":true}`, true},
	} {
		setting := DefaultApplicationSetting()
		if err := json.Unmarshal([]byte(tt.stored), &setting); err != nil {
			t.Fatal(err)
		}
		if setting.LoginCaptcha != tt.want {
			t.Errorf("setting %s: loginCaptcha=%v, want %v", tt.stored, setting.LoginCaptcha, tt.want)
		}
	}
}
