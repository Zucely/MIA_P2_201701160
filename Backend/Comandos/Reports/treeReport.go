package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
)

func GenerarReporteTree(pathDisco string, nombreParticion string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP TREE: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP TREE: No se pudo leer el MBR")
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
		fmt.Println("ERROR REP TREE: No se encontro la particion")
		return
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		fmt.Println("ERROR REP TREE: No se pudo leer el superbloque")
		return
	}

	var dot strings.Builder
	var bloquesDot strings.Builder
	var flechas strings.Builder
	bloquesDibujados := make(map[int32]bool)

	dot.WriteString("digraph G {\n")
	dot.WriteString("  rankdir=LR\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")

	hayInodosUsados := false

	for i := int32(0); i < sb.S_inodes_count; i++ {
		var bit Structs.Bite
		if err := Herramientas.ReadObject(disco, &bit, int64(sb.S_bm_inode_start+i)); err != nil {
			continue
		}
		if bit.Val[0] != '1' {
			continue
		}
		hayInodosUsados = true

		var inodo Structs.Inode
		posInodo := sb.S_inode_start + (i * sb.S_inode_size)
		if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
			continue
		}

		tipo := limpiarBytes(inodo.I_type[:])
		nombreNodoInodo := fmt.Sprintf("inodo%d", i)

		dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodoInodo))
		dot.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n")
		dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\" BGCOLOR=\"#e8a0a0\"><B>inodo %d</B></TD></TR>\n", i))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_uid</TD><TD>%d</TD></TR>\n", inodo.I_uid))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_gid</TD><TD>%d</TD></TR>\n", inodo.I_gid))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_size</TD><TD>%d</TD></TR>\n", inodo.I_size))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_atime</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_atime[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_ctime</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_ctime[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_mtime</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_mtime[:])))

		for b := 0; b < 15; b++ {
			numBloque := inodo.I_block[b]
			puerto := fmt.Sprintf("b%d", b+1)

			if numBloque != -1 {
				dot.WriteString(fmt.Sprintf("      <TR><TD>i_block_%d</TD><TD PORT=\"%s\">%d</TD></TR>\n", b+1, puerto, numBloque))

				var nombreNodoBloque string
				if b < 12 {
					if tipo == "0" {
						nombreNodoBloque = fmt.Sprintf("carpeta%d", numBloque)
						dibujarBloqueCarpeta(disco, sb, numBloque, nombreNodoBloque, &bloquesDot, &flechas, bloquesDibujados)
					} else {
						nombreNodoBloque = fmt.Sprintf("archivo%d", numBloque)
						dibujarBloqueArchivo(disco, sb, numBloque, nombreNodoBloque, &bloquesDot, bloquesDibujados)
					}
				} else {
					nombreNodoBloque = fmt.Sprintf("apuntadores%d", numBloque)
					dibujarBloqueApuntadores(disco, sb, numBloque, nombreNodoBloque, tipo, &bloquesDot, &flechas, bloquesDibujados)
				}

				flechas.WriteString(fmt.Sprintf("  %s:%s -> %s\n", nombreNodoInodo, puerto, nombreNodoBloque))
			} else {
				dot.WriteString(fmt.Sprintf("      <TR><TD>i_block_%d</TD><TD>%d</TD></TR>\n", b+1, numBloque))
			}
		}

		dot.WriteString(fmt.Sprintf("      <TR><TD>i_type</TD><TD>%s</TD></TR>\n", tipo))
		dot.WriteString(fmt.Sprintf("      <TR><TD>i_perm</TD><TD>%s</TD></TR>\n", limpiarBytes(inodo.I_perm[:])))
		dot.WriteString("    </TABLE>\n")
		dot.WriteString("  >]\n\n")
	}

	if !hayInodosUsados {
		dot.WriteString("  sininodos [label=\"No hay inodos utilizados\"]\n")
	}

	dot.WriteString(bloquesDot.String())
	dot.WriteString(flechas.String())
	dot.WriteString("}\n")

	guardarYConvertirDot(dot.String(), pathSalida, "TREE")
}

// Dibuja un bloque de carpeta (4 entradas). Si una entrada apunta a un inodo (B_inodo != -1)
// que NO sea "." ni ".." (para evitar ciclos visuales), genera una flecha bloque -> inodo
// usando PORT solo en el origen.
func dibujarBloqueCarpeta(disco *os.File, sb Structs.Superblock, numBloque int32, nombreNodo string, dot *strings.Builder, flechas *strings.Builder, dibujados map[int32]bool) {
	if dibujados[numBloque] {
		return
	}
	dibujados[numBloque] = true

	var bloque Structs.Folderblock
	posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
	if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
		return
	}

	dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
	dot.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n")
	dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\" BGCOLOR=\"#a0c4e8\"><B>b. carpeta %d</B></TD></TR>\n", numBloque))

	for idx, c := range bloque.B_content {
		nombre := Structs.GetB_name(string(c.B_name[:]))
		puerto := fmt.Sprintf("p%d", idx)

		if c.B_inodo != -1 {
			dot.WriteString(fmt.Sprintf("      <TR><TD>%s</TD><TD PORT=\"%s\">%d</TD></TR>\n", nombre, puerto, c.B_inodo))

			// Evitar flechas de "." y ".." porque generan ciclos visuales (apuntan a si mismo o al padre ya dibujado)
			if nombre != "." && nombre != ".." {
				flechas.WriteString(fmt.Sprintf("  %s:%s -> inodo%d\n", nombreNodo, puerto, c.B_inodo))
			}
		} else {
			dot.WriteString("      <TR><TD></TD><TD>-1</TD></TR>\n")
		}
	}

	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n\n")
}

func dibujarBloqueArchivo(disco *os.File, sb Structs.Superblock, numBloque int32, nombreNodo string, dot *strings.Builder, dibujados map[int32]bool) {
	if dibujados[numBloque] {
		return
	}
	dibujados[numBloque] = true

	var bloque Structs.Fileblock
	posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
	if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
		return
	}

	contenido := Structs.GetB_content(string(bloque.B_content[:]))

	dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
	dot.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n")
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR=\"#f5d020\"><B>b. archivo %d</B></TD></TR>\n", numBloque))
	dot.WriteString(fmt.Sprintf("      <TR><TD>%s</TD></TR>\n", contenido))
	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n\n")
}

func dibujarBloqueApuntadores(disco *os.File, sb Structs.Superblock, numBloque int32, nombreNodo string, tipoInodo string, dot *strings.Builder, flechas *strings.Builder, dibujados map[int32]bool) {
	if dibujados[numBloque] {
		return
	}
	dibujados[numBloque] = true

	var bloque Structs.Pointerblock
	posBloque := sb.S_block_start + (numBloque * sb.S_block_size)
	if err := Herramientas.ReadObject(disco, &bloque, int64(posBloque)); err != nil {
		return
	}

	dot.WriteString(fmt.Sprintf("  %s [label=<\n", nombreNodo))
	dot.WriteString("    <TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"4\">\n")
	dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN=\"2\" BGCOLOR=\"#c9a0e8\"><B>b. apuntadores %d</B></TD></TR>\n", numBloque))

	for idx, p := range bloque.B_pointers {
		puerto := fmt.Sprintf("p%d", idx)
		dot.WriteString(fmt.Sprintf("      <TR><TD>%d</TD><TD PORT=\"%s\">%d</TD></TR>\n", idx, puerto, p))

		if p != -1 {
			var nombreDestino string
			if tipoInodo == "0" {
				nombreDestino = fmt.Sprintf("carpeta%d", p)
				dibujarBloqueCarpeta(disco, sb, p, nombreDestino, dot, flechas, dibujados)
			} else {
				nombreDestino = fmt.Sprintf("archivo%d", p)
				dibujarBloqueArchivo(disco, sb, p, nombreDestino, dot, dibujados)
			}
			flechas.WriteString(fmt.Sprintf("  %s:%s -> %s\n", nombreNodo, puerto, nombreDestino))
		}
	}

	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n\n")
}
