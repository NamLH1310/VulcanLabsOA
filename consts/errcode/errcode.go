package errcode

const (
	Success = iota

	InvalidParameters
)

func Text(code int) string {
	switch code {
	case Success:
		return "Success"
	case InvalidParameters:
		return "Invalid parameters"
	default:
		return ""
	}
}
