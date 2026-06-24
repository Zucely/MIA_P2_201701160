package reports

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
)

func GenerarReporteLs(pathDisco string, nombreParticion string, pathSalida string, pathFileLs string) {
	if pathFileLs == "" {
		pathFileLs = "/"
	}

	disco, sb, _, err := Tools.AbrirDiscoYSuperbloque(pathDisco, nombreParticion)
	if err != nil {
		fmt.Println("ERROR REP LS:", err)
		return
	}
	defer disco.Close()

	numInodo, err := Tools.BuscarRuta(disco, sb, pathFileLs)
	if err != nil {
		fmt.Println("ERROR REP LS: No se encontro la ruta", pathFileLs)
		return
	}

	var inodo Structs.Inode
	posInodo := sb.S_inode_start + (numInodo * sb.S_inode_size)
	if err := Herramientas.ReadObject(disco, &inodo, int64(posInodo)); err != nil {
		fmt.Println("ERROR REP LS: No se pudo leer el inodo")
		return
	}

	type entradaLs struct {
		nombre   string
		numInodo int32
	}
	var entradas []entradaLs

	if limpiarBytes(inodo.I_type[:]) == "1" {
		// Es un archivo: se muestra unicamente esa entrada
		entradas = append(entradas, entradaLs{nombre: ultimoSegmento(pathFileLs), numInodo: numInodo})
	} else {
		// Es una carpeta: se listan sus hijos (sin . ni ..)
		for b := 0; b < 12; b++ {
			numBloque := inodo.I_block[b]
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
				if nombre == "" || nombre == "-" || nombre == "." || nombre == ".." {
					continue
				}
				entradas = append(entradas, entradaLs{nombre: nombre, numInodo: c.B_inodo})
			}
		}
	}

	usuarios, grupos := cargarUsuariosYGrupos(pathDisco, nombreParticion)

	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")
	dot.WriteString("  ls [label=<\n")
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='6'>\n")
	dot.WriteString("      <TR>")
	for _, encabezado := range []string{"Permisos", "Owner", "Grupo", "Size", "Fecha", "Hora", "Tipo", "Name"} {
		dot.WriteString(fmt.Sprintf("<TD BGCOLOR='#2c3e50'><FONT COLOR='white'><B>%s</B></FONT></TD>", encabezado))
	}
	dot.WriteString("</TR>\n")

	if len(entradas) == 0 {
		dot.WriteString("      <TR><TD COLSPAN='8'>La carpeta esta vacia</TD></TR>\n")
	}

	for _, e := range entradas {
		var inodoHijo Structs.Inode
		posHijo := sb.S_inode_start + (e.numInodo * sb.S_inode_size)
		if err := Herramientas.ReadObject(disco, &inodoHijo, int64(posHijo)); err != nil {
			continue
		}

		tipo := limpiarBytes(inodoHijo.I_type[:])
		perm := limpiarBytes(inodoHijo.I_perm[:])
		permTexto := permisosLinux(perm, tipo)

		owner, existeOwner := usuarios[inodoHijo.I_uid]
		if !existeOwner {
			owner = strconv.Itoa(int(inodoHijo.I_uid))
		}
		grupo, existeGrupo := grupos[inodoHijo.I_gid]
		if !existeGrupo {
			grupo = strconv.Itoa(int(inodoHijo.I_gid))
		}

		fechaHora := strings.Fields(limpiarBytes(inodoHijo.I_mtime[:]))
		fecha, hora := "", ""
		if len(fechaHora) >= 1 {
			fecha = fechaHora[0]
		}
		if len(fechaHora) >= 2 {
			hora = fechaHora[1]
		}

		tipoTexto := "Archivo"
		if tipo == "0" {
			tipoTexto = "Carpeta"
		}

		dot.WriteString("      <TR>")
		dot.WriteString(fmt.Sprintf("<TD>%s</TD><TD>%s</TD><TD>%s</TD><TD>%d</TD><TD>%s</TD><TD>%s</TD><TD>%s</TD><TD>%s</TD>",
			permTexto, owner, grupo, inodoHijo.I_size, fecha, hora, tipoTexto, e.nombre))
		dot.WriteString("</TR>\n")
	}

	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n")
	dot.WriteString("}\n")

	guardarYConvertirDot(dot.String(), pathSalida, "LS")
}
