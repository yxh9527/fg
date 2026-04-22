package tables

import (
	"app/tables/manager"
	"app/tables/player"
	"time"

	"gorm.io/gorm"
)

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
		games := []*manager.Game{
			{Name: "cjsgj", NameZH: "超级水果机", Number: 3001, ConfName: "cjsgj", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "shz", NameZH: "水浒传", Number: 3002, ConfName: "shz", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "lhdb", NameZH: "连环夺宝", Number: 3003, ConfName: "lhdb", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "tgpd", NameZH: "糖果派对", Number: 3004, ConfName: "tgpd", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "dfdc", NameZH: "多福多财", Number: 3005, ConfName: "dfdc", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jfn", NameZH: "巨富鸟", Number: 3008, ConfName: "jfn", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "xldb", NameZH: "寻龙夺宝", Number: 3009, ConfName: "xldb", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jqb", NameZH: "金钱豹", Number: 3010, ConfName: "jqb", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "hgxs", NameZH: "海龟先生", Number: 3011, ConfName: "hgxs", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "worldcup", NameZH: "世界杯", Number: 3012, ConfName: "worldcup", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "wcg", NameZH: "旺财狗", Number: 3013, ConfName: "wcg", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "lzhd", NameZH: "龙争虎斗", Number: 3014, ConfName: "lzhd", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "rhdb", NameZH: "狨猴寻宝", Number: 3015, ConfName: "rhdb", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "sbwh", NameZH: "桑巴舞会", Number: 3016, ConfName: "sbwh", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "cfmm", NameZH: "财富密码", Number: 3017, ConfName: "cfmm", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "stkh", NameZH: "圣徒狂欢", Number: 3018, ConfName: "stkh", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jbp", NameZH: "聚宝盆", Number: 3019, ConfName: "jbp", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "dwwg", NameZH: "动物王国", Number: 3020, ConfName: "dwwg", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "bdyds", NameZH: "钓鱼大师", Number: 3022, ConfName: "bdyds", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jlbz", NameZH: "金龙宝藏", Number: 3023, ConfName: "jlbz", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "hdbz", NameZH: "海盗宝藏", Number: 3024, ConfName: "hdbz", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "hshwk", NameZH: "黑神话悟空", Number: 3025, ConfName: "hshwk", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "fkseven", NameZH: "疯狂777", Number: 3026, ConfName: "fkseven", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "xldb2", NameZH: "寻龙夺宝2", Number: 3028, ConfName: "xldb2", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "cjsgj2", NameZH: "超级水果机2", Number: 3030, ConfName: "cjsgj2", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "hhsc", NameZH: "虎虎生财", Number: 3031, ConfName: "hhsc", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "sbjn", NameZH: "十倍金牛", Number: 3033, ConfName: "sbjn", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jqt", NameZH: "金钱兔", Number: 3035, ConfName: "jqt", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "sjnw", NameZH: "赏金女王", Number: 3036, ConfName: "sjnw", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "sjddj", NameZH: "赏金大对决", Number: 3037, ConfName: "sjddj", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jszc", NameZH: "金蛇招财", Number: 3038, ConfName: "jszc", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "xmwlj", NameZH: "寻梦亡灵节", Number: 3039, ConfName: "xmwlj", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "cjwp", NameZH: "超级王牌", Number: 3040, ConfName: "cjwp", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "ssff", NameZH: "鼠鼠福福", Number: 3042, ConfName: "ssff", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "jlbs", NameZH: "迦罗宝石", Number: 3051, ConfName: "jlbs", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "mjhl", NameZH: "麻将糊了", Number: 3029, ConfName: "mjhl", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "mjhl2", NameZH: "麻将糊了2", Number: 3032, ConfName: "mjhl2", GameClass: 1, GameType: 3, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "yfct", NameZH: "一飞冲天", Number: 5001, ConfName: "yfct", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "ld", NameZH: "幸运骰子", Number: 5002, ConfName: "ld", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "double", NameZH: "猜红黑", Number: 5003, ConfName: "double", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "dice", NameZH: "猜数字", Number: 5004, ConfName: "dice", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "bxsl", NameZH: "扫雷", Number: 5005, ConfName: "bxsl", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "hilo", NameZH: "高低纸牌", Number: 5006, ConfName: "hilo", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "circle", NameZH: "幸运转盘", Number: 5007, ConfName: "circle", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "plinko", NameZH: "叮咚球", Number: 5008, ConfName: "plinko", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "keno", NameZH: "基诺", Number: 5009, ConfName: "keno", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "limbo", NameZH: "火箭鸟", Number: 5010, ConfName: "limbo", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "tower", NameZH: "玛雅神殿", Number: 5011, ConfName: "tower", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "slide", NameZH: "幸运滑行", Number: 5012, ConfName: "slide", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "coin", NameZH: "幸运硬币", Number: 5013, ConfName: "coin", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "spiritParty", NameZH: "亡灵派对", Number: 5014, ConfName: "spiritParty", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "bbjl", NameZH: "至尊百家乐", Number: 5015, ConfName: "bbjl", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "roulette", NameZH: "轮盘", Number: 5016, ConfName: "roulette", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "bhjk", NameZH: "21点", Number: 5017, ConfName: "bhjk", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
			{Name: "baviator", NameZH: "飞行员", Number: 5018, ConfName: "baviator", GameClass: 1, GameType: 5, LimitTime: "10", IsFrozen: 0, State: 1, CreateTime: int(time.Now().Unix()), UpdateTime: int(time.Now().Unix()), Weight: 0, ShowType: 1, IsShow: 1},
		}
		m.Create(games)
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
