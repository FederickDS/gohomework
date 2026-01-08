package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"project/nameserver"
	"project/services"

	"github.com/redis/go-redis/v9"
)

const portTries int = 100

func main() {
	/*
		la porta del server o viene impostata come primo argomento,
		oppure viene messa di default con 12345
		oppure da 12345 controlla se possiamo connetterci a una porta da 12345 alle N successive,
		N numero definito come costante all'inizio del file
	*/
	port := ":12345"
	portInt, err := strconv.Atoi(strings.TrimPrefix(port, ":")) // numero di porta, in generale data la stringa della porta tolgo il prefisso e converto
	if err != nil {
		log.Fatalf("Errore di conversione:", err)
	}
	// server di default valore casuale tra 0 e 1 o da argomento
	serverWeight := rand.New(rand.NewSource(time.Now().UnixMilli())).Float64()
	// impostata priorita'massima se peso scelto casualmente non ha un peso tra 0 e 1
	if serverWeight <= 0 || serverWeight > 1 {
		serverWeight = 1.0
	}

	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}

	if len(os.Args) > 2 {
		var possibleWeight float64
		_, err := fmt.Sscanf(os.Args[2], "%f", &possibleWeight)
		if err != nil || possibleWeight <= 0 || possibleWeight > 1 {
			fmt.Printf("Peso non valido '%s', scelto peso %f\n", os.Args[2], serverWeight)
		} else {
			serverWeight = possibleWeight
		}
	}

	// configura connessione Redis
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
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Impossibile connettersi a Redis: %v", err)
	}
	fmt.Printf("Connesso a Redis su %s\n", redisAddr)

	// crea istanza del servizio Aritmetico
	aritmeticoService := new(services.Aritmetico)

	// crea istanza del servizio Contatore con Redis client
	contatoreService := &services.Contatore{
		RedisClient: redisClient,
	}

	// Registra i servizi con il server RPC
	server := rpc.NewServer()

	err = server.RegisterName("Aritmetico", aritmeticoService)
	if err != nil {
		log.Fatal("Impossibile registrare il servizio Aritmetico: ", err)
	}

	err = server.RegisterName("Contatore", contatoreService)
	if err != nil {
		log.Fatal("Impossibile registrare il servizio Contatore: ", err)
	}

	// ascolta connessioni TCP in ingresso
	var lis net.Listener //lis viene visto anche fuori, viene definito o usciamo con l'if a seguito del controllo dell'errore
	for i := 0; i < portTries; i++ {
		lis, err = net.Listen("tcp", port)
		if err == nil {
			break
		}
		portInt++
		port = fmt.Sprintf(":%d", portInt)
	}

	if err != nil {
		log.Fatal("Impossibile mettere in ascolto il server: ", err)
	} else {
		fmt.Printf("Server registrato con porta %s\n", port)
	}
	defer lis.Close()

	fmt.Printf("Server RPC in ascolto all'indirizzo %s\n", lis.Addr().String())
	fmt.Println("Servizi disponibili:")
	fmt.Println("  - Aritmetico.Fibonacci")
	fmt.Println("  - Contatore.Counter")

	// registra questo server sul NameServer prima di accettare richieste
	serverAddress := lis.Addr().String()
	registerWithNameServer(serverAddress, serverWeight)

	// setup signal handler per deregistrazione pulita
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// goroutine per gestire la deregistrazione su segnale
	go func() {
		sig := <-sigChan
		fmt.Printf("\nRicevuto segnale %v, deregistrazione in corso...\n", sig)
		deregisterFromNameServer(serverAddress)
		lis.Close()
		os.Exit(0)
	}()

	server.Accept(lis)
}

// registerWithNameServer registra questo server sul NameServer
func registerWithNameServer(serverAddr string, weight float64) {
	nameServerAddr := "localhost:9000"

	// connessione al NameServer
	client, err := rpc.Dial("tcp", nameServerAddr)
	if err != nil {
		log.Printf("Impossibile connettersi al NameServer: %v\n", err)
		log.Println("Il server continuerà comunque ad operare")
		return
	}
	defer client.Close()

	// prepara argomenti per la registrazione
	args := nameserver.RegisterArgs{
		Address: serverAddr,
		Weight:  weight,
	}
	var reply nameserver.RegisterReply

	// chiamata RPC per registrazione
	err = client.Call("NameServer.Register", &args, &reply)
	if err != nil {
		log.Fatalf("Errore durante la registrazione: %v", err)
	}

	if reply.Success {
		fmt.Printf("Registrazione completata: %s\n", reply.Message)
	} else {
		log.Fatalf("Registrazione fallita: %s", reply.Message)
	}
}

// deregisterFromNameServer deregistra questo server dal NameServer
func deregisterFromNameServer(serverAddr string) {
	nameServerAddr := "localhost:9000"

	fmt.Printf("Deregistrazione dal NameServer a %s in corso\n", nameServerAddr)

	// connessione al NameServer
	client, err := rpc.Dial("tcp", nameServerAddr)
	if err != nil {
		log.Fatalf("Impossibile connettersi al NameServer: %v", err)
	}
	defer client.Close()

	// prepara argomenti per la deregistrazione
	args := nameserver.DeregisterArgs{
		Address: serverAddr,
	}
	var reply nameserver.DeregisterReply

	// chiamata RPC per deregistrazione
	err = client.Call("NameServer.Deregister", &args, &reply)
	if err != nil {
		log.Fatalf("Errore durante la deregistrazione: %v", err)
	}

	if reply.Success {
		fmt.Printf("Deregistrazione completata: %s\n", reply.Message)
	} else {
		log.Fatalf("Deregistrazione fallita: %s", reply.Message)
	}
}
