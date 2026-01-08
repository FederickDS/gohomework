package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	"project/nameserver"
)

// NameServer gestisce la registrazione e discovery dei server
type NameServer struct {
	mu      sync.RWMutex
	servers map[string]nameserver.ServerInfo // Mappa: address -> ServerInfo
}

// NewNameServer crea una nuova istanza del NameServer
func NewNameServer() *NameServer {
	return &NameServer{
		servers: make(map[string]nameserver.ServerInfo),
	}
}

// registra un nuovo server
func (ns *NameServer) Register(args *nameserver.RegisterArgs, reply *nameserver.RegisterReply) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if args.Address == "" {
		reply.Success = false
		reply.Message = "Indirizzo server non valido"
		return fmt.Errorf("indirizzo vuoto")
	}

	// controlla se il peso e'tra 0 o 1
	weight := args.Weight
	if weight <= 0 || weight > 1 {
		weight = 1.0 // default
		fmt.Printf("Peso non valido per %s, impostato a default 1.0", args.Address)
	}

	// controlla se il server è già registrato
	if existing, exists := ns.servers[args.Address]; exists {
		// aggiorna il peso se diverso
		if existing.Weight != weight {
			ns.servers[args.Address] = nameserver.ServerInfo{
				Address: args.Address,
				Port:    extractPort(args.Address),
				Weight:  weight,
			}
			reply.Success = true
			reply.Message = fmt.Sprintf("Server %s già registrato, peso aggiornato a %.2f", args.Address, weight)
			fmt.Printf("Server %s peso aggiornato: %.2f", args.Address, weight)
		} else {
			reply.Success = true
			reply.Message = fmt.Sprintf("Server %s già registrato con peso %.2f", args.Address, weight)
		}
		return nil
	}

	// registra il nuovo server
	ns.servers[args.Address] = nameserver.ServerInfo{
		Address: args.Address,
		Port:    extractPort(args.Address),
		Weight:  weight,
	}

	reply.Success = true
	reply.Message = fmt.Sprintf("Server %s registrato con successo (peso: %.2f)", args.Address, weight)

	fmt.Printf("Nuovo server registrato: %s (peso: %.2f)", args.Address, weight)
	fmt.Printf("Totale server registrati: %d", len(ns.servers))

	return nil
}

// rimuove un server dalla lista
func (ns *NameServer) Deregister(args *nameserver.DeregisterArgs, reply *nameserver.DeregisterReply) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if args.Address == "" {
		reply.Success = false
		reply.Message = "Indirizzo server non valido"
		return fmt.Errorf("indirizzo vuoto")
	}

	// controlla se il server esiste
	if _, exists := ns.servers[args.Address]; !exists {
		reply.Success = false
		reply.Message = fmt.Sprintf("Server %s non trovato", args.Address)
		return fmt.Errorf("server non registrato")
	}

	// rimuovi il server
	delete(ns.servers, args.Address)

	reply.Success = true
	reply.Message = fmt.Sprintf("Server %s deregistrato con successo", args.Address)

	fmt.Printf("Server deregistrato: %s", args.Address)
	fmt.Printf("Totale server registrati: %d", len(ns.servers))

	return nil
}

// restituisce la lista di tutti i server connessi al nameserver
func (ns *NameServer) Lookup(args *nameserver.LookupArgs, reply *nameserver.LookupReply) error {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	reply.Servers = make([]nameserver.ServerInfo, 0, len(ns.servers))
	for _, serverInfo := range ns.servers {
		reply.Servers = append(reply.Servers, serverInfo)
	}

	fmt.Printf("Lookup richiesto: restituiti %d server", len(reply.Servers))

	return nil
}

// funzione per estrarre la porta da un indirizzo (es: "localhost:12345" -> "12345")
func extractPort(address string) string {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}
	return port
}

func main() {
	// porta per il nameserver
	port := ":9000"

	ns := NewNameServer()

	server := rpc.NewServer()

	//registra il nameserver
	err := server.RegisterName("NameServer", ns)
	if err != nil {
		log.Fatal("Registrazione nameserver fallita: ", err)
	}

	//nameserver in ascolto su porta del nameserver
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Errore ascolto del nameserver: ", err)
	}
	defer lis.Close()

	fmt.Printf("NameServer nameserver in ascolto all'indirizzo %s", lis.Addr().String())
	fmt.Println("Metodi disponibili:")
	fmt.Println("  - NameServer.Register")
	fmt.Println("  - NameServer.Deregister")
	fmt.Println("  - NameServer.Lookup")

	server.Accept(lis)
}
