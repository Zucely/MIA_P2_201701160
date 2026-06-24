package Comandos

import (
	Herramientas "MIA_P1/Herramientas"
	Structs "MIA_P1/Structs"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var contadorDiscos int32 = 0

func Mkdisk(parametros []string) string {
	salida := " "
	fmt.Println("Ejecutando mkdisk...")
	//fmt.Println()
	//fmt.Println("DEBUG: parametros =", parametros)
	fmt.Println()
	//parametros a recibir -size -unit -fit -path

	var size int      //obligatorio
	var path string   //obligatorio
	fit := "FF"       //opcional por defecto FF
	unit := 1048576   //1024 * 1024 = 1048576 por defecto MB (en bytes)
	paramC := true    //valida que todos los parametros sean correctos
	sizeInit := false //para saber si vino el parametro size
	pathInit := false //para saber si vino el parametro path

	for _, parametro := range parametros[1:] {
		//recorre los parametros desde el indice 1 (0 es mkdisk) hasta el final
		tmp2 := strings.TrimSpace(parametro) //Quita espacios a la derecha por si aun viniera
		tmp := strings.Split(tmp2, "=")      //separa el parametro del valor

		//fmt.Println()
		//fmt.Println("DEBUG: tmp2 =", tmp2) // Imprime el valor de tmp2
		//fmt.Println("DEBUG: tmp =", tmp)   // Imprime el valor de tmp como un slice
		//fmt.Println()

		//Si llegara a faltar el valor del parametro actual, reconoce el error y se interrumpe el proceso
		if len(tmp) != 2 {
			fmt.Println("MKDISK Error: Valor desconocido del parametro ", tmp[0])
			paramC = false
			//break //para finalizar el ciclo for con el error y no ejecutar lo que haga falta
			return "Valor desconocido del parametro"
		}

		if strings.ToLower(tmp[0]) == "size" { //si el parametro es size
			sizeInit = true
			var err error
			size, err = strconv.Atoi(tmp[1]) //convierte el valor a entero
			if err != nil || size <= 0 {     //si no se pudo convertir o es menor o igual a 0
				fmt.Println("MDISK ERROR: El parametro size debe ser un valor entero mayor a 0, se leyo:", tmp[1])
				paramC = false
				break //sale ya que encontro un error
			}

		} else if strings.ToLower(tmp[0]) == "fit" {
			if strings.ToLower(tmp[1]) == "bf" {
				//el ajuste es best fit
				fit = "BF"
			} else if strings.ToLower(tmp[1]) == "wf" {
				//el ajuste es worst fit
				fit = "WF"
			} else if strings.ToLower(tmp[1]) != "ff" {
				//el ajuste ff "first fit" es el por defecto, si no es ninguno de los 3
				fmt.Println("MDISK ERROR: El parametro fit solo acepta los valores: bf, wf o ff, se leyo:", tmp[1])
				paramC = false
				break //sale ya que encontro un error
			}
		} else if strings.ToLower(tmp[0]) == "unit" {
			if strings.ToLower(tmp[1]) == "k" {
				//el tamaño es en KB
				unit = 1024
			} else if strings.ToLower(tmp[1]) != "m" {
				//el tamaño en MB es el por defecto, si no es k o m
				fmt.Println("MDISK ERROR: El parametro unit solo acepta los valores: k o m, se leyo:", tmp[1])
				paramC = false
				break //sale ya que encontro un error
			}
		} else if strings.ToLower(tmp[0]) == "path" {
			pathInit = true
			path = tmp[1] //guarda la ruta tal cual venga
		} else {
			fmt.Println("MDISK ERROR: El parametro", tmp[0], "no es valido")
			paramC = false
			break //sale ya que encontro un error
		}

	}

	if paramC { //si todos los parametros son correctos
		if sizeInit && pathInit {
			tam := size * unit //calcula el tamaño real a crear. El comando trae el tamaño (1, 2, 3...) y la unidad (k o m), por eso se multiplican
			//verifica si trae comillas y las quita de ser el caso
			diskPath := strings.Trim(path, "\"")     //quita las comillas de la ruta si es que las trae
			diskName := strings.Split(diskPath, "/") //separa la ruta por las barras para obtener el nombre del disco
			disk := diskName[len(diskName)-1]        //el nombre del disco es el ultimo elemento del arreglo
			//Crear el archivo binario
			//verificar si la ruta existe, si no existe se crea, si existe se muestra un error
			//verificar si el disco ya existe, si existe se muestra un error, si no existe se crea el disco
			if Herramientas.VerificarDisco(diskPath) {
				fmt.Println("MDISK ERROR: El disco ya existe en la ruta especificada")
				return "El disco ya existe en la ruta especificada"
			}
			err := Herramientas.CrearDisco(diskPath)
			if err != nil {
				fmt.Println("MDISK ERROR: No se pudo crear el disco en la ruta especificada")
			}

			//Abrir el archivo binario
			file, err := Herramientas.OpenFile(diskPath)
			if err != nil {
				fmt.Println("MDISK ERROR: No se pudo abrir el disco en la ruta especificada")
			}

			//crear el array de bytes con el tamaño especificado
			datos := make([]byte, tam)
			newErr := Herramientas.WriteObject(file, datos, 0)
			if newErr != nil {
				fmt.Println("MDISK ERROR: No se pudo escribir en el disco,", newErr)
			}

			ahora := time.Now()

			//Id random
			//idRandom := rand.Intn(9) + 1

			//Obtener el Id del archivo, para eso se lee el ultimo id guardado en el archivo contador.txt, se le suma 1, se guarda de nuevo y se retorna el nuevo id
			//idGuardado := ObtenerSiguienteIdDisco()

			//Obtener Id mientras esta en ejecucion
			contadorDiscos++
			idTemp := contadorDiscos

			//crear el MBR
			var newMBR Structs.MBR
			newMBR.MbrSize = int32(tam)                              //atributo tamanio del MBR que es el tamanio total del disco
			newMBR.Id = int32(idTemp)                                //atributo id del MBR que es un numero random unico para cada disco, se obtiene del contador de discos
			copy(newMBR.Fit[:], fit)                                 //copia el fit al mbr
			copy(newMBR.FechaC[:], ahora.Format("02/01/2006 15:04")) //formato de fecha dd/mm/yyyy hh:mm

			//escribir el MBR en el archivo
			newErr = Herramientas.WriteObject(file, newMBR, 0)
			if newErr != nil {
				fmt.Println("MDISK ERROR: No se pudo escribir el MBR en el disco,", newErr)
			} else {
				salida = "MDISK: Disco " + disk + " creado exitosamente en la ruta: " + path
			}

			defer file.Close() //cerrar el archivo al finalizar todo

			//imprimir el MBR creado
			var tempMBR Structs.MBR
			if err := Herramientas.ReadObject(file, &tempMBR, 0); err != nil {
				fmt.Println("MDISK ERROR: No se pudo leer el MBR del disco,", err)
			} else {
				Structs.PrintMBR(tempMBR)
			}

			fmt.Println("\n ======= END MKDISK =======")
		} else {
			fmt.Println("MDISK ERROR: No se ha ingresado el parametro size o el parametro path")
		}
	}
	return salida
}

// Lee el ultimo id usado, le suma 1, lo guarda de nuevo y lo retorna

// Agregar "strings" al import existente si no lo tienes:
// import (
//     Herramientas "MIA_P1/Herramientas"
//     Structs "MIA_P1/Structs"
//     "fmt"
//     "strconv"
//     "strings"
//     "time"
// )

// CrearDiscoLogico contiene la logica de negocio pura para crear un disco,
// sin parseo de parametros CLI y sin fmt.Println de depuracion.
// La usan tanto Mkdisk (CLI) como el controller de la API.
// Retorna el MBR ya escrito en disco, o un error si algo falla.
func CrearDiscoLogico(size int, unit string, fit string, path string) (Structs.MBR, error) {
	var vacio Structs.MBR

	if size <= 0 {
		return vacio, fmt.Errorf("el tamaño debe ser un valor entero mayor a 0")
	}

	unitBytes := 1048576
	switch strings.ToUpper(unit) {
	case "", "M":
		unitBytes = 1048576
	case "K":
		unitBytes = 1024
	default:
		return vacio, fmt.Errorf("el parametro unit solo acepta los valores: k o m")
	}

	fitNormalizado := "FF"
	switch strings.ToUpper(fit) {
	case "", "FF":
		fitNormalizado = "FF"
	case "BF":
		fitNormalizado = "BF"
	case "WF":
		fitNormalizado = "WF"
	default:
		return vacio, fmt.Errorf("el parametro fit solo acepta los valores: bf, wf o ff")
	}

	if path == "" {
		return vacio, fmt.Errorf("falta el parametro path")
	}

	tam := size * unitBytes

	if Herramientas.VerificarDisco(path) {
		return vacio, fmt.Errorf("el disco ya existe en la ruta especificada")
	}

	if err := Herramientas.CrearDisco(path); err != nil {
		return vacio, fmt.Errorf("no se pudo crear el disco en la ruta especificada: %v", err)
	}

	file, err := Herramientas.OpenFile(path)
	if err != nil {
		return vacio, fmt.Errorf("no se pudo abrir el disco en la ruta especificada: %v", err)
	}
	defer file.Close()

	datos := make([]byte, tam)
	if err := Herramientas.WriteObject(file, datos, 0); err != nil {
		return vacio, fmt.Errorf("no se pudo escribir en el disco: %v", err)
	}

	ahora := time.Now()

	contadorDiscos++
	idTemp := contadorDiscos

	var newMBR Structs.MBR
	newMBR.MbrSize = int32(tam)
	newMBR.Id = int32(idTemp)
	copy(newMBR.Fit[:], fitNormalizado)
	copy(newMBR.FechaC[:], ahora.Format("02/01/2006 15:04"))

	if err := Herramientas.WriteObject(file, newMBR, 0); err != nil {
		return vacio, fmt.Errorf("no se pudo escribir el MBR en el disco: %v", err)
	}

	var mbrEscrito Structs.MBR
	if err := Herramientas.ReadObject(file, &mbrEscrito, 0); err != nil {
		return vacio, fmt.Errorf("no se pudo leer el MBR recien escrito: %v", err)
	}

	return mbrEscrito, nil
}
