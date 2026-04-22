package tables

import (
	"app/tables/manager"
	"app/tables/player"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type poolConfigItem struct {
	Min        string `json:"min"`
	Normal     string `json:"normal"`
	Max        string `json:"max"`
	NormalRate string `json:"normalRate"`
	MinRate    string `json:"minRate"`
	MaxRate    string `json:"maxRate"`
	Control    string `json:"control"`
	Revenue    string `json:"revenue"`
	Base       string `json:"base"`
}

type poolConfigValue struct {
	Symbol string                    `json:"symbol"`
	Name   string                    `json:"name"`
	NameZH string                    `json:"nameZH"`
	GameID int                       `json:"gameId"`
	Pool   map[string]poolConfigItem `json:"pool"`
}

// 初始化数据库结构
func InitMysqlDb(m, p *gorm.DB) {
	if !m.Migrator().HasTable(&manager.MsgType{}) {
		m.AutoMigrate(&manager.MsgType{})
		mt := []*manager.MsgType{
			{Id: 1, Title: "游戏公告", Class: 1},
			{Id: 2, Title: "维护公告", Class: 1},
			{Id: 3, Title: "管理消息", Class: 2},
		}
		m.Create(mt)
	}
	if !m.Migrator().HasTable(&manager.SystemConf{}) {
		m.AutoMigrate(&manager.SystemConf{})
		sc := []*manager.SystemConf{
			{Id: 1, SystemState: 1},
		}
		m.Create(sc)
	}

	if !m.Migrator().HasTable(&manager.SystemUser{}) {
		m.AutoMigrate(&manager.SystemUser{})
		su := []*manager.SystemUser{
			{Id: 1, Account: "admin1030", Password: "e10adc3949ba59abbe56e057f20f883e", UType: 1, AgentId: 0, UName: "", CreateTime: time.Now().Unix(), IsForzen: 0},
		}
		m.Create(su)
	}
	if !m.Migrator().HasTable(&manager.Game{}) {
		m.AutoMigrate(&manager.Game{})
		now := int(time.Now().Unix())
		newGame := func(number int, nameZH, symbol string) *manager.Game {
			return &manager.Game{
				Number:     number,
				Name:       symbol,
				NameZH:     nameZH,
				ConfName:   symbol,
				GameClass:  1,
				GameType:   3,
				LimitTime:  "10",
				IsFrozen:   0,
				State:      1,
				CreateTime: now,
				UpdateTime: now,
				Weight:     0,
				ShowType:   1,
				IsShow:     1,
			}
		}
		games := []*manager.Game{
			newGame(6000, "萌宠夺宝", "api"),
			newGame(6001, "飞禽走兽", "bbs"),
			newGame(6002, "奔驰宝马2", "bcbm"),
			newGame(6003, "奔驰宝马", "bzb"),
			newGame(6004, "糖果派对", "cp"),
			newGame(6005, "多福多财", "dfdc"),
			newGame(6006, "发发发", "fff"),
			newGame(6007, "浮岛历险记", "fia"),
			newGame(6008, "水果传奇", "fl"),
			newGame(6009, "水果机", "frt"),
			newGame(6010, "水果转盘", "fw"),
			newGame(6011, "黄金大转轮", "gw"),
			newGame(6012, "三国赛马", "hr"),
			newGame(6013, "幸运5", "lucky"),
			newGame(6014, "龙珠探宝", "lztb"),
			newGame(6015, "街机水浒传", "mawm"),
			newGame(6016, "百变猴子", "mcm"),
			newGame(6017, "猴子爬树", "mr"),
			newGame(6018, "新年", "ny"),
			newGame(6019, "水果派对", "party"),
			newGame(6020, "抢红包2", "qhb"),
			newGame(6021, "连环夺宝", "si"),
			newGame(6022, "金瓶梅2", "xjpm"),
			newGame(6023, "水浒英雄", "shyx"),
			newGame(6024, "弹弹丸", "ttw"),
			newGame(6025, "白蛇传", "bsz"),
			newGame(6026, "荣耀王者", "rywz"),
			newGame(6027, "埃及女王", "aijinw"),
			newGame(6028, "西游", "xy"),
			newGame(6029, "招财进宝", "zcjb"),
			newGame(6030, "捕鱼达人", "fishing"),
			newGame(6031, "埃及艳后", "mny"),
			newGame(6032, "守望英雄", "watch"),
			newGame(6033, "森林舞会", "mslwh"),
			newGame(6034, "植物大战僵尸", "pvz"),
			newGame(6035, "金刚", "jg"),
			newGame(6036, "阿凡达", "avatar"),
			newGame(6037, "钻石之恋", "mdl"),
			newGame(6038, "格斗之王", "gdzw"),
			newGame(6039, "火影忍者", "hyrz"),
			newGame(6040, "武松打虎", "wsdf"),
			newGame(6041, "抢红包", "mqhb"),
			newGame(6042, "女校体操队", "sgt"),
			newGame(6043, "速度与激情", "fastf"),
			newGame(6044, "怪物命运", "gwmy"),
			newGame(6045, "神秘东方", "smdf"),
			newGame(6046, "大秦帝国", "dqdg"),
			newGame(6047, "战争", "zz"),
			newGame(6048, "梦幻西游", "mhxy"),
			newGame(6049, "哪咤闹海", "nznh"),
			newGame(6050, "绿野仙踪", "lyxz"),
			newGame(6051, "粉红女郎", "mpg"),
			newGame(6052, "七夕情缘", "qxqy"),
			newGame(6053, "仙剑", "xj"),
			newGame(6054, "四大美女", "sdmn"),
			newGame(6055, "拉斯维加斯", "wjs"),
			newGame(6056, "幻想大陆", "hxdl"),
			newGame(6057, "幸运水果机", "mlfm"),
			newGame(6058, "海盗财富", "hdcf"),
			newGame(6059, "女校足球队", "nxzqd"),
			newGame(6060, "古墓丽影", "gmly"),
			newGame(6061, "女校游泳队", "sst"),
			newGame(6062, "黑珍珠号", "hzzh"),
			newGame(6063, "女校啦啦队", "scs"),
			newGame(6064, "武圣关云长", "wsgyc"),
			newGame(6065, "封神榜", "fsb"),
			newGame(6066, "指环王", "zhw"),
			newGame(6067, "性感女神", "pinup"),
			newGame(6068, "人猿泰山", "ryts"),
			newGame(6069, "鹿鼎记", "wldj"),
			newGame(6070, "富饶农场", "frnc"),
			newGame(6071, "剑侠情缘", "jxqy"),
			newGame(6072, "笑傲江湖", "wxajh"),
			newGame(6073, "刺客信条", "ckxt"),
			newGame(6074, "海岛奇兵", "mboombeach"),
			newGame(6075, "神雕侠侣", "wsdxl"),
			newGame(6076, "天龙八部", "wtlbb"),
			newGame(6077, "红楼梦", "hlm"),
			newGame(6078, "金瓶梅", "jpm"),
			newGame(6079, "女校橄榄球", "nxglq"),
			newGame(6080, "吸血鬼PK狼人", "xxgpklr"),
			newGame(6081, "鸟叔", "mpjs"),
			newGame(6082, "希腊传说", "xlcs"),
			newGame(6083, "十二星座", "constellation"),
			newGame(6084, "玛雅", "my"),
			newGame(6085, "加勒比海盗", "pirates"),
			newGame(6086, "众神之王", "mmm"),
			newGame(6087, "十二生肖", "zodiac"),
			newGame(6088, "愤怒的小鸟", "angrybirds"),
			newGame(6089, "瑞狗迎春", "rgyc"),
			newGame(6090, "侏罗纪", "jurassic"),
			newGame(6091, "女校剑道部", "nxjdb"),
			newGame(6092, "湛蓝深海", "xsjr"),
			newGame(6093, "印第安", "yda"),
			newGame(6094, "金狗旺财", "jgwc"),
			newGame(6095, "恭贺新春", "ghxc"),
			newGame(6096, "角斗士", "jiaods"),
			newGame(6097, "灌篮高手", "msd"),
			newGame(6098, "闹元宵", "mnys"),
			newGame(6099, "金球争霸", "jqzb"),
			newGame(6100, "黄金右脚", "hjyj"),
			newGame(6101, "世界杯吉祥物", "sjbjxw"),
			newGame(6102, "潘帕斯雄鹰", "panpasixiongying"),
			newGame(6103, "群星闪耀", "qunxshanyao"),
			newGame(6104, "金靴争霸", "mjxzb"),
			newGame(6105, "激情球迷", "mjqqm"),
			newGame(6106, "激情世界杯", "mjqsjb"),
			newGame(6107, "船长宝藏", "mczbz"),
			newGame(6108, "疯狂七", "mcrazy7"),
			newGame(6109, "鹊桥会", "qqh"),
			newGame(6110, "泰坦尼克号", "mtitanic"),
			newGame(6111, "蜘蛛侠", "zzx"),
			newGame(6112, "古怪猴子", "gghz"),
			newGame(6113, "变形金刚", "bxjg"),
			newGame(6114, "摸金校尉", "mjxw"),
			newGame(6115, "国王游戏", "gwyx"),
			newGame(6116, "百变猴子2", "mcm2"),
			newGame(6117, "森林舞会2", "mslwh2"),
			newGame(6118, "抢红包3", "qhb3"),
			newGame(6119, "淘金热", "tjr"),
			newGame(6120, "喜气洋洋", "xqyy"),
			newGame(6121, "66路", "route66"),
			newGame(6122, "阿兹特克", "azteke"),
			newGame(6123, "埃及", "egypt"),
			newGame(6124, "狂欢", "kh"),
			newGame(6125, "中世纪特权", "zsjtq"),
			newGame(6126, "欢乐水果", "sg"),
			newGame(6127, "街头霸王", "jtbw"),
			newGame(6128, "战舰少女", "zjsn"),
			newGame(6129, "东方国度", "zgf"),
			newGame(6130, "西部牛仔", "nz"),
			newGame(6131, "荒野大镖客", "nz2"),
			newGame(6132, "梦游仙境", "myxj"),
			newGame(6133, "80天旅行", "80tlx"),
			newGame(6134, "疯狂黄金城", "fkhjc"),
			newGame(6135, "麻将来了 Pro", "mjllpro"),
			newGame(6136, "埃菲尔", "eiffel"),
			newGame(6137, "新年符号", "nysymbols"),
			newGame(6138, "世界末日", "doomsday"),
			newGame(6139, "现代战争", "mowf"),
			newGame(6140, "甜水绿洲", "lushwater"),
			newGame(6141, "功夫熊猫", "panda"),
			newGame(6142, "哥谭魅影猫女", "gtmymn"),
			newGame(6143, "发发发2", "fff2"),
			newGame(6144, "水果派对2", "party2"),
		}
		m.Create(games)
	}
	if !m.Migrator().HasTable(&manager.PoolConfig{}) {
		m.AutoMigrate(&manager.PoolConfig{})
		var games []manager.Game
		m.Find(&games)

		poolConfigs := make([]*manager.PoolConfig, 0, len(games))
		for _, game := range games {
			value, err := json.Marshal(poolConfigValue{
				Symbol: game.ConfName,
				Name:   game.ConfName,
				NameZH: game.NameZH,
				GameID: game.Number,
				Pool: map[string]poolConfigItem{
					"1": {
						Min:        "300",
						Normal:     "900",
						Max:        "3000",
						NormalRate: "0.99",
						MinRate:    "0.99",
						MaxRate:    "0.99",
						Control:    "1",
						Revenue:    "0.03",
						Base:       "0",
					},
				},
			})
			if err != nil {
				panic(err)
			}

			poolConfigs = append(poolConfigs, &manager.PoolConfig{
				Key:   fmt.Sprintf("/config/pool/%s", game.ConfName),
				Value: string(value),
			})
		}
		m.Create(poolConfigs)
	}
	//manager
	m.AutoMigrate(&manager.Agent{},
		&manager.AgentConfig{},
		&manager.AgentGame{},
		&manager.AgentGameConf{},
		&manager.ApiConfig{},
		&manager.Feedback{},
		&manager.GameDataHour{},
		&manager.GameDataSummary{},
		&manager.Log{},
		&manager.Msg{},
		&manager.MsgType{},
		&manager.PoolConfig{},
		&manager.PlayerDataHour{},
		&manager.PlayerDataSummary{},
		&manager.PlayerProRank{},
		&manager.ProfitLoos{},
		&manager.Statistics{},
		&manager.SystemUserMsg{},
		&manager.User{},
		&manager.Web{},
		&manager.UserScoreLog{},
	)
	//player
	p.AutoMigrate(
		&player.Player{},
	)
}
