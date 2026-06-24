package UsuariosGrupos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func Rmgrp(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR RMGRP: No hay una sesion activa")
		return
	}

	if Structs.SesionActiva.Nombre != "root" {
		fmt.Println("ERROR RMGRP: Solo el usuario root puede ejecutar este comando")
		return
	}

	var nombreGrupo string
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR RMGRP: valor desconocido de parametro ", valores[0])
			return
		}

		if strings.ToLower(valores[0]) == "name" {
			nombreGrupo = strings.ReplaceAll(valores[1], "\"", "")
			nombreGrupo = strings.TrimSpace(nombreGrupo)
		} else {
			fmt.Println("ERROR RMGRP: Parametro desconocido: ", valores[0])
			paramC = false
			break
		}
	}

	if !paramC {
		return
	}
	if nombreGrupo == "" {
		fmt.Println("ERROR RMGRP: Falta el parametro name")
		return
	}

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	contenido, err := Tools.LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR RMGRP: No se pudo leer users.txt:", err)
		return
	}

	nuevoContenido, encontrado := eliminarGrupo(contenido, nombreGrupo)
	if !encontrado {
		fmt.Println("ERROR RMGRP: El grupo", nombreGrupo, "no existe")
		return
	}

	if err := Tools.EscribirUsersTxt(pathDisco, nombreParticion, nuevoContenido); err != nil {
		fmt.Println("ERROR RMGRP: No se pudo actualizar users.txt:", err)
		return
	}

	fmt.Println("RMGRP: Grupo", nombreGrupo, "eliminado exitosamente")
}

// Busca la linea del grupo (activo, gid != 0) y le cambia el GID a 0.
// Retorna el contenido actualizado y si se encontro el grupo.
func eliminarGrupo(contenido string, nombreGrupo string) (string, bool) {
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

		if len(camposLimpios) == 3 && strings.ToUpper(camposLimpios[1]) == "G" {
			gid, err := strconv.Atoi(camposLimpios[0])
			if err != nil || gid == 0 {
				continue
			}
			if camposLimpios[2] == nombreGrupo {
				lineas[i] = fmt.Sprintf("0,G,%s", camposLimpios[2])
				encontrado = true
			}
		}
	}

	if !encontrado {
		return contenido, false
	}

	return strings.Join(lineas, "\n"), true
}
