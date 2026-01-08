package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Aritmetico struct{}

// Args e'l'indice del numero di fibonacci
type Args struct {
	Value int
}

// Result e' il numero di fibonacci corrispondente ad Args.Value
type Result struct {
	Value int
}

func (t *Aritmetico) Fibonacci(args *Args, res *Result) error {
	if args.Value < 0 {
		return errors.New("indice fibonacci non puo' essere negativo")
	}

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

// servizio stateful
// struttura di riferimento per redis
type Contatore struct {
	RedisClient *redis.Client
}

// stringa da cercare nel servizio di persistenza, che invece dipende da Contatore
type CounterArgs struct {
	Word string
}

// oltre al numero di occorrenze della stringa abbiamo un messaggio personalizzato
type CounterResult struct {
	RequestCount int
	Message      string
}

// incrementa il contatore di richieste per l'utente specificato
func (c *Contatore) Counter(args *CounterArgs, res *CounterResult) error {
	if args.Word == "" {
		return errors.New("La parola da cercare non puo' essere vuota")
	}

	ctx := context.Background()

	// chiave Redis: "user:<Word>:count"
	key := fmt.Sprintf("user:%s:count", args.Word)

	// incrementa il contatore in Redis
	count, err := c.RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("errore Redis: %v", err)
	}

	res.RequestCount = int(count)
	res.Message = fmt.Sprintf("la parola %s e'stata cercata %d volte", args.Word, count)

	return nil
}
