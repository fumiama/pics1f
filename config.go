package pics1f

import (
	"net/http"
)

var (
	// UserAgent used in request
	UserAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.35`
	// HTTPClient used in request
	HTTPClient = http.Client{}
)
