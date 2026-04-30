package service

import (
	"MarafoNet/internal/model"
	gameLogic "MarafoNet/internal/utils/gameLogic"
	"context"
	"encoding/json"
	"fmt"
)

type GameService struct {
	etcdService *EtcdService
}

func NewGameService(etcdService *EtcdService) *GameService {
	return &GameService{
		etcdService: etcdService,
	}
}

func (gameService *GameService) StartGame(ctx context.Context, playerNames []string) (gameId string, err error) {
	gameId, err = gameService.etcdService.GetNextGameID(ctx)
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

	if err := gameService.etcdService.PutNewGame(ctx, gameId, gameJson); err != nil {
		return "", fmt.Errorf("failed to create new game in etcd: %w", err)
	}

	return gameId, nil
}

func (gameService *GameService) IsGameEnded(gameJson []byte) (bool, error) {
	var game model.Game
	if err := json.Unmarshal(gameJson, &game); err != nil {
		return false, err
	}
	return gameLogic.IsGameEnded(game), nil
}

func (gameService *GameService) ForfeitGame(ctx context.Context, gameId string, playerName string) error {
	return gameService.applyUpdate(ctx, gameId, func(game model.Game) (model.Game, error) {
		return gameLogic.ForfeitGame(game, playerName)
	})
}

func (gameService *GameService) GetGameView(gameJson []byte, playerName string) (gameViewJson []byte, err error) {
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

func (gameService *GameService) SetTrumpSuit(ctx context.Context, gameId string, playerName string, suit model.Suit) error {
	return gameService.applyUpdate(ctx, gameId, func(game model.Game) (model.Game, error) {
		return gameLogic.SetTrumpSuit(game, playerName, suit)
	})
}

func (gameService *GameService) PlayCard(ctx context.Context, gameId string, playerName string, card model.Card) error {
	return gameService.applyUpdate(ctx, gameId, func(game model.Game) (model.Game, error) {
		return gameLogic.PlayCard(game, playerName, card)
	})
}

func (gameService *GameService) applyUpdate(ctx context.Context, gameId string, updater func(model.Game) (model.Game, error)) error {
	gameJson, revision, err := gameService.etcdService.GetGameJsonAndRevision(ctx, gameId)
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

	return gameService.etcdService.PutUpdatedGameJsonIfRevisionMatch(ctx, gameId, updatedGameJson, revision)
}
