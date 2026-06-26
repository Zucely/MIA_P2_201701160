package reports

import (
	Tools "MIA_P1/Comandos/SistemaDeArchivos"
	"strconv"
	"strings"
)

// Convierte un digito octal de permiso (0-7) a su representacion rwx
func octalToRWX(digito byte) string {
	n := int(digito - '0')
	r, w, x := "-", "-", "-"
	if n&4 != 0 {
		r = "r"
	}
	if n&2 != 0 {
		w = "w"
	}
	if n&1 != 0 {
		x = "x"
	}
	return r + w + x
}

// Construye la cadena de permisos estilo Linux (ej: -rw-rw-r--) a partir
// del permiso octal de 3 digitos y el tipo de inodo ("0"=carpeta, "1"=archivo)
func permisosLinux(perm string, tipo string) string {
	tipoChar := "-"
	if tipo == "0" {
		tipoChar = "d"
	}
	if len(perm) != 3 {
		return tipoChar + "?????????"
	}
	return tipoChar + octalToRWX(perm[0]) + octalToRWX(perm[1]) + octalToRWX(perm[2])
}

// Retorna el ultimo componente de una ruta (ej: /home/user/a.txt -> a.txt)
func ultimoSegmento(path string) string {
	limpio := strings.TrimRight(path, "/")
	idx := strings.LastIndex(limpio, "/")
	if idx == -1 {
		if limpio == "" {
			return "/"
		}
		return limpio
	}
	nombre := limpio[idx+1:]
	if nombre == "" {
		return "/"
	}
	return nombre
}

// Escapa caracteres especiales de HTML para insertar contenido de archivo
// de forma segura dentro de un label de graphviz
func escaparHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// Lee users.txt y construye mapas uid->nombreUsuario y gid->nombreGrupo
func cargarUsuariosYGrupos(pathDisco string, nombreParticion string) (map[int32]string, map[int32]string) {
	usuarios := make(map[int32]string)
	grupos := make(map[int32]string)

	contenido, err := Tools.LeerUsersTxt(pathDisco, nombreParticion)
	if err != nil {
		return usuarios, grupos
	}

	for _, linea := range strings.Split(contenido, "\n") {
		linea = strings.TrimSpace(linea)
		if linea == "" {
			continue
		}
		campos := strings.Split(linea, ",")
		for i := range campos {
			campos[i] = strings.TrimSpace(campos[i])
		}
		if len(campos) < 3 {
			continue
		}
		idNum, err := strconv.Atoi(campos[0])
		if err != nil {
			continue
		}
		switch strings.ToUpper(campos[1]) {
		case "G":
			grupos[int32(idNum)] = campos[2]
		case "U":
			if len(campos) >= 4 {
				usuarios[int32(idNum)] = campos[3]
			}
		}
	}

	return usuarios, grupos
}
