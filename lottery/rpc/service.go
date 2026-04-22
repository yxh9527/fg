package rpc

import (
	"app/config"
	"app/entity"
	"app/entity/view"
	"app/tables/player"
	"context"
	"crypto/md5"
	"fmt"
	"lottery/dao"
	"strconv"
	"strings"
	"sync"
	"time"

	"micro_service/services"

	jsoniter "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type RecordItem struct {
	record  *entity.CacheRecordsReq
	TimeOut int64
}

type RecordCacheMgr struct {
	lock    *sync.RWMutex
	records map[string]*RecordItem
}

type GameObject struct {
	Record *entity.ClientRecordsReq `json:"record"`
}

type LotteryService struct {
	services.UnimplementedLotteryServiceServer
	db  *dao.DBDao
	rds *dao.RedisDao
	es  *dao.ESDao

	pcr        *PoolChangeRecord
	poolChange chan *view.PoolLogItem
	RecordChan chan *entity.CacheRecordsReq
	BillChan   chan *entity.CacheBillsReq

	recordsCache *RecordCacheMgr
}

type PoolChangeRecord struct {
	lock   *sync.RWMutex
	record map[string]decimal.Decimal //key: agentId-symbol value:pool value
}

func (p *PoolChangeRecord) Record(agentId int64, symbol string, v decimal.Decimal) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.record[fmt.Sprintf("%d-%s", agentId, symbol)] = v
}

func (p *PoolChangeRecord) Reset() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.record = make(map[string]decimal.Decimal)
}

func NewLotteryService(es *elastic.Client) *LotteryService {
	tmp := &LotteryService{
		db:         dao.NewDBDao(),
		rds:        dao.RedisIns(),
		es:         dao.NewESDao(es),
		poolChange: make(chan *view.PoolLogItem, 1024),
		RecordChan: make(chan *entity.CacheRecordsReq, 10240),
		BillChan:   make(chan *entity.CacheBillsReq, 1024),
		pcr: &PoolChangeRecord{
			lock:   &sync.RWMutex{},
			record: make(map[string]decimal.Decimal),
		},
	}
	tmp.initRecordsCache()
	tmp.producterPoolLog()
	tmp.consumerPool()
	tmp.consumerRecord()
	tmp.consumerBill()
	return tmp
}

func ConvertUserEntityToHumanPlayer(p *player.Player) *services.HumanPlayer {
	return &services.HumanPlayer{
		Id:             uint32(p.UserId),
		Nickname:       p.NickName,
		GameCurrency:   p.Score.StringFixed(2),
		Avatar:         p.Pic,
		Gender:         uint32(p.Sex),
		Exp:            p.Exp,
		AgentId:        uint32(p.ProxyId),
		LoginIP:        p.LoginIp,
		LoginTimeStamp: p.LoginTime,
		CurrencyLimit:  p.MoneyLimit.StringFixed(2),
		WebSiteId:      uint32(p.WebsiteId),
		Account:        p.Account,
		CurrencyType:   p.CurrencyType,
		AllTimes:       p.AllTimes,
	}
}

func (d *LotteryService) loadPlayerToCache(id uint32) services.ErrorCode {
	playerInfo, err := d.db.GetPlayer(id)
	if err != nil {
		zap.L().Error("load player from db failed", zap.Any("id", id), zap.Error(err))
		return services.ErrorCode_SYSTEM_ERROR
	}
	if err := d.rds.SetPlayer(ConvertUserEntityToHumanPlayer(playerInfo)); err != nil {
		zap.L().Error("set player cache failed", zap.Any("id", id), zap.Error(err))
		return services.ErrorCode_SYSTEM_ERROR
	}
	return services.ErrorCode_OK
}

func (d *LotteryService) getPlayerCurrency(id uint32) (int64, services.ErrorCode) {
	newCurrency, err := d.rds.GetPlayerCurrency(id)
	if err == nil {
		return newCurrency, services.ErrorCode_OK
	}
	if err == redis.Nil {
		if code := d.loadPlayerToCache(id); code != services.ErrorCode_OK {
			return 0, code
		}
		newCurrency, err = d.rds.GetPlayerCurrency(id)
		if err == nil {
			return newCurrency, services.ErrorCode_OK
		}
	}
	zap.L().Error("get player currency failed", zap.Any("id", id), zap.Error(err))
	return 0, services.ErrorCode_SYSTEM_ERROR
}

func validateUserRecordInfo(ur *entity.UserRecordInfo) bool {
	return ur != nil && ur.Common != nil && ur.BetRecord != nil
}

func (d *LotteryService) SaveBill(agentId, playerId uint32, delta decimal.Decimal, currencyScore float64, symbol, desc, currencyType string, roundID string) {
	now := time.Now()
	billNo := fmt.Sprintf("L%04d%02d%02d%02d%02d%02d%07d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()%10000000)
	eGame := dao.GamesManagerIns().Get(symbol)
	bill := &entity.CacheBillsReq{
		UserId:         playerId,
		GameId:         uint32(eGame.Number),
		AgentId:        uint32(agentId),
		Bet:            delta.InexactFloat64(),
		CurrentScore:   currencyScore,
		Currency:       currencyType,
		CreateTime:     now.Unix(),
		RoundID:        roundID,
		FlowingWaterOn: billNo,
		Symbol:         symbol,
		Desc:           desc,
	}
	d.BillChan <- bill
}

func (d *LotteryService) SaveRecord(record *entity.CacheRecordsReq) *entity.CacheRecordsReq {
	d.recordsCache.lock.Lock()
	defer d.recordsCache.lock.Unlock()

	hashStr := fmt.Sprintf("%d|%d|%s", record.AgentId, record.UserId, record.RoundID)
	record.Hash = fmt.Sprintf("%x", md5.Sum([]byte(hashStr)))
	d.recordsCache.records[record.Hash] = &RecordItem{
		TimeOut: time.Now().Unix() + 10,
		record:  record,
	}
	return record
}

func (d *LotteryService) initRecordsCache() {
	if d.recordsCache == nil {
		d.recordsCache = &RecordCacheMgr{
			lock:    &sync.RWMutex{},
			records: make(map[string]*RecordItem),
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					zap.L().Error("panic", zap.Any("err", err))
				}
			}()
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				tmp := make([]*entity.CacheRecordsReq, 0, 512)
				d.recordsCache.lock.Lock()
				now := time.Now()
				for key, item := range d.recordsCache.records {
					tmp = append(tmp, item.record)
					if now.Unix() > item.TimeOut || item.record.Complete {
						delete(d.recordsCache.records, key)
					}
				}
				d.recordsCache.lock.Unlock()
				for _, item := range tmp {
					d.RecordChan <- item
				}
			}
		}()
	}
}

func (d *LotteryService) consumerPool() {
	gw := &sync.WaitGroup{}
	gw.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("panic", zap.Any("err", err))
			}
		}()
		gw.Done()
		data := make([]*view.PoolLogItem, 0, 64)
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				if len(data) > 0 {
					d.BulkPoolLog(data)
					data = make([]*view.PoolLogItem, 0, 64)
				}
			case req := <-d.poolChange:
				data = append(data, req)
				if len(data) >= 32 {
					d.BulkPoolLog(data)
					data = make([]*view.PoolLogItem, 0, 64)
				}
			}
		}
	}()
	gw.Wait()
}

func (d *LotteryService) producterPoolLog() {
	gw := &sync.WaitGroup{}
	gw.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("panic", zap.Any("err", err))
			}
		}()
		gw.Done()
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			d.pcr.lock.Lock()
			for k, v := range d.pcr.record {
				arr := strings.Split(k, "-")
				agentId, _ := strconv.ParseInt(arr[0], 10, 64)
				symbol := arr[1]
				pcfg := config.CfgIns.GetPoolCfg(agentId, symbol)
				if pcfg != nil {
					d.poolChange <- &view.PoolLogItem{
						AgentId:    int(agentId),
						Symbol:     symbol,
						PoolValue:  v.Truncate(2).InexactFloat64(),
						Normal:     int(pcfg.Pool[1].Normal.IntPart()),
						NormalRate: pcfg.Pool[1].NormalRate,
						Min:        int(pcfg.Pool[1].Min.IntPart()),
						MinRate:    pcfg.Pool[1].MinRate,
						Max:        int(pcfg.Pool[1].Max.IntPart()),
						MaxRate:    pcfg.Pool[1].MaxRate,
						Ctl:        int(pcfg.Pool[1].Control.IntPart()),
						Revenue:    pcfg.Pool[1].Revenue,
						CreateTime: time.Now().Unix(),
					}
				}
			}
			d.pcr.record = make(map[string]decimal.Decimal)
			d.pcr.lock.Unlock()
		}
	}()
	gw.Wait()
}

func (d *LotteryService) consumerRecord() {
	gw := &sync.WaitGroup{}
	gw.Add(1)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				zap.L().Error("syncSaveGameState,数据落地协程panic", zap.Any("recover", e))
			}
		}()
		gw.Done()
		data := make([]*entity.CacheRecordsReq, 0, 64)
		t := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-t.C:
				if len(data) > 0 {
					d.es.BulkRecordsSave(data)
					data = make([]*entity.CacheRecordsReq, 0, 64)
				}
			case req := <-d.RecordChan:
				data = append(data, req)
				if len(data) >= 16 {
					d.es.BulkRecordsSave(data)
					data = make([]*entity.CacheRecordsReq, 0, 64)
				}
			}
		}
	}()
	gw.Wait()
}

func (d *LotteryService) consumerBill() {
	gw := &sync.WaitGroup{}
	gw.Add(1)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				zap.L().Error("syncSaveGameState,数据落地协程panic", zap.Any("recover", e))
			}
		}()
		gw.Done()
		data := make([]*entity.CacheBillsReq, 0, 64)
		t := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-t.C:
				if len(data) > 0 {
					d.es.BulkBillsSave(data)
					data = make([]*entity.CacheBillsReq, 0, 64)
				}
			case req := <-d.BillChan:
				data = append(data, req)
				if len(data) >= 16 {
					d.es.BulkBillsSave(data)
					data = make([]*entity.CacheBillsReq, 0, 64)
				}
			}
		}
	}()
	gw.Wait()
}

func (d *LotteryService) BulkPoolLog(data []*view.PoolLogItem) error {
	bulkService := d.es.Client.Bulk()
	records := make([]elastic.BulkableRequest, 0)
	for _, req := range data {
		records = append(records, elastic.NewBulkIndexRequest().Index("pp_pool_record_log").Doc(req))
	}
	bulkService.Add(records...)
	_, err := bulkService.Do(context.Background())
	if err != nil {
		zap.L().Error("BulkPoolLog,批量插入数据失败", zap.Any("err", err), zap.Any("data", data))
	}
	return nil
}

func (d *LotteryService) updatePlayerCurrency(id uint32, delta int64) (int64, services.ErrorCode) {
	newCurrency, err := d.rds.UpdatePlayerCurrency(id, delta, 0, 0, 0)
	if err != nil {
		if err == redis.Nil {
			if code := d.loadPlayerToCache(id); code != services.ErrorCode_OK {
				return 0, code
			}
			newCurrency, err = d.rds.UpdatePlayerCurrency(id, delta, 0, 0, 0)
			if err != nil {
				zap.L().Error("update player currency failed", zap.Any("id", id), zap.Error(err))
				return 0, services.ErrorCode_SYSTEM_ERROR
			}
			return newCurrency, services.ErrorCode_OK
		}
		zap.L().Error("update player currency failed", zap.Any("id", id), zap.Error(err))
		return 0, services.ErrorCode_SYSTEM_ERROR
	}
	return newCurrency, services.ErrorCode_OK
}

func (d *LotteryService) SlotsBet(webId uint32, exchange decimal.Decimal, ur *entity.UserRecordInfo, req *services.SlotsLotteryReq) (int64, bool, services.ErrorCode) {
	var newCurrency int64 = 0
	award, _ := decimal.NewFromString(req.ProfitLoss)
	bet, _ := decimal.NewFromString(req.Bet)
	exBet := bet.Mul(exchange)
	exAward := award.Mul(exchange)
	awardMax, _ := decimal.NewFromString(req.MaxProfitLoss)
	exAwardMax := awardMax.Mul(exchange)
	if exAwardMax.GreaterThan(decimal.Zero) {
		exAward = exAwardMax
		award = awardMax
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame.Number != int(req.GameId) {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", ur.Common.RecordId), zap.Any("req", req))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}
	pc := config.CfgIns.GetPoolCfg(req.AgentId, eGame.ConfName)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", ur.Common.RecordId), zap.Any("pc", pc))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}
	b := false
	zap.L().Debug("Bet:下注", zap.Any("agentId", req.AgentId),
		zap.Any("symbol", eGame.ConfName),
		zap.Any("gameId", req.GameId),
		zap.Any("roundId", ur.Common.RecordId),
		zap.Any("playerId", req.PlayerId),
		zap.Any("bet", bet),
		zap.Any("award", award),
		zap.Any("awardMax", awardMax),
		zap.Any("currenType", req.CurrencyType))
	if award.GreaterThan(decimal.Zero) {
		if bet.LessThan(award) {
			b = true
			_, ok := dao.CacheIns().Lottery(int64(req.AgentId), int32(req.PlayerId), pc, eGame.ConfName, req.CurrencyType, exBet, exAward, ur.Common.RecordId)
			if !ok {
				return 0, false, services.ErrorCode_NO_ENOUGH_POOL_MONEY
			}
		}
	}
	if bet.GreaterThan(decimal.Zero) {
		tmp, errCode := d.updatePlayerCurrency(req.PlayerId, bet.Neg().Mul(decimal.NewFromInt(100)).IntPart())
		if errCode != services.ErrorCode_OK {
			zap.L().Debug("Bet:下注失败,更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("gameId", req.GameId),
				zap.Any("roundId", ur.Common.RecordId),
				zap.Any("playerId", req.PlayerId),
				zap.Any("bet", bet),
				zap.Any("award", award),
				zap.Any("currenType", req.CurrencyType))
			return 0, false, errCode
		}
		newCurrency = tmp
	}
	nc := decimal.NewFromInt(newCurrency).Div(decimal.NewFromInt(100))
	if !b {
		dao.CacheIns().ChangePool(int64(req.AgentId), int32(req.PlayerId), eGame.ConfName, req.CurrencyType, req.RoundID, exBet, exAward, pc.Pool[1].Revenue)
	}
	if exAwardMax.GreaterThan(decimal.Zero) {
		dao.CacheIns().SaveRoundData(int64(req.AgentId), ur.Common.RecordId, exAwardMax, req.PlayerId)
	}
	if bet.GreaterThan(decimal.Zero) {
		d.SaveBill(uint32(req.AgentId), req.PlayerId, bet.Neg(), nc.Truncate(2).InexactFloat64(), eGame.ConfName, "下注", req.CurrencyType, ur.Common.RecordId)
	}
	d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
	zap.L().Debug("Bet:下注成功",
		zap.Any("agentId", req.AgentId),
		zap.Any("symbol", eGame.ConfName),
		zap.Any("roundId", ur.Common.RecordId),
		zap.Any("playerId", req.PlayerId))
	return newCurrency, true, services.ErrorCode_OK
}

func (d *LotteryService) Complete(webId uint32, exchange decimal.Decimal, ur *entity.UserRecordInfo, req *services.SlotsLotteryReq) (int64, bool, services.ErrorCode) {
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame.Number != int(req.GameId) {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", ur.Common.RecordId), zap.Any("req", req))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}

	bet := decimal.NewFromFloat(ur.BetRecord.TotalBetGold)
	award, _ := decimal.NewFromString(req.ProfitLoss)

	zap.L().Debug("Complete:收到注单结束请求",
		zap.Any("agentId", req.AgentId),
		zap.Any("playerId", req.PlayerId),
		zap.Any("roundId", ur.Common.RecordId),
		zap.Any("symbol", eGame.ConfName))

	pc := config.CfgIns.GetPoolCfg(req.AgentId, eGame.ConfName)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", ur.Common.RecordId), zap.Any("pc", pc))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}

	var newCurrency int64
	var code services.ErrorCode
	if award.GreaterThan(decimal.Zero) {
		newCurrency, code = d.updatePlayerCurrency(req.PlayerId, award.Mul(decimal.NewFromInt(100)).IntPart())
		if code != services.ErrorCode_OK {
			zap.L().Debug("Award:返奖失败,更新玩家积分失败!",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("gameId", req.GameId),
				zap.Any("roundId", ur.Common.RecordId),
				zap.Any("playerId", req.PlayerId),
				zap.Any("award", award),
				zap.Any("currenType", req.CurrencyType))
			return 0, false, code
		}
	} else {
		newCurrency, code = d.getPlayerCurrency(req.PlayerId)
		if code != services.ErrorCode_OK {
			return 0, false, code
		}
	}

	nc := decimal.NewFromInt(newCurrency).Div(decimal.NewFromInt(100))
	record := ConvertRecord(uint32(req.AgentId), req.PlayerId, ur.Common.RecordId, req.CurrencyType, eGame.ConfName, req.Account, req.State, nc, uint32(webId), req.Complete, ur.BetRecord.TotalBetGold, award.InexactFloat64())
	d.SaveRecord(record)

	zap.L().Debug("Award:返奖", zap.Any("agentId", req.AgentId),
		zap.Any("symbol", eGame.ConfName),
		zap.Any("roundId", ur.Common.RecordId),
		zap.Any("playerId", req.PlayerId),
		zap.Any("gameId", req.GameId),
		zap.Any("award", award),
		zap.Any("bet", bet),
		zap.Any("exAward", award),
		zap.Any("exBet", bet))

	d.SaveBill(uint32(req.AgentId), req.PlayerId, award, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "返奖", req.CurrencyType, ur.Common.RecordId)

	zap.L().Debug("Complete:游戏结束", zap.Any("agentId", req.AgentId), zap.Any("gameId", req.GameId), zap.Any("symbol", eGame.ConfName), zap.Any("roundId", ur.Common.RecordId), zap.Any("playerId", req.PlayerId), zap.Any("exAward", award), zap.Any("exBet", bet))
	dao.CacheIns().Complete(int64(req.AgentId), req.PlayerId, eGame.ConfName, bet.Mul(exchange), award.Mul(exchange), pc.Pool[1].Revenue)

	if ri := dao.CacheIns().FinishRoundData(int64(req.AgentId), ur.Common.RecordId); ri != nil {
		delta := ri.MaxPay.Round(2).Sub(award.Mul(exchange).Truncate(2))
		if delta.GreaterThanOrEqual(decimal.Zero) {
			zap.L().Debug("Complete:返还水池多扣的积分",
				zap.Any("agentId", req.AgentId),
				zap.Any("gameId", req.GameId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", ur.Common.RecordId),
				zap.Any("playerId", req.PlayerId),
				zap.Any("awardMax", ri.MaxPay),
				zap.Any("delta", delta))
			dao.CacheIns().ReturnPool(ri.AgentId, req.PlayerId, eGame.ConfName, delta)
		} else {
			zap.L().Error("返奖异常，预扣值比实际获奖小！！！",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", ur.Common.RecordId),
				zap.Any("playerId", req.PlayerId),
				zap.Any("award", award),
				zap.Any("ri", ri))
		}
	}
	return newCurrency, true, services.ErrorCode_OK
}

func ConvertRecord(agentId, userId uint32, recordId, currencyType, symbol, account, log string, newCurrency decimal.Decimal, webId uint32, complete bool, totalBet, win float64) *entity.CacheRecordsReq {
	rate, _ := config.CfgIns.GetExchange(currencyType)
	p := config.CfgIns.GetPoolCfg(int64(agentId), symbol)
	bet := decimal.NewFromFloat(totalBet)
	award := decimal.NewFromFloat(win)
	chips := bet
	if chips.LessThan(award) {
		chips = award
	}
	r := bet.Mul(p.Pool[1].Revenue)
	record := &entity.CacheRecordsReq{
		WebId:          webId,
		UserId:         userId,
		AgentId:        uint32(agentId),
		GameId:         uint32(p.GameId),
		Account:        account,
		NickName:       account,
		Bet:            bet.Truncate(4).InexactFloat64(),
		ExBet:          bet.Mul(rate).Truncate(4).InexactFloat64(),
		Currency:       currencyType,
		CurrencySymbol: currencyType,
		BaseBet:        totalBet,
		Win:            award.Truncate(4).InexactFloat64(),
		ExWin:          award.Mul(rate).Truncate(4).InexactFloat64(),
		PlayedDate:     time.Now().UnixMilli(),
		RoundID:        recordId,
		Symbol:         symbol,
		RowVersion:     time.Now().UnixNano(),
		Revenue:        r.Truncate(4).InexactFloat64(),
		ExRevenue:      r.Mul(rate).Truncate(4).InexactFloat64(),
		Log:            log,
		GameName:       p.Name,
		Balance:        newCurrency.Truncate(4).InexactFloat64(),
		BalanceCash:    newCurrency.Truncate(4).InexactFloat64(),
		Chips:          chips.Mul(rate).Truncate(4).InexactFloat64(),
		Complete:       complete,
	}
	return record
}

func (d *LotteryService) SlotsLottery(_ context.Context, req *services.SlotsLotteryReq) (resp *services.SlotsLotteryResp, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			zap.L().Error("panic", zap.Any("err", rec))
			if resp == nil {
				resp = &services.SlotsLotteryResp{}
			}
			resp.Result = false
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			resp.NewCurrency = ""
			err = nil
		}
	}()

	resp = &services.SlotsLotteryResp{Code: services.ErrorCode_OK}
	bet, _ := decimal.NewFromString(req.Bet)
	resp.Result = true
	resp.Code = services.ErrorCode_OK
	eAgent := dao.AgentManagerIns().Get(req.AgentId)
	if eAgent == nil {
		resp.Code = services.ErrorCode_AGENT_FROZEN
		resp.Result = false
		zap.L().Debug("获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.PlayerId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		resp.Result = false
		zap.L().Debug("获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.PlayerId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Result = false
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.PlayerId),
			zap.Any("gameId", req.GameId),
			zap.Any("symbol", eGame.ConfName))
		return resp, nil
	}

	ur := &entity.UserRecordInfo{}
	err = jsoniter.UnmarshalFromString(req.State, ur)
	if err == nil && !validateUserRecordInfo(ur) {
		zap.L().Error("invalid user record info",
			zap.Any("userId", req.PlayerId),
			zap.Any("symbol", eGame.ConfName),
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId),
			zap.Any("state", req.State))
		resp.Result = false
		resp.Code = services.ErrorCode_PARAMS_INVALID
		return resp, nil
	}
	if err != nil {
		zap.L().Error("从游戏状态中获取注单信息失败",
			zap.Any("userId", req.PlayerId),
			zap.Any("symbol", eGame.ConfName),
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId),
			zap.Any("state", req.State))
		resp.Result = false
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		return resp, nil
	}

	if bet.GreaterThan(decimal.Zero) {
		newCurrency, ok, code := d.SlotsBet(uint32(eAgent.WebId), exchange, ur, req)
		if !ok {
			resp.Result = false
			resp.Code = code
			return resp, nil
		}
		resp.NewCurrency = decimal.NewFromFloat(float64(newCurrency) / 100).String()
	}

	if req.Complete {
		newCurrency, ok, code := d.Complete(uint32(eAgent.WebId), exchange, ur, req)
		if !ok {
			resp.Result = false
			resp.Code = code
			return resp, nil
		}
		resp.NewCurrency = decimal.NewFromFloat(float64(newCurrency) / 100).String()
	}

	return resp, nil
}
