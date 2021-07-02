package fetcher

import (
	"net/http"
	"time"
)

type Fetcher interface {
	Fetch(userName string, since time.Time, until time.Time) http.Response
}