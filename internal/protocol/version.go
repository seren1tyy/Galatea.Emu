package protocol

// Константы, которые сервер ожидает от клиента при handshake
const (
	EVEVersionNumber   = 13.08
	MachoNetVersion    = uint16(414)
	EVEBuildVersion    = int32(958007)
	EVEProjectRegion   = "ccp"
	EVEProjectVersion  = "EVE-TRANQUILITY@ccp"
	EVEProjectCodename = "EVE-TRANQUILITY"
	EVEBirthday        = int32(170472)
)
