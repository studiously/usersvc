package svcerror

import (
	"testing"
	"github.com/go-errors/errors"
)

func TestNew(t *testing.T) {
	t.Parallel()

	se := New(404, "not found")

	_, ok := se.(error)
	if !ok {
		t.Fatalf("could not cast result of svcerror.New (%v) to error", se)
	}
	if se.Status() != 404 {
		t.Fatalf("bad status code: expected %d, got %d", 404, se.Status())
	}
	if se.Error() != "not found" {
		t.Fatalf("bad error text: expected '%s', got '%s'", "not found", se.Error())
	}
}

func TestWrap(t *testing.T) {
	t.Parallel()

	err := errors.New("not found")
	se := Wrap(404, err)

	_, ok := se.(error)
	if !ok {
		t.Fatalf("could not cast the result of svcerror.Wrap (%v) to error", se)
	}

	if se.Status() != 404 {
		t.Fatalf("bad status code: expected %d, got %d", 404, se.Status())
	}
	if se.Error() != "not found" {
		t.Fatalf("bad error text: expected '%s', got '%s'", "not found", se.Error())
	}
}

func TestServiceError_Status(t *testing.T) {
	t.Parallel()

	se := New(404, "not found")
	if se.Status() != 404 {
		t.Fatalf("bad status code: expected %d, got %d", 404, se.Status())
	}
}

func TestServiceError_Error(t *testing.T) {
	t.Parallel()

	se := New(404, "not found")
	if se.Error() != "not found" {
		t.Fatalf("bad error text: expected '%s', got '%s'", "not found", se.Error())
	}
}
