package UsuariosGrupos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func Mkusr(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR MKUSR: No hay una sesion activa")
		return
	}

	if Structs.SesionActiva.Nombre != "root" {
		fmt.Println("ERROR MKUSR: Solo el usuario root puede ejecutar este comando")
		return
	}

	var user, pass, grp string
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR MKUSR: valor desconocido de parametro ", valores[0])
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
		case "grp":
			grp = valor
		default:
			fmt.Println("ERROR MKUSR: Parametro desconocido: ", valores[0])
			paramC = false
		}
	}

	if !paramC {
		return
	}
	if user == "" || pass == "" || grp == "" {
		fmt.Println("ERROR MKUSR: Faltan parametros obligatorios (user, pass, grp)")
		return
	}

	if len(user) > 10 {
		fmt.Println("ERROR MKUSR: El nombre de usuario no puede exceder 10 caracteres")
		return
	}
	if len(pass) > 10 {
		fmt.Println("ERROR MKUSR: La contraseña no puede exceder 10 caracteres")
		return
	}
	if len(grp) > 10 {
		fmt.Println("ERROR MKUSR: El nombre de grupo no puede exceder 10 caracteres")
		return
	}

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	contenido, err := Tools.LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR MKUSR: No se pudo leer users.txt:", err)
		return
	}

	// Validar que el usuario no exista ya (entre los activos, uid != 0), sin importar el grupo
	if usuarioExiste(contenido, user) {
		fmt.Println("ERROR MKUSR: El usuario", user, "ya existe")
		return
	}

	// Validar que el grupo exista y este activo
	if !grupoActivo(contenido, grp) {
		fmt.Println("ERROR MKUSR: El grupo", grp, "no existe")
		return
	}

	siguienteUid := obtenerSiguienteUid(contenido)

	nuevaLinea := fmt.Sprintf("%d,U,%s,%s,%s\n", siguienteUid, grp, user, pass)
	nuevoContenido := contenido + nuevaLinea

	if err := Tools.EscribirUsersTxt(pathDisco, nombreParticion, nuevoContenido); err != nil {
		fmt.Println("ERROR MKUSR: No se pudo actualizar users.txt:", err)
		return
	}

	fmt.Println("MKUSR: Usuario", user, "creado exitosamente con UID", siguienteUid)
}

// Verifica si un usuario (por nombre) ya existe entre los registros activos (uid != 0)
func usuarioExiste(contenido string, user string) bool {
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

		if len(campos) == 5 && strings.ToUpper(campos[1]) == "U" {
			uid, err := strconv.Atoi(campos[0])
			if err != nil || uid == 0 {
				continue
			}
			if campos[3] == user {
				return true
			}
		}
	}
	return false
}

// Verifica si un grupo (por nombre) existe y esta activo (gid != 0)
func grupoActivo(contenido string, nombreGrupo string) bool {
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

		if len(campos) == 3 && strings.ToUpper(campos[1]) == "G" {
			gid, err := strconv.Atoi(campos[0])
			if err != nil || gid == 0 {
				continue
			}
			if campos[2] == nombreGrupo {
				return true
			}
		}
	}
	return false
}

// Calcula el siguiente UID disponible, basado en el maximo UID usado (activo o eliminado)
func obtenerSiguienteUid(contenido string) int {
	maximo := 0
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

		if len(campos) == 5 && strings.ToUpper(campos[1]) == "U" {
			uid, err := strconv.Atoi(campos[0])
			if err == nil && uid > maximo {
				maximo = uid
			}
		}
	}
	return maximo + 1
}
