package service

import (
	"MarafoNet/internal/model"
	gameLogic "MarafoNet/internal/utils/gameLogic"
	"context"
	"encoding/json"
	"errors"
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

func (gameService *GameService) StartGame(ctx context.Context, playerNames json.RawMessage) (matchId string, err error) {
	var players []model.Player
	err = json.Unmarshal(playerNames, &players)
	if err != nil {
		return "", err
	}
	matchId, err = gameService.etcdService.GetNextMatchID(ctx)
	if err != nil {
		return "", err
	}
	match, err := gameLogic.StartGame(players)
	if err != nil {
		return "", err
	}
	matchJson, err := json.Marshal(match)
	if err != nil {
		return "", err
	}
	success, err := gameService.etcdService.PutNewGame(ctx, matchId, matchJson)
	if err != nil || !success {
		return "", errors.New("failed to create new game in etcd")
	}
	return matchId, nil
}

func (gameService *GameService) SetTrumpSuit(ctx context.Context, matchId string, playerName string, suit model.Suit) error {
	err := gameService.applyUpdate(ctx, matchId, func(m model.Match) (model.Match, error) {
		return gameLogic.SetTrumpSuit(m, playerName, suit)
	})
	if err != nil {
		return fmt.Errorf("failed to set trump suit: %w", err)
	}
	return nil
}

func (gameService *GameService) PlayCard(ctx context.Context, matchId string, playerName string, card model.Card) error {
	err := gameService.applyUpdate(ctx, matchId, func(m model.Match) (model.Match, error) {
		return gameLogic.PlayCard(m, playerName, card)
	})
	if err != nil {
		return fmt.Errorf("failed to play card: %w", err)
	}
	return nil
}

func (gameService *GameService) applyUpdate(ctx context.Context, matchId string, updater func(model.Match) (model.Match, error)) error {
	matchJson, revision, err := gameService.etcdService.GetValueAndRevision(ctx, matchId)
	if err != nil {
		return err
	}
	var match model.Match
	if err := json.Unmarshal(matchJson, &match); err != nil {
		return err
	}
	updatedMatch, err := updater(match)
	if err != nil {
		return fmt.Errorf("mossa non valida: %w", err)
	}
	updatedMatchJson, err := json.Marshal(updatedMatch)
	if err != nil {
		return err
	}
	success, err := gameService.etcdService.PutUpdatedGameIfRevisionMatch(ctx, matchId, updatedMatchJson, revision)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("failed to update match")
	}
	return nil
}
