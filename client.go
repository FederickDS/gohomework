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

// lista dei server disponibili ottenuta dal NameServer
var availableServers []nameserver.ServerInfo

// indirizzo server selezionato dal load balancer
var serverAddr string

// tipo di load balancing scelto
var loadBalancingType string

// indice per Round Robin in algoritmo stateless
var roundRobinIndex int = 0

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Apri l'appilcazione scrivendo: %s <type of load balancing: stateless or stateful>\n", os.Args[0])
		fmt.Printf("Esempio: %s stateless", os.Args[0])
		os.Exit(1)
	}

	// lookup dei server disponibili
	lookup()

	for {
		// scelta server ad ogni nuova richiesta del servizio
		selectServer()
		// scelta servizio
		var serviceType int
		fmt.Printf("\nIl metodo richiesto verrà eseguito su un server remoto.\nScegli il servizio da eseguire con il numero a sinistra della riga\n")
		fmt.Println("0 : Calcolo indice di fibonacci")
		fmt.Println("1 : Conta quante volte una parola è stata richiesta da ogni client")
		fmt.Println("2 : Esci")
		_, err := fmt.Scan(&serviceType)
		if err != nil {
			fmt.Println("Invalid input: %v\n", err)
			continue
		}
		switch serviceType {
		case 0:
			fibonacci()
		case 1:
			counter()
		case 2:
			fmt.Println("Alla prossima")
			os.Exit(0)
		default:
			fmt.Println("Tipo di servizio scelto invalido. Riprova")
		}
	}
}

// lookup contatta il NameServer per ottenere la lista dei server disponibili
func lookup() {
	nameServerAddr := "localhost:9000" // indirizzo del NameServer

	// connessione al NameServer
	client, err := rpc.Dial("tcp", nameServerAddr)
	if err != nil {
		log.Fatalf("Impossibile connettersi al NameServer: %v", err)
	}
	defer client.Close()

	// inizializza argomenti per la lookup
	args := nameserver.LookupArgs{}
	var reply nameserver.LookupReply

	// chiamata RPC per lookup
	err = client.Call("NameServer.Lookup", &args, &reply)
	if err != nil {
		log.Fatalf("Errore durante la lookup: %v", err)
	}

	// verifica che ci siano server disponibili
	if len(reply.Servers) == 0 {
		log.Fatalf("Nessun server disponibile. Almeno un server deve essere attivo prima del client.")
	}

	availableServers = reply.Servers
	fmt.Printf("Trovati %d server disponibili\n", len(availableServers))
	for i, server := range availableServers {
		fmt.Printf("  [%d] %s\n", i, server.Address)
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
		fmt.Println("Algoritmo sbagliato scelto in input. Riparti con go run client.go [stateless/stateful]")
		os.Exit(1)
	}
}

func selectServerStateless() {
	serverAddr = availableServers[roundRobinIndex].Address
	roundRobinIndex = (roundRobinIndex + 1) % len(availableServers)
}

func selectServerStateful() {
	// calcolo somma totale dei pesi
	totalWeight := 0.0
	for _, server := range availableServers {
		totalWeight += server.Weight
	}

	// numero casuale tra 0 e totalWeight
	randomValue := rand.New(rand.NewSource(time.Now().UnixMilli())).Float64() * totalWeight
	// raggiungi server scelto casualmente
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
		fmt.Printf("Numero di fibonacci a posizione: ")
		// ottieni indice fibonacci
		_, err := fmt.Scan(&n)
		if err != nil {
			fmt.Printf("L'input non è un numero. Riprova: %v\n", err)
			// pulizia buffer di input per sicurezza
			var discard string
			fmt.Scanln(&discard)
			n = -1
			continue
		}
		if n < 0 {
			fmt.Println("La posizione è un numero non negativo. Riprova")
		}
	}

	// connessione con server RPC
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Connessione al server %s e'fallita: %v", serverAddr, err)
	}
	defer client.Close()

	fmt.Printf("Connessione al server con indirizzo %s\n", serverAddr)

	// argomenti per chiamata RPC
	args := services.Args{Value: n}
	var result services.Result

	// chiamata RPC per il numero di fibonacci
	err = client.Call("Aritmetico.Fibonacci", &args, &result)
	if err != nil {
		log.Fatalf("La chiamata RPC ha fallito: %v", err)
	}

	fmt.Printf("Fibonacci(%d) = %d\n", n, result.Value)
}

func counter() {
	var word string

	fmt.Printf("Inserisci parola di cui contare le occorrenze: ")
	_, err := fmt.Scan(&word)
	if err != nil {
		log.Fatalf("Impossibile salvare la parola: %v", err)
	}

	// connessione con server RPC
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Connessione al server all'indirizzo %s fallita: %v", serverAddr, err)
	}
	defer client.Close()

	fmt.Printf("Connesso al server all'indirizzo %s\n", serverAddr)

	// prepara argomenti per chiamata RPC
	args := services.CounterArgs{
		Word: word,
	}
	var result services.CounterResult

	// Chiamata RPC al contatore
	err = client.Call("Contatore.Counter", &args, &result)
	if err != nil {
		log.Fatalf("Chiamata RPC fallita: %v", err)
	}

	fmt.Printf("La parola: %s e' stata richiesta %d volte\n", word, result.RequestCount)
}
