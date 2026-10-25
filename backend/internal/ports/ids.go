package ports

// IDGenerator permite inyectar la generación de IDs (UUIDv7) para pruebas deterministas.
type IDGenerator interface {
	NewID() string
}
