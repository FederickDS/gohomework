package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"project/nameserver"
	"project/services"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Porta di default o da argomento
	port := ":12345"
	serverWeight := 1.0

	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	if len(os.Args) > 2 {
		_, err := fmt.Sscanf(os.Args[2], "%f", &serverWeight)
		if err != nil || serverWeight <= 0 || serverWeight > 1 {
			log.Printf("Peso non valido '%s', uso default 1.0", os.Args[2])
			serverWeight = 1.0
		}
	}

	// Configura connessione Redis
	redisAddr := "localhost:6379"
	if redisEnv := os.Getenv("REDIS_ADDR"); redisEnv != "" {
		redisAddr = redisEnv
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password di default
		DB:       0,  // database di default
	})

	// Verifica connessione Redis
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Impossibile connettersi a Redis: %v", err)
	}
	log.Printf("Connesso a Redis su %s", redisAddr)

	// Crea istanza del servizio Aritmetico
	aritmeticoService := new(services.Aritmetico)

	// Crea istanza del servizio Contatore con Redis client
	contatoreService := &services.Contatore{
		RedisClient: redisClient,
	}

	// Registra i servizi con il server RPC
	server := rpc.NewServer()

	err = server.RegisterName("Aritmetico", aritmeticoService)
	if err != nil {
		log.Fatal("Failed to register Aritmetico service: ", err)
	}

	err = server.RegisterName("Contatore", contatoreService)
	if err != nil {
		log.Fatal("Failed to register Contatore service: ", err)
	}

	// Ascolta connessioni TCP in ingresso
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	defer lis.Close()

	fmt.Println("RPC server listening on %s", lis.Addr().String())
	fmt.Println("Available services:")
	fmt.Println("  - Aritmetico.Fibonacci")
	fmt.Println("  - Contatore.Counter")

	// Registra questo server sul NameServer prima di accettare richieste
	serverAddress := lis.Addr().String()
	registerWithNameServer(serverAddress, serverWeight)

	// Setup signal handler per deregistrazione pulita
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Goroutine per gestire la deregistrazione su segnale
	go func() {
		sig := <-sigChan
		log.Printf("\nRicevuto segnale %v, deregistrazione in corso...", sig)
		deregisterFromNameServer(serverAddress)
		lis.Close()
		os.Exit(0)
	}()

	// Accetta e serve le richieste
	server.Accept(lis)
}

// registerWithNameServer registra questo server sul NameServer
func registerWithNameServer(serverAddr string, weight float64) {
	nameServerAddr := "localhost:9000" // Indirizzo hardcoded del NameServer

	log.Printf("Tentativo di registrazione sul NameServer a %s con peso %.2f", nameServerAddr, weight)

	// Connessione al NameServer
	client, err := rpc.Dial("tcp", nameServerAddr)
	if err != nil {
		log.Printf("ATTENZIONE: Impossibile connettersi al NameServer: %v", err)
		log.Println("Il server continuerà comunque ad operare")
		return
	}
	defer client.Close()

	// Prepara argomenti per la registrazione
	args := nameserver.RegisterArgs{
		Address: serverAddr,
		Weight:  weight,
	}
	var reply nameserver.RegisterReply

	// Chiamata RPC per registrazione
	err = client.Call("NameServer.Register", &args, &reply)
	if err != nil {
		log.Printf("ERRORE durante la registrazione: %v", err)
		return
	}

	if reply.Success {
		log.Printf("Registrazione completata: %s", reply.Message)
	} else {
		log.Printf("Registrazione fallita: %s", reply.Message)
	}
}

// deregisterFromNameServer deregistra questo server dal NameServer
func deregisterFromNameServer(serverAddr string) {
	nameServerAddr := "localhost:9000" // Indirizzo hardcoded del NameServer

	log.Printf("Deregistrazione dal NameServer a %s...", nameServerAddr)

	// Connessione al NameServer
	client, err := rpc.Dial("tcp", nameServerAddr)
	if err != nil {
		log.Printf("ATTENZIONE: Impossibile connettersi al NameServer: %v", err)
		return
	}
	defer client.Close()

	// Prepara argomenti per la deregistrazione
	args := nameserver.DeregisterArgs{
		Address: serverAddr,
	}
	var reply nameserver.DeregisterReply

	// Chiamata RPC per deregistrazione
	err = client.Call("NameServer.Deregister", &args, &reply)
	if err != nil {
		log.Printf("ERRORE durante la deregistrazione: %v", err)
		return
	}

	if reply.Success {
		log.Printf("Deregistrazione completata: %s", reply.Message)
	} else {
		log.Printf("Deregistrazione fallita: %s", reply.Message)
	}
}
