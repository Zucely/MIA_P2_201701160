package Controllers

import (
	Comandos "MIA_P1/Comandos/AdmDeDiscos"
	Models "MIA_P1/Models"
	"encoding/json"
	"net/http"
	"strings"
)

// MkdiskHandler atiende peticiones POST /api/mkdisk.
// Recibe un JSON con size, unit, fit y path; crea el disco; responde con JSON.
func MkdiskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Models.ErrorResponse{Error: "Metodo no permitido, use POST"})
		return
	}

	var req Models.MkdiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Models.ErrorResponse{Error: "JSON invalido: " + err.Error()})
		return
	}

	mbr, err := Comandos.CrearDiscoLogico(req.Size, req.Unit, req.Fit, req.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Models.ErrorResponse{Error: err.Error()})
		return
	}

	respuesta := Models.MkdiskResponse{
		Mensaje:       "Disco creado exitosamente",
		Path:          req.Path,
		TamañoBytes:   mbr.MbrSize,
		FechaCreacion: limpiarBytesNulos(string(mbr.FechaC[:])),
		Id:            mbr.Id,
		Fit:           limpiarBytesNulos(string(mbr.Fit[:])),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(respuesta)
}

// limpiarBytesNulos quita los bytes nulos de relleno de campos [N]byte,
// igual que Structs.GetName, pero local a este controller para no acoplar
// demasiado el paquete Controllers con Structs internos.
func limpiarBytesNulos(s string) string {
	if pos := strings.IndexByte(s, 0); pos != -1 {
		s = s[:pos]
	}
	return s
}
