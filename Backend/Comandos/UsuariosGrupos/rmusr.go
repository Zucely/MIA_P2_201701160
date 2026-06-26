package UsuariosGrupos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func Rmusr(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR RMUSR: No hay una sesion activa")
		return
	}

	if Structs.SesionActiva.Nombre != "root" {
		fmt.Println("ERROR RMUSR: Solo el usuario root puede ejecutar este comando")
		return
	}

	var nombreUsuario string
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR RMUSR: valor desconocido de parametro ", valores[0])
			return
		}

		if strings.ToLower(valores[0]) == "user" {
			nombreUsuario = strings.ReplaceAll(valores[1], "\"", "")
			nombreUsuario = strings.TrimSpace(nombreUsuario)
		} else {
			fmt.Println("ERROR RMUSR: Parametro desconocido: ", valores[0])
			paramC = false
			break
		}
	}

	if !paramC {
		return
	}
	if nombreUsuario == "" {
		fmt.Println("ERROR RMUSR: Falta el parametro user")
		return
	}

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	contenido, err := Tools.LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR RMUSR: No se pudo leer users.txt:", err)
		return
	}

	nuevoContenido, encontrado := eliminarUsuario(contenido, nombreUsuario)
	if !encontrado {
		fmt.Println("ERROR RMUSR: El usuario", nombreUsuario, "no existe")
		return
	}

	if err := Tools.EscribirUsersTxt(pathDisco, nombreParticion, nuevoContenido); err != nil {
		fmt.Println("ERROR RMUSR: No se pudo actualizar users.txt:", err)
		return
	}

	fmt.Println("RMUSR: Usuario", nombreUsuario, "eliminado exitosamente")
}

// Busca la linea del usuario (activo, uid != 0) y le cambia el UID a 0.
// Retorna el contenido actualizado y si se encontro el usuario.
func eliminarUsuario(contenido string, nombreUsuario string) (string, bool) {
	lineas := strings.Split(contenido, "\n")
	encontrado := false

	for i, linea := range lineas {
		lineaTrim := strings.TrimSpace(linea)
		if lineaTrim == "" {
			continue
		}
		campos := strings.Split(lineaTrim, ",")
		camposLimpios := make([]string, len(campos))
		for j := range campos {
			camposLimpios[j] = strings.TrimSpace(campos[j])
		}

		if len(camposLimpios) == 5 && strings.ToUpper(camposLimpios[1]) == "U" {
			uid, err := strconv.Atoi(camposLimpios[0])
			if err != nil || uid == 0 {
				continue
			}
			if camposLimpios[3] == nombreUsuario {
				lineas[i] = fmt.Sprintf("0,U,%s,%s,%s", camposLimpios[2], camposLimpios[3], camposLimpios[4])
				encontrado = true
			}
		}
	}

	if !encontrado {
		return contenido, false
	}

	return strings.Join(lineas, "\n"), true
}
