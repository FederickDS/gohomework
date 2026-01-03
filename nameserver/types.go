package nameserver

// ServerInfo contiene le informazioni di un server registrato
type ServerInfo struct {
	Address string // Indirizzo completo (es: "192.168.1.10:12345")
	Port    string // Porta (es: "12345")
}

// RegisterArgs - Argomenti per la registrazione di un server
type RegisterArgs struct {
	Address string // Indirizzo IP + porta del server (es: "localhost:12345")
}

// RegisterReply - Risposta alla registrazione
type RegisterReply struct {
	Success bool
	Message string
}

// DeregisterArgs - Argomenti per la deregistrazione di un server
type DeregisterArgs struct {
	Address string // Indirizzo del server da rimuovere
}

// DeregisterReply - Risposta alla deregistrazione
type DeregisterReply struct {
	Success bool
	Message string
}

// LookupArgs - Argomenti per la query di lookup (fase 4)
type LookupArgs struct {
	ServiceName string // Nome del servizio richiesto (opzionale per ora)
}

// LookupReply - Risposta con la lista dei server disponibili
type LookupReply struct {
	Servers []string // Lista di indirizzi server disponibili
}
