package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"
)

type bloqueDisk struct {
	nombre      string
	size        int32
	porcentaje  float64 // porcentaje respecto al tamaño TOTAL del disco
	esLibre     bool
	esExtendida bool
	logicas     []bloqueDisk // solo si esExtendida == true
}

func GenerarReporteDisk(pathDisco string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP DISK: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP DISK: No se pudo leer el MBR")
		return
	}

	tamDisco := mbr.MbrSize
	sizeMBR := int32(binary.Size(mbr))

	bloques := calcularBloques(disco, mbr, tamDisco, sizeMBR)

	dot := construirDotDisk(bloques)
	guardarYConvertirDot(dot, pathSalida, "DISK")
}

// -------------------- CALCULO DE BLOQUES --------------------

func calcularBloques(disco *os.File, mbr Structs.MBR, tamDisco int32, sizeMBR int32) []bloqueDisk {
	var particiones []Structs.Partition
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Size > 0 {
			particiones = append(particiones, mbr.Partitions[i])
		}
	}

	sort.Slice(particiones, func(i, j int) bool {
		return particiones[i].Start < particiones[j].Start
	})

	var bloques []bloqueDisk
	cursor := sizeMBR

	for _, part := range particiones {
		// Espacio libre antes de esta particion
		if part.Start > cursor {
			gap := part.Start - cursor
			bloques = append(bloques, bloqueDisk{nombre: "Libre", size: gap, esLibre: true})
		}

		nombre := Structs.GetName(string(part.Name[:]))
		tipo := limpiarBytes(part.Type[:])

		if tipo == "E" {
			logicas := calcularBloquesEBR(disco, part)
			bloques = append(bloques, bloqueDisk{nombre: nombre, size: part.Size, esExtendida: true, logicas: logicas})
		} else {
			bloques = append(bloques, bloqueDisk{nombre: nombre, size: part.Size})
		}

		cursor = part.Start + part.Size
	}

	if cursor < tamDisco {
		bloques = append(bloques, bloqueDisk{nombre: "Libre", size: tamDisco - cursor, esLibre: true})
	}

	for i := range bloques {
		bloques[i].porcentaje = (float64(bloques[i].size) / float64(tamDisco)) * 100
		for j := range bloques[i].logicas {
			// porcentaje de cada logica respecto al disco TOTAL (no solo la extendida)
			bloques[i].logicas[j].porcentaje = (float64(bloques[i].logicas[j].size) / float64(tamDisco)) * 100
		}
	}

	return bloques
}

func calcularBloquesEBR(disco *os.File, partExtendida Structs.Partition) []bloqueDisk {
	var logicas []bloqueDisk

	var ebr Structs.EBR
	if err := Herramientas.ReadObject(disco, &ebr, int64(partExtendida.Start)); err != nil {
		return logicas
	}

	sizeEBRStruct := int32(binary.Size(ebr))
	cursor := partExtendida.Start

	for {
		if ebr.Start > cursor {
			gap := ebr.Start - cursor
			logicas = append(logicas, bloqueDisk{nombre: "Libre", size: gap, esLibre: true})
		}

		nombre := Structs.GetName(string(ebr.Name[:]))
		tamanoTotal := ebr.Size + sizeEBRStruct
		logicas = append(logicas, bloqueDisk{nombre: nombre, size: tamanoTotal})

		cursor = ebr.Start + tamanoTotal

		if ebr.Next == -1 {
			break
		}
		if err := Herramientas.ReadObject(disco, &ebr, int64(ebr.Next)); err != nil {
			break
		}
	}

	finExtendida := partExtendida.Start + partExtendida.Size
	if cursor < finExtendida {
		logicas = append(logicas, bloqueDisk{nombre: "Libre", size: finExtendida - cursor, esLibre: true})
	}

	return logicas
}

// -------------------- CONSTRUCCION DEL .DOT --------------------

func construirDotDisk(bloques []bloqueDisk) string {
	var dot strings.Builder

	totalColumnas := len(bloques)

	dot.WriteString("digraph G {\n")
	dot.WriteString("  node [shape=plaintext fontname=\"Helvetica\"]\n\n")
	dot.WriteString("  disk [label=<\n")
	dot.WriteString("    <TABLE BORDER='1' CELLBORDER='1' CELLSPACING='0' CELLPADDING='8'>\n")

	// Titulo general
	dot.WriteString(fmt.Sprintf("      <TR><TD COLSPAN='%d' BGCOLOR='#2c3e50'><FONT COLOR='white'><B>REPORTE DISK</B></FONT></TD></TR>\n", totalColumnas+1))

	// Fila 1: MBR + cada particion (bloques del mismo ancho)
	dot.WriteString("      <TR>\n")
	dot.WriteString("        <TD BGCOLOR='#34495e'><FONT COLOR='white'><B>MBR</B></FONT></TD>\n")

	for _, b := range bloques {
		if b.esExtendida {
			dot.WriteString(construirCeldaExtendida(b))
		} else {
			color := colorBloque(b)
			dot.WriteString(fmt.Sprintf(
				"        <TD BGCOLOR='%s'>%s<BR/>%.1f%%</TD>\n",
				color, b.nombre, b.porcentaje,
			))
		}
	}
	dot.WriteString("      </TR>\n")

	dot.WriteString("    </TABLE>\n")
	dot.WriteString("  >]\n")
	dot.WriteString("}\n")

	return dot.String()
}

// Celda de la extendida: dividida en 2 niveles (arriba titulo, abajo lista de logicas)
func construirCeldaExtendida(b bloqueDisk) string {
	var inner strings.Builder

	inner.WriteString("        <TD>\n")
	inner.WriteString("          <TABLE BORDER='0' CELLBORDER='1' CELLSPACING='0' CELLPADDING='6'>\n")

	numColLogicas := maxInt(len(b.logicas), 1)

	// Fila superior: EXTENDIDA + nombre
	inner.WriteString(fmt.Sprintf("            <TR><TD COLSPAN='%d' BGCOLOR='#8e44ad'><FONT COLOR='white'><B>EXTENDIDA<BR/>%s</B></FONT></TD></TR>\n", numColLogicas, b.nombre))

	// Fila inferior: lista de EBRs/logicas
	inner.WriteString("            <TR>\n")
	if len(b.logicas) == 0 {
		inner.WriteString("              <TD>(sin particiones logicas)</TD>\n")
	} else {
		for _, log := range b.logicas {
			color := colorBloque(log)
			etiqueta := "EBR"
			if log.esLibre {
				etiqueta = "Libre"
			}
			inner.WriteString(fmt.Sprintf(
				"              <TD BGCOLOR='%s'>%s<BR/>%s<BR/>%.1f%%</TD>\n",
				color, etiqueta, log.nombre, log.porcentaje,
			))
		}
	}
	inner.WriteString("            </TR>\n")

	inner.WriteString("          </TABLE>\n")
	inner.WriteString("        </TD>\n")

	return inner.String()
}

func colorBloque(b bloqueDisk) string {
	if b.esLibre {
		return "#ecf0f1"
	}
	return "#a9dfbf"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
