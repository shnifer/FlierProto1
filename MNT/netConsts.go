package MNT

const (
	CMD_BROADCAST = "broadcast"
	CMD_CHECKROOM = "checkRoom"
	IN_MSG = "msg"
	RES_CHECKROOM = "resCheckRoom"
	ERR_UNKNOWNCMD = "unknownCmd"
	RES_LOGIN = "LoginAccepted"
	CMD_GETGALAXY = "getGalaxy"
)

const (
	ROLE_PILOT = "pilot"
	ROLE_ENGINEER = "engineer"
	ROLE_NAVIGATOR = "navigator"
	ROLE_CARGO = "cargo"
)

const ServerName  = "localhost"
const TcpPort = ":6666"