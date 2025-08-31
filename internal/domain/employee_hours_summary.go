package domain

// Period structure reused from ExtraHourSummary

type EmployeeInfo struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
}

type EmployeeHourBlock struct {
	Absence        float64 `json:"absence"` // en horas
	Novelty        float64 `json:"novelty"` // en horas
	DiurnalExtra   float64 `json:"diurnal_extra"`
	NocturnalExtra float64 `json:"nocturnal_extra"`
}

type EmployeeHourSummary struct {
	Employee EmployeeInfo      `json:"employee"`
	Period   Period            `json:"period"`
	Ordinary EmployeeHourBlock `json:"ordinary"`
	Sunday   EmployeeHourBlock `json:"sunday"`
	Holiday  EmployeeHourBlock `json:"holiday"`
}
