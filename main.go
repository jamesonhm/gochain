package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jamesonhm/gochain/internal/tasty"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// defer functions are processed LIFO, context cancel must run before scheduler shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Env Variable Load
	godotenv.Load()
	tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastySandbox)
	login := tasty.LoginInfo{
		Login:      mustEnv("TASTY_USER"),
		Password:   mustEnv("SB_PASSWORD"),
		RememberMe: true,
	}
	err := tastyClient.CreateSession(ctx, login)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Tasty Session", slog.String("error creating session", err.Error()))
	}
	logger.Info("Tasty Session", "tasty user", tastyClient.GetUser())

	customer, err := tastyClient.GetCustomer(ctx)
	fmt.Println(customer)

	accts, err := tastyClient.GetAccounts(ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", accts)
	}

	acctNum := accts[0].Account.AccountNumber

	acct, err := tastyClient.GetAccount(ctx, acctNum)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", acct)
	}

	status, err := tastyClient.GetAccountTradingStatus(ctx, acctNum)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", status)
	}

	pos, err := tastyClient.GetAccountPositions(ctx, acctNum, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", pos)
	}

	bal, err := tastyClient.GetAccountBalances(ctx, acctNum)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", bal)
	}

	//chain, err := tastyClient.GetOptionCompact(ctx, "SPY")
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", chain)
	//}

	//act := true
	//syms := []string{"XSP 250430P00529000"}
	eoSymbol := tasty.EquityOptionsSymbology{
		Symbol:     "XSP",
		OptionType: tasty.Call,
		Strike:     550,
		Expiration: time.Date(2025, 04, 25, 0, 0, 0, 0, time.UTC),
	}
	eqOpParams := tasty.EquityOptionsParams{
		Symbols: []string{eoSymbol.Build()},
	}
	eqOpts, err := tastyClient.GetEquityOptions(ctx, &eqOpParams)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", eqOpts)
	}

	//eqOpSym := tasty.EquityOptionSymbol{
	//	Active: &act,
	//}
	//eqOpt, err := tastyClient.GetEquityOption(ctx, sym, &eqOpSym)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", eqOpt)
	//}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Env variable %s required", key))
	}
	return val
}
