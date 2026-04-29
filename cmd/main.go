package main

import (
	"MarafoNet/internal/model"
	"MarafoNet/internal/service"
	user "MarafoNet/model"
	"context"
	"encoding/json"
	"log"
	"time"
)

const ETCD_SERVER_ADDRESS = "localhost:2379"
const SERVICE_DIAL_TIMEOUT = 5 * time.Second
const GAME_DIAL_TIMEOUT = 3 * time.Second
const PLAY_DIAL_TIMEOUT = 3 * time.Second

func main() {
	playerInformations := json.RawMessage(`[
		{"Name": "Alice"},
		{"Name": "Bob"},
		{"Name": "Carol"},
		{"Name": "Dave"}
	]`)
	_ = playerInformations

	etcdService, err := service.NewEtcdService([]string{ETCD_SERVER_ADDRESS}, SERVICE_DIAL_TIMEOUT)
	if err != nil {
		log.Fatalf("[Main] Etcd error: %v", err)
	}
	defer etcdService.Close()

	gameService := service.NewGameService(etcdService)
	ctx, cancel := context.WithTimeout(context.Background(), GAME_DIAL_TIMEOUT)
	defer cancel()

	watchQueueCtx, cancelWatchQueue := context.WithCancel(context.Background())
	defer cancelWatchQueue() // Assicura che il watcher venga chiuso quando la connessione cade
	updateQueueChannel, _ := etcdService.WatchUserQueue(watchQueueCtx)
	go func() {
		for users := range updateQueueChannel {
			log.Printf("[Watcher] Coda degli utenti: %v\n", users)
			for _, user := range users {
				log.Printf("[Watcher] Utente in coda: %s\n", user)
			}
		}
	}()

	player := "player1"
	err = etcdService.PutUserIntoQueue(ctx, player)
	queue, err := etcdService.GetUserQueue(ctx)
	log.Printf("[Main] queue: %v", queue)
	player = "player2"
	err = etcdService.PutUserIntoQueue(ctx, player)
	queue, err = etcdService.GetUserQueue(ctx)
	log.Printf("[Main] queue: %v", queue)
	err = etcdService.RemoveUserFromQueue(ctx, player)
	queue, err = etcdService.GetUserQueue(ctx)
	log.Printf("[Main] queue: %v", queue)

	user := user.User{Name: "Alice", Password: "ciao"}
	err = etcdService.RegisterUser(ctx, user)
	if err != nil {
		log.Printf("[Main] Error registering user: %v", err)
	}
	err = etcdService.RegisterUser(ctx, user)
	if err != nil {
		log.Printf("[Main] Error registering user: %v", err)
	}

	playerNames := []string{"Alice", "Bob", "Carol", "Dave"}
	gameId, err := gameService.StartGame(ctx, playerNames)
	if err != nil {
		log.Fatalf("[Main] Error starting game: %v", err)
	}
	log.Printf("[Main] Game created successfully! ID: %s\n", string(gameId))

	gameJson, _, err := etcdService.GetGameJsonAndRevision(context.Background(), gameId)
	if err == nil {
		log.Printf("[Main] Dati della partita: %s\n", string(gameJson))
	}
	// 2. Avviamo il Watcher per ascoltare le modifiche future
	watchCtx, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch() // Assicura che il watcher venga chiuso quando la connessione cade
	updateChannel, _ := etcdService.WatchGame(watchCtx, gameId)
	// 3. Goroutine per INVIARE i dati aggiornati al client
	go func() {
		for updatedGameJson := range updateChannel {
			// Invia il JSON aggiornato al client
			log.Printf("[Watcher] Dati aggiornati della partita: %s\n", string(updatedGameJson))
			// CONTROLLO VITTORIA: Se vuoi puoi fare un Unmarshal per verificare se c'è un vincitore
			// e magari chiudere la connessione o inviare un messaggio speciale.
			// (Oppure fai gestire la schermata di vittoria direttamente al frontend leggendo il JSON)
			var game model.Game
			err := json.Unmarshal(gameJson, &game)

			card := game.Players[0].Hand[0]
			ctxMove, cancelMove := context.WithTimeout(context.Background(), PLAY_DIAL_TIMEOUT)
			defer cancelMove()
			err = gameService.PlayCard(ctxMove, gameId, "Alice", card)
			log.Printf("Carta giocata: %s\n", card)
			if err != nil {
				log.Printf("Errore durante la mossa: %v\n", err)
			}
		}
	}()

	suit := model.Swords
	ctxMove, cancelMove := context.WithTimeout(context.Background(), PLAY_DIAL_TIMEOUT)
	defer cancelMove()
	err = gameService.SetTrumpSuit(ctxMove, gameId, "Alice", suit)
	log.Printf("[Main] Seme impostato: %s\n", suit)
	if err != nil {
		log.Printf("[Main] Errore durante la mossa: %v\n", err)
	}

	time.Sleep(3 * time.Second)

	for {
		time.Sleep(1 * time.Second)
	}
}
