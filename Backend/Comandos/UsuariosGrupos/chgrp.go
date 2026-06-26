package UsuariosGrupos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func Chgrp(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR CHGRP: No hay una sesion activa")
		return
	}

	if Structs.SesionActiva.Nombre != "root" {
		fmt.Println("ERROR CHGRP: Solo el usuario root puede ejecutar este comando")
		return
	}

	var user, grp string
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR CHGRP: valor desconocido de parametro ", valores[0])
			return
		}

		clave := strings.ToLower(valores[0])
		valor := strings.ReplaceAll(valores[1], "\"", "")
		valor = strings.TrimSpace(valor)

		switch clave {
		case "user":
			user = valor
		case "grp":
			grp = valor
		default:
			fmt.Println("ERROR CHGRP: Parametro desconocido: ", valores[0])
			paramC = false
		}
	}

	if !paramC {
		return
	}
	if user == "" || grp == "" {
		fmt.Println("ERROR CHGRP: Faltan parametros obligatorios (user, grp)")
		return
	}

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	contenido, err := Tools.LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR CHGRP: No se pudo leer users.txt:", err)
		return
	}

	if !usuarioActivo(contenido, user) {
		fmt.Println("ERROR CHGRP: El usuario", user, "no existe")
		return
	}

	if !grupoActivo(contenido, grp) {
		fmt.Println("ERROR CHGRP: El grupo", grp, "no existe")
		return
	}

	nuevoContenido := cambiarGrupoUsuario(contenido, user, grp)

	if err := Tools.EscribirUsersTxt(pathDisco, nombreParticion, nuevoContenido); err != nil {
		fmt.Println("ERROR CHGRP: No se pudo actualizar users.txt:", err)
		return
	}

	fmt.Println("CHGRP: El usuario", user, "ahora pertenece al grupo", grp)
}

// Verifica si un usuario (por nombre) esta activo (uid != 0)
func usuarioActivo(contenido string, user string) bool {
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

// Cambia el campo Grupo de la linea del usuario indicado
func cambiarGrupoUsuario(contenido string, user string, nuevoGrupo string) string {
	lineas := strings.Split(contenido, "\n")

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
			if camposLimpios[3] == user {
				lineas[i] = fmt.Sprintf("%s,U,%s,%s,%s", camposLimpios[0], nuevoGrupo, camposLimpios[3], camposLimpios[4])
			}
		}
	}

	return strings.Join(lineas, "\n")
}
