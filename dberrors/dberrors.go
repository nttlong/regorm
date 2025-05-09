package dberrors

import "strings"

type ErrorCode int
type DbAction int

// Định nghĩa các hằng số cho các giá trị enum.
// Define enum values for ErrorCode and DbAction.
const (
	Unknown ErrorCode = iota // Lỗi không xác định, Unknown error
	// Lỗi trùng lặp, Duplicate error
	// Duplicate error
	Duplicate // iota tự động tăng giá trị của mỗi hằng số
	// Lỗi tham chiếu, Reference error, lỗ này thường gây ra do relation ship giữa các bảng.
	// Reference error, usually caused by relation ship between tables.

	Reference
	// Lỗi bắt buộc, Required error, lỗi này thường gây ra khi một thông tin cần thiết bị không được cung cấp.
	// Required error, usually caused by missing required data.
	Require
	// Lỗi kích thước, Invalid length error, lỗi này thường gây ra khi một thông tin có kích thước không hợp lệ.
	// Invalid length error, usually caused by invalid data length.
	InvalidLen
)

// Định nghĩa các hằng số cho các thao tác trên database gây ra lỗi.
const (
	// nguyên nhân do insert
	Insert DbAction = iota
	// nguyên nhân do update
	Update
	// nguyên nhân do delete
	Delete
)

// Hàm String để làm cho enum dễ đọc hơn khi in. make ErrorCode enum more readable
// String() method for ErrorCode enum
func (e ErrorCode) String() string {
	switch e {
	case Duplicate:
		return "Duplicate"
	case Reference:
		return "Reference"
	case Require:
		return "Require"
	case InvalidLen:
		return "InvalidLen"
	default:
		return "Unknown"
	}
}

// Hàm String để làm cho enum dễ đọc hơn khi in. make DbAction enum more readable
func (e DbAction) String() string {
	switch e {
	case Insert:
		return "Insert"
	case Update:
		return "Update"
	case Delete:
		return "Delete"
	default:
		return "Unknown"
	}
}

// Cấu trúc ghi nhận lại toàn bộ lỗi gây ra khi thao tác trên database.
// Structure to catch all errors occurred when operate on database.
type DataActionError struct {
	Err error
	// Thao tác gây ra lỗi, action cause error
	Action string
	// Mã lỗi, error code
	Code ErrorCode
	// các cột liên quan đến lỗi nếu có, reference columns cause error
	RefColumns []string // Các cột liên quan đến lỗi nếu có , Reference columns cause error
	// các bảng liên quan đến lỗi nếu có, reference table name cause error
	RefTableName string // Tên bảng liên quan đến lỗi nếu có, Reference table name cause error

}

// hàm diễn dịch lại lỗi gây ra khi thao tác trên database.
// translate DataActionError to readable message
func (e *DataActionError) Error() string {
	msg := "Error " + e.Action + " " + e.Code.String() + " " + e.Err.Error()
	if len(e.RefColumns) > 0 {
		msg += " RefColumns: " + strings.Join(e.RefColumns, ",")
	}
	if e.RefTableName != "" {
		msg += " RefTableName: " + e.RefTableName
	}
	return msg
}
