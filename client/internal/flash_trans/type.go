package flash_trans

type appid_t uint32

const (
	PrivateRecord     appid_t = 1402
	GroupRecord       appid_t = 1403
	PrivateVideo      appid_t = 1413
	PrivateVideoThumb appid_t = 1414
	GroupVideo        appid_t = 1415
	GroupVideoThumb   appid_t = 1416
	PrivateImage      appid_t = 1406
	GroupImage        appid_t = 1407
	FlashTrans        appid_t = 14901
	FlashTransThumb   appid_t = 14903
)

type appid_body_t struct{ ID, ID101, ID102, ID103, ID200, IDXXX uint32 }

// 首位是子业务id 随之是101/102/103/200 最后一个是fileid同级6号位
var SceneInfoMap = map[appid_t]appid_body_t{
	1402:  {4717, 1, 3, 0, 1, 0},   // 私信语音
	1403:  {4718, 1, 3, 0, 2, 0},   // 群语音
	1413:  {4585, 2, 2, 0, 1, 0},   // 私信视频
	1414:  {4585, 2, 2, 0, 1, 100}, // 私信视频封面
	1415:  {4586, 2, 2, 0, 2, 0},   // 群视频
	1416:  {4586, 2, 2, 0, 2, 100}, // 群视频封面
	1406:  {4549, 2, 1, 0, 1, 0},   // 私信图片
	1407:  {4548, 2, 1, 0, 2, 0},   // 群聊图片
	14901: {4777, 2, 4, 22, 5, 0},  // 闪传
	14903: {4777, 2, 4, 23, 5, 0},  // 闪传封面
}
