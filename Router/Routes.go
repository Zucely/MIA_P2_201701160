package Router

import (
	Controllers "MIA_P1/Controllers"
	"net/http"
)

// RegistrarRutas registra todos los endpoints de la API en el mux dado.
// Por ahora solo tenemos /api/mkdisk; aqui se iran agregando los demas
// comandos (fdisk, mount, mkfs, mkdir, mkfile, etc.) a medida que se migren.
func RegistrarRutas(mux *http.ServeMux) {
	mux.HandleFunc("/api/mkdisk", Controllers.MkdiskHandler)
}
