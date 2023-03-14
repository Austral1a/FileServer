package command

// auth commands
const USER = "USER"
const PASS = "PASS"

const PWD = "PWD"
const CWD = "CWD"
const LIST = "LIST"
const MLSD = "MLSD"

const STOR = "STOR"
const DELE = "DELE"
const RETP = "RETP"
const RNFR = "RNFR"
const RNTO = "RNTO"

const SYST = "SYST"
const STAT = "STAT"
const OPTS = "OPTS"
const QUIT = "QUIT"
const FEAT = "FEAT"
const TYPE = "TYPE"

// connection modes
const EPSV = "EPSV"
const PASV = "PASV"
const EPRT = "EPRT"
const PORT = "PORT"

var CommandsEnablingActiveConnType = []string{EPRT, PORT}
var CommandsEnablingPassiveConnType = []string{EPSV, PASV}
