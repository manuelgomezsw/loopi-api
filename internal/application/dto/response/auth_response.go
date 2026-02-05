package response

// LoginResponse represents the login response.
type LoginResponse struct {
	Token    string           `json:"token"`
	Employee EmployeeResponse `json:"employee"`
}

// EmployeeResponse represents an employee in API responses.
type EmployeeResponse struct {
	ID       uint16 `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}
