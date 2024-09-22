package gomts

import "context"

// DepartmentClient interfaces with Department related MyTimeStation API
// methods.
type DepartmentClient interface {
	// Create a new department.
	Create(ctx context.Context, req *DepartmentCreateRequest) (*Department, error)

	List(ctx context.Context) ([]Department, error)

	Delete(ctx context.Context, id string) (*Department, error)
}

// Department represents a department at a customer company in the
// MyTimeStation system.
type Department struct {
	// ID is the unique identifier for the department within the MyTimeStation
	// system.
	ID string `json:"department_id"`

	// Name is the name of the department.
	Name string `json:"name"`
}

type DepartmentCreateRequest struct {
	// Name is the name of the department.
	// This field is required.
	Name string `url:"name"`
}

// form implements formRequest.
func (DepartmentCreateRequest) form() {}

// DepartmentListResponse is the response used for the List API method.
type DepartmentListResponse struct {
	// Departments is the list of departments
	Departments []Department `json:"departments"`
}

// DepartmentResponse is the response used for the Create, Get, Update and
// Delete API methods.
type DepartmentResponse struct {
	// Department is the department of subject.
	Department Department `json:"department"`
}

// depertmentClient implements DepartmentClient.
type departmentClient struct {
	*client
}

func (c *departmentClient) Create(ctx context.Context, req *DepartmentCreateRequest) (*Department, error) {
	resp, err := httpPost[DepartmentResponse](ctx, c.client, "/departments", req)
	if err != nil {
		return nil, err
	}

	return &resp.Department, nil
}

func (c *departmentClient) List(ctx context.Context) ([]Department, error) {
	resp, err := httpGet[DepartmentListResponse](ctx, c.client, "/departments")
	if err != nil {
		return nil, err
	}

	return resp.Departments, nil
}

func (c *departmentClient) Delete(ctx context.Context, id string) (*Department, error) {
	resp, err := httpDelete[DepartmentResponse](ctx, c.client, "/departments/"+id)
	if err != nil {
		return nil, err
	}

	return &resp.Department, nil
}

// compile-time assertion that departmentClient implementation fulfils
// DepartmentClient interface.
var _ DepartmentClient = (*departmentClient)(nil)
