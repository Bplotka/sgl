package sgl

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleLogger_FatalLevel(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(b)
	s.SetLevel(FatalLevel)
	s.Debug("test_debug")
	s.Info("test_info")
	s.Error("test_error")
	assert.Equal(t, 0, b.Len())
}

func TestSimpleLogger_ErrorLevel(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(b)
	s.SetLevel(ErrorLevel)
	s.Debug("test_debug")
	s.Info("test_info")
	assert.Equal(t, 0, b.Len())

	now := time.Now()
	s.timeNow = func() time.Time {
		return now
	}

	s.Error("test_error")
	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=error msg=\"test_error\" file=log_test.go:", now.String()))
}

func TestSimpleLogger_InfoLevel(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(b)
	s.SetLevel(InfoLevel)
	s.Debug("test_debug")
	assert.Equal(t, 0, b.Len())

	now := time.Now()
	s.timeNow = func() time.Time {
		return now
	}

	s.Info("test_info")
	s.Error("test_error")
	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=info msg=\"test_info\" file=log_test.go:", now.String()))
	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=error msg=\"test_error\" file=log_test.go:", now.String()))
}

func TestSimpleLogger_DebugLevel(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(b)
	s.SetLevel(DebugLevel)

	now := time.Now()
	s.timeNow = func() time.Time {
		return now
	}

	s.Debug("test_debug")
	s.Info("test_info")
	s.Error("test_error")

	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=debug msg=\"test_debug\" file=log_test.go:", now.String()))
	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=info msg=\"test_info\" file=log_test.go:", now.String()))
	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=error msg=\"test_error\" file=log_test.go:", now.String()))

}

func TestSimpleLogger_WithField(t *testing.T) {
	b := &bytes.Buffer{}
	s := New(b)
	s.SetLevel(InfoLevel)

	now := time.Now()
	s.timeNow = func() time.Time {
		return now
	}

	muted := s.WithField("testKey", "testValue")
	muted.Info("test_info")
	assert.Contains(t, b.String(), fmt.Sprintf(
		"time=\"%s\" lvl=info msg=\"test_info\" testKey=\"testValue\" file=log_test.go:", now.String()))
}
