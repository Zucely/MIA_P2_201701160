package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"strings"
)

func GenerarReporteInode(pathDisco string, nombreParticion string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP INODE: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP INODE: No se pudo leer el MBR")
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
		fmt.Println("ERROR REP INODE: No se encontro la particion")
		return
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		fmt.Println("ERROR REP INODE: No se pudo leer el superbloque")
		return
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("  rankdir=LR\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")

	hayInodosUsados := false
	var nombreNodoAnterior string

	for i := int32(0); i < sb.S_inodes_count; i++ {
		// Verificar si este inodo esta ocupado segun el bitmap
		var bit Structs.Bite
		if err := Herramientas.ReadObject(disco, &bit, int64(sb.S_bm_inode_start+i)); err != nil {
			continue
		}
		if bit.Val[0] != '1' {
			continue // inodo libre, no se muestra
		}

		hayInodosUsados = true

		var inodo Structs.Inode
		posInodo := sb.S_inode_start + (i * sb.S_inode_size)
		if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
			continue
		}

		nombreNodo := fmt.Sprintf("inodo%d", i)

		dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
		dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='4'>\n")
		dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN='2' BGCOLOR='#2c3e50'><FONT COLOR='white'><B>Inodo %d</B></FONT></TD></TR>\n", i))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_uid</TD><TD>%d</TD></TR>\n", inodo.I_uid))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_gid</TD><TD>%d</TD></TR>\n", inodo.I_gid))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_size</TD><TD>%d</TD></TR>\n", inodo.I_size))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_atime</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_atime[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_ctime</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_ctime[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_mtime</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_mtime[:])))
		for b := 0; b < 15; b++ {
			dot.WriteString(fmt.Sprintf("      <TR><TD>i_block_%d</TD><TD>%d</TD></TR>\n", b+1, inodo.I_block[b]))
		}
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_type</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_type[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_perm</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_perm[:])))
		dot.WriteString("    </TABLE>\n")
		dot.WriteString("  >]\n\n")

		// Flecha desde el inodo anterior hacia este (orden de creacion)
		if nombreNodoAnterior != "" {
			dot.WriteString(fmt.Sprintf("  %s -> %s\n", nombreNodoAnterior, nombreNodo))
		}
		nombreNodoAnterior = nombreNodo
	}

	if !hayInodosUsados {
		dot.WriteString("  sininodos [label=\"No hay inodos utilizados\"]\n")
	}

	dot.WriteString("}\n")

	guardarYConvertirDot(dot.String(), pathSalida, "INODE")
}
