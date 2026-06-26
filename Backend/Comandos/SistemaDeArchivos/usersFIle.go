package SistemaDeArchivos

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
)

// Lee el contenido completo de users.txt (inodo 1, por convencion del mkfs)
// recorriendo sus bloques directos.
func LeerUsersTxt(pathDisco string, nombreParticion string) (string, error) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		return "", err
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return "", err
	}

	var partStart int32 = -1
	for i := 0; i < 4; i++ {
		nombre := Structs.GetName(string(mbr.Partitions[i].Name[:]))
		if nombre == nombreParticion {
			partStart = mbr.Partitions[i].Start
			break
		}
	}
	if partStart == -1 {
		return "", fmt.Errorf("particion no encontrada")
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		return "", err
	}

	// Buscar el inodo de users.txt dentro de la carpeta raiz (inodo 0)
	numInodoUsers, err := BuscarEnCarpeta(disco, sb, 0, "users.txt")
	if err != nil {
		return "", err
	}

	// Leer el inodo de users.txt
	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodoUsers * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		return "", err
	}

	// Leer el contenido de sus bloques directos, hasta completar I_size bytes
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

	return contenido.String(), nil
}

// Busca un nombre dentro del bloque de carpeta de un inodo dado, retorna el numero de inodo encontrado
func BuscarEnCarpeta(disco *os.File, sb Structs.Superblock, numInodoCarpeta int32, nombreBuscado string) (int32, error) {
	var inodoCarpeta Structs.Inode
	posInodo := sb.S_inode_start + (numInodoCarpeta * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodoCarpeta, int64(posInodo)); err != nil {
		return -1, err
	}

	for b := 0; b < 12; b++ {
		numBloque := inodoCarpeta.I_block[b]
		//fmt.Println("DEBUG BuscarEnCarpeta: inodo", numInodoCarpeta, "I_block[", b, "] =", numBloque)
		if numBloque == -1 {
			continue
		}

		var bloque Structs.Folderblock
		posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
		if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
			continue
		}

		for _, c := range bloque.B_content {
			nombre := Structs.GetB_name(string(c.B_name[:]))
			//fmt.Printf("DEBUG contenido bloque %d: nombre='%s' inodo=%d buscando='%s'\n", numBloque, nombre, c.B_inodo, nombreBuscado)
			if nombre == nombreBuscado {
				//fmt.Println("DEBUG: MATCH encontrado, retornando inodo", c.B_inodo)
				return c.B_inodo, nil

			}
		}
	}

	return -1, fmt.Errorf("no se encontro %s", nombreBuscado)
}
