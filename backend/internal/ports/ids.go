package ports

import "strings"

// IDGenerator permite inyectar la generación de IDs (UUID v4/v7 compatible con PostgreSQL)
// para pruebas deterministas.
type IDGenerator interface {
	NewID() string
}

// IsValidUUID verifica que un identificador cumpla el formato canónico UUID
// de 36 caracteres (8-4-4-4-12 en hexadecimal), alineado con el tipo UUID de PostgreSQL.
func IsValidUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHexChar(c) {
				return false
			}
		}
	}
	return !strings.EqualFold(id, "00000000-0000-0000-0000-000000000000")
}

func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
