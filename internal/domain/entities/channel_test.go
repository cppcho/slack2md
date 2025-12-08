package entities

import "testing"

func TestNewChannel(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		chName  string
		wantErr bool
	}{
		{
			name:    "valid channel",
			id:      "C123456",
			chName:  "general",
			wantErr: false,
		},
		{
			name:    "valid channel with special chars",
			id:      "C7890ABC",
			chName:  "team-announcements",
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			chName:  "general",
			wantErr: true,
		},
		{
			name:    "empty name",
			id:      "C123456",
			chName:  "",
			wantErr: true,
		},
		{
			name:    "both empty",
			id:      "",
			chName:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := NewChannel(tt.id, tt.chName)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewChannel() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if ch == nil {
					t.Error("Expected channel to be created, got nil")
					return
				}
				if ch.ID != tt.id {
					t.Errorf("Expected ID = %s, got = %s", tt.id, ch.ID)
				}
				if ch.Name != tt.chName {
					t.Errorf("Expected Name = %s, got = %s", tt.chName, ch.Name)
				}
			} else {
				if ch != nil {
					t.Error("Expected nil channel on error, got non-nil")
				}
			}
		})
	}
}

func TestNewChannel_ErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		chName         string
		expectedErrMsg string
	}{
		{
			name:           "empty id error message",
			id:             "",
			chName:         "general",
			expectedErrMsg: "channel ID cannot be empty",
		},
		{
			name:           "empty name error message",
			id:             "C123",
			chName:         "",
			expectedErrMsg: "channel name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewChannel(tt.id, tt.chName)
			if err == nil {
				t.Errorf("Expected error, got nil")
				return
			}
			if err.Error() != tt.expectedErrMsg {
				t.Errorf("Expected error message = %q, got = %q", tt.expectedErrMsg, err.Error())
			}
		})
	}
}
