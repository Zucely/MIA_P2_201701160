package SistemaDeArchivos

import (
	Structs "MIA_P1/Structs"
	"fmt"
)

func Logout() {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR LOGOUT: No hay una sesion activa")
		return
	}

	fmt.Println("LOGOUT: Sesion de", Structs.SesionActiva.Nombre, "cerrada correctamente")

	Structs.SesionActiva = Structs.UserInfo{}
}
