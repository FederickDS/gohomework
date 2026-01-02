package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	
	"project/services"
)

// Indirizzo server (all'inizio preimpostato)
var serverAddr string = "localhost:12345"

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <type of service (0 - fibonacci, 1 - counter)> [<other args>]\n", os.Args[0])
		os.Exit(1)
	}
	
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
	// TODO: Implementare nella fase 2
	fmt.Println("Counter service not yet implemented")
	os.Exit(1)
}
