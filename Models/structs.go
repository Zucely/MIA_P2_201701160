package Models

// MkdiskRequest representa el cuerpo JSON que el cliente envia para crear un disco.
// Ejemplo de uso:
//
//	{
//	  "size": 10,
//	  "unit": "M",
//	  "fit": "FF",
//	  "path": "/home/zucely/mia/cali/discoApi1.dsk"
//	}
type MkdiskRequest struct {
	Size int    `json:"size"`
	Unit string `json:"unit"`
	Fit  string `json:"fit"`
	Path string `json:"path"`
}

// MkdiskResponse representa la respuesta JSON que se le devuelve al cliente
// luego de crear el disco exitosamente.
type MkdiskResponse struct {
	Mensaje       string `json:"mensaje"`
	Path          string `json:"path"`
	TamañoBytes   int32  `json:"tamañoBytes"`
	FechaCreacion string `json:"fechaCreacion"`
	Id            int32  `json:"id"`
	Fit           string `json:"fit"`
}

// ErrorResponse representa una respuesta de error generica para cualquier endpoint.
type ErrorResponse struct {
	Error string `json:"error"`
}
