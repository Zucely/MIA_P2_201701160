package UsuariosGrupos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func Mkgrp(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR MKGRP: No hay una sesion activa")
		return
	}

	if Structs.SesionActiva.Nombre != "root" {
		fmt.Println("ERROR MKGRP: Solo el usuario root puede ejecutar este comando")
		return
	}

	var nombreGrupo string
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR MKGRP: valor desconocido de parametro ", valores[0])
			return
		}

		if strings.ToLower(valores[0]) == "name" {
			nombreGrupo = strings.ReplaceAll(valores[1], "\"", "")
			nombreGrupo = strings.TrimSpace(nombreGrupo)
		} else {
			fmt.Println("ERROR MKGRP: Parametro desconocido: ", valores[0])
			paramC = false
			break
		}
	}

	if !paramC {
		return
	}
	if nombreGrupo == "" {
		fmt.Println("ERROR MKGRP: Falta el parametro name")
		return
	}

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	contenido, err := Tools.LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR MKGRP: No se pudo leer users.txt:", err)
		return
	}

	// Validar que el grupo no exista ya (entre los activos, gid != 0)
	if grupoExiste(contenido, nombreGrupo) {
		fmt.Println("ERROR MKGRP: El grupo", nombreGrupo, "ya existe")
		return
	}

	// Calcular el siguiente GID disponible
	siguienteGid := obtenerSiguienteGid(contenido)

	nuevaLinea := fmt.Sprintf("%d,G,%s\n", siguienteGid, nombreGrupo)
	nuevoContenido := contenido + nuevaLinea

	if err := Tools.EscribirUsersTxt(pathDisco, nombreParticion, nuevoContenido); err != nil {
		fmt.Println("ERROR MKGRP: No se pudo actualizar users.txt:", err)
		return
	}

	fmt.Println("MKGRP: Grupo", nombreGrupo, "creado exitosamente con GID", siguienteGid)
}

// Verifica si un grupo (por nombre) ya existe entre los registros activos (gid != 0)
func grupoExiste(contenido string, nombreGrupo string) bool {
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

// Calcula el siguiente GID disponible, basado en el maximo GID usado (activo o eliminado)
func obtenerSiguienteGid(contenido string) int {
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

		if len(campos) == 3 && strings.ToUpper(campos[1]) == "G" {
			gid, err := strconv.Atoi(campos[0])
			if err == nil && gid > maximo {
				maximo = gid
			}
		}
	}
	return maximo + 1
}
