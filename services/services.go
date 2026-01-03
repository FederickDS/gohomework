package services

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/redis/go-redis/v9"
)

type Aritmetico struct{}

// Valori di input dipendono dal numero massimo di argomenti forniti da funzione
type Args struct {
	Value int
}

// Valore di ritorno per metodo Fibonacci
type Result struct {
	Value int
}

func (t *Aritmetico) Fibonacci(args *Args, res *Result) error {
	if args.Value < 0 {
		return errors.New("indice fibonacci non puo' essere negativo")
	}
	
	// Casi immediati (avendo visto che sono >= 0)
	if args.Value < 2 {
		res.Value = args.Value
		return nil
	}
	
	var f1, f2 int = 0, 1
	for i := 2; i <= args.Value; i++ {
		f1, f2 = f2, f1+f2
	}
	
	res.Value = f2
	return nil
}

// --- SERVIZIO STATEFUL CON REDIS ---

// Contatore service per tenere traccia delle richieste per utente
type Contatore struct {
	RedisClient *redis.Client
}

// CounterArgs - Input per il servizio Counter
type CounterArgs struct {
	Username string
	Password string
}

// CounterResult - Risultato del servizio Counter
type CounterResult struct {
	RequestCount int
	Message      string
}

// Counter incrementa il contatore di richieste per l'utente specificato
func (c *Contatore) Counter(args *CounterArgs, res *CounterResult) error {
	if args.Username == "" || args.Password == "" {
		return errors.New("username e password non possono essere vuoti")
	}
	
	ctx := context.Background()
	
	// Chiave Redis: "user:<username>:count"
	key := fmt.Sprintf("user:%s:count", args.Username)
	
	// Incrementa il contatore in Redis
	count, err := c.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("errore Redis: %v", err)
	}
	
	res.RequestCount = int(count)
	res.Message = fmt.Sprintf("Richiesta #%d per l'utente %s", count, args.Username)
	
	return nil
}
