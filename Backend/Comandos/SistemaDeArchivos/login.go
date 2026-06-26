package SistemaDeArchivos

import (
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func Login(parametros []string) {
	var user, pass, id string
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR LOGIN: valor desconocido de parametro ", valores[0])
			return
		}

		clave := strings.ToLower(valores[0])
		valor := strings.ReplaceAll(valores[1], "\"", "")
		valor = strings.TrimSpace(valor)

		switch clave {
		case "user":
			user = valor
		case "pass":
			pass = valor
		case "id":
			id = valor
		default:
			fmt.Println("ERROR LOGIN: Parametro desconocido: ", valores[0])
			paramC = false
		}
	}

	if !paramC {
		return
	}
	if user == "" || pass == "" || id == "" {
		fmt.Println("ERROR LOGIN: Faltan parametros obligatorios (user, pass, id)")
		return
	}

	if Structs.SesionActiva.Status {
		fmt.Println("ERROR LOGIN: Ya existe una sesion activa, debe cerrar sesion primero")
		return
	}

	idx := Structs.BuscarMontadaPorId(id)
	if idx == -1 {
		fmt.Println("ERROR LOGIN: No existe una particion montada con el id ", id)
		return
	}
	pathDisco := Structs.Montadas[idx].PathM
	nombreParticion := Structs.Montadas[idx].Name

	contenido, err := LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR LOGIN: No se pudo leer users.txt:", err)
		return
	}

	encontrado, passCorrecta, idUsr, idGrp := validarCredenciales(contenido, user, pass)

	if !encontrado {
		fmt.Println("ERROR LOGIN: El usuario no existe")
		return
	}
	if !passCorrecta {
		fmt.Println("ERROR LOGIN: Autenticación fallida, contraseña incorrecta")
		return
	}

	Structs.SesionActiva = Structs.UserInfo{
		Id:         id,
		IdGrp:      idGrp,
		IdUsr:      idUsr,
		Nombre:     user,
		Status:     true,
		PathD:      pathDisco,
		NombrePart: nombreParticion,
	}

	fmt.Println("LOGIN: Sesion iniciada correctamente como", user)
}

// Retorna: encontrado, passCorrecta, idUsr, idGrp
func validarCredenciales(contenido string, user string, pass string) (bool, bool, int32, int32) {
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}
		campos := strings.Split(linea, ",")
		for i := range campos {
			campos[i] = strings.TrimSpace(campos[i])
		}

		// Linea de usuario: UID, Tipo, Grupo, Usuario, Contraseña
		if len(campos) == 5 && strings.ToUpper(campos[1]) == "U" {
			uidStr := campos[0]
			grupoNombre := campos[2]
			usuarioLinea := campos[3]
			passLinea := campos[4]

			uid, err := strconv.Atoi(uidStr)
			if err != nil || uid == 0 {
				continue // eliminado o invalido
			}

			if usuarioLinea == user {
				if passLinea == pass {
					idGrp := buscarIdGrupo(contenido, grupoNombre)
					return true, true, int32(uid), idGrp
				}
				return true, false, 0, 0
			}
		}
	}
	return false, false, 0, 0
}

// Busca el GID de un grupo por su nombre, recorriendo las lineas de tipo G
func buscarIdGrupo(contenido string, nombreGrupo string) int32 {
	lineas := strings.Split(contenido, "\n")
	for _, linea := range lineas {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}
		campos := strings.Split(linea, ",")
		for i := range campos {
			campos[i] = strings.TrimSpace(campos[i])
		}

		// Linea de grupo: GID, Tipo, Grupo
		if len(campos) == 3 && strings.ToUpper(campos[1]) == "G" {
			gidStr := campos[0]
			nombreLinea := campos[2]

			gid, err := strconv.Atoi(gidStr)
			if err != nil || gid == 0 {
				continue
			}

			if nombreLinea == nombreGrupo {
				return int32(gid)
			}
		}
	}
	return -1
}
