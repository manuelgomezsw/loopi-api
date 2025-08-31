package dto

type ShiftProjectionRequest struct {
	ShiftID int `json:"shift_id"`
	Year    int `json:"year"`
	Month   int `json:"month"`
}
