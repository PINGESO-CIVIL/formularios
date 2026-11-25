// Paquete store implementa el almacenamiento que consume internal/service.
// En el Hito 1 los datos viven en memoria; en el Hito 2 se agrega PostgreSQL con las
// queries generadas por sqlc en internal/repository. No contiene reglas de negocio.
package store
