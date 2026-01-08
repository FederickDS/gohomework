package nameserver

// informazioni di un server registrato
type ServerInfo struct {
	Address string  // indirizzo completo (es: "192.168.1.10:12345")
	Port    string  // porta (es: "12345")
	Weight  float64 // peso per load balancing (0.0 - 1.0)
}

// argomenti per la registrazione di un server
type RegisterArgs struct {
	Address string  // indirizzo IP + porta del server (es: "localhost:12345")
	Weight  float64 // peso del server per load balancing
}

// risposta alla registrazione di un server
type RegisterReply struct {
	Success bool
	Message string
}

// argomenti per la deregistrazione di un server
type DeregisterArgs struct {
	Address string // indirizzo del server da rimuovere
}

// risposta alla deregistrazione di un server
type DeregisterReply struct {
	Success bool
	Message string
}

// argomenti per la query di lookup
type LookupArgs struct {
	ServiceName string // nome del servizio richiesto
}

// risposta del nameserver al client
type LookupReply struct {
	Servers []ServerInfo // lista con info dei server
}
