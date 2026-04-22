package dao

import (
	"app/tables/manager"
	"app/tables/player"
	"context"
	"micro_service/services"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GamesManager struct {
	lock       *sync.RWMutex
	games      map[string]*manager.Game
	gamesIdMap map[int64]*manager.Game
}

func (gm *GamesManager) load(db *gorm.DB) {
	gm.lock.Lock()
	defer gm.lock.Unlock()

	tmp := make([]*manager.Game, 0, 128)
	db.Model(manager.Game{}).Debug().Find(&tmp)
	for _, item := range tmp {
		gm.games[item.ConfName] = item
		gm.gamesIdMap[item.Id] = item
	}
}

func (gm *GamesManager) Get(symbol string) *manager.Game {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	return gm.games[symbol]
}

func (gm *GamesManager) GetById(id int64) *manager.Game {
	gm.lock.RLock()
	defer gm.lock.RUnlock()

	return gm.gamesIdMap[id]
}

func newGameManager(mdb *gorm.DB) *GamesManager {
	tmp := &GamesManager{lock: &sync.RWMutex{}, games: make(map[string]*manager.Game), gamesIdMap: make(map[int64]*manager.Game)}
	tmp.load(mdb)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("panic", zap.Any("err", err))
			}
		}()
		ticker := time.NewTicker(5 * 60 * time.Second)
		for range ticker.C {
			tmp.load(mdb)
		}
	}()
	return tmp
}

type DBDao struct {
	db      *gorm.DB
	manager *gorm.DB
	GM      *GamesManager
}

var dbDao *DBDao = nil

func DbIns() *DBDao {
	return dbDao
}

func NewDBDao(db, manager *gorm.DB) *DBDao {
	if dbDao == nil {
		dbDao = &DBDao{db: db, manager: manager, GM: newGameManager(manager)}
	}
	return dbDao
}

func (dd *DBDao) GetPlayer(ctx context.Context, playerId uint32) (*player.Player, error) {
	var user player.Player
	err := dd.db.Model(player.Player{}).Debug().Where("user_id=?", playerId).Take(&user).Error
	return &user, err
}

func (dd *DBDao) GameManager() *GamesManager {
	return dd.GM
}

// 获取游戏列表
func (dd *DBDao) GetGameList() *services.GetGameListResp {
	dd.GM.lock.Lock()
	defer dd.GM.lock.Unlock()

	result := &services.GetGameListResp{
		Code:      services.ErrorCode_OK,
		All:       make([]*services.GetGameListResp_GameInfoItem, 0, 64),
		New:       make([]string, 0, 64),
		Hot:       make([]string, 0, 64),
		Recommend: make([]string, 0, 64),
	}
	for _, g := range dd.GM.games {
		if g.State <= 0 {
			continue
		}
		result.All = append(result.All, &services.GetGameListResp_GameInfoItem{
			Symbol: g.ConfName,
			State:  int32(g.State),
		})
		switch g.ShowType {
		case 1:
			result.Hot = append(result.Hot, g.ConfName)
		case 2:
			result.New = append(result.New, g.ConfName)
		case 3:
			result.Recommend = append(result.Recommend, g.ConfName)
		}
	}
	return result
}

// 更新游戏状态
func (dd *DBDao) ChangeGameState(symbol string, state, showType int32) {
	dd.GM.lock.Lock()
	defer dd.GM.lock.Unlock()

	g := dd.GM.games[symbol]
	if g != nil {
		g.State = int16(state)
		g.ShowType = int(showType)
	}
}

func (dd *DBDao) GetGame(symbol string) *manager.Game {
	dd.GM.lock.Lock()
	defer dd.GM.lock.Unlock()
	return dd.GM.games[symbol]
}

// 添加游戏配置
func (dd *DBDao) AddGame(g *manager.Game) {
	dd.GM.lock.Lock()
	defer dd.GM.lock.Unlock()

	dd.GM.games[g.ConfName] = g
}
