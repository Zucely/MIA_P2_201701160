package reports

import (
	//Structs "MIA_P1/Structs"
	"MIA_P1/Structs"
	"fmt"
	"strings"
)

func Rep(parametros []string) {
	var id string
	var path string
	var name string
	paramC := true

	//para reporte ls y file
	var pathFileLs string

	//parametros: rep -name=MBR -path=/home/”su_usuario”/Calificacion_MIA/Reportes/MBR_Disco1.dot -id=Disco1
	for _, parametro := range parametros[1:] {
		tmp := strings.TrimSpace(parametro)
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR REP, valor desconocido de parametros ", valores[1])
			return //Finaliza comando
		}

		//******************* NAME *************
		if strings.ToLower(valores[0]) == "name" {
			name = strings.ReplaceAll(valores[1], "\"", "")
			name = strings.ToLower(strings.TrimSpace(name))

			//******************* PATH *************
		} else if strings.ToLower(valores[0]) == "path" {
			path = strings.ReplaceAll(valores[1], "\"", "")

			//******************* ID *************
		} else if strings.ToLower(valores[0]) == "id" {
			id = strings.ReplaceAll(valores[1], "\"", "")
			id = strings.TrimSpace(id)

			//******************* PATH_FILE_LS ************* (este puede o no venir)
		} else if strings.ToLower(valores[0]) == "path_file_ls" {
			pathFileLs = strings.ToLower(strings.TrimSpace(valores[1]))
			//******************* ERROR EN LOS PARAMETROS *************
		} else {
			fmt.Println("ERROR REP: Parametro desconocido: ", valores[0])
			paramC = false
			break //por si en el camino reconoce algo invalido de una vez se sale
		}
	}

	if !paramC {
		fmt.Println("ERROR REP: Parametros incorrectos")
		return
	}

	if id == "" {
		fmt.Println("ERROR REP: Falta parametro ID")
		return
	}
	if path == "" {
		fmt.Println("ERROR REP: Falta parametro PATH")
		return
	}
	if name == "" {
		fmt.Println("ERROR REP: Falta parametro NAME")
		return
	}
	if (name == "ls" || name == "file") && pathFileLs == "" {
		fmt.Println("ERROR REP: Falta parametro path_file_ls para el reporte ", name)
		return
	}

	idToLookFor := Structs.BuscarMontadaPorId(id)
	if idToLookFor == -1 {
		fmt.Println("ERROR REP: No existe una particion montada con el ID ", id)
		return
	}

	pathDisco := Structs.Montadas[idToLookFor].PathM

	switch name {
	case "MBR", "mbr":
		GenerarReporteMBR(pathDisco, path)
	case "Disk", "disk":
		GenerarReporteDisk(pathDisco, path)
	case "sb", "SB":
		GenerarReporteSB(pathDisco, Structs.Montadas[idToLookFor].Name, path)
	case "bm_inode":
		GenerarReporteBMInode(pathDisco, Structs.Montadas[idToLookFor].Name, path)
	case "inode":
		GenerarReporteInode(pathDisco, Structs.Montadas[idToLookFor].Name, path)
	case "block":
		GenerarReporteBlock(pathDisco, Structs.Montadas[idToLookFor].Name, path)
	case "bm_block":
		GenerarReporteBMBlock(pathDisco, Structs.Montadas[idToLookFor].Name, path)
	case "tree":
		GenerarReporteTree(pathDisco, Structs.Montadas[idToLookFor].Name, path)
	case "ls":
		GenerarReporteLs(pathDisco, Structs.Montadas[idToLookFor].Name, path, pathFileLs)
	case "file":
		GenerarReporteFile(pathDisco, Structs.Montadas[idToLookFor].Name, path, pathFileLs)
	default:
		fmt.Println("ERROR REP: No se reconoce el nombre del reporte ", name)
		return
	}

}
