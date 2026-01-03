package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	
	"project/nameserver"
	"project/services"
)

// Lista dei server disponibili ottenuta dal NameServer
var availableServers []string

// Indirizzo server selezionato dal load balancer
var serverAddr string

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <type of service (0 - fibonacci, 1 - counter)> [<other args>]\n", os.Args[0])
		os.Exit(1)
	}
	
	// Esegui lookup dei server disponibili dal NameServer
	lookup()
	
	// Controlla il tipo di funzione
	serviceType := os.Args[1]
	switch serviceType {
	case "0":
		fibonacci()
	case "1":
		counter()
	default:
		fmt.Printf("Invalid service type. Use 0 for fibonacci or 1 for counter\n")
		os.Exit(1)
	}
}

// lookup contatta il NameServer per ottenere la lista dei server disponibili
func lookup() {
	nameServerAddr := "localhost:9000" // Indirizzo hardcoded del NameServer
	
	log.Printf("Contatto il NameServer su %s per ottenere i server disponibili...", nameServerAddr)
	
	// Connessione al NameServer
	client, err := rpc.Dial("tcp", nameServerAddr)
	if err != nil {
		log.Fatalf("ERRORE: Impossibile connettersi al NameServer: %v", err)
	}
	defer client.Close()
	
	// Prepara argomenti per la lookup (per ora vuoti)
	args := nameserver.LookupArgs{}
	var reply nameserver.LookupReply
	
	// Chiamata RPC per lookup
	err = client.Call("NameServer.Lookup", &args, &reply)
	if err != nil {
		log.Fatalf("ERRORE durante la lookup: %v", err)
	}
	
	// Verifica che ci siano server disponibili
	if len(reply.Servers) == 0 {
		log.Fatalf("ERRORE: Nessun server disponibile. Avvia almeno un server prima del client.")
	}
	
	availableServers = reply.Servers
	log.Printf("Trovati %d server disponibili:", len(availableServers))
	for i, server := range availableServers {
		log.Printf("  [%d] %s", i, server)
	}
	
	// Per ora selezioniamo sempre il primo server (fase 5: load balancing)
	serverAddr = availableServers[0]
	log.Printf("Selezionato server: %s", serverAddr)
}

func fibonacci() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s 0 <positive integer>\n", os.Args[0])
		fmt.Printf("Example: client 0 10\n")
		os.Exit(1)
	}
	
	// Ottieni indice fibonacci
	n, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid Fibonacci index: %v", err)
	}
	
	// Connessione con server RPC
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server at %s: %v", serverAddr, err)
	}
	defer client.Close()
	
	log.Printf("Connected to server at %s", serverAddr)
	
	// Prepare arguments
	args := services.Args{Value: n}
	var result services.Result
	
	// Chiamata RPC sincrona
	log.Printf("Calling Aritmetico.Fibonacci with N=%d", n)
	err = client.Call("Aritmetico.Fibonacci", &args, &result)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}
	
	// Display result
	fmt.Printf("Fibonacci(%d) = %d\n", n, result.Value)
}

func counter() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s 1 <username> <password>\n", os.Args[0])
		fmt.Printf("Example: client 1 mario rossi123\n")
		os.Exit(1)
	}
	
	// Ottieni credenziali
	username := os.Args[2]
	password := os.Args[3]
	
	// Connessione con server RPC
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server at %s: %v", serverAddr, err)
	}
	defer client.Close()
	
	log.Printf("Connected to server at %s", serverAddr)
	
	// Prepare arguments
	args := services.CounterArgs{
		Username: username,
		Password: password,
	}
	var result services.CounterResult
	
	// Chiamata RPC sincrona
	log.Printf("Calling Contatore.Counter for user %s", username)
	err = client.Call("Contatore.Counter", &args, &result)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}
	
	// Display result
	fmt.Printf("User: %s\n", username)
	fmt.Printf("Request count: %d\n", result.RequestCount)
	fmt.Printf("%s\n", result.Message)
}
