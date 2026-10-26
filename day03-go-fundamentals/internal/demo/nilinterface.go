package demo

// MyErr เป็น error ที่รับด้วย pointer
type MyErr struct{ Msg string }

func (e *MyErr) Error() string { return e.Msg }

// BuggyFail: ตัวแปร p เป็น *MyErr (nil) แล้ว return เป็น error
// interface ที่ได้มี (type=*MyErr, value=nil) ไม่ใช่ nil interface -> err != nil เป็น true
func BuggyFail() error {
	var p *MyErr
	return p
}

// FixedFail: คืน nil ตรงๆ
func FixedFail() error {
	return nil
}

// IsNilInterface บอกว่า err เป็น nil interface จริงไหม
func IsNilInterface(err error) bool { return err == nil }

// Describe ใช้ type switch
func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int:
		return "int"
	case string:
		return "string:" + x
	case error:
		return "error:" + x.Error()
	default:
		return "other"
	}
}
