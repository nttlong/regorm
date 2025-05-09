package compiler

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SimpleExprTree đại diện cho một nút trong cây biểu thức
type SimpleExprTree struct {
	V  string            // Giá trị biểu thức (tối giản hoặc tên hàm nếu Nt là "func")
	Op string            // Toán tử
	Ns []*SimpleExprTree // Các nút con
	Nt string            // Node type: "func", "param", "const", "field"
}

func ParseExpr(expr string) (*SimpleExprTree, error) {
	return parseToSimpleExprTree(expr)
}
func ReconstructedSimpleExprTree(t *SimpleExprTree) string {
	return reconstructExpressionSimple(t)
}

func (t *SimpleExprTree) String() string {
	return reconstructExpression(t)
}
func IsValidColumnName(name string) bool {
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return re.MatchString(name)
}
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}
	if !IsValidColumnName(s) {
		return s
	}

	// Kiểm tra xem chuỗi có phải toàn chữ hoa (hoặc chữ hoa + số) không
	isAllUpper := true
	for _, r := range s {
		if !unicode.IsUpper(r) && !unicode.IsNumber(r) && unicode.IsLetter(r) {
			isAllUpper = false
			break
		}
	}

	// Nếu toàn chữ hoa, chỉ cần chuyển thành chữ thường
	if isAllUpper {
		return strings.ToLower(s)
	}

	var result strings.Builder
	runes := []rune(s)

	// Vị trí bắt đầu của chuỗi chữ hoa
	upperRunStart := -1

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if unicode.IsUpper(r) {
			if i == 0 {
				// Ký tự đầu tiên là chữ hoa, không thêm _
				result.WriteRune(unicode.ToLower(r))
				upperRunStart = i
			} else {
				// Kiểm tra ranh giới từ
				prevIsLower := unicode.IsLower(runes[i-1])
				nextIsLower := (i+1 < len(runes)) && unicode.IsLower(runes[i+1])

				if prevIsLower || (nextIsLower && upperRunStart != i-1) {
					// Thêm _ nếu trước đó là chữ thường hoặc đây là chữ hoa bắt đầu từ mới
					if result.Len() > 0 && result.String()[result.Len()-1] != '_' {
						result.WriteRune('_')
					}
				}
				result.WriteRune(unicode.ToLower(r))
				upperRunStart = i
			}
		} else if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			// Thay ký tự đặc biệt bằng dấu gạch dưới
			if result.Len() > 0 && result.String()[result.Len()-1] != '_' {
				result.WriteRune('_')
			}
			upperRunStart = -1
		} else {
			// Chữ thường hoặc số
			result.WriteRune(r)
			upperRunStart = -1
		}
	}

	// Loại bỏ dấu gạch dưới ở đầu và cuối, thay thế nhiều dấu gạch dưới liên tiếp bằng một
	snake := strings.Trim(result.String(), "_")
	snake = strings.ReplaceAll(snake, "__", "_")
	return snake
}

func Resolve(node *SimpleExprTree, resolver func(node *SimpleExprTree) error) (string, error) {
	return resolve(node, resolver)
}

// reconstructExpression tái tạo lại biểu thức ban đầu từ cây (giữ nguyên cấu trúc với ngoặc)
func resolve(node *SimpleExprTree, resolver func(node *SimpleExprTree) error) (string, error) {
	if node == nil {
		return "", nil
	}
	err := resolver(node)
	if err != nil {
		return "", err
	}
	//)
	// Nếu là nút lá (không có Ns)
	if len(node.Ns) == 0 {
		return node.V, nil
	}

	// Xử lý dựa trên toán tử
	if node.Op == "()" {
		// Nếu là biểu thức trong ngoặc, thêm ngoặc bao quanh
		subExpr, errOp := resolve(node.Ns[0], resolver)
		if errOp != nil {
			return "", errOp
		}
		return "(" + subExpr + ")", nil
	} else if node.Nt == "func" {
		// Xử lý hàm: ghép tên hàm với các đối số
		var args []string
		for _, child := range node.Ns {
			argN, errArgN := resolve(child, resolver)
			if errArgN != nil {
				return "", errArgN
			}

			args = append(args, argN)
		}
		return node.V + "(" + strings.Join(args, ", ") + ")", nil
	} else {
		// Ghép các biểu thức con với toán tử
		var parts []string
		for _, child := range node.Ns {
			argN, errArgN := resolve(child, resolver)
			if errArgN != nil {
				return "", errArgN
			}
			parts = append(parts, argN)
		}
		return strings.Join(parts, " "+node.Op+" "), nil
	}
}

// parseToSimpleExprTree chuyển đổi biểu thức phức tạp thành cây SimpleExprTree
func parseToSimpleExprTree(expr string) (*SimpleExprTree, error) {
	if expr == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	// Loại bỏ khoảng trắng thừa
	expr = strings.TrimSpace(expr)

	// Tạo nút gốc
	root := &SimpleExprTree{V: expr}

	// Kiểm tra nếu biểu thức đã tối giản
	if isSimpleExpr(expr) {
		setNodeType(root)
		return root, nil
	}

	// Xử lý dấu ngoặc bao ngoài (hỗ trợ nhiều lớp ngoặc)
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		openCount := 0
		matched := true
		for i, char := range expr {
			if char == '(' {
				openCount++
			} else if char == ')' {
				openCount--
			}
			if openCount == 0 && i != len(expr)-1 {
				matched = false
				break
			}
		}
		if matched && openCount == 0 {
			// Phân tích nội dung bên trong
			innerExpr := expr[1 : len(expr)-1]
			innerNode, err := parseToSimpleExprTree(innerExpr)
			if err != nil {
				return nil, err
			}
			root.Op = "()"
			root.Ns = append(root.Ns, innerNode)
			return root, nil
		}
	}

	// Tìm toán tử chính (ưu tiên thấp nhất: or, and, ==, like, v.v.)
	op, opPos := findMainOperator(expr)
	if op != "" {
		root.Op = op
		leftExpr := strings.TrimSpace(expr[:opPos])
		rightExpr := strings.TrimSpace(expr[opPos+len(op):])

		// Phân tích đệ quy cho các biểu thức con
		if leftExpr != "" {
			leftNode, err := parseToSimpleExprTree(leftExpr)
			if err != nil {
				return nil, err
			}
			root.Ns = append(root.Ns, leftNode)
		}
		if rightExpr != "" {
			rightNode, err := parseToSimpleExprTree(rightExpr)
			if err != nil {
				return nil, err
			}
			root.Ns = append(root.Ns, rightNode)
		}
		return root, nil
	}

	// Nếu là hàm, phân tích tham số
	if isFunction(expr) {
		funcNameEnd := strings.Index(expr, "(")
		if funcNameEnd == -1 {
			return nil, fmt.Errorf("invalid function format")
		}
		funcName := strings.TrimSpace(expr[:funcNameEnd])
		funcNode := &SimpleExprTree{V: funcName, Nt: "func"}
		args, err := parseFunctionArgs(expr)
		if err != nil {
			return nil, err
		}
		for _, arg := range args {
			argNode, err := parseToSimpleExprTree(arg)
			if err != nil {
				return nil, err
			}
			funcNode.Ns = append(funcNode.Ns, argNode)
		}
		*root = *funcNode
		return root, nil
	}

	return root, nil
}

// parseFunctionArgs tách các tham số của hàm
func parseFunctionArgs(expr string) ([]string, error) {
	var args []string
	if !strings.Contains(expr, "(") || !strings.Contains(expr, ")") {
		return nil, fmt.Errorf("invalid function format")
	}

	start := strings.Index(expr, "(") + 1
	end := strings.LastIndex(expr, ")")
	if start >= end {
		return nil, fmt.Errorf("invalid parentheses")
	}

	argStr := expr[start:end]
	parenDepth := 0
	currentArg := ""
	inQuotes := false

	for i := 0; i < len(argStr); i++ {
		char := argStr[i]
		if char == '\'' {
			inQuotes = !inQuotes
		} else if char == '(' && !inQuotes {
			parenDepth++
		} else if char == ')' && !inQuotes {
			parenDepth--
		} else if char == ',' && parenDepth == 0 && !inQuotes {
			args = append(args, strings.TrimSpace(currentArg))
			currentArg = ""
			continue
		}
		currentArg += string(char)
	}

	if currentArg != "" {
		args = append(args, strings.TrimSpace(currentArg))
	}

	return args, nil
}

// isSimpleExpr kiểm tra xem biểu thức có phải là tối giản không
func isSimpleExpr(expr string) bool {
	// 1. Tham số
	if expr == "?" {
		return true
	}
	// 2. Hằng số
	if isConstant(expr) {
		return true
	}
	// 3. Đối số (Field)
	if isOperand(expr) {
		return true
	}
	return false
}

// setNodeType đặt loại nút dựa trên giá trị
func setNodeType(node *SimpleExprTree) {
	if node.V == "?" {
		node.Nt = "param"
	} else if isConstant(node.V) {
		node.Nt = "const"
	} else if isOperand(node.V) {
		node.Nt = "field" // Đối số đơn giản xem như field
	}
}

// findMainOperator tìm toán tử chính (ưu tiên thấp nhất) trong biểu thức
func findMainOperator(expr string) (string, int) {
	parenDepth := 0
	inQuotes := false
	lowestPrecedence := 999
	opPos := -1
	op := ""

	i := 0
	for i < len(expr) {
		char := expr[i]

		if char == '\'' {
			inQuotes = !inQuotes
		} else if char == '(' {
			parenDepth++
		} else if char == ')' {
			parenDepth--
		} else if !inQuotes && parenDepth == 0 {
			// Kiểm tra các toán tử
			if strings.HasPrefix(expr[i:], "||") || strings.HasPrefix(expr[i:], "or") {
				precedence := 1
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					if strings.HasPrefix(expr[i:], "||") {
						op = "||"
					} else {
						op = "or"
					}
					opPos = i
				}
				i += len(op) - 1
			} else if strings.HasPrefix(expr[i:], "&&") || strings.HasPrefix(expr[i:], "and") {
				precedence := 2
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					if strings.HasPrefix(expr[i:], "&&") {
						op = "&&"
					} else {
						op = "and"
					}
					opPos = i
				}
				i += len(op) - 1
			} else if strings.HasPrefix(expr[i:], "==") || strings.HasPrefix(expr[i:], "=") {
				precedence := 3
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					if strings.HasPrefix(expr[i:], "==") {
						op = "=="
					} else {
						op = "="
					}
					opPos = i
				}
				i += len(op) - 1
			} else if strings.HasPrefix(expr[i:], "<=") || strings.HasPrefix(expr[i:], ">=") {
				precedence := 3
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					op = expr[i : i+2]
					opPos = i
				}
				i += 1
			} else if strings.HasPrefix(expr[i:], "like") {
				precedence := 3
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					op = "like"
					opPos = i
				}
				i += 3
			} else if char == '<' || char == '>' {
				precedence := 3
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					op = string(char)
					opPos = i
				}
			} else if char == '+' || char == '-' {
				precedence := 4
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					op = string(char)
					opPos = i
				}
			} else if char == '*' || char == '/' || char == '%' || char == '^' {
				precedence := 5
				if precedence < lowestPrecedence {
					lowestPrecedence = precedence
					op = string(char)
					opPos = i
				}
			}
		}
		i++
	}

	return op, opPos
}

// hasOuterParentheses kiểm tra xem biểu thức có dấu ngoặc bao ngoài không
func hasOuterParentheses(expr string) bool {
	if !strings.HasPrefix(expr, "(") || !strings.HasSuffix(expr, ")") {
		return false
	}
	openCount := 0
	for i, char := range expr {
		if char == '(' {
			openCount++
		} else if char == ')' {
			openCount--
		}
		if openCount == 0 && i != len(expr)-1 {
			return false
		}
	}
	return true
}

// isConstant kiểm tra xem biểu thức có phải là hằng số không
func isConstant(expr string) bool {
	// Hằng số số
	if _, err := fmt.Sscanf(expr, "%d", new(int)); err == nil {
		return true
	}
	// Hằng số chuỗi (bao trong dấu nháy đơn, hỗ trợ thoát '' cho ')
	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") {
		expr = expr[1 : len(expr)-1] // Loại bỏ dấu nháy ngoài
		return strings.Count(expr, "''") > 0 || !strings.ContainsAny(expr, "()+-*/=<>!&|,")
	}
	return false
}

// isFunction kiểm tra xem biểu thức có phải là hàm không
func isFunction(expr string) bool {
	// Hàm phải có dạng tên hàm + dấu ngoặc, ví dụ: concat(...)
	if !strings.Contains(expr, "(") || !strings.Contains(expr, ")") {
		return false
	}
	parenDepth := 0
	for i, char := range expr {
		if char == '(' {
			parenDepth++
			if parenDepth == 1 {
				funcName := expr[:i]
				return isOperand(funcName)
			}
		} else if char == ')' {
			parenDepth--
		}
	}
	return false
}

// isOperand kiểm tra xem chuỗi có phải là toán hạng không
func isOperand(s string) bool {
	// Toán hạng là biến (chữ cái, số, dấu gạch dưới)
	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

// printTree in cây biểu thức dưới dạng JSON với indent
func printTree(node *SimpleExprTree) {
	jsonData, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// reconstructExpression tái tạo lại biểu thức ban đầu từ cây (giữ nguyên cấu trúc với ngoặc)
func reconstructExpression(node *SimpleExprTree) string {
	if node == nil {
		return ""
	}

	// Nếu là nút lá (không có Ns)
	if len(node.Ns) == 0 {
		return node.V
	}

	// Xử lý dựa trên toán tử
	if node.Op == "()" {
		// Nếu là biểu thức trong ngoặc, thêm ngoặc bao quanh
		subExpr := reconstructExpression(node.Ns[0])
		return "(" + subExpr + ")"
	} else if node.Nt == "func" {
		// Xử lý hàm: ghép tên hàm với các đối số
		var args []string
		for _, child := range node.Ns {
			args = append(args, reconstructExpression(child))
		}
		return node.V + "(" + strings.Join(args, ", ") + ")"
	} else {
		// Ghép các biểu thức con với toán tử
		var parts []string
		for _, child := range node.Ns {
			parts = append(parts, reconstructExpression(child))
		}
		return strings.Join(parts, " "+node.Op+" ")
	}
}

// reconstructExpressionSimple tái tạo biểu thức mà không dùng ngoặc không cần thiết
func reconstructExpressionSimple(node *SimpleExprTree) string {
	if node == nil {
		return ""
	}

	// Nếu là nút lá (không có Ns)
	if len(node.Ns) == 0 {
		return node.V
	}

	// Xử lý dựa trên toán tử
	if node.Nt == "func" {
		// Xử lý hàm: ghép tên hàm với các đối số
		var args []string
		for _, child := range node.Ns {
			args = append(args, reconstructExpressionSimple(child))
		}
		return node.V + "(" + strings.Join(args, ", ") + ")"
	} else {
		// Ghép các biểu thức con với toán tử, bỏ qua ngoặc không cần thiết
		var parts []string
		for _, child := range node.Ns {
			parts = append(parts, reconstructExpressionSimple(child))
		}
		return strings.Join(parts, " "+node.Op+" ")
	}
}
