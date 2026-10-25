package ports

import "time"

// Clock permite inyectar el tiempo para pruebas deterministas.
type Clock interface {
	Now() time.Time
}
