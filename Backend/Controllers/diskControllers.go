package Controllers

import (
	Models "MIA_P1/Models"
	DiskService "MIA_P1/Services/DiskService"
	"encoding/json"
	"net/http"
)

func MkdiskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Models.MkdiskResponse{Success: false, Message: "Método no permitido"})
		return
	}

	var req Models.MkdiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Models.MkdiskResponse{Success: false, Message: "JSON inválido"})
		return
	}

	_, err := DiskService.CrearDiscoLogico(req.Size, req.Unit, req.Fit, req.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Models.MkdiskResponse{Success: false, Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(Models.MkdiskResponse{Success: true, Message: "Disco creado exitosamente"})
}
