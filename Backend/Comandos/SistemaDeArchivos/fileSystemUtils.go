package SistemaDeArchivos

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
	"time"
)

// Abre el disco y lee su superbloque, dado el path del disco y nombre de particion
func AbrirDiscoYSuperbloque(pathDisco string, nombreParticion string) (*os.File, Structs.Superblock, int32, error) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		return nil, Structs.Superblock{}, -1, fmt.Errorf("no se pudo abrir el disco")
	}

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		disco.Close()
		return nil, Structs.Superblock{}, -1, fmt.Errorf("no se pudo leer el MBR")
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
		disco.Close()
		return nil, Structs.Superblock{}, -1, fmt.Errorf("particion no encontrada")
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		disco.Close()
		return nil, Structs.Superblock{}, -1, fmt.Errorf("no se pudo leer el superbloque")
	}

	return disco, sb, partStart, nil
}

// Resuelve una ruta absoluta (ej: /home/user/a.txt) navegando desde la raiz (inodo 0)
// y retorna el numero de inodo del elemento final.
func BuscarRuta(disco *os.File, sb Structs.Superblock, ruta string) (int32, error) {
	//fmt.Println("DEBUG BuscarRuta: ruta recibida:", ruta)
	ruta = strings.TrimPrefix(ruta, "/")
	//fmt.Println("DEBUG BuscarRuta: ruta limpia:", ruta)
	partes := strings.Split(ruta, "/")
	//fmt.Println("DEBUG BuscarRuta: partes:", partes)

	numInodoActual := int32(0)
	for _, parte := range partes {
		if parte == "" {
			continue
		}
		//fmt.Println("DEBUG BuscarRuta: buscando parte:", parte, "en inodo:", numInodoActual)
		siguiente, err := BuscarEnCarpeta(disco, sb, numInodoActual, parte)
		if err != nil {
			//fmt.Println("DEBUG BuscarRuta: no encontro:", parte, "error:", err)
			return -1, fmt.Errorf("no se encontro %s", parte)
		}
		//fmt.Println("DEBUG BuscarRuta: encontro:", parte, "-> inodo:", siguiente)
		numInodoActual = siguiente
	}
	return numInodoActual, nil
}

// Quita los bytes nulos (\x00) que quedan al final de los campos [N]byte
func LimpiarBytesFS(b []byte) string {
	s := string(b)
	if pos := strings.IndexByte(s, 0); pos != -1 {
		s = s[:pos]
	}
	return strings.TrimSpace(s)
}

// Sobrescribe el contenido completo de users.txt (inodo 1), liberando los bloques viejos
// y asignando nuevos bloques segun el contenido actualizado.
func EscribirUsersTxt(pathDisco string, nombreParticion string, nuevoContenido string) error {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el disco")
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		return fmt.Errorf("no se pudo leer el MBR")
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
		return fmt.Errorf("particion no encontrada")
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		return fmt.Errorf("no se pudo leer el superbloque")
	}

	// Buscar el inodo de users.txt (siempre en la raiz)
	numInodoUsers, err := BuscarEnCarpeta(disco, sb, 0, "users.txt")
	if err != nil {
		return fmt.Errorf("no se encontro users.txt")
	}

	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodoUsers * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		return fmt.Errorf("no se pudo leer el inodo de users.txt")
	}

	// Liberar los bloques viejos que ocupaba (marcarlos como libres en el bitmap)
	bloquesLiberados := int32(0)
	for b := 0; b < 12; b++ {
		if inodo.I_block[b] != -1 {
			if err := MarcarBitFS(disco, sb.S_bm_block_start, inodo.I_block[b], '0'); err != nil {
				return err
			}
			inodo.I_block[b] = -1
			bloquesLiberados++
		}
	}

	// Calcular cuantos bloques nuevos se necesitan
	tamañoNuevo := int32(len(nuevoContenido))
	bloquesNecesarios := (tamañoNuevo + sb.S_block_size - 1) / sb.S_block_size // division redondeando hacia arriba
	if tamañoNuevo == 0 {
		bloquesNecesarios = 0
	}

	if bloquesNecesarios > 12 {
		return fmt.Errorf("users.txt excede el tamaño maximo soportado (bloques directos)")
	}

	// Buscar bloques libres (primer disponible) y escribir el contenido
	bytesEscritos := int32(0)
	bloquesAsignados := int32(0)

	for b := int32(0); b < bloquesNecesarios; b++ {
		numBloqueLibre, err := BuscarPrimerBloqueLibre(disco, sb)
		if err != nil {
			return fmt.Errorf("no hay bloques libres suficientes")
		}

		// Marcar como ocupado
		if err := MarcarBitFS(disco, sb.S_bm_block_start, numBloqueLibre, '1'); err != nil {
			return err
		}

		// Escribir el pedazo correspondiente del contenido
		var bloque Structs.Fileblock
		restante := tamañoNuevo - bytesEscritos
		aEscribir := sb.S_block_size
		if restante < aEscribir {
			aEscribir = restante
		}
		copy(bloque.B_content[:], nuevoContenido[bytesEscritos:bytesEscritos+aEscribir])

		posBloque := sb.S_block_start + (numBloqueLibre * sb.S_block_size)
		if err := Herramientas.WriteObject(disco, bloque, int64(posBloque)); err != nil {
			return err
		}

		inodo.I_block[b] = numBloqueLibre
		bytesEscritos += aEscribir
		bloquesAsignados++
	}

	// Actualizar el inodo
	inodo.I_size = tamañoNuevo
	ahora := time.Now()
	copy(inodo.I_mtime[:], ahora.Format("02/01/2006 15:04"))

	if err := Herramientas.WriteObject(disco, inodo, int64(posInodo)); err != nil {
		return fmt.Errorf("no se pudo actualizar el inodo de users.txt")
	}

	// Actualizar contadores del superbloque
	sb.S_free_blocks_count = sb.S_free_blocks_count + bloquesLiberados - bloquesAsignados
	if err := Herramientas.WriteObject(disco, sb, int64(partStart)); err != nil {
		return fmt.Errorf("no se pudo actualizar el superbloque")
	}

	return nil
}

// Busca el primer bloque libre en el bitmap (estrategia first-fit a nivel de bloques)
func BuscarPrimerBloqueLibre(disco *os.File, sb Structs.Superblock) (int32, error) {
	for i := sb.S_first_blo; i < sb.S_blocks_count; i++ {
		var bit Structs.Bite
		if err := Herramientas.ReadObject(disco, &bit, int64(sb.S_bm_block_start+i)); err != nil {
			return -1, err
		}
		if bit.Val[0] == '0' {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no hay bloques libres")
}

// Marca un bit especifico del bitmap (de inodos o bloques) como ocupado o libre
func MarcarBitFS(disco *os.File, inicioBitmap int32, posicion int32, valor byte) error {
	var bit Structs.Bite
	bit.Val[0] = valor
	return Herramientas.WriteObject(disco, bit, int64(inicioBitmap+posicion))
}
