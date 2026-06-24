package Structs

import "fmt"

/*
Mapa: path del disco -> ultima letra usada para particiones montadas de ese disco
Como funciona un mapa:
- Se declara un mapa con make(map[TipoClave]TipoValor)
- Se agrega un elemento al mapa con map[clave] = valor
- Se obtiene un valor del mapa con map[clave]
- Se elimina un elemento del mapa con delete(map, clave)
- Se verifica si una clave existe en el mapa con valor, existe := map[clave]
- Si existe es true, si no existe es false
Ejemplo:
LetrasMontaje := make(map[string]byte)
LetrasMontaje["/home/user/disk.dsk"] = 'A'
letra, existe := LetrasMontaje["/home/user/disk.dsk"]

	if existe {
		fmt.Println("La letra es: ", letra)
	}
*/
var LetrasMontaje = make(map[string]byte)

// ==============================================================================
// Almacena la informacion de cada particion montada
var Montadas []mountAlready

type mountAlready struct {
	Id    string //Id de la particion 60+id+letra
	PathM string //Path del disco al que pertenece la particion
	Name  string //Nombre de la particion
}

// La funcion append agrega el elemento nuevo al final de un slice, pero retorna un nuevo slice con el elemento agregado, por lo que se debe asignar
// nuevamente a la variable Montadas para que tenga el nuevo slice con el elemento agregado
func AddMontadas(id string, path string, name string) {
	Montadas = append(Montadas, mountAlready{Id: id, PathM: path, Name: name})
}

// Busca una particion montada por su Id, retorna el indice (-1 si no existe)
// Busca si una particion (identificada por disco+nombre) ya esta montada
func BuscarMontadaPorPathYNombre(path string, name string) int {
	for i := range Montadas {
		if Montadas[i].PathM == path && Montadas[i].Name == name {
			return i
		}
	}
	return -1
}

func BuscarMontadaPorId(id string) int {
	for i := range Montadas {
		if Montadas[i].Id == id {
			return i
		}
	}
	return -1
}

// Imprime todas las particiones actualmente montadas en el sistema
func ListarMontadas() {
	fmt.Println("\n===== Particiones Montadas =====")
	if len(Montadas) == 0 {
		fmt.Println("No hay particiones montadas")
	} else {
		for _, m := range Montadas {
			fmt.Println("ID:", m.Id, "| Disco:", m.PathM, "| Particion:", m.Name)
		}
	}
	fmt.Println("=================================")
}
