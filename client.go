package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"time"

	"project/nameserver"
	"project/services"
)

// Lista dei server disponibili ottenuta dal NameServer
var availableServers []nameserver.ServerInfo

// Indirizzo server selezionato dal load balancer
var serverAddr string

// Tipo di load balancing scelto
var loadBalancingType string

// Indice per Round Robin in algoritmo stateless
var roundRobinIndex int = 0

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: %s <type of load balancing: stateless or stateful>\n", os.Args[0])
		fmt.Println("Example: %s stateless", os.Args[0])
		os.Exit(1)
	}

	// Esegui lookup dei server disponibili dal NameServer
	lookup()

	//ciclo infinito per ricevere le richieste
	for {
		// selezionamento server
		selectServer()
		// selezionamento servizio
		var serviceType int
		fmt.Println("Service lookup. You can ask:")
		fmt.Println("0 : Fibonacci of a number \"n\"")
		fmt.Println("1 : Counting the occurrences of a word over every client")
		fmt.Println("2 : Exit")
		_, err := fmt.Scan(&serviceType)
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			continue
		}
		switch serviceType {
		case 0:
			fibonacci()
		case 1:
			counter()
		case 2:
			fmt.Println("See you next time")
			os.Exit(0)
		default:
			fmt.Println("Invalid service type. Use 0 for fibonacci, 1 for counter, 2 for exit\n")
		}
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
}

func selectServer() {
	loadBalancingType = os.Args[1]

	switch loadBalancingType {
	case "stateful":
		selectServerStateful()
	case "stateless":
		selectServerStateless()
	default:
		fmt.Println("ERRORE: algoritmo sbagliato scelto in input")
		os.Exit(1)
	}
}

func selectServerStateless() {
	serverAddr = availableServers[roundRobinIndex].Address
	fmt.Println("serverAddr: %s", serverAddr)
	roundRobinIndex = (roundRobinIndex + 1) % len(availableServers)
}

func selectServerStateful() {
	// per generare numeri casuali
	rand.Seed(time.Now().UnixNano())
	//calcolo somma totale dei pesi
	totalWeight := 0.0
	for _, server := range availableServers {
		totalWeight += server.Weight
	}

	//numero casuale tra 0 e totalWeight
	randomValue := rand.New(rand.NewSource(time.Now().UnixMilli())).Float64() * totalWeight
	fmt.Printf("somma pesi: %.f, valore casuale: %.f\n", totalWeight, randomValue)
	//raggiungi server scelto casualmente
	cumulativeWeight := 0.0
	for _, server := range availableServers {
		cumulativeWeight += server.Weight
		if randomValue <= cumulativeWeight {
			serverAddr = server.Address
			return
		}
	}
}

func fibonacci() {
	var n int = -1
	for n < 0 {
		fmt.Println("Insert the fibonacci index (non-negative number): ")
		// Ottieni indice fibonacci
		_, err := fmt.Scan(&n)
		if err != nil {
			fmt.Printf("Invalid input: %v\n", err)
			// Pulisci il buffer di input
			var discard string
			fmt.Scanln(&discard)
			n = -1
			continue
		}
		if n < 0 {
			fmt.Println("Error: Index must be non-negative")
		}
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
	var word string

	// Ottieni credenziali
	fmt.Printf("word to count the occurrences: ")
	_, err := fmt.Scan(&word)
	if err != nil {
		log.Fatalf("Failed to recieve word to search occurrences: %v", err)
	}

	// Connessione con server RPC
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server at %s: %v", serverAddr, err)
	}
	defer client.Close()

	log.Printf("Connected to server at %s", serverAddr)

	// Prepare arguments
	args := services.CounterArgs{
		Word: word,
	}
	var result services.CounterResult

	// Chiamata RPC sincrona
	log.Printf("Calling Contatore.Counter for user %s", word)
	err = client.Call("Contatore.Counter", &args, &result)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}

	// Display result
	fmt.Printf("User: %s\n", word)
	fmt.Printf("Request count: %d\n", result.RequestCount)
	fmt.Printf("%s\n", result.Message)
}
