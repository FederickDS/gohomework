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

// Register registra un nuovo server
func (ns *NameServer) Register(args *nameserver.RegisterArgs, reply *nameserver.RegisterReply) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if args.Address == "" {
		reply.Success = false
		reply.Message = "Indirizzo server non valido"
		return fmt.Errorf("indirizzo vuoto")
	}

	// Valida il peso (deve essere tra 0 e 1)
	weight := args.Weight
	if weight <= 0 || weight > 1 {
		weight = 1.0 // Peso di default
		log.Printf("Peso non valido per %s, impostato a default 1.0", args.Address)
	}

	// Controlla se il server è già registrato
	if existing, exists := ns.servers[args.Address]; exists {
		// Aggiorna il peso se diverso
		if existing.Weight != weight {
			ns.servers[args.Address] = nameserver.ServerInfo{
				Address: args.Address,
				Port:    extractPort(args.Address),
				Weight:  weight,
			}
			reply.Success = true
			reply.Message = fmt.Sprintf("Server %s già registrato, peso aggiornato a %.2f", args.Address, weight)
			log.Printf("Server %s peso aggiornato: %.2f", args.Address, weight)
		} else {
			reply.Success = true
			reply.Message = fmt.Sprintf("Server %s già registrato con peso %.2f", args.Address, weight)
			log.Printf("Server %s già presente", args.Address)
		}
		return nil
	}

	// Registra il nuovo server
	ns.servers[args.Address] = nameserver.ServerInfo{
		Address: args.Address,
		Port:    extractPort(args.Address),
		Weight:  weight,
	}

	reply.Success = true
	reply.Message = fmt.Sprintf("Server %s registrato con successo (peso: %.2f)", args.Address, weight)

	log.Printf("Nuovo server registrato: %s (peso: %.2f)", args.Address, weight)
	log.Printf("Totale server registrati: %d", len(ns.servers))

	return nil
}

// Deregister rimuove un server dalla lista
func (ns *NameServer) Deregister(args *nameserver.DeregisterArgs, reply *nameserver.DeregisterReply) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if args.Address == "" {
		reply.Success = false
		reply.Message = "Indirizzo server non valido"
		return fmt.Errorf("indirizzo vuoto")
	}

	// Controlla se il server esiste
	if _, exists := ns.servers[args.Address]; !exists {
		reply.Success = false
		reply.Message = fmt.Sprintf("Server %s non trovato", args.Address)
		return fmt.Errorf("server non registrato")
	}

	// Rimuovi il server
	delete(ns.servers, args.Address)

	reply.Success = true
	reply.Message = fmt.Sprintf("Server %s deregistrato con successo", args.Address)

	log.Printf("Server deregistrato: %s", args.Address)
	log.Printf("Totale server registrati: %d", len(ns.servers))

	return nil
}

// Lookup restituisce la lista di tutti i server registrati con i loro pesi
func (ns *NameServer) Lookup(args *nameserver.LookupArgs, reply *nameserver.LookupReply) error {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	reply.Servers = make([]nameserver.ServerInfo, 0, len(ns.servers))
	for _, serverInfo := range ns.servers {
		reply.Servers = append(reply.Servers, serverInfo)
	}

	log.Printf("Lookup richiesto: restituiti %d server", len(reply.Servers))

	return nil
}

// extractPort estrae la porta da un indirizzo (es: "localhost:12345" -> "12345")
func extractPort(address string) string {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}
	return port
}

func main() {
	// Porta fissa per il nameserver (hardcoded)
	port := ":9000"

	// Crea istanza del NameServer
	ns := NewNameServer()

	// Registra il servizio RPC
	server := rpc.NewServer()
	err := server.RegisterName("NameServer", ns)
	if err != nil {
		log.Fatal("Failed to register NameServer: ", err)
	}

	// Ascolta connessioni TCP
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	defer lis.Close()

	fmt.Printf("NameServer listening on %s", lis.Addr().String())
	fmt.Println("Available methods:")
	fmt.Println("  - NameServer.Register")
	fmt.Println("  - NameServer.Deregister")
	fmt.Println("  - NameServer.Lookup")

	// Accetta e serve le richieste
	server.Accept(lis)
}
