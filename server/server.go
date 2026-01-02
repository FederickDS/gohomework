package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	
	"project/services"
)

func main() {
	// Porta di default o da argomento
	port := ":12345"
	if len(os.Args) > 1 {
		port = ":" + os.Args[1]
	}
	
	// Crea istanza del servizio Aritmetico
	aritmeticoService := new(services.Aritmetico)
	
	// Registra il servizio con il server RPC
	server := rpc.NewServer()
	err := server.RegisterName("Aritmetico", aritmeticoService)
	if err != nil {
		log.Fatal("Failed to register service: ", err)
	}
	
	// Ascolta connessioni TCP in ingresso
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	defer lis.Close()
	
	log.Printf("RPC server listening on %s", lis.Addr().String())
	log.Println("Available services: Aritmetico.Fibonacci")
	
	// Accetta e serve le richieste
	server.Accept(lis)
}
