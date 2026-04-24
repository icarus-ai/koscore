package client

import "github.com/kernel-ai/koscore/client/packets/pb/v2/service/operation"

func (m OnlineStatus) Status() (ret operation.SetStatus) {
	if m > 100 {
		ret.Status = uint32(StatusOnline)
		ret.ExtStatus = uint32(m)
		if ret.ExtStatus == uint32(StatusBattery) {
			ret.BatteryStatus = 100
		}
	} else {
		ret.Status = uint32(m)
	}
	return
}

func (m OnlineStatusMap) Status(name string) (ret operation.SetStatus) {
	if status, ok := m[name]; ok {
		return status.Status()
	}
	return StatusInvisible.Status()
}

type OnlineStatusMap map[string]OnlineStatus

var UserOnlineStatusStrMap = OnlineStatusMap{
	"在线":      StatusOnline,
	"离开":      StatusAway,
	"隐身":      StatusInvisible,
	"忙碌":      StatusBusy,
	"离线":      StatusOffline,
	"Q我吧":     StatusQme,
	"请勿打扰":    StatusDoNotDisturb,
	"当前电量":    StatusBattery,
	"信号弱":     StatusSignal,
	"睡觉中":     StatusSleep,
	"游戏中":     StatusGaming,
	"学习中":     StatusStudy,
	"干饭中":     StatusCookedRice,
	"健身中":     StatusFitness,
	"追剧中":     StatusWatchingTV,
	"度假中":     StatusVacationing,
	"在线学习":    StatusStudyOnline,
	"timi中":   StatusTimi,
	"听歌中":     StatusListening,
	"今日天气":    StatusWeather,
	"熬夜中":     StatusStayUp,
	"星座运势":    StatusConstellation,
	"打球中":     StatusPlayBall,
	"恋爱中":     StatusLoving,
	"我没事":     StatusIMFine,
	"汪汪汪":     StatusWangWang,
	"嗨到飞起":    StatusHiToFly,
	"元气满满":    StatusFullOfEnergy,
	"悠哉哉":     StatusLeisurely,
	"无聊中":     StatusBoredom,
	"想静静":     StatusIWantToBeQuiet,
	"我太难了":    StatusItSTooHardForMe,
	"一言难尽":    StatusItSHardToPutIntoWords,
	"吃鸡中":     StatusEatChicken,
	"遇见春天":    StatusMeetSpring,
	"好运锦鲤":    StatusGoodLuckKoi,
	"水逆退散":    StatusTheWaterRetreats,
	"摸鱼中":     StatusTouchingTheFish,
	"emo中":    StatusEmo,
	"难得糊涂":    StatusItSHardToGetConfused,
	"出去浪":     StatusGetOutOnTheWaves,
	"爱你":      StatusLoveYou,
	"肝作业":     StatusLiverWork,
	"我想开了":    StatusIWantToOpenIt,
	"被掏空":     StatusHollowedOut,
	"去旅行":     StatusGoOnATrip,
	"今日步数":    StatusTodayStepCount,
	"我crush了": StatusCrushed,
	"搬砖中":     StatusMovingBricks,
	"一起元梦":    StatusTotherStar,
	"求星搭子":    StatusSeekPartner,
	"做好事":     StatusDoGood,
}

type OnlineStatus uint16

const (
	StatusOnline       OnlineStatus = 10 // 在线
	StatusOffline      OnlineStatus = 21 // 离线
	StatusAway         OnlineStatus = 30 // 离开
	StatusInvisible    OnlineStatus = 40 // 隐身
	StatusBusy         OnlineStatus = 50 // 忙碌
	StatusQme          OnlineStatus = 60 // Q我吧
	StatusDoNotDisturb OnlineStatus = 70 // 请勿打扰
	// status 10 ext_status
	StatusBattery               OnlineStatus = 1000 // 当前电量
	StatusSignal                OnlineStatus = 1011 // 信号弱
	StatusSleep                 OnlineStatus = 1016 // 睡觉中
	StatusGaming                OnlineStatus = 1017 // . 游戏中
	StatusStudy                 OnlineStatus = 1018 // 学习中
	StatusCookedRice            OnlineStatus = 1019 // . 干饭中
	StatusFitness               OnlineStatus = 1020 // . 健身中
	StatusWatchingTV            OnlineStatus = 1021 // 追剧中
	StatusVacationing           OnlineStatus = 1022 // . 度假中
	StatusStudyOnline           OnlineStatus = 1024 // . 在线学习
	StatusTimi                  OnlineStatus = 1027 // timi中
	StatusListening             OnlineStatus = 1028 // 听歌中
	StatusWeather               OnlineStatus = 1030 // 今日天气
	StatusStayUp                OnlineStatus = 1032 // 熬夜中
	StatusConstellation         OnlineStatus = 1040 // . 星座运势
	StatusPlayBall              OnlineStatus = 1050 // . 打球中
	StatusLoving                OnlineStatus = 1051 // 恋爱中
	StatusIMFine                OnlineStatus = 1052 // 我没事
	StatusWangWang              OnlineStatus = 1053 // . 汪汪汪
	StatusHiToFly               OnlineStatus = 1056 // 嗨到飞起
	StatusFullOfEnergy          OnlineStatus = 1058 // 元气满满
	StatusLeisurely             OnlineStatus = 1059 // 悠哉哉
	StatusBoredom               OnlineStatus = 1060 // 无聊中
	StatusIWantToBeQuiet        OnlineStatus = 1061 // 想静静
	StatusItSTooHardForMe       OnlineStatus = 1062 // 我太难了
	StatusItSHardToPutIntoWords OnlineStatus = 1063 // 一言难尽
	StatusEatChicken            OnlineStatus = 1064 // . 吃鸡中
	StatusMeetSpring            OnlineStatus = 1069 // . 遇见春天
	StatusGoodLuckKoi           OnlineStatus = 1071 // 好运锦鲤
	StatusTheWaterRetreats      OnlineStatus = 1201 // 水逆退散
	StatusTouchingTheFish       OnlineStatus = 1300 // 摸鱼中
	StatusEmo                   OnlineStatus = 1401 // emo中
	StatusItSHardToGetConfused  OnlineStatus = 2001 // 难得糊涂
	StatusGetOutOnTheWaves      OnlineStatus = 2003 // 出去浪
	StatusLoveYou               OnlineStatus = 2006 // 爱你
	StatusLiverWork             OnlineStatus = 2012 // 肝作业
	StatusIWantToOpenIt         OnlineStatus = 2013 // 我想开了
	StatusHollowedOut           OnlineStatus = 2014 // 被掏空
	StatusGoOnATrip             OnlineStatus = 2015 // 去旅行
	StatusTodayStepCount        OnlineStatus = 2017 // 今日步数
	StatusCrushed               OnlineStatus = 2019 // 我crush了
	StatusMovingBricks          OnlineStatus = 2023 // 搬砖中
	StatusTotherStar            OnlineStatus = 2025 // 一起元梦
	StatusSeekPartner           OnlineStatus = 2026 // 求星搭子
	StatusDoGood                OnlineStatus = 2047 // 做好事
)
