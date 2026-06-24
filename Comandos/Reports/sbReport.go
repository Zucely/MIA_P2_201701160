package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"strings"
)

func GenerarReporteSB(pathDisco string, nombreParticion string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP SB: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP SB: No se pudo leer el MBR")
		return
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
		fmt.Println("ERROR REP SB: No se encontro la particion")
		return
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		fmt.Println("ERROR REP SB: No se pudo leer el superbloque")
		return
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")
	dot.WriteString("  sb [label=<\n")
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='6'>\n")
	dot.WriteString("      <TR><TD COLSPAN='2' BGCOLOR='#2c3e50'><FONT COLOR='white'><B>REPORTE SUPERBLOQUE</B></FONT></TD></TR>\n")

	fila := func(nombre string, valor interface{}) string {
		return fmt.Sprintf("      <TR><TD BGCOLOR='#dfe6e9'><B>%s</B></TD><TD>%v</TD></TR>\n", nombre, valor)
	}

	dot.WriteString(fila("Filesystem Type", sb.S_filesystem_type))
	dot.WriteString(fila("Inodes Count", sb.S_inodes_count))
	dot.WriteString(fila("Blocks Count", sb.S_blocks_count))
	dot.WriteString(fila("Free Blocks Count", sb.S_free_blocks_count))
	dot.WriteString(fila("Free Inodes Count", sb.S_free_inodes_count))
	dot.WriteString(fila("Mtime", limpiarBytes(sb.S_mtime[:])))
	dot.WriteString(fila("Umtime", limpiarBytes(sb.S_umtime[:])))
	dot.WriteString(fila("Mnt Count", sb.S_mnt_count))
	dot.WriteString(fila("Magic", fmt.Sprintf("0x%X", sb.S_magic)))
	dot.WriteString(fila("Inode Size", sb.S_inode_size))
	dot.WriteString(fila("Block Size", sb.S_block_size))
	dot.WriteString(fila("First Ino", sb.S_first_ino))
	dot.WriteString(fila("First Blo", sb.S_first_blo))
	dot.WriteString(fila("Bm Inode Start", sb.S_bm_inode_start))
	dot.WriteString(fila("Bm Block Start", sb.S_bm_block_start))
	dot.WriteString(fila("Inode Start", sb.S_inode_start))
	dot.WriteString(fila("Block Start", sb.S_block_start))

	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n")
	dot.WriteString("}\n")

	guardarYConvertirDot(dot.String(), pathSalida, "SUPERBLOQUE")
}
