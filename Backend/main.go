package main

import (
	Router "MIA_P1/Router"
	"fmt"
	"net/http"
)

func main() {
	Router.SetupRoutes()
	fmt.Println("Servidor corriendo en puerto 8080...")
	http.ListenAndServe(":8080", nil)
}

// ========================================== Funcion main de la fase 1 =============================================
/*
import (
	DM "MIA_P1/Comandos/AdmDeDiscos"
	FD "MIA_P1/Comandos/CarpetasArchivos"
	REP "MIA_P1/Comandos/Reports"
	FS "MIA_P1/Comandos/SistemaDeArchivos"
	UG "MIA_P1/Comandos/UsuariosGrupos"
	Router "MIA_P1/Router"
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Si se ejecuta como: go run main.go api
	// arranca el servidor HTTP en vez del CLI interactivo.
	// Sin argumentos, se comporta exactamente igual que antes (modo CLI).
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "api" {
		iniciarServidorAPI()
		return
	}

	iniciarCLI()
}

// iniciarServidorAPI levanta el servidor HTTP con todas las rutas registradas en Router.
func iniciarServidorAPI() {
	mux := http.NewServeMux()
	Router.RegistrarRutas(mux)

	puerto := ":8080"
	fmt.Println("Servidor API escuchando en http://localhost" + puerto)
	fmt.Println("Endpoint disponible: POST /api/mkdisk")

	if err := http.ListenAndServe(puerto, mux); err != nil {
		log.Fatal("Error al iniciar el servidor: ", err)
	}
}

// iniciarCLI contiene exactamente la misma logica que ya tenias: el loop
// interactivo que lee comandos desde la terminal.
func iniciarCLI() {
	//MENSAJES DE INICIO
	Ms_inicio := "Bienvenido escriba un comando"
	Ms_info := "Si desesa salir escriba el comando exit"

	fmt.Println(Ms_inicio)
	fmt.Println(Ms_info)

	reader := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n$: ")
		reader.Scan()

		entrada := strings.TrimRight(reader.Text(), " ")
		linea := strings.Split(entrada, "#")

		if strings.ToLower(linea[0]) != "exit" {
			analizar(linea[0])
		} else {
			fmt.Println("SALIENDO DEL PROGRAMA...")
			break
		}
	}
}

func analizar(entrada string) {
	//separar los parametros -size=10000 -path=ruta

	tmp := strings.TrimRight(entrada, " ")
	parametros := strings.Split(tmp, " -")
	//------------------------------- EJECUTAR SCRIPT ----------------------------------------------------------------
	if strings.ToLower(parametros[0]) == "script" {
		if len(parametros) == 2 {
			tmpParametro := strings.Split(parametros[1], "=")
			if strings.ToLower(tmpParametro[0]) == "path" && len(tmpParametro) == 2 {
				archivo, err := os.Open(tmpParametro[1])
				if err != nil {
					fmt.Println("Error al leer el script: ", err)
					return
				}
				defer archivo.Close()
				lector := bufio.NewScanner(archivo)
				for lector.Scan() {
					linea := strings.Split(lector.Text(), "#")
					if len(linea[0]) != 0 {
						fmt.Println("\n*********************************************************************************************")
						fmt.Println("Linea en ejecucion: ", linea[0])
						analizar(linea[0])
					}
				}
			} else {
				fmt.Println("SCRIPT ERROR: parametro path no encontrado")
			}
		}
		//------------------------------------------------------------- COMANDO PAUSE --------------------------------------------------------------------
	} else if strings.ToLower(parametros[0]) == "pause" {
		fmt.Println("\n===== PAUSE =====")
		fmt.Println("Presione ENTER para continuar...")
		fmt.Print("\n\n")
		time.Sleep(5 * time.Second)

	} else if strings.ToLower(parametros[0]) == "mkdisk" {
		if len(parametros) > 1 {
			DM.Mkdisk(parametros)
		} else {
			fmt.Println("MDISK ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "fdisk" {
		if len(parametros) > 1 {
			DM.Fdisk(parametros)
		} else {
			fmt.Println("MDISK ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "rmdisk" {
		if len(parametros) > 1 {
			DM.Rmdisk(parametros)
		} else {
			fmt.Println("MDISK ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "mount" {
		if len(parametros) > 1 {
			DM.Mount(parametros)
		} else {
			fmt.Println("MOUNT ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "mkfs" {
		if len(parametros) > 1 {
			FS.Mkfs(parametros)
		} else {
			fmt.Println("MKFS ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "login" {
		if len(parametros) > 1 {
			FS.Login(parametros)
		} else {
			fmt.Println("LOGIN ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "logout" {
		FS.Logout()

	} else if strings.ToLower(parametros[0]) == "cat" {
		if len(parametros) > 1 {
			FS.Cat(parametros)
		} else {
			fmt.Println("CAT ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "mkgrp" {
		if len(parametros) > 1 {
			UG.Mkgrp(parametros)
		} else {
			fmt.Println("MKGRP ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "mkusr" {
		if len(parametros) > 1 {
			UG.Mkusr(parametros)
		} else {
			fmt.Println("MKUSR ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "rmgrp" {
		if len(parametros) > 1 {
			UG.Rmgrp(parametros)
		} else {
			fmt.Println("RMGRP ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "rmusr" {
		if len(parametros) > 1 {
			UG.Rmusr(parametros)
		} else {
			fmt.Println("RMUSR ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "chgrp" {
		if len(parametros) > 1 {
			UG.Chgrp(parametros)
		} else {
			fmt.Println("CHGRP ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "mkdir" {
		if len(parametros) > 1 {
			FD.Mkdir(parametros)
		} else {
			fmt.Println("MKDIR ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "mkfile" {
		if len(parametros) > 1 {
			FD.Mkfile(parametros)
		} else {
			fmt.Println("MKFILE ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "rep" {
		if len(parametros) > 1 {
			REP.Rep(parametros)
		} else {
			fmt.Println("REP ERROR: No se han ingresado parametros")
		}

	} else if strings.ToLower(parametros[0]) == "exit" {
		fmt.Println("SALIENDO DEL PROGRAMA...")
		os.Exit(0)
	} else {
		fmt.Println("ERROR: Comando no reconocido")
	}

}
*/
