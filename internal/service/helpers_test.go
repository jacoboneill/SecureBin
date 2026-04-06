package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/db"
)

var ErrDBMock = errors.New("mock db error")

type QuerierMock struct {
	db.Querier
	Calls int
}

type ResultMock struct {
	rowsAffected int64
}

func (r ResultMock) LastInsertId() (int64, error) { return 0, nil }
func (r ResultMock) RowsAffected() (int64, error) { return r.rowsAffected, nil }

func Modify(in string) string {
	return fmt.Sprintf("%s1", in)
}

func AssertErrorsEqual(t testing.TB, expectedErr, capturedErr error) {
	t.Helper()
	if !errors.Is(capturedErr, expectedErr) {
		if expectedErr == nil {
			t.Errorf("expected no errors, got %v", capturedErr)
		} else {
			t.Errorf("expected error: %v, got %v", expectedErr, capturedErr)
		}
	}
}

func AssertCallCountsEqual(t testing.TB, expectedCallCount, capturedCallCount int) {
	t.Helper()
	if expectedCallCount != capturedCallCount {
		var suffix string
		if expectedCallCount > 1 {
			suffix = "s"
		}
		t.Errorf("expected %d %s, got %d", expectedCallCount, suffix, capturedCallCount)
	}
}
