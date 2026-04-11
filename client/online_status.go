package client

import "github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"

var (
	StatusOnline                = operation.SetStatus{Status: 10}                  // 在线
	StatusAway                  = operation.SetStatus{Status: 30}                  // 离开
	StatusInvisible             = operation.SetStatus{Status: 40}                  // 隐身
	StatusBusy                  = operation.SetStatus{Status: 50}                  // 忙碌
	StatusQme                   = operation.SetStatus{Status: 60}                  // Q我吧
	StatusDoNotDisturb          = operation.SetStatus{Status: 70}                  // 请勿打扰
	StatusBattery               = operation.SetStatus{Status: 10, ExtStatus: 1000} // 当前电量
	StatusStatusSignal          = operation.SetStatus{Status: 10, ExtStatus: 1011} // 信号弱
	StatusStatusSleep           = operation.SetStatus{Status: 10, ExtStatus: 1016} // 睡觉中
	StatusStatusStudy           = operation.SetStatus{Status: 10, ExtStatus: 1018} // 学习中
	StatusWatchingTV            = operation.SetStatus{Status: 10, ExtStatus: 1021} // 追剧中
	StatusTimi                  = operation.SetStatus{Status: 10, ExtStatus: 1027} // timi中
	StatusListening             = operation.SetStatus{Status: 10, ExtStatus: 1028} // 听歌中
	StatusWeather               = operation.SetStatus{Status: 10, ExtStatus: 1030} // 今日天气
	StatusStayUp                = operation.SetStatus{Status: 10, ExtStatus: 1032} // 熬夜中
	StatusLoving                = operation.SetStatus{Status: 10, ExtStatus: 1051} // 恋爱中
	StatusIMFine                = operation.SetStatus{Status: 10, ExtStatus: 1052} // 我没事
	StatusHiToFly               = operation.SetStatus{Status: 10, ExtStatus: 1056} // 嗨到飞起
	StatusFullOfEnergy          = operation.SetStatus{Status: 10, ExtStatus: 1058} // 元气满满
	StatusLeisurely             = operation.SetStatus{Status: 10, ExtStatus: 1059} // 悠哉哉
	StatusBoredom               = operation.SetStatus{Status: 10, ExtStatus: 1060} // 无聊中
	StatusIWantToBeQuiet        = operation.SetStatus{Status: 10, ExtStatus: 1061} // 想静静
	StatusItSTooHardForMe       = operation.SetStatus{Status: 10, ExtStatus: 1062} // 我太难了
	StatusItSHardToPutIntoWords = operation.SetStatus{Status: 10, ExtStatus: 1063} // 一言难尽
	StatusGoodLuckKoi           = operation.SetStatus{Status: 10, ExtStatus: 1071} // 好运锦鲤
	StatusTheWaterRetreats      = operation.SetStatus{Status: 10, ExtStatus: 1201} // 水逆退散
	StatusTouchingTheFish       = operation.SetStatus{Status: 10, ExtStatus: 1300} // 摸鱼中
	StatusStatusEmo             = operation.SetStatus{Status: 10, ExtStatus: 1401} // emo中
	StatusItSHardToGetConfused  = operation.SetStatus{Status: 10, ExtStatus: 2001} // 难得糊涂
	StatusGetOutOnTheWaves      = operation.SetStatus{Status: 10, ExtStatus: 2003} // 出去浪
	StatusLoveYou               = operation.SetStatus{Status: 10, ExtStatus: 2006} // 爱你
	StatusLiverWork             = operation.SetStatus{Status: 10, ExtStatus: 2012} // 肝作业
	StatusIWantToOpenIt         = operation.SetStatus{Status: 10, ExtStatus: 2013} // 我想开了
	StatusHollowedOut           = operation.SetStatus{Status: 10, ExtStatus: 2014} // 被掏空
	StatusGoOnATrip             = operation.SetStatus{Status: 10, ExtStatus: 2015} // 去旅行
	StatusTodayStepCount        = operation.SetStatus{Status: 10, ExtStatus: 2017} // 今日步数
	StatusCrushed               = operation.SetStatus{Status: 10, ExtStatus: 2019} // 我crush了
	StatusMovingBricks          = operation.SetStatus{Status: 10, ExtStatus: 2023} // 搬砖中
	StatusTotherStar            = operation.SetStatus{Status: 10, ExtStatus: 2025} // 一起元梦
	StatusSeekPartner           = operation.SetStatus{Status: 10, ExtStatus: 2026} // 求星搭子
	StatusDoGood                = operation.SetStatus{Status: 10, ExtStatus: 2047} // 做好事
)
