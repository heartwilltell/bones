package id

import (
	"testing"
)

func TestULID(t *testing.T) {
	t.Run("ULID", func(t *testing.T) {
		got := ULID()
		if len(got) != 26 {
			t.Errorf("the len of ULID() = %v, doesn't equal to 26 characters", got)
		}
	})
}

func TestValidateULID(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		id := ULID()
		if err := ValidateULID(id); err != nil {
			t.Errorf("ULID should be valid but its not: %v", err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		id := "invalid id"
		if err := ValidateULID(id); err == nil || err != ErrInvalidID {
			t.Errorf("ULID should not be valid. Expected ErrInvalidID")
		}
	})
}

func TestDigiCode(t *testing.T) {
	t.Run("DigiCode", func(t *testing.T) {
		got := DigiCode()
		t.Log(got)
	})
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		e    Error
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestULID1(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ULID(); got != tt.want {
				t.Errorf("ULID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateULID1(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateULID(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("ValidateULID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateXID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateXID(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("ValidateXID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestXID(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := XID(); got != tt.want {
				t.Errorf("XID() = %v, want %v", got, tt.want)
			}
		})
	}
}
