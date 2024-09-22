package gomts

import "context"

// EmployeeClient interfaces with Employee related MyTimeStation API methods.
type EmployeeClient interface {
	// Create a new employee.
	Create(ctx context.Context, req *EmployeeCreateRequest) (*Employee, error)

	// Get an employee by id.
	Get(ctx context.Context, id string) (*Employee, error)

	// List all employees.
	List(ctx context.Context) ([]Employee, error)

	// Update an employee by id.
	Update(ctx context.Context, id string, req *EmployeeUpdateRequest) (*Employee, error)

	// Delete an employee by id.
	Delete(ctx context.Context, id string) (*Employee, error)
}

// EmployeeStatus represents the employee's clock-in/out state.
// Valid values *should* only be "in" or "out".
type EmployeeStatus string

const (
	// EmployeeInStatus signals the employee is clocked in.
	EmployeeInStatus EmployeeStatus = "in"

	// EmployeeOutStatus signals the employee is clocked out.
	EmployeeOutStatus EmployeeStatus = "out"
)

// Employee represents an employee working for a customer company in the
// MyTimeStation system.
type Employee struct {
	// ID is the unique identifier for the employee within the MyTimeStation
	// system.
	ID string `json:"employee_id"`

	// Name is the full name of the employee.
	Name string `json:"name"`

	// Title is the job title of the employee (e.g., Payroll Manager).
	Title string `json:"title"`

	// PrimaryDepartment is the main department where the employee works.
	PrimaryDepartment string `json:"primary_department"`

	// PrimaryDepartmentID is the unique identifier for the primary department.
	PrimaryDepartmentID string `json:"primary_department_id"`

	// CurrentDepartment is the department where the employee is currently
	// working (can be different from primary).
	CurrentDepartment string `json:"current_department"`

	// CurrentDepartmentID is the unique identifier for the current department.
	CurrentDepartmentID string `json:"current_department_id"`

	// Status represents the employee's current clock-in status (in or out).
	Status EmployeeStatus `json:"status"`

	// CustomEmployeeID is the company-defined employee ID, which may differ
	// from the system-generated ID.
	CustomEmployeeID string `json:"custom_employee_id"`

	// PIN is the employee's assigned personal identification number.
	PIN string `json:"pin"`

	// CardNumber is the employee's physical card number used for clocking
	// in/out.
	CardNumber string `json:"card_number"`

	// CardQRCode is the QR code associated with the employee's card, used for
	// clocking in/out.
	CardQRCode string `json:"card_qr_code"`

	// CustomFields is a map of additional employee-specific fields, such as
	// phone number or start date.
	CustomFields map[string]string `json:"custom_fields"`
}

// EmployeeListResponse is the response used for the List API method.
type EmployeeListResponse struct {
	// Employees is the list of employees.
	Employees []Employee `json:"employees"`
}

// EmployeeResponse is the response used for the Create, Get, Update and Delete
// API methods.
type EmployeeResponse struct {
	// Employee is the employee of subject.
	Employee Employee `json:"employee"`
}

// EmployeeCreateRequest represents the request body to create a new employee
// in the MyTimeStation system.
type EmployeeCreateRequest struct {
	// Name is the full name of the employee.
	// This field is required.
	Name string `url:"name"`

	// DepartmentID is the ID of the primary department to assign the employee.
	// Either DepartmentID or DepartmentName must be supplied.
	DepartmentID string `url:"department_id,omitempty"`

	// DepartmentName is the name of the department to assign the employee.
	// It can either create a new department or use an existing one.
	// Either DepartmentID or DepartmentName must be supplied.
	DepartmentName string `url:"department_name,omitempty"`

	// CustomEmployeeID is an optional second ID to associate the employee with
	// another system.
	CustomEmployeeID string `url:"custom_employee_id,omitempty"`

	// Title is the job title of the employee (e.g., Payroll Manager).
	Title string `url:"title,omitempty"`

	// HourlyRate is the hourly wage rate of the employee.
	HourlyRate float64 `url:"hourly_rate,omitempty"`

	// PIN is the 4-digit personal identification number for the employee.
	PIN string `url:"pin,omitempty"`

	// CustomFields allows setting one or more custom fields for the employee.
	// The key is the custom field name, and the value is the field value.
	CustomFields map[string]string `url:"custom_fields,omitempty"`
}

func (EmployeeCreateRequest) form() {}

// EmployeeUpdateRequest represents the request body to update an existing
// employee in the MyTimeStation system.
type EmployeeUpdateRequest struct {
	// Name is the full name of the employee.
	Name *string `json:"name"`

	// DepartmentID is the ID of the primary department to assign the employee.
	// Either DepartmentID or DepartmentName must be supplied.
	DepartmentID *string `json:"department_id,omitempty"`

	// DepartmentName is the name of the department to assign the employee.
	// It can either create a new department or use an existing one.
	// Either DepartmentID or DepartmentName must be supplied.
	DepartmentName *string `json:"department_name,omitempty"`

	// CustomEmployeeID is an optional second ID to associate the employee
	// with another system.
	CustomEmployeeID *string `json:"custom_employee_id,omitempty"`

	// Title is the job title of the employee (e.g., Payroll Manager).
	Title *string `json:"title,omitempty"`

	// HourlyRate is the hourly wage rate of the employee.
	HourlyRate *float64 `json:"hourly_rate,omitempty"`

	// PIN is the 4-digit personal identification number for the employee.
	PIN *string `json:"pin,omitempty"`

	// CustomFields allows setting one or more custom fields for the employee.
	// The key is the custom field name, and the value is the field value.
	CustomFields map[string]string `json:"custom_fields,omitempty"`

	// ConvertPrimaryDepartment indicates if the previous primary department
	// should be retained as a secondary department when the primary department
	// is changed. This parameter applies only to the current API request.
	ConvertPrimaryDepartment *bool `json:"convert_primary_department,omitempty"`
}

// employeeService implements EmployeeClient
type employeeClient = client

func (c *employeeClient) Create(ctx context.Context, req *EmployeeCreateRequest) (*Employee, error) {
	resp, err := httpPost[EmployeeResponse](ctx, c, "/employees", req)
	if err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

func (c *employeeClient) Get(ctx context.Context, id string) (*Employee, error) {
	resp, err := httpGet[EmployeeResponse](ctx, c, "/employees/"+id)
	if err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

func (c *employeeClient) Update(ctx context.Context, id string, req *EmployeeUpdateRequest) (*Employee, error) {
	resp, err := httpPut[EmployeeResponse](ctx, c, "/employees/"+id, req)
	if err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

func (c *employeeClient) Delete(ctx context.Context, id string) (*Employee, error) {
	resp, err := httpDelete[EmployeeResponse](ctx, c, "/employees/"+id)
	if err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

func (c *employeeClient) List(ctx context.Context) ([]Employee, error) {
	resp, err := httpGet[EmployeeListResponse](ctx, c, "/employees")
	if err != nil {
		return nil, err
	}

	return resp.Employees, nil
}

// compile-time assertion that employeeClient implementation fulfils
// EmployeeClient interface.
var _ EmployeeClient = (*employeeClient)(nil)
