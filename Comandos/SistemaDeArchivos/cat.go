package SistemaDeArchivos

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Cat(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR CAT: No hay una sesion activa")
		return
	}

	// Recolectar los parametros -file1, -file2, -file3... en un mapa para luego ordenarlos
	archivos := make(map[int]string)
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR CAT: valor desconocido de parametro ", valores[0])
			return
		}

		clave := strings.ToLower(valores[0])
		valorPath := strings.ReplaceAll(valores[1], "\"", "")
		valorPath = strings.TrimSpace(valorPath)

		if strings.HasPrefix(clave, "file") {
			numStr := strings.TrimPrefix(clave, "file")
			num, err := strconv.Atoi(numStr)
			if err != nil {
				fmt.Println("ERROR CAT: Parametro de archivo invalido: ", valores[0])
				paramC = false
				break
			}
			archivos[num] = valorPath
		} else {
			fmt.Println("ERROR CAT: Parametro desconocido: ", valores[0])
			paramC = false
			break
		}
	}

	if !paramC {
		return
	}
	if len(archivos) == 0 {
		fmt.Println("ERROR CAT: Debe especificar al menos un archivo (-file1, -file2, ...)")
		return
	}

	// Ordenar las claves numericas (file1, file2, file3...) para respetar el orden
	claves := make([]int, 0, len(archivos))
	for k := range archivos {
		claves = append(claves, k)
	}
	sort.Ints(claves)

	var resultado strings.Builder

	for idx, k := range claves {
		rutaArchivo := archivos[k]

		contenido, err := leerArchivoConPermiso(rutaArchivo)
		if err != nil {
			fmt.Println("ERROR CAT:", err)
			return
		}

		resultado.WriteString(contenido)
		if idx < len(claves)-1 {
			resultado.WriteString("\n")
		}
	}

	fmt.Println(resultado.String())
}

// Busca el archivo por su ruta absoluta desde la raiz, valida permisos de lectura,
// y retorna su contenido completo.
func leerArchivoConPermiso(rutaArchivo string) (string, error) {
	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	disco, sb, _, err := AbrirDiscoYSuperbloque(pathDisco, nombreParticion) // fileSystemUtils.go
	if err != nil {
		return "", err
	}
	defer disco.Close()

	numInodo, err := BuscarRuta(disco, sb, rutaArchivo) // fileSystemUtils.go
	if err != nil {
		return "", fmt.Errorf("el archivo %s no existe", rutaArchivo)
	}

	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodo * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		return "", fmt.Errorf("no se pudo leer el inodo de %s", rutaArchivo)
	}

	// Validar que sea un archivo, no una carpeta
	if LimpiarBytesFS(inodo.I_type[:]) != "1" {
		return "", fmt.Errorf("%s es una carpeta, no un archivo", rutaArchivo)
	}

	// Validar permiso de lectura
	if !tienePermisoLectura(inodo) {
		return "", fmt.Errorf("no tiene permiso de lectura sobre %s", rutaArchivo)
	}

	return leerContenidoInodo(disco, sb, inodo)
}

// Valida si el usuario de la sesion activa tiene permiso de LECTURA sobre el inodo dado
func tienePermisoLectura(inodo Structs.Inode) bool {
	// Root siempre tiene acceso total
	if Structs.SesionActiva.Nombre == "root" {
		return true
	}

	perm := LimpiarBytesFS(inodo.I_perm[:]) // ej "664"
	if len(perm) != 3 {
		return false
	}

	var digito byte
	if inodo.I_uid == Structs.SesionActiva.IdUsr {
		digito = perm[0] // propietario
	} else if inodo.I_gid == Structs.SesionActiva.IdGrp {
		digito = perm[1] // grupo
	} else {
		digito = perm[2] // otros
	}

	valor := int(digito - '0') // convierte el caracter '0'-'7' a su valor numerico
	return valor >= 4          // 4,5,6,7 tienen el bit de lectura activado
}

// Lee el contenido completo de un inodo de tipo archivo, recorriendo sus bloques directos
func leerContenidoInodo(disco *os.File, sb Structs.Superblock, inodo Structs.Inode) (string, error) {
	var contenido strings.Builder
	bytesLeidos := int32(0)

	for b := 0; b < 12 && bytesLeidos < inodo.I_size; b++ {
		numBloque := inodo.I_block[b]
		if numBloque == -1 {
			break
		}

		var bloque Structs.Fileblock
		posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
		if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
			return "", err
		}

		restante := inodo.I_size - bytesLeidos
		aLeer := sb.S_block_size
		if restante < aLeer {
			aLeer = restante
		}

		contenido.Write(bloque.B_content[:aLeer])
		bytesLeidos += aLeer
	}

	// TODO: si el archivo usa bloques indirectos (I_block[12..14]), falta leerlos aqui

	return contenido.String(), nil
}
