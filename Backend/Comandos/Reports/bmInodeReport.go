package reports

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strings"
)

func GenerarReporteBMInode(pathDisco string, nombreParticion string, pathSalida string) {
	disco, err := Herramientas.OpenFile(pathDisco)
	if err != nil {
		fmt.Println("ERROR REP BM_INODE: No se pudo abrir el disco")
		return
	}
	defer disco.Close()

	var mbr Structs.MBR
	if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
		fmt.Println("ERROR REP BM_INODE: No se pudo leer el MBR")
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
		fmt.Println("ERROR REP BM_INODE: No se encontro la particion")
		return
	}

	var sb Structs.Superblock
	if err := Herramientas.ReadObject(disco, &sb, int64(partStart)); err != nil {
		fmt.Println("ERROR REP BM_INODE: No se pudo leer el superbloque")
		return
	}

	var salida strings.Builder
	for i := int32(0); i < sb.S_inodes_count; i++ {
		var bit Structs.Bite
		if err := Herramientas.ReadObject(disco, &bit, int64(sb.S_bm_inode_start+i)); err != nil {
			fmt.Println("ERROR REP BM_INODE: No se pudo leer un bit del bitmap")
			return
		}
		salida.WriteByte(bit.Val[0])

		if (i+1)%20 == 0 {
			salida.WriteString("\n")
		} else {
			salida.WriteString(" ")
		}
	}

	if err := os.WriteFile(pathSalida, []byte(salida.String()), 0644); err != nil {
		fmt.Println("ERROR REP BM_INODE: No se pudo guardar el archivo de salida")
		return
	}

	fmt.Println("REP BM_INODE: Reporte generado exitosamente en:", pathSalida)
}
