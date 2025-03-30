package paginate

var (
	DefaultMinPageSize uint64 = 10
	DefaultMaxPageSize uint64 = 100
)

// FilterParameter defines a key for filtering paginated results.
// Custom filter parameters can be defined as needed by the application in service layer.
// Example values include:
//   - "status": To filter results by status (e.g., "active", "inactive").
//   - "user_id": To filter results by a specific user ID.
//   - "created_at": To filter results based on creation date.
type FilterParameter string

// Paginated struct represents a paginated response that will be passed to the repository layer
// to retrieve data based on pagination settings, filters, sorting, etc.
type Paginated struct {
	Page        uint64                     `json:"page"`
	PerPage     uint64                     `json:"per_page"`
	Total       uint64                     `json:"total"`
	Filters     map[FilterParameter]Filter `json:"filters"`
	SortColumn  string                     `json:"sort_column"`
	Decscending bool                       `json:"descending"`
}

func (p *Paginated) Validate() error {
	if p.Page == 0 {
		p.Page = 1
	}

	if p.PerPage < DefaultMinPageSize {
		p.PerPage = DefaultMinPageSize
	}

	if p.PerPage > DefaultMaxPageSize {
		p.PerPage = DefaultMaxPageSize
	}

	return nil
}

/*
// Example of how the Paginated struct can be used in a database implementation:
// type PaginationSupportDB struct {
//     GetPaginated(ctx context.Context, p Paginated) ([]Result, error)
// }
// This interface defines how pagination and filtering data can be fetched from the database.
*/

type Filter struct {
	Operator FilterOperator
	Values   []interface{}
}

type FilterOperator int

const (
	FilterOperatorEqual FilterOperator = iota
	FilterOpratorNotEqual
	FilterOperatorGreater
	FilterOperatorGreaterEqual
	FilterOperatorLess
	FilterOperatorLessEqual
	FilterOperatorIN
	FilterOperatorNotIn
	FilterOperatorBetween
)

// the base structure for pagination requests from the client
type PaginateRequestBase struct {
	CurrentPage uint64                     `json:"current_page"`
	PageSize    uint64                     `json:"page_size"`
	Filters     map[FilterParameter]Filter `json:"filters"`
	SortColumn  string                     `json:"sort_column"`
	Decscending bool                       `json:"descending"`
}

// the base structure for paginated responses sent to the client
type PaginatedResponseBase struct {
	CurrentPage  uint64 `json:"current_page"`
	PageSize     uint64 `json:"page_size"`
	TotalNumbers uint64 `json:"total_numbers"`
	TotalPages   uint64 `json:"total_pages"`
}

func (r *PaginateRequestBase) Validate() error {
	if r.CurrentPage < 1 {
		r.CurrentPage = 1
	}
	if r.PageSize < DefaultMinPageSize {
		r.PageSize = DefaultMinPageSize
	}

	if r.PageSize > DefaultMaxPageSize {
		r.PageSize = DefaultMaxPageSize
	}

	// TODO: Add filters validation to ensure that the filters are well-formed

	return nil
}
