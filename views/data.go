package views

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
	AlertMsgGeneric = "Ooops... Something Went Wrong!"
)

// Alert ...
type Alert struct {
	Level   string
	Message string
}

// Data ...
type Data struct {
	Alert *Alert
	Yield interface{}
}
