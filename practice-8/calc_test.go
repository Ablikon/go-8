package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	got := Add(2, 3)
	want := 5
	if got != want {
		t.Errorf("Add(2, 3) = %d; want %d", got, want)
	}
}

func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"both positive", 2, 3, 5},
		{"positive + zero", 5, 0, 5},
		{"negative + positive", -1, 4, 3},
		{"both negative", -2, -3, -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			assert.Equal(t, tt.want, got, "they should be equal")
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name      string
		a, b      int
		want      int
		wantError bool
		errorMsg  string
	}{
		{"both positive", 10, 2, 5, false, ""},
		{"positive / negative", 10, -2, -5, false, ""},
		{"division by zero", 10, 0, 0, true, "division by zero"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)
			if tt.wantError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"Both positive numbers", 10, 5, 5},
		{"Positive minus zero", 10, 0, 10},
		{"Negative minus positive", -5, 5, -10},
		{"Both negative", -5, -10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
