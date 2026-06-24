package Structs

// Representa la sesion activa del sistema. Solo puede haber UNA sesion a la vez.
var SesionActiva UserInfo

type UserInfo struct {
	Id         string //id de la particion montada (ej: 601A)
	IdGrp      int32  //id del grupo al que pertenece el usuario logueado
	IdUsr      int32  //id del usuario logueado
	Nombre     string //nombre del usuario (identifica si es root o cualquier otro)
	Status     bool   //si esta iniciada la sesion
	PathD      string //path del disco
	NombrePart string //nombre de la particion (para no buscarlo de nuevo en Montadas)
}
