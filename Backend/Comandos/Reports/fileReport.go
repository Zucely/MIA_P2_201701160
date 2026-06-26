package reports

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
)

func GenerarReporteFile(pathDisco string, nombreParticion string, pathSalida string, pathFileLs string) {
	if pathFileLs == "" {
		fmt.Println("ERROR REP FILE: Falta el parametro path_file_ls")
		return
	}

	disco, sb, _, err := Tools.AbrirDiscoYSuperbloque(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR REP FILE:", err)
		return
	}
	defer disco.Close()

	fmt.Println("DEBUG FILE: buscando ruta:", pathFileLs)
	fmt.Println("DEBUG FILE: en disco:", pathDisco, "particion:", nombreParticion)

	numInodo, err := Tools.BuscarRuta(disco, sb, pathFileLs)
	if err != nil {
		fmt.Println("ERROR REP FILE: No se encontro el archivo", pathFileLs)
		return
	}

	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodo * sb.S_inode_size)

	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		fmt.Println("ERROR REP FILE: No se pudo leer el inodo")
		return
	}

	if limpiarBytes(inodo.I_type[:]) != "1" {
		fmt.Println("ERROR REP FILE: La ruta", pathFileLs, "no corresponde a un archivo")
		return
	}

	var contenido strings.Builder
	bytesLeidos := int32(0)

	for i := 0; i < 12 && bytesLeidos < inodo.I_size; i++ {
		numBloque := inodo.I_block[i]

		if numBloque == -1 {
			break
		}

		var bloque Structs.Fileblock
		posBloque := sb.S_block_start + (numBloque * sb.S_block_size)

		if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
			fmt.Println("ERROR REP FILE: No se pudo leer el bloque", numBloque)
			return
		}

		restante := inodo.I_size - bytesLeidos
		aLeer := int32(len(bloque.B_content))

		if restante < aLeer {
			aLeer = restante
		}

		contenido.Write(bloque.B_content[:aLeer])
		bytesLeidos += aLeer
	}

	contenidoStr := contenido.String()

	if contenidoStr == "" {
		contenidoStr = "(archivo vacio)"
	}

	// Nombre del archivo + contenido
	salida := fmt.Sprintf(
		"Archivo: %s\n\n%s",
		ultimoSegmento(pathFileLs),
		contenidoStr,
	)

	if err := os.WriteFile(pathSalida, []byte(salida), 0644); err != nil {
		fmt.Println("ERROR REP FILE: No se pudo guardar el archivo:", err)
		return
	}

	fmt.Println("REP FILE: Reporte generado exitosamente en:", pathSalida)
}
