package entity

// MeasurementUnit represents a unit of measure for items (e.g. Unidad, Gramos, Litros).
type MeasurementUnit struct {
	ID   uint16 `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
