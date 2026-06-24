package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
)

func GenerarReporteBlock(pathDisco string, nombreParticion string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP BLOCK: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP BLOCK: No se pudo leer el MBR")
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
		fmt.Println("ERROR REP BLOCK: No se encontro la particion")
		return
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		fmt.Println("ERROR REP BLOCK: No se pudo leer el superbloque")
		return
	}

	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("  rankdir=LR\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")

	visitados := make(map[int32]bool)
	hayBloques := false
	var nombreNodoAnterior string // 👈 nuevo: para encadenar flechas

	for i := int32(0); i < sb.S_inodes_count; i++ {
		var bit Structs.Bite
		if err := Herramientas.ReadObject(disco, &bit, int64(sb.S_bm_inode_start+i)); err != nil {
			continue
		}
		if bit.Val[0] != '1' {
			continue
		}

		var inodo Structs.Inode
		posInodo := sb.S_inode_start + (i * sb.S_inode_size)
		if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
			continue
		}

		esCarpeta := limpiarBytes(inodo.I_type[:]) == "0"

		for b := 0; b < 12; b++ {
			numBloque := inodo.I_block[b]
			if numBloque == -1 || visitados[numBloque] {
				continue
			}
			visitados[numBloque] = true
			hayBloques = true

			posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
			nombreNodo := fmt.Sprintf("bloque%d", numBloque)

			if esCarpeta {
				escribirBloqueCarpeta(&dot, disco, numBloque, posBloque)
			} else {
				escribirBloqueArchivo(&dot, disco, numBloque, posBloque)
			}

			// 👈 nuevo: conectar con el bloque anterior
			if nombreNodoAnterior != "" {
				dot.WriteString(fmt.Sprintf("  %s -> %s\n", nombreNodoAnterior, nombreNodo))
			}
			nombreNodoAnterior = nombreNodo
		}

		for b := 12; b < 15; b++ {
			numBloque := inodo.I_block[b]
			if numBloque == -1 || visitados[numBloque] {
				continue
			}
			visitados[numBloque] = true
			hayBloques = true

			posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
			nombreNodo := fmt.Sprintf("bloque%d", numBloque)

			escribirBloqueApuntadores(&dot, disco, numBloque, posBloque)

			// 👈 nuevo: conectar con el bloque anterior
			if nombreNodoAnterior != "" {
				dot.WriteString(fmt.Sprintf("  %s -> %s\n", nombreNodoAnterior, nombreNodo))
			}
			nombreNodoAnterior = nombreNodo
		}
	}

	if !hayBloques {
		dot.WriteString("  sinbloques [label=\"No hay bloques utilizados\"]\n")
	}

	dot.WriteString("}\n")

	guardarYConvertirDot(dot.String(), pathSalida, "BLOCK")
}
func escribirBloqueCarpeta(dot *strings.Builder, disco *os.File, numBloque int32, posBloque int32) {
	var bloque Structs.Folderblock
	if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
		return
	}

	nombreNodo := fmt.Sprintf("bloque%d", numBloque)
	dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='4'>\n")
	dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN='2' BGCOLOR='#27ae60'><FONT COLOR='white'><B>Bloque Carpeta %d</B></FONT></TD></TR>\n", numBloque))
	dot.WriteString("      <TR><TD><B>b_name</B></TD><TD><B>b_inodo</B></TD></TR>\n")
	for _, c := range bloque.B_content {
		nombre := Structs.GetB_name(string(c.B_name[:]))
		if nombre == "" {
			continue
		}
		dot.WriteString(fmt.Sprintf("      <TR><TD>%s</TD><TD>%d</TD></TR>\n", nombre, c.B_inodo))
	}
	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n\n")
}

func escribirBloqueArchivo(dot *strings.Builder, disco *os.File, numBloque int32, posBloque int32) {
	var bloque Structs.Fileblock
	if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
		return
	}

	contenido := Structs.GetB_content(string(bloque.B_content[:]))

	nombreNodo := fmt.Sprintf("bloque%d", numBloque)
	dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='4'>\n")
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#e67e22'><FONT COLOR='white'><B>Bloque Archivo %d</B></FONT></TD></TR>\n", numBloque))
	dot.WriteString(fmt.Sprintf("      <TR><TD>%s</TD></TR>\n", contenido))
	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n\n")
}

func escribirBloqueApuntadores(dot *strings.Builder, disco *os.File, numBloque int32, posBloque int32) {
	var bloque Structs.Pointerblock
	if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
		return
	}

	nombreNodo := fmt.Sprintf("bloque%d", numBloque)
	dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='4'>\n")
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#8e44ad'><FONT COLOR='white'><B>Bloque Apuntadores %d</B></FONT></TD></TR>\n", numBloque))
	for _, p := range bloque.B_pointers {
		dot.WriteString(fmt.Sprintf("      <TR><TD>%d</TD></TR>\n", p))
	}
	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n\n")
}
