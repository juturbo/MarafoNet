package service

import (
	"MarafoNet/internal/model"
	"MarafoNet/internal/repository"
	gameLogic "MarafoNet/internal/utils/gamelogic"
	"context"
	"encoding/json"
	"fmt"
)

type GameService interface {
	StartGame(ctx context.Context, playerNames []string) (gameId string, err error)
	IsGameEnded(gameJson []byte) (bool, error)
	GetGameView(gameJson []byte, playerName string) (gameViewJson []byte, err error)
	ForfeitGame(ctx context.Context, gameId string, playerName string) error
	SetTrumpSuit(ctx context.Context, gameId string, playerName string, suit model.Suit) error
	PlayCard(ctx context.Context, gameId string, playerName string, card model.Card) error
}

type gameService struct {
	storage repository.GameStorage
}

func NewGameService(storage repository.GameStorage) GameService {
	return &gameService{
		storage: storage,
	}
}

func (gameSvc *gameService) StartGame(ctx context.Context, playerNames []string) (gameId string, err error) {
	gameId, err = gameSvc.storage.GetNextGameID()
	if err != nil {
		return "", err
	}

	game, err := gameLogic.StartGame(playerNames)
	if err != nil {
		return "", err
	}

	gameJson, err := json.Marshal(game)
	if err != nil {
		return "", err
	}

	if err := gameSvc.storage.PutNewGame(ctx, gameId, gameJson); err != nil {
		return "", fmt.Errorf("failed to create new game in etcd: %w", err)
	}

	return gameId, nil
}

func (gameSvc *gameService) IsGameEnded(gameJson []byte) (bool, error) {
	var game model.Game
	if err := json.Unmarshal(gameJson, &game); err != nil {
		return false, err
	}
	return gameLogic.IsGameEnded(game), nil
}

func (gameSvc *gameService) GetGameView(gameJson []byte, playerName string) (gameViewJson []byte, err error) {
	var game model.Game
	if err = json.Unmarshal(gameJson, &game); err != nil {
		return nil, err
	}
	gameView, err := game.ViewForPlayer(playerName)
	if err != nil {
		return nil, err
	}
	gameViewJson, err = json.Marshal(gameView)
	if err != nil {
		return nil, err
	}
	return gameViewJson, nil
}

func (gameSvc *gameService) ForfeitGame(ctx context.Context, gameId string, playerName string) error {
	return gameSvc.applyUpdate(ctx, gameId, func(game model.Game) (model.Game, error) {
		return gameLogic.ForfeitGame(game, playerName)
	})
}

func (gameSvc *gameService) SetTrumpSuit(ctx context.Context, gameId string, playerName string, suit model.Suit) error {
	return gameSvc.applyUpdate(ctx, gameId, func(game model.Game) (model.Game, error) {
		return gameLogic.SetTrumpSuit(game, playerName, suit)
	})
}

func (gameSvc *gameService) PlayCard(ctx context.Context, gameId string, playerName string, card model.Card) error {
	return gameSvc.applyUpdate(ctx, gameId, func(game model.Game) (model.Game, error) {
		return gameLogic.PlayCard(game, playerName, card)
	})
}

func (gameSvc *gameService) applyUpdate(ctx context.Context, gameId string, updater func(model.Game) (model.Game, error)) error {
	gameJson, revision, err := gameSvc.storage.GetGameJsonAndRevision(ctx, gameId)
	if err != nil {
		return err
	}

	var game model.Game
	if err := json.Unmarshal(gameJson, &game); err != nil {
		return err
	}

	updatedGame, err := updater(game)
	if err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	updatedGameJson, err := json.Marshal(updatedGame)
	if err != nil {
		return err
	}

	return gameSvc.storage.PutUpdatedGameJsonIfRevisionMatch(ctx, gameId, updatedGameJson, revision)
}
