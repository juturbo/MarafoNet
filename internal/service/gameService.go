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

func (gameService *GameService) StartGame(ctx context.Context, playerNames []string) (matchId string, err error) {
	matchId, err = gameService.etcdService.GetNextMatchID(ctx)
	if err != nil {
		return "", err
	}

	match, err := gameLogic.StartGame(playerNames)
	if err != nil {
		return "", err
	}

	matchJson, err := json.Marshal(match)
	if err != nil {
		return "", err
	}

	if err := gameService.etcdService.PutNewGame(ctx, matchId, matchJson); err != nil {
		return "", fmt.Errorf("failed to create new game in etcd: %w", err)
	}

	return matchId, nil
}

func (gameService *GameService) SetTrumpSuit(ctx context.Context, matchId string, playerName string, suit model.Suit) error {
	return gameService.applyUpdate(ctx, matchId, func(m model.Game) (model.Game, error) {
		return gameLogic.SetTrumpSuit(m, playerName, suit)
	})
}

func (gameService *GameService) PlayCard(ctx context.Context, matchId string, playerName string, card model.Card) error {
	return gameService.applyUpdate(ctx, matchId, func(m model.Game) (model.Game, error) {
		return gameLogic.PlayCard(m, playerName, card)
	})
}

func (gameService *GameService) applyUpdate(ctx context.Context, matchId string, updater func(model.Game) (model.Game, error)) error {
	matchJson, revision, err := gameService.etcdService.GetMatchJsonAndRevision(ctx, matchId)
	if err != nil {
		return err
	}

	var match model.Game
	if err := json.Unmarshal(matchJson, &match); err != nil {
		return err
	}

	updatedMatch, err := updater(match)
	if err != nil {
		return fmt.Errorf("invalid move: %w", err)
	}

	updatedMatchJson, err := json.Marshal(updatedMatch)
	if err != nil {
		return err
	}

	return gameService.etcdService.PutUpdatedGameJsonIfRevisionMatch(ctx, matchId, updatedMatchJson, revision)
}
