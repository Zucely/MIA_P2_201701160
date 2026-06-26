package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// path salida es el path donde se va a guardar el reporte, path disco es el path del disco del que se va a generar el reporte
func GenerarReporteMBR(pathDisco string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP MBR: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP MBR: No se pudo leer el MBR")
		return
	}

	var dot strings.Builder

	dot.WriteString("digraph G {\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")

	dot.WriteString("  mbr [label=<\n")
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='6'>\n")

	// Titulo
	dot.WriteString("      <TR><TD COLSPAN='2' BGCOLOR='#2c3e50'><FONT COLOR='white'><B>REPORTE MBR</B></FONT></TD></TR>\n")

	// Datos del MBR
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#dfe6e9'><B>Tamaño</B></TD><TD>%d bytes</TD></TR>\n", mbr.MbrSize))
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#dfe6e9'><B>Fecha Creación</B></TD><TD>%s</TD></TR>\n", limpiarBytes(mbr.FechaC[:])))
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#dfe6e9'><B>ID</B></TD><TD>%d</TD></TR>\n", mbr.Id))
	dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#dfe6e9'><B>Fit</B></TD><TD>%s</TD></TR>\n", limpiarBytes(mbr.Fit[:])))

	// Particiones
	for i := 0; i < 4; i++ {
		part := mbr.Partitions[i]
		nombre := Structs.GetName(string(part.Name[:]))
		tipo := limpiarBytes(part.Type[:])

		if nombre == "" {
			continue // particion vacia, no se muestra
		}

		colorTitulo := "#e67e22" // primaria
		tituloPart := "PARTICIÓN PRIMARIA"
		if tipo == "E" {
			colorTitulo = "#8e44ad"
			tituloPart = "PARTICIÓN EXTENDIDA"
		}

		dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN='2' BGCOLOR='%s'><FONT COLOR='white'><B>%s</B></FONT></TD></TR>\n", colorTitulo, tituloPart))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#fdebd0'><B>Status</B></TD><TD>%s</TD></TR>\n", limpiarBytes(part.Status[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#fdebd0'><B>Type</B></TD><TD>%s</TD></TR>\n", tipo))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#fdebd0'><B>Fit</B></TD><TD>%s</TD></TR>\n", limpiarBytes(part.Fit[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#fdebd0'><B>Start</B></TD><TD>%d</TD></TR>\n", part.Start))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#fdebd0'><B>Size</B></TD><TD>%d bytes</TD></TR>\n", part.Size))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#fdebd0'><B>Name</B></TD><TD>%s</TD></TR>\n", nombre))

		// Si es extendida, agregar sus EBRs encadenados
		if tipo == "E" {
			agregarEBRsAlReporte(&dot, disco, part)
		}
	}

	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n")
	dot.WriteString("}\n")

	guardarYConvertirDot(dot.String(), pathSalida, "MBR")
}

func agregarEBRsAlReporte(dot *strings.Builder, disco *os.File, partExtendida Structs.Partition) {
	var ebr Structs.EBR
	if err := Herramientas.ReadObject(disco, &ebr, int64(partExtendida.Start)); err != nil {
		fmt.Println("ERROR REP MBR: No se pudo leer el EBR inicial")
		return
	}

	numero := 1
	for {
		nombre := Structs.GetName(string(ebr.Name[:]))

		dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN='2' BGCOLOR='#27ae60'><FONT COLOR='white'><B>EBR #%d (LÓGICA)</B></FONT></TD></TR>\n", numero))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Status</B></TD><TD>%s</TD></TR>\n", limpiarBytes(ebr.Status[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Type</B></TD><TD>%s</TD></TR>\n", limpiarBytes(ebr.Type[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Fit</B></TD><TD>%s</TD></TR>\n", limpiarBytes(ebr.Fit[:])))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Start</B></TD><TD>%d</TD></TR>\n", ebr.Start))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Size</B></TD><TD>%d bytes</TD></TR>\n", ebr.Size))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Name</B></TD><TD>%s</TD></TR>\n", nombre))
		dot.WriteString(fmt.Sprintf("      <TR><TD BGCOLOR='#d5f5e3'><B>Next</B></TD><TD>%d</TD></TR>\n", ebr.Next))

		if ebr.Next == -1 {
			break
		}

		if err := Herramientas.ReadObject(disco, &ebr, int64(ebr.Next)); err != nil {
			fmt.Println("ERROR REP MBR: No se pudo leer el siguiente EBR")
			break
		}
		numero++
	}
}

// Quita los bytes nulos (\x00) que quedan al final de los campos [N]byte
func limpiarBytes(b []byte) string {
	s := string(b)
	if pos := strings.IndexByte(s, 0); pos != -1 {
		s = s[:pos]
	}
	return strings.TrimSpace(s)
}

func guardarYConvertirDot(contenidoDot string, pathSalida string, nombreReporte string) {
	carpeta := filepath.Dir(pathSalida)
	if err := os.MkdirAll(carpeta, os.ModePerm); err != nil {
		fmt.Println("ERROR REP: No se pudo crear la carpeta de salida:", err)
		return
	}
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(pathSalida)), ".")
	if extension == "" {
		extension = "jpg"
	}
	pathDot := strings.TrimSuffix(pathSalida, filepath.Ext(pathSalida)) + ".dot"
	if err := os.WriteFile(pathDot, []byte(contenidoDot), 0644); err != nil {
		fmt.Println("ERROR REP: No se pudo guardar el archivo .dot:", err)
		return
	}
	cmd := exec.Command("dot", "-T"+extension, pathDot, "-o", pathSalida)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("ERROR REP: Graphviz no pudo generar la imagen:", err)
		fmt.Println("Output:", string(output))
		return
	}
	fmt.Println("REP", nombreReporte, ": Reporte generado exitosamente en:", pathSalida) // Cambiar el mensaje para indicar el tipo de reporte generado
}
