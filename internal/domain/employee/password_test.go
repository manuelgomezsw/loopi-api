package employee

import (
	"testing"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

func TestDefaultPasswordForReset(t *testing.T) {
	docNum := "12345"
	birthYear2000 := time.Date(2000, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		employee *entity.Employee
		want     string
	}{
		{
			name:     "nil employee -> fallback",
			employee: nil,
			want:     defaultPasswordFallback,
		},
		{
			name:     "document_number + birth_year",
			employee: &entity.Employee{DocumentNumber: &docNum, BirthDate: &birthYear2000},
			want:     "123452000",
		},
		{
			name:     "document_number only",
			employee: &entity.Employee{DocumentNumber: &docNum},
			want:     "12345",
		},
		{
			name:     "birth_date only (no document) -> fallback",
			employee: &entity.Employee{BirthDate: &birthYear2000},
			want:     defaultPasswordFallback,
		},
		{
			name:     "no document no birth -> fallback",
			employee: &entity.Employee{},
			want:     defaultPasswordFallback,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultPasswordForReset(tt.employee)
			if got != tt.want {
				t.Errorf("DefaultPasswordForReset() = %q, want %q", got, tt.want)
			}
		})
	}
}
