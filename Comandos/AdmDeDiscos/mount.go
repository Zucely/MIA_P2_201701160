package Comandos

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Mount(parametros []string) {
	var name string //Nombre de la particion a montar
	var path string //Path del Disco
	paramC := true

	for _, parametro := range parametros[1:] {
		tmp := strings.TrimRight(parametro, " ")
		valores := strings.Split(tmp, "=")

		if len(valores) != 2 {
			fmt.Println("ERROR MOUNT, valor desconocido de parametros ", valores[1])
			return //Finaliza comando
		}
		//mount -path=/home/”su_usuario”/Calificacion_MIA/Discos/Disco1.mia -name=Part11

		//******************* PATH *************
		if strings.ToLower(valores[0]) == "path" {
			path = strings.ReplaceAll(valores[1], "\"", "")
			_, err := os.Stat(path)
			if os.IsNotExist(err) {
				fmt.Println("ERROR MOUNT: El disco no existe")
				paramC = false
				break // Terminar el bucle porque encontramos un nombre único
			}
			//********************  NAME *****************
		} else if strings.ToLower(valores[0]) == "name" {
			// Eliminar comillas
			name = strings.ReplaceAll(valores[1], "\"", "")
			// Eliminar espacios en blanco al final
			name = strings.TrimSpace(name)

			//******************* ERROR EN LOS PARAMETROS *************
		} else {
			fmt.Println("ERROR MOUNT: Parametro desconocido: ", valores[0])
			paramC = false
			break //por si en el camino reconoce algo invalido de una vez se sale
		}
	}

	if paramC {
		if path != "" {
			if name != "" {
				// Abrir y cargar el disco
				disco, err := Herramientas.OpenFile(path)
				if err != nil {
					fmt.Println("ERROR NO SE PUEDE LEER EL DISCO ")
					return
				}

				//Se crea un mbr para cargar el mbr del disco
				var mbr Structs.MBR
				//Guardo el mbr leido
				if err := Herramientas.ReadObject(disco, &mbr, 0); err != nil {
					return
				}

				// cerrar el archivo del disco (cuando termine completa la funcion)
				defer disco.Close()

				montar := true // para guardar error si no se puede montar
				for i := 0; i < 4; i++ {
					nombre := Structs.GetName(string(mbr.Partitions[i].Name[:]))
					if nombre == name {
						montar = false
						if string(mbr.Partitions[i].Type[:]) != "E" { //no se puede montar una particion extendida

							//buscar en la lista de particiones montadas si ya esta montada esta particion
							if Structs.BuscarMontadaPorPathYNombre(path, name) == -1 { //1 = Mounted, -1 = Unmounted

								var idDisco int32 = mbr.Id

								//se le manda el path del disco y se obtiene la ultima letra usada para particiones montadas de ese disco
								//ejemplo: /home/user/disk.dsk -> A
								letraActual, existe := Structs.LetrasMontaje[path]

								var letra byte
								if !existe {
									letra = 'A'
								} else {
									letra = letraActual + 1
								}

								//carne = 2017011 - 60
								//guardar en el mapa la ultima letra usada para particiones montadas de ese disco
								Structs.LetrasMontaje[path] = letra
								id := "60" + strconv.Itoa(int(idDisco)) + string(letra) //Id de particion

								/// Actualizar en memoria (RAM)
								//creo que no lo estoy usando xd
								copy(mbr.Partitions[i].Status[:], "M")
								copy(mbr.Partitions[i].Id[:], id)
								mbr.Partitions[i].Correlative = idDisco

								Structs.AddMontadas(id, path, name)

								//sobreescribir el mbr para guardar los cambios
								//if err := Herramientas.WriteObject(disco, mbr, 0); err != nil {
								//	return
								//}

								fmt.Println("Particion con nombre ", name, " montada correctamente. ID: ", id)
								Structs.ListarMontadas()
							} else {
								fmt.Println("ERROR MOUNT. Esta particion ya esta montada")
								return
							}
						} else {
							fmt.Println("ERROR MOUNT. No se puede montar una particion extendida")
							return
						}
					}
				}

				if montar {
					fmt.Println("ERROR MOUNT. No se pudo montar la particion ", name)
					fmt.Println("ERROR MOUNT. No se encontro la particion")
					return
				}

			} else {
				fmt.Println("ERROR: Falta parametro NAME en MOUNT")
			}
		} else {
			fmt.Println("ERROR: Falta parametro PATH en MOUNT")
		}
	}
}
