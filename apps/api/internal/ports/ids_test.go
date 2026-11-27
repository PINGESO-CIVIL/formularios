package ports_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/ports"
)

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "acepta UUID v7 válido",
			id:   "01933b4e-7c90-7000-8000-000000000001",
			want: true,
		},
		{
			name: "acepta UUID v4 en mayúsculas",
			id:   "550E8400-E29B-41D4-A716-446655440000",
			want: true,
		},
		{
			name: "rechaza nil UUID",
			id:   "00000000-0000-0000-0000-000000000000",
			want: false,
		},
		{
			name: "rechaza ObjectID de MongoDB (24 chars hex)",
			id:   "507f1f77bcf86cd799439011",
			want: false,
		},
		{
			name: "rechaza string con caracteres no hexadecimales",
			id:   "01933b4e-7c90-7000-8000-00000000000g",
			want: false,
		},
		{
			name: "rechaza separadores en posición incorrecta",
			id:   "01933b4e7-c90-7000-8000-000000000001",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ports.IsValidUUID(tc.id))
		})
	}
}
