package storage

import "fmt"

type PathBuilder interface {
	GameCounterPath() string
	GamePath(gameID string) string
	GameTimeoutPath(gameID, playerName string) string
	GameTimeoutPrefix() string

	UserQueuePath(playerName string) string
	UserQueuePrefix() string

	UserPath(playerName string) string
	UserPasswordPath(playerName string) string
	UserConnectionPath(playerName string) string
	UserCurrentGamePath(playerName string) string
}

type pathBuilderImpl struct{}

func NewPathBuilder() PathBuilder {
	return &pathBuilderImpl{}
}

func (p *pathBuilderImpl) GameCounterPath() string {
	return "global/game_counter"
}

func (p *pathBuilderImpl) GamePath(gameID string) string {
	return fmt.Sprintf("game/%s", gameID)
}

func (p *pathBuilderImpl) GameTimeoutPath(gameID, playerName string) string {
	return fmt.Sprintf("game_timeout/%s/%s", gameID, playerName)
}

func (p *pathBuilderImpl) GameTimeoutPrefix() string {
	return "game_timeout/"
}

func (p *pathBuilderImpl) UserQueuePath(playerName string) string {
	return "user_queue/" + playerName
}

func (p *pathBuilderImpl) UserQueuePrefix() string {
	return "user_queue/"
}

func (p *pathBuilderImpl) UserPath(playerName string) string {
	return fmt.Sprintf("users/%s", playerName)
}

func (p *pathBuilderImpl) UserPasswordPath(playerName string) string {
	return fmt.Sprintf("users/%s/password", playerName)
}

func (p *pathBuilderImpl) UserConnectionPath(playerName string) string {
	return fmt.Sprintf("users/%s/is_connected", playerName)
}

func (p *pathBuilderImpl) UserCurrentGamePath(playerName string) string {
	return fmt.Sprintf("users/%s/current_game", playerName)
}
