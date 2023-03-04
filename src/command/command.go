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

const SYST = "SYST"
const OPTS = "OPTS"
const QUIT = "QUIT"
const FEAT = "FEAT"
const TYPE = "TYPE"

// connection modes
// TODO: needs support
const EPSV = "EPSV"
const PASV = "PASV"
const EPRT = "EPRT"
const PORT = "PORT"

var CommandsEnablingActiveConnType = []string{EPRT, PORT}
var CommandsEnablingPassiveConnType = []string{EPSV, PASV}
