package client

type logger interface {
	LOGI(format string, args ...any)
	LOGD(format string, args ...any)
	LOGW(format string, args ...any)
	LOGE(format string, args ...any)
	DUMP(data []byte, format string, args ...any)
}

type log_t struct{}

func (m log_t) LOGI(format string, args ...any)              {}
func (m log_t) LOGD(format string, args ...any)              {}
func (m log_t) LOGW(format string, args ...any)              {}
func (m log_t) LOGE(format string, args ...any)              {}
func (m log_t) DUMP(data []byte, format string, args ...any) {}
