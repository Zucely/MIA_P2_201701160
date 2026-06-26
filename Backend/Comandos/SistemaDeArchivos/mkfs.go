package SistemaDeArchivos

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

func Mkfs(parametros []string) {
	var id string
	tipo := "full" // valor por defecto si no viene -type
	fs := "2fs"    // valor por defecto: ext2
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimSpace(parametro)
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR MKFS: valor desconocido de parametro ", valores[0])
			return
		}

		if strings.ToLower(valores[0]) == "id" {
			id = strings.ReplaceAll(valores[1], "\"", "")
			id = strings.TrimSpace(id)
			id = strings.ToUpper(id)

		} else if strings.ToLower(valores[0]) == "type" {
			tipo = strings.ToLower(strings.TrimSpace(valores[1]))
		} else {
			fmt.Println("ERROR MKFS: Parametro desconocido: ", valores[0])
			paramC = false
			break
		}
	}

	if !paramC {
		return
	}
	if id == "" {
		fmt.Println("ERROR MKFS: Falta parametro ID")
		return
	}

	// Buscar la particion montada por su id
	idx := Structs.BuscarMontadaPorId(id)
	if idx == -1 {
		fmt.Println("ERROR MKFS: No existe una particion montada con el id ", id)
		return
	}
	pathDisco := Structs.Montadas[idx].PathM
	nombreParticion := Structs.Montadas[idx].Name

	// Abrir el disco
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR MKFS: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR MKFS: No se pudo leer el MBR")
		return
	}

	// Buscar la particion (primaria) por nombre para obtener Start y Size
	var partStart, partSize int32 = -1, -1
	for i := 0; i < 4; i++ {
		nombre := Structs.GetName(string(mbr.Partitions[i].Name[:]))
		if nombre == nombreParticion {
			partStart = mbr.Partitions[i].Start
			partSize = mbr.Partitions[i].Size
			break
		}
	}

	if partStart == -1 {
		fmt.Println("ERROR MKFS: No se encontro la particion ", nombreParticion, " en el disco")
		return
	}

	fmt.Println("Formateando particion", nombreParticion, "como", fs, "(type:", tipo, ")")

	// ----------- Calcular estructuras -----------
	n, numBloques, sizeSB, sizeInodo, sizeBlock := calcularEstructurasFS(partSize)
	if n <= 2 {
		fmt.Println("ERROR MKFS: La particion es demasiado pequeña para crear el sistema de archivos")
		return
	}

	bmInodeStart, bmBlockStart, inodeStart, blockStart := calcularPosiciones(partStart, n, sizeSB, sizeInodo, sizeBlock)

	ahora := time.Now()

	// ----------- Construir y escribir el Superbloque -----------
	sb := construirSuperbloque(n, numBloques, sizeSB, sizeInodo, sizeBlock, bmInodeStart, bmBlockStart, inodeStart, blockStart, ahora)

	if err := Herramientas.WriteObject(disco, sb, int64(partStart)); err != nil {
		fmt.Println("ERROR MKFS: No se pudo escribir el superbloque")
		return
	}

	// ----------- Inicializar bitmaps en '0' (libre) -----------
	if err := inicializarBitmap(disco, bmInodeStart, n); err != nil {
		fmt.Println("ERROR MKFS: No se pudo inicializar el bitmap de inodos")
		return
	}
	if err := inicializarBitmap(disco, bmBlockStart, numBloques); err != nil {
		fmt.Println("ERROR MKFS: No se pudo inicializar el bitmap de bloques")
		return
	}

	// ----------- Crear inodo 0 (carpeta raiz) -----------
	var inodoRaiz Structs.Inode
	// Inicializar el inodo raiz con valores por defecto
	//-1 ->no esta en uso
	// Todos los I_block en -1, excepto el primero que apunta al bloque 0 (carpeta raiz)
	inodoRaiz.Inicializar()
	// El propietario del inodo raiz es el usuario root (UID=1, GID=1)
	inodoRaiz.I_uid = 1
	inodoRaiz.I_gid = 1
	// Tamaño del archivo (carpeta raiz) en bytes, que es el tamaño de un bloque de carpeta, 64 bytes
	inodoRaiz.I_size = int32(binary.Size(Structs.Folderblock{}))
	copy(inodoRaiz.I_atime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoRaiz.I_ctime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoRaiz.I_mtime[:], ahora.Format("02/01/2006 15:04"))
	// Se sobrescribe el primer bloque del inodo raiz con el bloque de carpeta raiz, que contiene los apuntadores a los archivos
	// y carpetas dentro de la carpeta raiz. Ahora la posicion 0, antes en -1, apunta al bloque 0 (bloque carpeta de la raiz)
	inodoRaiz.I_block[0] = 0         // apunta al bloque 0 (bloque carpeta de la raiz)
	copy(inodoRaiz.I_type[:], "0")   // Tipo 0 -> carpeta
	copy(inodoRaiz.I_perm[:], "664") // Permisos UGO en octal, ej "664"

	if err := Herramientas.WriteObject(disco, inodoRaiz, int64(inodeStart)); err != nil {
		fmt.Println("ERROR MKFS: No se pudo escribir el inodo raiz")
		return
	}

	// ----------- Crear inodo 1 (archivo users.txt) -----------
	contenidoUsers := "1,G,root\n1,U,root,root,123\n"

	var inodoUsers Structs.Inode
	inodoUsers.Inicializar() //Todos los apuntadores en -1, excepto el primero que apunta al bloque 1 (archivo users.txt)
	inodoUsers.I_uid = 1     // Propietario del archivo users.txt es el usuario root (UID=1, GID=1)
	inodoUsers.I_gid = 1
	inodoUsers.I_size = int32(len(contenidoUsers))
	copy(inodoUsers.I_atime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoUsers.I_ctime[:], ahora.Format("02/01/2006 15:04"))
	copy(inodoUsers.I_mtime[:], ahora.Format("02/01/2006 15:04"))
	inodoUsers.I_block[0] = 1       // apunta al bloque 1 (bloque archivo)
	copy(inodoUsers.I_type[:], "1") // archivo
	copy(inodoUsers.I_perm[:], "664")

	if err := Herramientas.WriteObject(disco, inodoUsers, int64(inodeStart+sizeInodo)); err != nil {
		fmt.Println("ERROR MKFS: No se pudo escribir el inodo de users.txt")
		return
	}

	// ----------- Crear bloque 0: carpeta raiz -----------
	var bloqueRaiz Structs.Folderblock
	bloqueRaiz.Inicializar()
	copy(bloqueRaiz.B_content[0].B_name[:], ".")         // se copia el nombre "." en la primera posicion del bloque de carpeta raiz
	bloqueRaiz.B_content[0].B_inodo = 0                  // el primer apuntador del bloque de carpeta raiz apunta al inodo 0 (carpeta raiz)
	copy(bloqueRaiz.B_content[1].B_name[:], "..")        // se copia el nombre ".." en la segunda posicion del bloque de carpeta raiz
	bloqueRaiz.B_content[1].B_inodo = 0                  // el segundo apuntador del bloque de carpeta raiz apunta al inodo 0 (carpeta raiz)
	copy(bloqueRaiz.B_content[2].B_name[:], "users.txt") // se copia el nombre "users.txt" en la tercera posicion del bloque de carpeta raiz
	bloqueRaiz.B_content[2].B_inodo = 1                  // el tercer apuntador del bloque de carpeta raiz apunta al inodo 1 (archivo users.txt)

	if err := Herramientas.WriteObject(disco, bloqueRaiz, int64(blockStart)); err != nil {
		fmt.Println("ERROR MKFS: No se pudo escribir el bloque de la carpeta raiz")
		return
	}

	// ----------- Crear bloque 1: archivo users.txt -----------
	var bloqueUsers Structs.Fileblock
	copy(bloqueUsers.B_content[:], contenidoUsers)

	if err := Herramientas.WriteObject(disco, bloqueUsers, int64(blockStart+sizeBlock)); err != nil {
		fmt.Println("ERROR MKFS: No se pudo escribir el bloque de users.txt")
		return
	}

	// ----------- Marcar inodos 0 y 1 como ocupados en el bitmap -----------
	marcarBit(disco, bmInodeStart, 0, '1')
	marcarBit(disco, bmInodeStart, 1, '1')

	// ----------- Marcar bloques 0 y 1 como ocupados en el bitmap -----------
	marcarBit(disco, bmBlockStart, 0, '1')
	marcarBit(disco, bmBlockStart, 1, '1')

	fmt.Println("MKFS: Particion", nombreParticion, "formateada exitosamente como EXT2")
	fmt.Println("Inodos:", n, "| Bloques:", numBloques)
}

// FUNCIONES AUXILIARES

func calcularEstructurasFS(tamañoParticion int32) (n int32, numBloques int32, sizeSB int32, sizeInodo int32, sizeBlock int32) {
	//tamañoParticion -> tamaño de la particion en bytes
	//sizeSB -> tamaño del superbloque en bytes
	//sizeInodo -> tamaño de un inodo en bytes
	//sizeBlock -> tamaño de un bloque en bytes
	//n -> cantidad de inodos y bloques que se pueden crear en la particion

	//tamañoParticion = sizeSB + n + 3*n + n*sizeInodo + 3*n*sizeBlock
	//tamañoParticion = sizeSB + n*(1 + 3 + sizeInodo + 3*sizeBlock)
	//n = (tamañoParticion - sizeSB) / (1 + 3 + sizeInodo + 3*sizeBlock)
	sizeSB = int32(binary.Size(Structs.Superblock{}))
	sizeInodo = int32(binary.Size(Structs.Inode{}))
	sizeBlock = int32(binary.Size(Structs.Fileblock{}))

	numerador := float64(tamañoParticion - sizeSB)
	denominador := float64(1 + 3 + sizeInodo + 3*sizeBlock)

	n = int32(math.Floor(numerador / denominador))
	numBloques = 3 * n

	return n, numBloques, sizeSB, sizeInodo, sizeBlock
}

// Calcula las posiciones de los bitmaps, tabla de inodos y tabla de bloques
//
//	| Inicio Particion
//	| Superbloque | Bitmap Inodos | Bitmap Bloques | Tabla Inodos | Tabla Bloques |
func calcularPosiciones(startParticion int32, n int32, sizeSB int32, sizeInodo int32, sizeBlock int32) (bmInodeStart, bmBlockStart, inodeStart, blockStart int32) {
	numBloques := 3 * n

	bmInodeStart = startParticion + sizeSB //suma la cantidad desde donde inicia la particion hasta el final del superbloque
	bmBlockStart = bmInodeStart + n        //suma la cantidad desde donde inicia el bitmap de inodos hasta el final del bitmap de inodos (cantidad de inodos)
	inodeStart = bmBlockStart + numBloques
	blockStart = inodeStart + (n * sizeInodo)

	return bmInodeStart, bmBlockStart, inodeStart, blockStart
}

func construirSuperbloque(n int32, numBloques int32, sizeSB int32, sizeInodo int32, sizeBlock int32, bmInodeStart int32, bmBlockStart int32, inodeStart int32, blockStart int32, ahora time.Time) Structs.Superblock {
	var sb Structs.Superblock

	sb.S_filesystem_type = 2
	sb.S_inodes_count = n
	sb.S_blocks_count = numBloques
	//inodo 0 -> carpeta raiz
	//inodo 1 -> archivo users.txt
	sb.S_free_blocks_count = numBloques - 2
	//bloque 0 -> contenido carpeta raiz
	//bloque 1 -> contenido archivo users.txt
	sb.S_free_inodes_count = n - 2
	copy(sb.S_mtime[:], ahora.Format("02/01/2006 15:04"))
	//cuantas veces se ha montado el sistema de archivos, al formatear es la primera vez
	sb.S_mnt_count = 1
	//magic number para identificar el sistema de archivos ext2
	sb.S_magic = 0xEF53
	sb.S_inode_size = sizeInodo
	sb.S_block_size = sizeBlock
	//primer inodo libre (inodo 0 y 1 ya ocupados)
	sb.S_first_ino = 2
	//primer bloque libre (bloque 0 y 1 ya ocupados)
	sb.S_first_blo = 2
	//inicio del bitmap de inodos
	sb.S_bm_inode_start = bmInodeStart
	//inicio del bitmap de bloques
	sb.S_bm_block_start = bmBlockStart
	//inicio de la tabla de inodos
	sb.S_inode_start = inodeStart
	//inicio de la tabla de bloques
	sb.S_block_start = blockStart

	return sb
}

// Escribe '0' (libre) en cada posicion del bitmap, desde "inicio" hasta "inicio+cantidad-1"
// (disco, bmInodeStart, n)
func inicializarBitmap(disco *os.File, inicio int32, cantidad int32) error {
	for i := int32(0); i < cantidad; i++ {
		var bit Structs.Bite
		//al inicializar el bitmap, todas las posiciones se marcan como libres ('0')
		bit.Val[0] = '0'
		//disco -> archivo donde se va a escribir
		//bit -> valor que se va a escribir
		//inicio+i -> posicion donde se va a escribir el bit (inicio al princpio la posicion del bitmap, i va aumentando hasta cantidad-1)
		if err := Herramientas.WriteObject(disco, bit, int64(inicio+i)); err != nil {
			return err
		}
	}
	return nil
}

// Marca una posicion especifica del bitmap como ocupada ('1') o libre ('0')
// marcarBit(disco, bmInodeStart, 0, '1')
// marcarBit(disco, bmInodeStart, 1, '1')
func marcarBit(disco *os.File, inicioBitmap int32, posicion int32, valor byte) {
	var bit Structs.Bite
	bit.Val[0] = valor
	Herramientas.WriteObject(disco, bit, int64(inicioBitmap+posicion))
}
