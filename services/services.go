package services

import "errors"

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
