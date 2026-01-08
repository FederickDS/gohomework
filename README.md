# gohomework
Esercizio assegnato dal corso di sistemi distribuiti e cloud computing dell'anno 2025/2026.

# Traccia dell'esercizio
Sviluppare un sistema distribuito con client-side service discovery. 

1. Il server mette a disposizione 2 servizi al client. Supponiamo che il client conosca già i servizi esposti dai server e che ogni server offra tutti i servizi. Un servizio (Fibonacci) è stateless, l'altro (Counter) è stateful.
2. Un nameserver espone ai server il servizio di registrazione e deregistrazione.
3. Il nameserver espone ai client il servizio di lookup per ricevere la lista con i server che possono soddisfare i servizi.
4. Il client per scegliere il server che svolgerà il suo servizio implementa un algoritmo stateful e un algoritmo stateless. L'utente può scegliere all'inizio l'algoritmo del load balancer scrivendo per eseguire il client se l'algoritmo è stateless o se è stateful:
   ```bash
   go run client.go <stateless/stateful>
   ```

Si è scelto di esporre il server e il nameserver semplicemente assegnando a ciascuno una porta TCP separata. Se un server è già in ascolto su una porta, l'ascolto sulla stessa porta da parte di un nuovo server fallirà. Il comportamento di default è assegnare la porta successiva fino ad un numero configurabile nell'applicazione, a seguito del quale non è possibile istanziare nuovi server senza terminare almeno uno dei server già attivi. 

Un server viene deregistrato dal nameserver se viene terminato con CTRL+C.

# Modalità d'uso in locale
1. Installare Go e Redis sul proprio sistema

Nel caso di Ubuntu:
   ```bash
   sudo snap install --classic --channel=1.23/stable go

   go version

   sudo apt-get install redis
   sudo systemctl enable redis-server
   sudo systemctl start redis-server
   ```
2. Clonare la repository
3. All'interno della cartella della repository eseguire:
   ```bash
   go mod tidy
   ```
4. Avere un **terminale** per il nameserver, almeno uno per il server e uno per il client.
---
Nel terminale per il **nameserver** eseguire:
   ```bash
   go run nameserver.go
   ```
---
Nel terminale per il **server** eseguire:
   ```bash
   go run server.go [port number] [server weight]
   ```
Nel terminale per il **client** eseguire:
   ```bash
   go run client.go <stateless/stateful>
   ```
E scegliere il servizio desiderato con il numero ad esso associato (0 - numero di Fibonacci, 1 - contatore delle occorrenze).

# Sviluppatori
 * FederickDS