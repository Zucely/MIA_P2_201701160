package Router

import (
	Controllers "MIA_P1/Controllers"
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/api/mkdisk", Controllers.MkdiskHandler)
}
