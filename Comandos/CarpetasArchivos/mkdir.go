package CarpetasArchivos

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
	"time"
)

func Mkdir(parametros []string) {
	if !Structs.SesionActiva.Status {
		fmt.Println("ERROR MKDIR: No hay una sesion activa")
		return
	}

	var path string
	crearPadres := false
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		clave := strings.ToLower(valores[0])

		if clave == "p" {
			if len(valores) != 1 {
				fmt.Println("ERROR MKDIR: El parametro p no recibe valor")
				return
			}
			crearPadres = true
		} else if clave == "path" {
			if len(valores) != 2 {
				fmt.Println("ERROR MKDIR: valor desconocido de parametro ", valores[0])
				return
			}
			path = strings.ReplaceAll(valores[1], "\"", "")
			path = strings.TrimSpace(path)
		} else {
			fmt.Println("ERROR MKDIR: Parametro desconocido: ", valores[0])
			paramC = false
			break
		}
	}

	if !paramC {
		return
	}
	if path == "" {
		fmt.Println("ERROR MKDIR: Falta el parametro path")
		return
	}

	pathDisco := Structs.SesionActiva.PathD
	nombreParticion := Structs.SesionActiva.NombrePart

	disco, sb, partStart, err := Tools.AbrirDiscoYSuperbloque(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR MKDIR:", err)
		return
	}
	defer disco.Close()

	path = strings.TrimPrefix(path, "/")
	partes := strings.Split(path, "/")

	if len(partes) == 0 || partes[0] == "" {
		fmt.Println("ERROR MKDIR: Ruta invalida")
		return
	}

	// Validar longitud de cada componente del path
	for _, parte := range partes {
		if len(parte) > 12 {
			fmt.Println("ERROR MKDIR: El nombre '" + parte + "' excede el maximo de 12 caracteres")
			return
		}
	}

	numInodoActual := int32(0) // empezar en la raiz
	seCreoAlgo := false

	for i, parte := range partes {
		esUltima := i == len(partes)-1

		siguiente, err := Tools.BuscarEnCarpeta(disco, sb, numInodoActual, parte)

		if err == nil {
			// ya existe esa carpeta/archivo
			if esUltima {
				fmt.Println("ERROR MKDIR: La carpeta /" + path + " ya existe")
				return
			}
			numInodoActual = siguiente
			continue
		}

		// no existe esta parte de la ruta
		if !esUltima && !crearPadres {
			fmt.Println("ERROR MKDIR: La carpeta padre no existe. Use -p para crearla")
			return
		}

		if !tienePermisoEscritura(disco, sb, numInodoActual) {
			fmt.Println("ERROR MKDIR: No tiene permiso de escritura en la carpeta padre")
			return
		}

		nuevoInodo, err := crearCarpetaNueva(disco, &sb, numInodoActual, parte)
		if err != nil {
			fmt.Println("ERROR MKDIR:", err)
			return
		}

		numInodoActual = nuevoInodo
		seCreoAlgo = true
	}

	if seCreoAlgo {
		if err := Herramientas.WriteObject(disco, sb, int64(partStart)); err != nil {
			fmt.Println("ERROR MKDIR: No se pudo actualizar el superbloque")
			return
		}
	}

	fmt.Println("MKDIR: Carpeta /" + path + " creada exitosamente")
}

// Valida si el usuario de la sesion tiene permiso de ESCRITURA sobre el inodo (carpeta) dado
func tienePermisoEscritura(disco *os.File, sb Structs.Superblock, numInodo int32) bool {
	if Structs.SesionActiva.Nombre == "root" {
		return true
	}

	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodo * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		return false
	}

	perm := limpiarBytesFS(inodo.I_perm[:])
	if len(perm) != 3 {
		return false
	}

	var digito byte
	if inodo.I_uid == Structs.SesionActiva.IdUsr {
		digito = perm[0]
	} else if inodo.I_gid == Structs.SesionActiva.IdGrp {
		digito = perm[1]
	} else {
		digito = perm[2]
	}

	valor := int(digito - '0')
	return valor&2 != 0 // bit de escritura: 2,3,6,7
}

// Crea un nuevo inodo de tipo carpeta (con su bloque . y ..) y lo agrega
// como entrada en el bloque de carpeta del padre. Retorna el numero del inodo nuevo.
func crearCarpetaNueva(disco *os.File, sb *Structs.Superblock, numInodoPadre int32, nombre string) (int32, error) {
	numInodoNuevo, err := buscarPrimerInodoLibre(disco, *sb)
	if err != nil {
		return -1, err
	}
	//fmt.Println("DEBUG: creando", nombre, "- inodo asignado:", numInodoNuevo, "- inodo padre:", numInodoPadre)

	numBloqueNuevo, err := Tools.BuscarPrimerBloqueLibre(disco, *sb)
	if err != nil {
		return -1, err
	}
	//fmt.Println("DEBUG: bloque asignado para", nombre, ":", numBloqueNuevo)

	ahora := time.Now()

	var inodoNuevo Structs.Inode
	inodoNuevo.Inicializar()
	inodoNuevo.I_uid = Structs.SesionActiva.IdUsr
	inodoNuevo.I_gid = Structs.SesionActiva.IdGrp
	inodoNuevo.I_size = int32(64)
	copy(inodoNuevo.I_atime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoNuevo.I_ctime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoNuevo.I_mtime[:], ahora.Format("02/01/2006 15:04"))
	inodoNuevo.I_block[0] = numBloqueNuevo
	copy(inodoNuevo.I_type[:], "0")
	copy(inodoNuevo.I_perm[:], "664")

	posInodoNuevo := sb.S_inode_start + (numInodoNuevo * sb.S_inode_size)
	if err := Herramientas.WriteObject(disco, inodoNuevo, int64(posInodoNuevo)); err != nil {
		return -1, fmt.Errorf("no se pudo escribir el inodo nuevo")
	}

	var bloqueNuevo Structs.Folderblock
	bloqueNuevo.Inicializar()
	copy(bloqueNuevo.B_content[0].B_name[:], ".")
	bloqueNuevo.B_content[0].B_inodo = numInodoNuevo
	copy(bloqueNuevo.B_content[1].B_name[:], "..")
	bloqueNuevo.B_content[1].B_inodo = numInodoPadre

	posBloqueNuevo := sb.S_block_start + (numBloqueNuevo * sb.S_block_size)
	if err := Herramientas.WriteObject(disco, bloqueNuevo, int64(posBloqueNuevo)); err != nil {
		return -1, fmt.Errorf("no se pudo escribir el bloque nuevo")
	}
	//fmt.Println("DEBUG: escrito bloque", numBloqueNuevo, "con . ->", numInodoNuevo, "y .. ->", numInodoPadre)

	if err := Tools.MarcarBitFS(disco, sb.S_bm_inode_start, numInodoNuevo, '1'); err != nil {
		return -1, err
	}
	if err := Tools.MarcarBitFS(disco, sb.S_bm_block_start, numBloqueNuevo, '1'); err != nil {
		return -1, err
	}
	//fmt.Println("DEBUG: marcados como ocupados - inodo", numInodoNuevo, "bloque", numBloqueNuevo)

	if err := agregarEntradaEnCarpeta(disco, sb, numInodoPadre, nombre, numInodoNuevo); err != nil {
		return -1, err
	}

	sb.S_free_inodes_count--
	sb.S_free_blocks_count--

	return numInodoNuevo, nil
}

// Busca el primer inodo libre en el bitmap de inodos
func buscarPrimerInodoLibre(disco *os.File, sb Structs.Superblock) (int32, error) {
	//
	for i := int32(0); i < sb.S_inodes_count; i++ {
		var bit Structs.Bite
		if err := Herramientas.ReadObject(disco, &bit, int64(sb.S_bm_inode_start+i)); err != nil {
			return -1, err
		}
		if bit.Val[0] == '0' {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no hay inodos libres")
}

// Agrega una entrada nueva (nombre, numInodo) en el primer espacio libre del bloque
// de carpeta del inodo dado.
// Agrega una entrada nueva (nombre, numInodo) en el primer espacio libre del bloque
// de carpeta del inodo dado. Si todos los bloques existentes estan llenos,
// crea un bloque nuevo y lo asigna al siguiente I_block libre del inodo.
func agregarEntradaEnCarpeta(disco *os.File, sb *Structs.Superblock, numInodoCarpeta int32, nombre string, numInodoNuevo int32) error {
	var inodoCarpeta Structs.Inode
	posInodoCarpeta := sb.S_inode_start + (numInodoCarpeta * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodoCarpeta, int64(posInodoCarpeta)); err != nil {
		return err
	}

	// 1. Intentar encontrar espacio en algun bloque YA existente
	for b := 0; b < 12; b++ {
		numBloque := inodoCarpeta.I_block[b]
		if numBloque == -1 {
			continue
		}

		var bloque Structs.Folderblock
		posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
		if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
			continue
		}

		for i := range bloque.B_content {
			nombreActual := Structs.GetB_name(string(bloque.B_content[i].B_name[:]))
			if nombreActual == "" || nombreActual == "-" {
				copy(bloque.B_content[i].B_name[:], nombre)
				bloque.B_content[i].B_inodo = numInodoNuevo

				if err := Herramientas.WriteObject(disco, bloque, int64(posBloque)); err != nil {
					return err
				}

				//fmt.Println("DEBUG agregarEntrada: escrito", nombre, "-> inodo:", numInodoNuevo, "en bloque:", numBloque, "posicion:", i)
				return nil
			}
		}
	}

	// 2. No hay espacio en los bloques existentes: buscar el primer I_block libre (-1)
	//    del inodo para asignarle un bloque NUEVO

	//fmt.Println("DEBUG: no hay espacio en bloques existentes de inodo padre", numInodoCarpeta, "- creando bloque nuevo para entrada:", nombre)

	indiceLibre := -1
	for b := 0; b < 12; b++ {
		if inodoCarpeta.I_block[b] == -1 {
			indiceLibre = b
			break
		}
	}

	if indiceLibre == -1 {
		return fmt.Errorf("la carpeta alcanzo el maximo de bloques directos (12)")
	}

	numBloqueNuevo, err := Tools.BuscarPrimerBloqueLibre(disco, *sb)
	if err != nil {
		return err
	}
	//fmt.Println("DEBUG: nuevo bloque para el padre", numInodoCarpeta, ":", numBloqueNuevo, "- contendra entrada:", nombre, "->", numInodoNuevo)

	// Crear el nuevo Folderblock con la entrada solicitada
	var bloqueNuevo Structs.Folderblock
	bloqueNuevo.Inicializar()
	copy(bloqueNuevo.B_content[0].B_name[:], nombre)
	bloqueNuevo.B_content[0].B_inodo = numInodoNuevo

	posBloqueNuevo := sb.S_block_start + (numBloqueNuevo * sb.S_block_size)
	if err := Herramientas.WriteObject(disco, bloqueNuevo, int64(posBloqueNuevo)); err != nil {
		return err
	}

	// Marcar el bloque como ocupado en el bitmap
	if err := Tools.MarcarBitFS(disco, sb.S_bm_block_start, numBloqueNuevo, '1'); err != nil {
		return err
	}
	sb.S_first_blo = numBloqueNuevo + 1 // actualizar el primer bloque libre en el superbloque

	// Actualizar el inodo de la carpeta padre con el nuevo I_block
	inodoCarpeta.I_block[indiceLibre] = numBloqueNuevo
	if err := Herramientas.WriteObject(disco, inodoCarpeta, int64(posInodoCarpeta)); err != nil {
		return err
	}

	// Reflejar el bloque gastado en el superbloque (ahora si se propaga, porque sb es puntero)
	sb.S_free_blocks_count--

	//fmt.Println("DEBUG agregarEntrada: bloque nuevo", numBloqueNuevo, "creado con entrada:", nombre, "-> inodo:", numInodoNuevo)

	//fmt.Println("DEBUG agregarEntrada: creando bloque nuevo", numBloqueNuevo, "para entrada:", nombre, "-> inodo:", numInodoNuevo)

	return nil
}

// Quita los bytes nulos (\x00) que quedan al final de los campos [N]byte
func limpiarBytesFS(b []byte) string {
	s := string(b)
	if pos := strings.IndexByte(s, 0); pos != -1 {
		s = s[:pos]
	}
	return strings.TrimSpace(s)
}
