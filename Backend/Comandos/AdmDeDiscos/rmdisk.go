package Comandos

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Rmdisk(parametros []string) {
	//PARAMETROS:  -path
	var path string //obligatorio (ruta del disco a eliminar)

	for _, parametro := range parametros[1:] {
		//quito los espacios en blanco despues de cada parametro
		tmp2 := strings.TrimSpace(parametro)
		//quito las comillas si es que vienen
		tmp2 = strings.Trim(tmp2, "\"")
		//separo el parametro del valor
		tmp := strings.Split(tmp2, "=")

		//Si falta el valor del parametro actual lo reconoce como error e interrumpe el proceso
		if len(tmp) != 2 {
			fmt.Println("RMDISK Error: Valor desconocido del parametro ", tmp[0])
			break
		}

		if strings.ToLower(tmp[0]) == "path" {
			path = tmp[1]
			path = strings.Trim(path, "\"") //quito las comillas de la ruta si es que las trae

			// Verificar que el archivo exista antes de pedir confirmacion
			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Println("RMDISK Error: El disco no existe en la ruta especificada")
				return
			}

			diskName := strings.Split(path, "/")
			disk := diskName[len(diskName)-1]

			// ----------- Pedir confirmacion -----------
			fmt.Printf("¿Esta seguro que desea eliminar el disco %s? (Y/N): ", disk)
			reader := bufio.NewScanner(os.Stdin)
			reader.Scan()
			respuesta := strings.ToUpper(strings.TrimSpace(reader.Text()))

			if respuesta != "Y" {
				fmt.Println("RMDISK: Operacion cancelada por el usuario")
				return
			}

			// ----------- Eliminar el disco -----------
			err := os.Remove(path)
			if err != nil {
				fmt.Println("RMDISK Error: No se pudo eliminar el disco. Verifique la ruta.")
				return
			}

			fmt.Println("Disco eliminado exitosamente:", disk)
		} else {
			fmt.Println("RMDISK Error: Parametro desconocido ", tmp[0])
			break
		}
	}
}
