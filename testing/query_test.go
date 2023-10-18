package testing

import (
	"context"
	"encoding/json"
	"github.com/hasura/go-graphql-client"
	"log"
	"sync"
	"testing"
)

type resultOvbInq struct {
	AbcsOvbInqTrrefn struct {
		AmountCredited string `graphql:"amountCredited"`
	} `graphql:"abcsOvbInqTrrefn(tellerId: $tellerId, trrefn: $trrefn)"`
}

type ResultOvbLocal struct {
	AbcsOvbInternal struct {
		StatusCode int    `graphql:"statusCode" json:"status_code"`
		TrRefn     string `graphql:"trrefn" json:"trrefn"`
	} `graphql:"abcsOvbInternal(request: { channelId: $channelId amountTrx: $amountTrx remark: $remark tellerId: $tellerId currencyCredit: $currencyCredit currencyDebit: $currencyDebit accountDebit: $accountDebit accountCredit: $accountCredit})"`
}

type ResultSimpleOverbooking struct {
	AbcsOvbInternal struct {
		StatusCode int    `graphql:"statusCode" json:"status_code"`
		TrRefn     string `graphql:"trrefn" json:"trrefn"`
	} `graphql:"testOvbSync(request: { channelId: $channelId amountTrx: $amountTrx remark: $remark tellerId: $tellerId currencyCredit: $currencyCredit currencyDebit: $currencyDebit accountDebit: $accountDebit accountCredit: $accountCredit})"`
}

func TestHitGraphQL(t *testing.T) {
	t.Run("hit graphql", func(t *testing.T) {
		client := graphql.NewClient("https://localhost:8003/graphql", nil)

		inqOvbTrf := resultOvbInq{}
		variables := map[string]any{
			"tellerId": "0999999",
			"trrefn":   "099999911102300011840",
		}
		client.Query(context.Background(), &inqOvbTrf, variables)

		log.Println(inqOvbTrf.AbcsOvbInqTrrefn)
	})
	t.Run("test norek1 ke norek2", func(t *testing.T) {
		client := graphql.NewClient("https://localhost:8003/graphql", nil)

		result1 := ResultOvbLocal{}

		norek1 := "045202000009807"
		norek2 := "045202000001809"
		tellerId := "0999999"
		currency := "USD"
		channelId := "mb"
		remark := "test"

		// norek1 transfer ke norek2
		variables1 := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek1,
			"accountCredit":  norek2,
		}

		if err := client.Mutate(context.Background(), &result1, variables1); err != nil {
			log.Println(err.Error())
		}

		log.Println(result1.AbcsOvbInternal)
	})
	t.Run("test norek2 ke norek1", func(t *testing.T) {
		client := graphql.NewClient("https://localhost:8003/graphql", nil)

		result1 := ResultOvbLocal{}

		norek1 := "045202000009807"
		norek2 := "045202000001809"
		tellerId := "0999999"
		currency := "USD"
		channelId := "mb"
		remark := "test"

		// norek1 transfer ke norek2
		variables1 := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek2,
			"accountCredit":  norek1,
		}

		if err := client.Mutate(context.Background(), &result1, variables1); err != nil {
			log.Println(err.Error())
		}

		log.Println(result1.AbcsOvbInternal)
	})
}

func TestHitOvb(t *testing.T) {
	norek1 := "045202000009807"
	norek2 := "045202000001809"
	tellerId := "0999999"
	currency := "USD"
	channelId := "mb"
	remark := "test"

	wg := &sync.WaitGroup{}
	mtx := &sync.Mutex{}
	wg.Add(2)
	go func(wg *sync.WaitGroup, mtx *sync.Mutex) {
		defer func() {
			mtx.Unlock()
			wg.Done()
		}()

		mtx.Lock()

		result := ResultOvbLocal{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek1,
			"accountCredit":  norek2,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 1 :", string(resultJson))
		}
	}(wg, mtx) // norek1 transkfer ke norek2
	go func(wg *sync.WaitGroup, mtx *sync.Mutex) {
		defer func() {
			mtx.Unlock()
			wg.Done()
		}()

		mtx.Lock()

		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		result := ResultOvbLocal{}
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek2,
			"accountCredit":  norek1,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 2 :", string(resultJson))
		}
	}(wg, mtx) // norek2 transfer ke norek1

	wg.Wait()

	log.Println("overbooking done!!")
}

func TestSimpleOverbooking(t *testing.T) {
	norek1 := "045202000009807"
	norek2 := "045202000001809"
	tellerId := "0999999"
	currency := "USD"
	channelId := "mb"
	remark := "test"

	wg := &sync.WaitGroup{}

	wg.Add(8)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek1,
			"accountCredit":  norek2,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 1 :", string(resultJson))
		}
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek2,
			"accountCredit":  norek1,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 2 :", string(resultJson))
		}
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek1,
			"accountCredit":  norek2,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 1 :", string(resultJson))
		}
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek2,
			"accountCredit":  norek1,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 2 :", string(resultJson))
		}
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek1,
			"accountCredit":  norek2,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 1 :", string(resultJson))
		}
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek2,
			"accountCredit":  norek1,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 2 :", string(resultJson))
		}
	}(wg)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek1,
			"accountCredit":  norek2,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 1 :", string(resultJson))
		}
	}(wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		result := ResultSimpleOverbooking{}
		client := graphql.NewClient("https://localhost:8003/graphql", nil)
		variables := map[string]any{
			"channelId":      channelId,
			"amountTrx":      2.0,
			"remark":         remark,
			"tellerId":       tellerId,
			"currencyCredit": currency,
			"currencyDebit":  currency,
			"accountDebit":   norek2,
			"accountCredit":  norek1,
		}

		if err := client.Mutate(context.Background(), &result, variables); err != nil {
			log.Println(err.Error())
			return
		}
		if resultJson, err := json.Marshal(&result.AbcsOvbInternal); err == nil {
			log.Println("result 2 :", string(resultJson))
		}
	}(wg)

	wg.Wait()
	log.Println("overbooking done!")
}
