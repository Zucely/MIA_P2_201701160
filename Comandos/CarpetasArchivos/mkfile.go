package CarpetasArchivos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Mkfile(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR MKFILE: No hay una sesion activa")
		return
	}

	var path, contPath string
	size := 0
	sizeEspecificado := false
	crearPadres := false
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		clave := strings.ToLower(valores[0])

		switch clave {
		case "r":
			if len(valores) != 1 {
				fmt.Println("ERROR MKFILE: El parametro r no recibe valor")
				return
			}
			crearPadres = true
		case "path":
			if len(valores) != 2 {
				fmt.Println("ERROR MKFILE: valor desconocido de parametro ", valores[0])
				return
			}
			path = strings.ReplaceAll(valores[1], "\"", "")
			path = strings.TrimSpace(path)
		case "size":
			if len(valores) != 2 {
				fmt.Println("ERROR MKFILE: valor desconocido de parametro ", valores[0])
				return
			}
			var err error
			size, err = strconv.Atoi(valores[1])
			if err != nil || size < 0 {
				fmt.Println("ERROR MKFILE: El parametro size debe ser un entero mayor o igual a 0")
				return
			}
			sizeEspecificado = true
		case "cont":
			if len(valores) != 2 {
				fmt.Println("ERROR MKFILE: valor desconocido de parametro ", valores[0])
				return
			}
			contPath = strings.ReplaceAll(valores[1], "\"", "")
			contPath = strings.TrimSpace(contPath)
		default:
			fmt.Println("ERROR MKFILE: Parametro desconocido: ", valores[0])
			paramC = false
		}
	}

	if !paramC {
		return
	}
	if path == "" {
		fmt.Println("ERROR MKFILE: Falta el parametro path")
		return
	}

	// Determinar el contenido final del archivo (cont tiene prioridad sobre size)
	var contenido string
	if contPath != "" {
		data, err := os.ReadFile(contPath)
		if err != nil {
			fmt.Println("ERROR MKFILE: No se pudo leer el archivo de contenido:", contPath)
			return
		}
		contenido = string(data)
	} else if sizeEspecificado {
		contenido = generarContenidoNumerico(size)
	}
	// si ninguno se especifico, contenido se queda en "" (0 bytes)

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	disco, sb, partStart, err := Tools.AbrirDiscoYSuperbloque(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR MKFILE:", err)
		return
	}
	defer disco.Close()

	pathLimpio := strings.TrimPrefix(path, "/")
	partes := strings.Split(pathLimpio, "/")

	if len(partes) == 0 || partes[0] == "" {
		fmt.Println("ERROR MKFILE: Ruta invalida")
		return
	}

	nombreArchivo := partes[len(partes)-1]
	partesPadre := partes[:len(partes)-1]

	if len(nombreArchivo) > 12 {
		fmt.Println("ERROR MKFILE: El nombre del archivo excede el maximo de 12 caracteres")
		return
	}

	for _, parte := range partesPadre {
		if len(parte) > 12 {
			fmt.Println("ERROR MKFILE: El nombre '" + parte + "' excede el maximo de 12 caracteres")
			return
		}
	}

	// Navegar/crear las carpetas padre (igual que en mkdir, pero solo para los componentes intermedios)
	numInodoActual := int32(0)
	seCreoAlgo := false

	for _, parte := range partesPadre {
		siguiente, err := Tools.BuscarEnCarpeta(disco, sb, numInodoActual, parte)

		if err == nil {
			numInodoActual = siguiente
			continue
		}

		if !crearPadres {
			fmt.Println("ERROR MKFILE: La carpeta padre no existe. Use -r para crearla")
			return
		}

		if !tienePermisoEscritura(disco, sb, numInodoActual) {
			fmt.Println("ERROR MKFILE: No tiene permiso de escritura en la carpeta padre")
			return
		}

		nuevoInodo, err := crearCarpetaNueva(disco, &sb, numInodoActual, parte)
		if err != nil {
			fmt.Println("ERROR MKFILE:", err)
			return
		}
		numInodoActual = nuevoInodo
		seCreoAlgo = true
	}

	// Verificar si el archivo ya existe
	numInodoExistente, errBusqueda := Tools.BuscarEnCarpeta(disco, sb, numInodoActual, nombreArchivo)

	if errBusqueda == nil {
		// Ya existe: preguntar si se desea sobreescribir
		fmt.Printf("El archivo %s ya existe. ¿Desea sobreescribirlo? (Y/N): ", path)
		reader := bufio.NewScanner(os.Stdin)
		reader.Scan()
		respuesta := strings.ToUpper(strings.TrimSpace(reader.Text()))

		if respuesta != "Y" {
			fmt.Println("MKFILE: Operacion cancelada por el usuario")
			return
		}

		if err := sobreescribirArchivo(disco, &sb, numInodoExistente, contenido); err != nil {
			fmt.Println("ERROR MKFILE:", err)
			return
		}

		if err := Herramientas.WriteObject(disco, sb, int64(partStart)); err != nil {
			fmt.Println("ERROR MKFILE: No se pudo actualizar el superbloque")
			return
		}

		fmt.Println("MKFILE: Archivo", path, "sobreescrito exitosamente")
		return
	}

	// No existe: validar permiso de escritura en la carpeta padre inmediata
	if !tienePermisoEscritura(disco, sb, numInodoActual) {
		fmt.Println("ERROR MKFILE: No tiene permiso de escritura en la carpeta padre")
		return
	}

	if err := crearArchivoNuevo(disco, &sb, numInodoActual, nombreArchivo, contenido); err != nil {
		fmt.Println("ERROR MKFILE:", err)
		return
	}
	seCreoAlgo = true
	//fmt.Println("DEBUG MKFILE: archivo creado, free_blocks:", sb.S_free_blocks_count, "free_inodes:", sb.S_free_inodes_count)

	if seCreoAlgo {
		if err := Herramientas.WriteObject(disco, sb, int64(partStart)); err != nil {
			fmt.Println("ERROR MKFILE: No se pudo actualizar el superbloque")
			return
		}
	}

	fmt.Println("MKFILE: Archivo", path, "creado exitosamente")
}

// Genera contenido con digitos 0-9 repetidos ciclicamente hasta llegar al tamaño indicado
func generarContenidoNumerico(size int) string {
	var sb strings.Builder
	for i := 0; i < size; i++ {
		sb.WriteByte(byte('0' + (i % 10)))
	}
	return sb.String()
}

// Crea un inodo nuevo de tipo ARCHIVO, le asigna los bloques necesarios para el contenido,
// y agrega la entrada correspondiente en la carpeta padre.
func crearArchivoNuevo(disco *os.File, sb *Structs.Superblock, numInodoPadre int32, nombre string, contenido string) error {
	numInodoNuevo, err := buscarPrimerInodoLibre(disco, *sb)
	if err != nil {
		return err
	}

	tamañoContenido := int32(len(contenido))
	bloquesNecesarios := int32(0)
	if tamañoContenido > 0 {
		bloquesNecesarios = (tamañoContenido + sb.S_block_size - 1) / sb.S_block_size
	}
	if bloquesNecesarios > 12 {
		return fmt.Errorf("el archivo excede el tamaño maximo soportado (bloques directos)")
	}

	ahora := time.Now()

	var inodoNuevo Structs.Inode
	inodoNuevo.Inicializar()
	inodoNuevo.I_uid = Structs.SesionActiva.IdUsr
	inodoNuevo.I_gid = Structs.SesionActiva.IdGrp
	inodoNuevo.I_size = tamañoContenido
	copy(inodoNuevo.I_atime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoNuevo.I_ctime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoNuevo.I_mtime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoNuevo.I_type[:], "1") // archivo
	copy(inodoNuevo.I_perm[:], "664")

	bytesEscritos := int32(0)
	for b := int32(0); b < bloquesNecesarios; b++ {
		numBloqueNuevo, err := Tools.BuscarPrimerBloqueLibre(disco, *sb)
		if err != nil {
			return err
		}

		var bloque Structs.Fileblock
		restante := tamañoContenido - bytesEscritos
		aEscribir := sb.S_block_size
		if restante < aEscribir {
			aEscribir = restante
		}
		copy(bloque.B_content[:], contenido[bytesEscritos:bytesEscritos+aEscribir])

		posBloque := sb.S_block_start + (numBloqueNuevo * sb.S_block_size)
		if err := Herramientas.WriteObject(disco, bloque, int64(posBloque)); err != nil {
			return err
		}

		if err := Tools.MarcarBitFS(disco, sb.S_bm_block_start, numBloqueNuevo, '1'); err != nil {
			return err
		}
		sb.S_first_blo = numBloqueNuevo + 1
		inodoNuevo.I_block[b] = numBloqueNuevo
		bytesEscritos += aEscribir
		sb.S_free_blocks_count--
	}

	posInodoNuevo := sb.S_inode_start + (numInodoNuevo * sb.S_inode_size)
	if err := Herramientas.WriteObject(disco, inodoNuevo, int64(posInodoNuevo)); err != nil {
		return fmt.Errorf("no se pudo escribir el inodo nuevo")
	}

	if err := Tools.MarcarBitFS(disco, sb.S_bm_inode_start, numInodoNuevo, '1'); err != nil {
		return err
	}
	sb.S_free_inodes_count--

	if err := agregarEntradaEnCarpeta(disco, sb, numInodoPadre, nombre, numInodoNuevo); err != nil {

		return err
	}

	return nil
}

// Sobrescribe el contenido de un archivo existente, liberando sus bloques viejos
// y asignando nuevos bloques segun el contenido actualizado.
func sobreescribirArchivo(disco *os.File, sb *Structs.Superblock, numInodo int32, nuevoContenido string) error {
	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodo * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		return err
	}

	// Liberar bloques viejos
	for b := 0; b < 12; b++ {
		if inodo.I_block[b] != -1 {
			if err := Tools.MarcarBitFS(disco, sb.S_bm_block_start, inodo.I_block[b], '0'); err != nil {
				return err
			}
			inodo.I_block[b] = -1
			sb.S_free_blocks_count++
		}
	}

	tamañoNuevo := int32(len(nuevoContenido))
	bloquesNecesarios := int32(0)
	if tamañoNuevo > 0 {
		bloquesNecesarios = (tamañoNuevo + sb.S_block_size - 1) / sb.S_block_size
	}
	if bloquesNecesarios > 12 {
		return fmt.Errorf("el archivo excede el tamaño maximo soportado")
	}

	bytesEscritos := int32(0)
	for b := int32(0); b < bloquesNecesarios; b++ {
		numBloqueNuevo, err := Tools.BuscarPrimerBloqueLibre(disco, *sb)
		if err != nil {
			return err
		}

		var bloque Structs.Fileblock
		restante := tamañoNuevo - bytesEscritos
		aEscribir := sb.S_block_size
		if restante < aEscribir {
			aEscribir = restante
		}
		copy(bloque.B_content[:], nuevoContenido[bytesEscritos:bytesEscritos+aEscribir])

		posBloque := sb.S_block_start + (numBloqueNuevo * sb.S_block_size)
		if err := Herramientas.WriteObject(disco, bloque, int64(posBloque)); err != nil {
			return err
		}

		if err := Tools.MarcarBitFS(disco, sb.S_bm_block_start, numBloqueNuevo, '1'); err != nil {
			return err
		}

		inodo.I_block[b] = numBloqueNuevo
		bytesEscritos += aEscribir
		sb.S_free_blocks_count--
	}

	inodo.I_size = tamañoNuevo
	ahora := time.Now()
	copy(inodo.I_mtime[:], ahora.Format("02/01/2006 15:04"))

	if err := Herramientas.WriteObject(disco, inodo, int64(posInodo)); err != nil {
		return fmt.Errorf("no se pudo actualizar el inodo")
	}

	return nil
}
