package Structs

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// NOTA: Recordar que los atributos de los struct deben iniciar con mayuscula
type MBR struct {
	MbrSize    int32        //mbr_tamano
	FechaC     [16]byte     //mbr_fecha_creacion
	Id         int32        //mbr_dsk_signature (random de forma unica)
	Fit        [2]byte      //dsk_fit
	Partitions [4]Partition //mbr_partitions
}

type Partition struct {
	Status      [1]byte  //part_status
	Type        [1]byte  //part_type
	Fit         [2]byte  //part_fit
	Start       int32    //part_start
	Size        int32    //part_s
	Name        [16]byte //part_name
	Correlative int32    //part_correlative
	Id          [5]byte  //part_id
}

//contador de IDs para discos

// Setear valores de la particion
// La estrucutura (p *Partition) indica que el metodo SetInfo pertenece al struct Partition, la p es una referencia a la particion que se esta modificando
// el *partition indica que se esta trabajando con una referencia a la particion, no con una copia
// Si no se usa el *, se trabaja con una copia de la particion, y los cambios no se reflejan en la particion original
// Por eso se usa el * para trabajar con la referencia y modificar la particion original
// SetInfo es el nombre de la funcion, y trae los parametros necesarios para setear la particion. EN este caso, no devuelve nada (void)
func (p *Partition) SetInfo(newType string, fit string, newStart int32, newSize int32, name string, correlativo int32) {
	p.Size = newSize
	p.Start = newStart
	p.Correlative = correlativo
	copy(p.Name[:], name)
	copy(p.Fit[:], fit)
	copy(p.Status[:], "I")
	copy(p.Type[:], newType)
}

// Metodos de Partition
// Cuando se guarda el nombre en el struct, si el nombre es menor a 16 caracteres, se rellenan los bytes faltantes con \x00 (byte nulo)
// Por lo tanto, al obtener el nombre y transformarlo en string, incluye los caracteres nulos al final (si es que los hay)
// se debe eliminar los bytes nulos para que no aparezcan al imprimir el nombre
// La funcion IndexByte de strings retorna la posicion del primer byte nulo
// Si no hay bytes nulos, retorna -1
// Si hay bytes nulos, se guarda la cadena hasta el primer byte nulo (elimina los bytes nulos)
// Si el nombre es "particion1", se guarda como "particion1\x00\x00\x00\x00\x00\x00\x00" (16 bytes)
// Al llamar a GetName("particion1\x00\x00\x00\x00\x00\x00\x00") retorna "particion1"
func GetName(nombre string) string {
	posicionNulo := strings.IndexByte(nombre, 0)
	//Si posicionNulo retorna -1 no hay bytes nulos
	if posicionNulo != -1 {
		//guarda la cadena hasta el primer byte nulo (elimina los bytes nulos)
		nombre = nombre[:posicionNulo]
	}
	return nombre
}

func GetId(nombre string) string {
	//si existe id, no contiene bytes nulos
	posicionNulo := strings.IndexByte(nombre, 0)
	//si posicionNulo  no es -1, no existe id.
	if posicionNulo != -1 {
		nombre = "-"
	}
	return nombre
}

func (p *Partition) GetEnd() int32 {
	return p.Start + p.Size
}

type EBR struct {
	Status [1]byte //part_mount (si esta montada)
	Type   [1]byte
	Fit    [2]byte  //part_fit
	Start  int32    //part_start
	Size   int32    //part_s
	Name   [16]byte //part_name
	Next   int32    //part_next
}

func (e *EBR) SetInfo(fit string, newStart int32, newSize int32, name string, newNext int32) {
	e.Size = newSize
	e.Start = newStart
	e.Next = newNext
	copy(e.Name[:], name)
	copy(e.Fit[:], fit)
	copy(e.Status[:], "I")
	copy(e.Type[:], "L")
}

func (e *EBR) GetEnd() int32 {
	return e.Start + e.Size + int32(binary.Size(e))
}

/*func GetIdMount(data Mount) string {
	return data.MPath
}*/

// Reportes de los Structs
func PrintMBR(data MBR) {
	fmt.Println("\n     Disco")
	fmt.Printf("CreationDate: %s, fit: %s, size: %d, id: %d\n", string(data.FechaC[:]), string(data.Fit[:]), data.MbrSize, data.Id)
	for i := 0; i < 4; i++ {
		fmt.Printf("Partition %d: %s, %s, %d, %d, %s, %d\n", i, string(data.Partitions[i].Name[:]), string(data.Partitions[i].Type[:]), data.Partitions[i].Start, data.Partitions[i].Size, string(data.Partitions[i].Fit[:]), data.Partitions[i].Correlative)
	}
}

func PrintEbr(data EBR) {
	fmt.Println("part_status ", string(data.Status[:]))
	fmt.Println("part_type ", string(data.Type[:]))
	fmt.Println("part_fit: ", string(data.Fit[:]))
	fmt.Println("part_start: ", data.Start)
	fmt.Println("part_s ", data.Size)
	fmt.Println("part_name: ", string(data.Name[:]))
	fmt.Println("next_part: ", data.Next)
}
