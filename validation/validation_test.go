package validation

import "testing"

func TestSignupInput_Valid(t *testing.T) {
	in := SignupInput{Username: "reza_1385", Password: "secret123"}
	if err := in.Validate(); err != nil {
		t.Fatalf("expected valid input, got %v", err)
	}
}

func TestSignupInput_ShortUsername(t *testing.T) {
	in := SignupInput{Username: "ab", Password: "secret123"}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error for short username")
	}
}

func TestSignupInput_BadChars(t *testing.T) {
	in := SignupInput{Username: "bad user!", Password: "secret123"}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error for invalid characters")
	}
}

func TestSignupInput_ShortPassword(t *testing.T) {
	in := SignupInput{Username: "reza", Password: "12345"}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestSignupInput_TrimsSpaces(t *testing.T) {
	in := SignupInput{Username: "  reza  ", Password: "  secret123  "}
	if err := in.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.Username != "reza" || in.Password != "secret123" {
		t.Fatal("expected fields to be trimmed")
	}
}

func TestTaskInput_Valid(t *testing.T) {
	in := TaskInput{Title: "Buy milk", Description: "from the store"}
	if err := in.Validate(); err != nil {
		t.Fatalf("expected valid input, got %v", err)
	}
}

func TestTaskInput_EmptyTitle(t *testing.T) {
	in := TaskInput{Title: "   "}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestTaskInput_TooLongTitle(t *testing.T) {
	long := make([]byte, 151)
	for i := range long {
		long[i] = 'a'
	}
	in := TaskInput{Title: string(long)}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error for too-long title")
	}
}

func TestTaskInput_TooLongDescription(t *testing.T) {
	long := make([]byte, 501)
	for i := range long {
		long[i] = 'a'
	}
	in := TaskInput{Title: "ok title", Description: string(long)}
	if err := in.Validate(); err == nil {
		t.Fatal("expected error for too-long description")
	}
}
