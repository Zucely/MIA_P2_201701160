package herramientas

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func CrearDisco(path string) error {
	//asegurarse que la carpeta exista
	dir := filepath.Dir(path)                             //obtiene la ruta sin el nombre del archivo
	if err := os.MkdirAll(dir, os.ModePerm); err != nil { //crea la carpeta si no existe
		fmt.Println("Error al crear el disco, path: ", err)
		return err
	}

	//crear el archivo si no existe todavia
	if _, err := os.Stat(path); os.IsNotExist(err) {
		newFile, err := os.Create(path)
		if err != nil {
			fmt.Println("Error al crear el disco, archivo: ", err)
			return err
		}
		defer newFile.Close()
	}
	return nil
}

func VerificarDisco(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error al abrir el disco: ", err)
		return nil, err
	}
	return file, nil
}

func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0) //(posicion, desde donde) -> (5,0) significa llevar a la posicion 5 desde el inicio del archivo (posicion 0). C
	//contar 5 espacios a partir de la posicion 0
	//position-> a donde va a llegar, 0-> desde donde empieza a contar
	//binary.LittleEndian -> orden de los bytes (de menor a mayor)
	//data -> que se va a escribir
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Error al escribir en el disco: ", err)
		return err
	}
	return nil
}

func ReadObject(file *os.File, data interface{}, position int64) error { //paramtros que recibe: file, data (donde se va a guardar lo leido), posicion (desde donde se va a leer)
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Error al leer el disco: ", err)
		return err
	}
	return nil
}
