package rpc

import (
	"app/config"
	"app/entity"
	"app/entity/view"
	"app/esindex"
	"app/tables/manager"
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
		poolChange: make(chan *view.PoolLogItem, 10240*5),
		RecordChan: make(chan *entity.CacheRecordsReq, 10240*5),
		BillChan:   make(chan *entity.CacheBillsReq, 10240*5),
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
		IsTourist:      p.IsTourist,
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

// SaveBill 写流水：bet 为下注（负），award 为返奖（>=0）。
func (d *LotteryService) SaveBill(agentId, playerId uint32, bet, award decimal.Decimal, currencyScore float64, symbol, desc, currencyType string, roundID string) {
	now := time.Now()
	billNo := fmt.Sprintf("L%04d%02d%02d%02d%02d%02d%07d", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()%10000000)
	eGame := dao.GamesManagerIns().Get(symbol)
	if award.LessThan(decimal.Zero) {
		award = decimal.Zero
	}
	bill := &entity.CacheBillsReq{
		UserId:         playerId,
		GameId:         uint32(eGame.Number),
		AgentId:        uint32(agentId),
		Bet:            bet.InexactFloat64(),
		Award:          award.InexactFloat64(),
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
				if len(data) >= 40 {
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
				if len(data) >= 50 {
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
		records = append(records, elastic.NewBulkIndexRequest().Index(esindex.PoolRecordLog()).Doc(req))
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

func (d *LotteryService) SlotsBet(webId uint32, exchange decimal.Decimal, req *slotsLotteryReq) (int64, bool, services.ErrorCode) {
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
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("req", req))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}
	pc := config.CfgIns.GetPoolCfg(req.AgentId, eGame.ConfName)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}
	b := false
	zap.L().Debug("Bet:下注", zap.Any("agentId", req.AgentId),
		zap.Any("symbol", eGame.ConfName),
		zap.Any("gameId", req.GameId),
		zap.Any("roundId", req.RoundID),
		zap.Any("playerId", req.PlayerId),
		zap.Any("bet", bet),
		zap.Any("award", award),
		zap.Any("awardMax", awardMax),
		zap.Any("currenType", req.CurrencyType))
	if award.GreaterThan(decimal.Zero) {
		if bet.LessThan(award) {
			b = true
			_, ok := dao.CacheIns().Lottery(int64(req.AgentId), int32(req.PlayerId), pc, eGame.ConfName, req.CurrencyType, exBet, exAward, req.RoundID)
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
				zap.Any("roundId", req.RoundID),
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
		dao.CacheIns().SaveRoundData(int64(req.AgentId), req.RoundID, exAwardMax, req.PlayerId)
	}
	if bet.GreaterThan(decimal.Zero) {
		d.SaveBill(uint32(req.AgentId), req.PlayerId, bet.Neg(), decimal.Zero, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "下注", req.CurrencyType, req.RoundID)
	}
	d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
	zap.L().Debug("Bet:下注成功",
		zap.Any("agentId", req.AgentId),
		zap.Any("symbol", eGame.ConfName),
		zap.Any("roundId", req.RoundID),
		zap.Any("playerId", req.PlayerId))
	return newCurrency, true, services.ErrorCode_OK
}

func (d *LotteryService) Complete(webId uint32, exchange decimal.Decimal, req *slotsLotteryReq) (int64, bool, services.ErrorCode) {
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame.Number != int(req.GameId) {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("req", req))
		return 0, false, services.ErrorCode_SYSTEM_ERROR
	}

	bet, _ := decimal.NewFromString(req.Bet)
	award, _ := decimal.NewFromString(req.ProfitLoss)

	zap.L().Debug("Complete:收到注单结束请求",
		zap.Any("agentId", req.AgentId),
		zap.Any("playerId", req.PlayerId),
		zap.Any("roundId", req.RoundID),
		zap.Any("symbol", eGame.ConfName))

	pc := config.CfgIns.GetPoolCfg(req.AgentId, eGame.ConfName)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
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
				zap.Any("roundId", req.RoundID),
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
	record := ConvertRecord(uint32(req.AgentId), req.PlayerId, req.RoundID, req.CurrencyType, eGame.ConfName, req.Account, req.State, nc, uint32(webId), req.Complete, bet.InexactFloat64(), award.InexactFloat64())
	d.SaveRecord(record)
	if req.Complete {
		zap.L().Info("Complete:注单信息",
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.PlayerId),
			zap.Any("gameId", req.GameId),
			zap.Any("symbol", eGame.ConfName),
			zap.Any("roundId", req.RoundID),
			zap.Any("record", record),
			zap.Any("userRecord", req.State))
	}

	zap.L().Debug("Award:返奖", zap.Any("agentId", req.AgentId),
		zap.Any("symbol", eGame.ConfName),
		zap.Any("roundId", req.RoundID),
		zap.Any("playerId", req.PlayerId),
		zap.Any("gameId", req.GameId),
		zap.Any("award", award),
		zap.Any("bet", bet),
		zap.Any("exAward", award),
		zap.Any("exBet", bet))

	// 返奖流水：bet 记本金（负），award 记返奖（>=0）
	billBet := decimal.Zero
	billAward := decimal.Zero
	if bet.GreaterThan(decimal.Zero) {
		billBet = bet.Neg()
	}
	if award.GreaterThan(decimal.Zero) {
		billAward = award
	}
	d.SaveBill(uint32(req.AgentId), req.PlayerId, billBet, billAward, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "返奖", req.CurrencyType, req.RoundID)

	zap.L().Debug("Complete:游戏结束", zap.Any("agentId", req.AgentId), zap.Any("gameId", req.GameId), zap.Any("symbol", eGame.ConfName), zap.Any("roundId", req.RoundID), zap.Any("playerId", req.PlayerId), zap.Any("exAward", award), zap.Any("exBet", bet))
	dao.CacheIns().Complete(int64(req.AgentId), req.PlayerId, eGame.ConfName, bet.Mul(exchange), award.Mul(exchange), pc.Pool[1].Revenue)

	if ri := dao.CacheIns().FinishRoundData(int64(req.AgentId), req.RoundID); ri != nil {
		delta := ri.MaxPay.Round(2).Sub(award.Mul(exchange).Truncate(2))
		if delta.GreaterThanOrEqual(decimal.Zero) {
			zap.L().Debug("Complete:返还水池多扣的积分",
				zap.Any("agentId", req.AgentId),
				zap.Any("gameId", req.GameId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.PlayerId),
				zap.Any("awardMax", ri.MaxPay),
				zap.Any("delta", delta))
			dao.CacheIns().ReturnPool(ri.AgentId, req.PlayerId, eGame.ConfName, delta)
		} else {
			zap.L().Error("返奖异常，预扣值比实际获奖小！！！",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
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

	if account == "" {
		account = dao.CacheIns().GetPlayerAccount(int64(agentId), int64(userId))
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

func (d *LotteryService) SlotsLottery(_ context.Context, req *slotsLotteryReq) (resp *slotsLotteryResp, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			zap.L().Error("panic", zap.Any("err", rec))
			if resp == nil {
				resp = &slotsLotteryResp{}
			}
			resp.Result = false
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			resp.NewCurrency = ""
			err = nil
		}
	}()

	resp = &slotsLotteryResp{Code: services.ErrorCode_OK}
	bet, _ := decimal.NewFromString(req.Bet)
	resp.Result = true
	resp.Code = services.ErrorCode_OK
	eAgent := dao.AgentManagerIns().Get(req.AgentId)
	if eAgent == nil {
		// agentId=0 为测试代理：跳过存在性校验，WebId 记 0。
		if req.AgentId == 0 {
			eAgent = &manager.Agent{Id: 0, WebId: 0}
		} else {
			resp.Code = services.ErrorCode_AGENT_FROZEN
			resp.Result = false
			zap.L().Debug("获取代理信息失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.PlayerId),
				zap.Any("gameId", req.GameId))
			return resp, nil
		}
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

	// ur := &entity.UserRecordInfo{}
	// err = jsoniter.UnmarshalFromString(req.State, ur)
	// if err == nil && !validateUserRecordInfo(ur) {
	// 	zap.L().Error("invalid user record info",
	// 		zap.Any("userId", req.PlayerId),
	// 		zap.Any("symbol", eGame.ConfName),
	// 		zap.Any("agentId", req.AgentId),
	// 		zap.Any("gameId", req.GameId),
	// 		zap.Any("state", req.State))
	// 	resp.Result = false
	// 	resp.Code = services.ErrorCode_PARAMS_INVALID
	// 	return resp, nil
	// }
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
		newCurrency, ok, code := d.SlotsBet(uint32(eAgent.WebId), exchange, req)
		if !ok {
			resp.Result = false
			resp.Code = code
			return resp, nil
		}
		resp.NewCurrency = decimal.NewFromFloat(float64(newCurrency) / 100).String()
	}

	if req.Complete {
		newCurrency, ok, code := d.Complete(uint32(eAgent.WebId), exchange, req)
		if !ok {
			resp.Result = false
			resp.Code = code
			return resp, nil
		}
		resp.NewCurrency = decimal.NewFromFloat(float64(newCurrency) / 100).String()
	}

	return resp, nil
}

func currencyFromCent(v int64) decimal.Decimal {
	return decimal.NewFromInt(v).Div(decimal.NewFromInt(100))
}

func (d *LotteryService) rollbackPoolChange(agentId, userId uint32, symbol, currencyType, recordId string, bet, award, rate decimal.Decimal) {
	dao.CacheIns().ChangePool(int64(agentId), int32(userId), symbol, currencyType, recordId, bet.Neg(), award.Neg(), rate)
}

func (d *LotteryService) deductBet(agentId, userId uint32, exchange decimal.Decimal, symbol, recordId, betStr, currencyType string) (decimal.Decimal, bool) {
	var newCurrency int64 = 0
	bet, _ := decimal.NewFromString(betStr)
	exBet := bet.Mul(exchange)
	pc := config.CfgIns.GetPoolCfg(int64(agentId), symbol)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", recordId), zap.Any("pc", pc))
		return decimal.Zero, false
	}
	zap.L().Debug("deductBet:开始下注", zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId),
		zap.Any("bet", bet),
		zap.Any("currenType", currencyType))
	if bet.GreaterThan(decimal.Zero) {
		//首先扣减用户金额
		tmp, err := d.updatePlayerCurrency(userId, (bet.Neg()).Mul(decimal.NewFromInt(100)).IntPart())
		if err != services.ErrorCode_OK {
			zap.L().Debug("deductBet:下注失败,更新玩家积分失败",
				zap.Any("agentId", agentId),
				zap.Any("symbol", symbol),
				zap.Any("roundId", recordId),
				zap.Any("playerId", userId),
				zap.Any("bet", bet),
				zap.Any("currenType", currencyType))
			return decimal.Zero, false
		}
		newCurrency = tmp
	}
	//
	nc := currencyFromCent(newCurrency)
	dao.CacheIns().ChangePool(int64(agentId), int32(userId), symbol, currencyType, recordId, exBet, decimal.Zero, pc.Pool[1].Revenue)
	user := dao.CacheIns().GetUser(int64(agentId), int64(userId))
	if user != nil && user.IsTourist == 0 {
		if bet.GreaterThan(decimal.Zero) {
			//下注流水
			d.SaveBill(uint32(agentId), userId, bet.Neg(), decimal.Zero, nc.Truncate(2).InexactFloat64(), symbol, "下注", currencyType, recordId)
		}
		//打点水池记录
		d.pcr.Record(int64(agentId), symbol, dao.CacheIns().GetPool(int64(agentId), symbol))
	}

	zap.L().Debug("deductBet:下注成功",
		zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId))
	return nc, true
}

func (d *LotteryService) refundBet(agentId, userId uint32, exchange decimal.Decimal, symbol, recordId, betStr, currencyType string) (decimal.Decimal, bool) {
	var newCurrency int64 = 0
	bet, _ := decimal.NewFromString(betStr)
	exBet := bet.Mul(exchange)
	pc := config.CfgIns.GetPoolCfg(int64(agentId), symbol)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", recordId), zap.Any("pc", pc))
		return decimal.Zero, false
	}
	zap.L().Debug("refundBet:回退", zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId),
		zap.Any("bet", bet),
		zap.Any("currenType", currencyType))
	if bet.GreaterThan(decimal.Zero) {
		//首先扣减用户金额
		tmp, err := d.updatePlayerCurrency(userId, bet.Mul(decimal.NewFromInt(100)).IntPart())
		if err != services.ErrorCode_OK {
			zap.L().Debug("refundBet:回退失败,更新玩家积分失败",
				zap.Any("agentId", agentId),
				zap.Any("symbol", symbol),
				zap.Any("roundId", recordId),
				zap.Any("playerId", userId),
				zap.Any("bet", bet),
				zap.Any("currenType", currencyType))
			return decimal.Zero, false
		}
		newCurrency = tmp
	}
	//
	nc := currencyFromCent(newCurrency)
	dao.CacheIns().ChangePool(int64(agentId), int32(userId), symbol, currencyType, recordId, exBet.Abs().Neg(), decimal.Zero, pc.Pool[1].Revenue)
	user := dao.CacheIns().GetUser(int64(agentId), int64(userId))
	if user != nil && user.IsTourist == 0 {
		if bet.GreaterThan(decimal.Zero) {
			// 回退：bet=0，award=回退金额(>0)
			d.SaveBill(uint32(agentId), userId, decimal.Zero, bet, nc.Truncate(2).InexactFloat64(), symbol, "回退", currencyType, recordId)
		}
		//打点水池记录
		d.pcr.Record(int64(agentId), symbol, dao.CacheIns().GetPool(int64(agentId), symbol))
	}

	zap.L().Debug("refundBet:回退成功",
		zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId))
	return nc, true
}

func (d *LotteryService) PoolAmountResult(_ context.Context, req *services.PoolAmountResultReq) (resp *services.PoolAmountResultResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.PoolAmountResultResp{Code: services.ErrorCode_OK}
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("PoolAmountResult:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("PoolAmountResult:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("PoolAmountResult:获取Pool配置文件失败", zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("PoolAmountResult:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	p := dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName)
	//换算成对应币种的积分   cny->[currencyType]
	resp.Currency = p.Div(exchange).Truncate(2).String()
	return resp, nil
}

func (d *LotteryService) doSingleBet(req *singleBetReq) (resp *singleBetResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &singleBetResp{Code: services.ErrorCode_OK}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doSingleBet:游戏被冻结",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("doSingleBet", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	bet, _ := decimal.NewFromString(req.Bet)
	win, _ := decimal.NewFromString(req.Win)
	// FG：win 可为负（玩家净亏损），仅 >0 走派奖分支；<=0 走下注/完局分支记玩家亏损。
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doSingleBet:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
	if user == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doSingleBet:获取用户信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("doSingleBet:获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("doSingleBet:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	newCurrency := decimal.Zero

	if win.GreaterThan(decimal.Zero) {
		revenue := bet.Mul(exchange).Mul(pc.Pool[1].Revenue).Truncate(4)
		//判断pool是否足够 足够立马扣除
		if !dao.CacheIns().CheckPoolWithChange(int64(req.AgentId), eGame.ConfName, req.RoundID, req.CurrencyType, win.Mul(exchange), bet.Mul(exchange), revenue, req.UserId) {
			//不够赔 不可以开
			resp.Code = services.ErrorCode_NO_ENOUGH_POOL_MONEY
			return resp, nil
		}
		var tmp int64 = 0
		var code services.ErrorCode = services.ErrorCode_OK
		tmp, code = d.updatePlayerCurrency(req.UserId, bet.Neg().Mul(decimal.NewFromInt(100)).IntPart())
		if code != services.ErrorCode_OK {
			d.rollbackPoolChange(req.AgentId, req.UserId, eGame.ConfName, req.CurrencyType, req.RoundID, bet.Mul(exchange), win.Mul(exchange), pc.Pool[1].Revenue)
			zap.L().Error("doSingleBet:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("bet", bet),
				zap.Any("currenType", req.CurrencyType))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		if bet.GreaterThan(decimal.Zero) && user.IsTourist == 0 {
			nc := currencyFromCent(tmp)
			d.SaveBill(uint32(req.AgentId), req.UserId, bet.Neg(), decimal.Zero, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "下注", req.CurrencyType, req.RoundID)
		}

		tmp, code = d.updatePlayerCurrency(req.UserId, win.Mul(decimal.NewFromInt(100)).IntPart())
		if code != services.ErrorCode_OK {
			d.rollbackPoolChange(req.AgentId, req.UserId, eGame.ConfName, req.CurrencyType, req.RoundID, bet.Mul(exchange), win.Mul(exchange), pc.Pool[1].Revenue)
			zap.L().Error("doSingleBet:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("win", win),
				zap.Any("currenType", req.CurrencyType))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}

		//新余额
		nc := currencyFromCent(tmp)
		if win.GreaterThan(decimal.Zero) && user.IsTourist == 0 {
			billBet := decimal.Zero
			if bet.GreaterThan(decimal.Zero) {
				billBet = bet.Neg()
			}
			d.SaveBill(uint32(req.AgentId), req.UserId, billBet, win, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "结算", req.CurrencyType, req.RoundID)
		}
		resp.Currency = nc.String()
		newCurrency = nc
	} else {
		if bet.GreaterThan(decimal.Zero) {
			if nc, ok := d.deductBet(req.AgentId, req.UserId, exchange, eGame.ConfName, req.RoundID, req.Bet, req.CurrencyType); ok {
				resp.Currency = nc.Truncate(2).String()
				newCurrency = nc
			} else {
				resp.Code = services.ErrorCode_SYSTEM_ERROR
			}
		}
	}
	if user.IsTourist == 0 {
		if len(req.Result) > 0 {
			record := ConvertRecord(
				uint32(req.AgentId),
				req.UserId,
				req.RoundID,
				req.CurrencyType,
				eGame.ConfName,
				user.Account,
				req.Result,
				newCurrency,
				uint32(eAgent.WebId),
				true,
				bet.InexactFloat64(),
				win.InexactFloat64())
			d.SaveRecord(record)
			if req.Complete {
				zap.L().Info("doSingleBet:注单信息",
					zap.Any("agentId", req.AgentId),
					zap.Any("playerId", req.UserId),
					zap.Any("gameId", req.GameId),
					zap.Any("symbol", eGame.ConfName),
					zap.Any("roundId", req.RoundID),
					zap.Any("record", record),
					zap.Any("userRecord", req.Result))
			}
		}

		if req.Complete {
			dao.CacheIns().Complete(int64(req.AgentId), req.UserId, eGame.ConfName, bet.Mul(exchange), win.Mul(exchange), decimal.Zero)
		}

		//打点水池记录
		d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
	}
	return resp, nil
}

func (d *LotteryService) doMultiBet(req *multiBetReq) (resp *multiBetResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &multiBetResp{Code: services.ErrorCode_OK}
	roundId := fmt.Sprintf("%s#%d", req.RoundID, req.UserId)
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doMultiBet:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("doMultiBet", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	initBet, _ := decimal.NewFromString(req.InitBet)

	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doMultiBet:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doMultiBet:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("doMultiBet:获取Pool配置文件失败", zap.Any("roundId", roundId), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("doMultiBet:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", roundId),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	if initBet.Abs().GreaterThan(decimal.Zero) {
		if nc, ok := d.deductBet(req.AgentId, req.UserId, exchange, eGame.ConfName, roundId, req.InitBet, req.CurrencyType); ok {
			resp.Currency = nc.Truncate(2).String()
		} else {
			zap.L().Error("doMultiBet:获取汇率配置失败",
				zap.Any("currencyType", req.CurrencyType),
				zap.Any("roundId", roundId),
				zap.Any("agentId", req.AgentId),
				zap.Any("playerId", req.UserId),
				zap.Any("gameId", req.GameId))
		}
	}
	return resp, nil
}

func (d *LotteryService) doMultiRefund(req *multiRefundReq) (resp *multiRefundResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &multiRefundResp{Code: services.ErrorCode_OK}
	roundId := fmt.Sprintf("%s#%d", req.RoundID, req.UserId)
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doMultiRefund:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("doMultiRefund", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doMultiRefund:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("doMultiRefund:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("doMultiRefund:获取Pool配置文件失败", zap.Any("roundId", roundId), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("doMultiRefund:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", roundId),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}

	if nc, ok := d.refundBet(req.AgentId, req.UserId, exchange, eGame.ConfName, roundId, req.Bet, req.CurrencyType); ok {
		resp.Currency = nc.Truncate(2).String()
	}
	return resp, nil
}

func (d *LotteryService) doMultiSettle(req *multiSettleReq) (resp *multiSettleResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &multiSettleResp{
		Code:      services.ErrorCode_SYSTEM_ERROR,
		Currencys: make([]*newCurrencyItem, 0),
	}
	if len(req.Records) <= 0 {
		zap.L().Error("doMultiSettle:批量结算", zap.Any("record count", len(req.Records)))
		return resp, nil
	}
	newCurrencys := make(map[uint32]*newCurrencyItem)
	deltas := make(map[uint32]int64)
	totalWin, _ := decimal.NewFromString(req.TotalWin)
	exchange, ok := config.CfgIns.GetExchange(req.Records[0].CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("doMultiSettle:获取汇率配置失败",
			zap.Any("currencyType", req.Records[0].CurrencyType),
			zap.Any("agentId", req.Records[0].AgentId),
			zap.Any("playerId", req.Records[0].UserId),
			zap.Any("gameId", req.Records[0].GameId))
		return resp, nil
	}
	totalWin = totalWin.Mul(exchange)
	agentId, gameId := int(req.Records[0].AgentId), int(req.Records[0].GameId)
	//首先判断水池是否足够赔付
	for _, item := range req.Records {
		user := dao.CacheIns().GetUser(int64(item.AgentId), int64(item.UserId))
		if user == nil || user.IsTourist != 0 {
			continue
		}
		win, _ := decimal.NewFromString(item.Win)
		if win.LessThanOrEqual(decimal.Zero) {
			continue
		}
	}

	// 百人类的 可以这么写 没有并发问题
	game := dao.GamesManagerIns().GetById(int64(gameId))
	if game == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		return resp, nil
	}

	zap.L().Debug("doMultiSettle:批量结算", zap.Any("symbol", game.ConfName), zap.Any("record count", len(req.Records)), zap.Any("req", req))

	pool := dao.CacheIns().GetPool(int64(agentId), game.ConfName)
	if pool.LessThan(totalWin) {
		zap.L().Debug("doMultiSettle:赔付失败", zap.Any("req", req))
		resp.Code = services.ErrorCode_NO_ENOUGH_POOL_MONEY
		return resp, nil
	}
	//批量更新积分
	for _, item := range req.Records {
		roundId := fmt.Sprintf("%s#%d", item.RoundID, item.UserId)
		win, _ := decimal.NewFromString(item.Win)
		game := dao.GamesManagerIns().GetById(int64(item.GameId))
		bet, _ := decimal.NewFromString(item.Bet)
		if win.GreaterThan(decimal.Zero) {
			win = win.Add(bet)
		} else {
			win = decimal.Zero
		}
		//换算成redis里面的单位 这里不能累加 一个玩家 同一局游戏 只能产生一条游戏记录 如果出现多个游戏记录就是bug  不能累加
		deltas[item.UserId] = win.Mul(decimal.NewFromInt(100)).Truncate(0).IntPart()
		if len(deltas) >= 100 {
			tmp, err := dao.RedisIns().BatchUpdatePlayerCurrencys(deltas)
			if err != nil {
				zap.L().Error("doMultiSettle:更新玩家积分失败",
					zap.Any("agentId", item.AgentId),
					zap.Any("symbol", game.ConfName),
					zap.Any("roundId", roundId),
					zap.Any("playerId", item.UserId),
					zap.Any("win", win),
					zap.Any("currenType", item.CurrencyType))
				resp.Code = services.ErrorCode_SYSTEM_ERROR
				return resp, nil
			}
			for k, v := range tmp {
				newCurrencys[k] = &newCurrencyItem{UserId: k, Currency: decimal.NewFromInt(v).Div(decimal.NewFromInt(100)).Truncate(2).String()}
			}
			deltas = make(map[uint32]int64)
		}
	}
	if len(deltas) > 0 {
		tmp, err := dao.RedisIns().BatchUpdatePlayerCurrencys(deltas)
		if err != nil {
			zap.L().Error("doMultiSettle:更新玩家积分失败", zap.Any("data", deltas))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		for k, v := range tmp {
			newCurrencys[k] = &newCurrencyItem{UserId: k, Currency: decimal.NewFromInt(v).Div(decimal.NewFromInt(100)).Truncate(2).String()}
		}
	}
	//批量保存注单信息
	for _, item := range req.Records {
		roundId := fmt.Sprintf("%s#%d", item.RoundID, item.UserId)
		exchange, ok := config.CfgIns.GetExchange(item.CurrencyType)
		if !ok {
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			zap.L().Error("doMultiSettle:获取汇率配置失败",
				zap.Any("currencyType", item.CurrencyType),
				zap.Any("roundId", roundId),
				zap.Any("agentId", item.AgentId),
				zap.Any("playerId", item.UserId),
				zap.Any("gameId", item.GameId))
			continue
		}
		agent := dao.AgentManagerIns().Get(int64(item.AgentId))
		game := dao.GamesManagerIns().GetById(int64(item.GameId))
		if game == nil || agent == nil {
			resp.Code = services.ErrorCode_PARAMS_INVALID
			continue
		}
		win, _ := decimal.NewFromString(item.Win)
		account := dao.CacheIns().GetPlayerAccount(int64(item.AgentId), int64(item.UserId))
		bet, _ := decimal.NewFromString(item.Bet)
		if tmp := newCurrencys[item.UserId]; tmp != nil {
			user := dao.CacheIns().GetUser(int64(item.AgentId), int64(item.UserId))
			if user == nil || user.IsTourist != 0 {
				continue
			}
			billAward := decimal.Zero
			if win.GreaterThan(decimal.Zero) {
				billAward = win
				win = win.Add(bet)
			} else {
				win = decimal.Zero
			}
			//新余额
			nc, _ := decimal.NewFromString(tmp.Currency)
			defRevenue := decimal.NewFromFloat(dao.DefaultRevenue)
			pc := config.CfgIns.GetPoolCfg(int64(item.AgentId), game.ConfName)
			if pc != nil {
				defRevenue = pc.Pool[1].Revenue
			}
			dao.CacheIns().ChangePool(int64(item.AgentId), int32(item.UserId), game.ConfName, item.CurrencyType, item.RoundID, decimal.Zero, win.Mul(exchange), defRevenue)
			dao.CacheIns().Complete(int64(item.AgentId), item.UserId, game.ConfName, bet.Mul(exchange), win.Mul(exchange), defRevenue)
			if billAward.GreaterThan(decimal.Zero) || bet.GreaterThan(decimal.Zero) {
				billBet := decimal.Zero
				if bet.GreaterThan(decimal.Zero) {
					billBet = bet.Neg()
				}
				d.SaveBill(uint32(item.AgentId), item.UserId, billBet, billAward, nc.Truncate(2).InexactFloat64(), game.ConfName, "结算", item.CurrencyType, roundId)
			}
			newCurrencys[item.UserId] = &newCurrencyItem{UserId: item.UserId, Currency: nc.Truncate(2).String()}
			if game.Number > 0 {
				//增加结算注单
				record := ConvertRecord(
					uint32(item.AgentId),
					item.UserId,
					roundId,
					item.CurrencyType,
					game.ConfName,
					account,
					item.Log,
					nc,
					uint32(agent.WebId),
					true,
					bet.Truncate(2).InexactFloat64(),
					win.Truncate(2).InexactFloat64())
				d.SaveRecord(record)
				zap.L().Info("doMultiSettle:注单信息",
					zap.Any("agentId", item.AgentId),
					zap.Any("playerId", item.UserId),
					zap.Any("gameId", item.GameId),
					zap.Any("symbol", game.ConfName),
					zap.Any("roundId", roundId),
					zap.Any("record", record))
			}
		}
	}
	resArr := make([]*newCurrencyItem, 0, 64)
	for _, v := range newCurrencys {
		resArr = append(resArr, v)
	}
	resp.Currencys = resArr
	resp.Code = services.ErrorCode_OK
	return resp, nil
}
