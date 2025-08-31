package domain

type Period struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

type ExtraHourBlock struct {
	DiurnalExtra   float64 `json:"diurnal_extra"`
	NocturnalExtra float64 `json:"nocturnal_extra"`
}

type ExtraHourSummary struct {
	Period   Period         `json:"period"`
	Ordinary ExtraHourBlock `json:"ordinary"`
	Sunday   ExtraHourBlock `json:"sunday"`
	Holiday  ExtraHourBlock `json:"holiday"`
}
