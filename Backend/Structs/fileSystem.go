package Structs

import (
	"strings"
)

// SISTEMA DE ARCHIVOS EXT2

// SUPERBLOQUE
// Se escribe 1 sola vez en la particion, al inicio del espacio destinado al FS
type Superblock struct {
	S_filesystem_type   int32    //0->no formateada; 2->ext2; 3->ext3
	S_inodes_count      int32    //numero total de inodos
	S_blocks_count      int32    //numero total de bloques
	S_free_blocks_count int32    //numero de bloques libres
	S_free_inodes_count int32    //numero de inodos libres
	S_mtime             [16]byte //ultima fecha en que el sistema fue montado "02/01/2006 15:04"
	S_umtime            [16]byte //ultima fecha en que el sistema fue desmontado
	S_mnt_count         int32    //numero de veces que se ha montado el sistema
	S_magic             int32    //0xEF53
	S_inode_size        int32    //tamaño de la estructura inodo
	S_block_size        int32    //tamaño de la estructura bloque (64)
	S_first_ino         int32    //primer inodo libre
	S_first_blo         int32    //primer bloque libre
	S_bm_inode_start    int32    //inicio del bitmap de inodos
	S_bm_block_start    int32    //inicio del bitmap de bloques
	S_inode_start       int32    //inicio de la tabla de inodos
	S_block_start       int32    //inicio de la tabla de bloques
}

// INODO
type Inode struct {
	I_uid   int32     //UID del propietario
	I_gid   int32     //GID del grupo
	I_size  int32     //tamaño del archivo en bytes
	I_atime [16]byte  //ultima fecha de lectura sin modificar
	I_ctime [16]byte  //fecha de creacion
	I_mtime [16]byte  //ultima fecha de modificacion
	I_block [15]int32 //12 directos + simple/doble/triple indirecto. -1 si no se usa
	I_type  [1]byte   //"1"=archivo, "0"=carpeta
	I_perm  [3]byte   //permisos UGO en octal, ej "664"
}

// Inicializa un inodo nuevo con valores por defecto (I_block en -1)
func (i *Inode) Inicializar() {
	for idx := range i.I_block {
		i.I_block[idx] = -1
	}
}

// BLOQUES (todos deben pesar 64 bytes)

// BLOQUE DE CARPETAS
// tamaño: 4 * (12+4) = 64 bytes
type Folderblock struct {
	B_content [4]Content
}

type Content struct {
	B_name  [12]byte //nombre de carpeta/archivo
	B_inodo int32    //apuntador al inodo asociado (-1 si no se usa)
}

// Inicializa un bloque de carpeta con sus 4 posiciones vacias (-1)
func (f *Folderblock) Inicializar() {
	for idx := range f.B_content {
		f.B_content[idx].B_inodo = -1
	}
}

func GetB_name(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)
	if posicionNulo != -1 {
		if posicionNulo != 0 {
			nombre = nombre[:posicionNulo]
		} else {
			nombre = "-"
		}
	}
	return nombre
}

// BLOQUE DE ARCHIVOS
// tamaño: 64 bytes
type Fileblock struct {
	B_content [64]byte
}

func GetB_content(nombre string) string {
	nombre = strings.ReplaceAll(nombre, "\n", "<br/>")
	posicionNulo := strings.IndexByte(nombre, 0)
	if posicionNulo != -1 {
		if posicionNulo != 0 {
			nombre = nombre[:posicionNulo]
		} else {
			nombre = "-"
		}
	}
	return nombre
}

// BLOQUE DE APUNTADORES INDIRECTOS
// tamaño: 16 * 4 = 64 bytes
type Pointerblock struct {
	B_pointers [16]int32
}

// Inicializa un bloque de apuntadores con sus 16 posiciones vacias (-1)
func (p *Pointerblock) Inicializar() {
	for idx := range p.B_pointers {
		p.B_pointers[idx] = -1
	}
}

// AUXILIAR PARA BITMAPS (1 byte por posicion, '0'=libre, '1'=ocupado)

type Bite struct {
	Val [1]byte
}
