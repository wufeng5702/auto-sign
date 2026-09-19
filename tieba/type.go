package tieba

type (
	// Gcon 贴吧信息.
	Gcon struct {
		// fid
		Id string `json:"id"`

		// kw
		Name string `json:"name"`

		FavoType string `json:"favo_type"`

		// 等级
		LevelId string `json:"level_id"`

		// 等级名称
		LevelName string `json:"level_name"`

		// 当前经验
		CurScore string `json:"cur_score"`

		// 升级所需经验
		LevelupScore string `json:"levelup_score"`

		IsForbidden string `json:"is_forbidden"`

		// 贴吧头像
		Avatar string `json:"avatar"`

		// slogan
		Slogan string `json:"slogan"`
	}
)

type (
	FavoriteRequest struct {
		// POST 请求

		// 从 1 开始
		pageNo int64

		// 每页数量，应当设置 100-200
		PageSize int64 `json:"page_size"`
	}

	FavoriteResponse struct {
		ForumList struct {
			NonGconForum []Gcon `json:"non-gconforum"`

			GconForum []Gcon `json:"gcon_forum"`
		} `json:"forum_list"`

		ServerTime string `json:"server_time"`

		Time int `json:"time"`

		Ctime int `json:"ctime"`

		Logid int `json:"logid"`

		ErrorCode string `json:"error_code"`

		// 分页数据
		HasMore string `json:"has_more"`

		PageNo string `json:"page_no"`
	}
)

type (
	SignRequest struct {
		// what?
		Fid string `json:"fid"`

		// 贴吧名称
		KW string `json:"kw"`
	}

	SignResponse struct {
		Info       []any `json:"info"`
		ServerTime any   `json:"server_time"`
		Time       any   `json:"time"`
		Ctime      any   `json:"ctime"`
		Logid      any   `json:"logid"`
	}
)

// LoginRequest 百度登录请求
//
// username 用户名
// password 密码
// verifyCode 验证码
// codeString 验证码图片code
// cookie cookie
// token token
type LoginRequest struct {
	UserName   string `json:"username"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
	CodeString string `json:"code_string"`
	Cookie     string `json:"cookie"`
	Token      string `json:"token"`
}

// LoginResponse 百度登录响应
//
// bduss 百度登录凭证
// stoken 百度登录凭证
// ptoken 百度登录凭证
// status 登录状态 0:成功 -1:需要验证码 -2:账号开启了登录保护 -3:用户名或密码错误 -4:其他错误
// message 登录消息
// imgUrl 验证码图片URL
// cookies cookie
// codestring 验证码code
// token token
type LoginResponse struct {
	BDUSS      string `json:"bduss"`
	SToken     string `json:"stoken"`
	PToken     string `json:"ptoken"`
	Status     int    `json:"status"`
	Message    string `json:"message"`
	ImgURL     string `json:"img_url"`
	Cookies    string `json:"cookies"`
	CodeString string `json:"codestring"`
	Token      string `json:"token"`
}

// UserInfoRequest 用户信息请求
type UserInfoRequest struct {
	// 用户名
	Username string `json:"username"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	// 用户UID
	UID string `json:"uid"`
	// 用户头像URL
	HeadImg string `json:"head_img"`
	// 用户礼物数量
	GiftNum string `json:"gift_num"`
	// 用户信息
	Info map[string]interface{} `json:"info"`
}

// FollowRequest 关注用户请求
//
// portrait 用户portrait
type FollowRequest struct {
	Portrait string `json:"portrait"`
}

// FollowResponse 关注用户响应
type FollowResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// UnfollowRequest 取消关注用户请求
//
// portrait 用户portrait
type UnfollowRequest struct {
	Portrait string `json:"portrait"`
}

// UnfollowResponse 取消关注用户响应
type UnfollowResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// FocusRequest 关注贴吧请求
//
// kw 贴吧名称
type FocusRequest struct {
	KW string `json:"kw"`
}

// FocusResponse 关注贴吧响应
type FocusResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// UnfocusRequest 取消关注贴吧请求
//
// kw 贴吧名称
type UnfocusRequest struct {
	KW string `json:"kw"`
}

// UnfocusResponse 取消关注贴吧响应
type UnfocusResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// Post 帖子信息
type Post struct {
	// 帖子ID
	Tid string `json:"tid"`
	// 帖子标题
	Title string `json:"title"`
	// 回复数
	ReplyNum int64 `json:"reply_num"`
	// 是否置顶
	IsTop bool `json:"is_top"`
	// 帖子链接
	Url string `json:"url"`
}

// IndexTListRequest 获取贴吧首页帖子列表请求
//
// tbName 贴吧名称
// replyNum 回复数筛选 0为筛选回复为0的帖子
type IndexTListRequest struct {
	TbName   string `json:"tb_name"`
	ReplyNum int64  `json:"reply_num"`
}

// FidRequest 获取贴吧FID请求
//
// tbName 贴吧名称
type FidRequest struct {
	TbName string `json:"tb_name"`
}

// FidResponse 获取贴吧FID响应
type FidResponse struct {
	Data struct {
		Fid         int `json:"fid"`
		CanSendPics int `json:"can_send_pics"`
	} `json:"data"`
}

// ClientType 客户端类型
//
// iPhone 苹果客户端
// Android 安卓客户端
// WP WindowsPhone客户端
// W8 Windows 8客户端
type ClientType int

const (
	// iPhone 苹果客户端
	IPhone ClientType = 1
	// Android 安卓客户端
	Android ClientType = 2
	// WP WindowsPhone客户端
	WP ClientType = 3
	// W8 Windows 8客户端
	W8 ClientType = 4
)

// ReplyRequest 回帖请求
//
// tbs tbs
// tid 帖子id
// tbName 贴吧名字
// content 回帖内容
type ReplyRequest struct {
	Tbs     string `json:"tbs"`
	Tid     string `json:"tid"`
	TbName  string `json:"tb_name"`
	Content string `json:"content"`
}

// ReplyResponse 回帖响应
type ReplyResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// ReplyFloorRequest 楼中楼回复请求
//
// tbs tbs
// tid 帖子id
// pid 楼层id
// tbName 贴吧名称
// content 回帖内容
type ReplyFloorRequest struct {
	Tbs     string `json:"tbs"`
	Tid     string `json:"tid"`
	Pid     string `json:"pid"`
	TbName  string `json:"tb_name"`
	Content string `json:"content"`
}

// ReplyFloorResponse 楼中楼回复响应
type ReplyFloorResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// AddThreadRequest 发帖请求
//
// tbs tbs
// tbName 贴吧名字
// title 标题，如果没有标题请设置为空字符串
// content 发帖内容
type AddThreadRequest struct {
	Tbs     string `json:"tbs"`
	TbName  string `json:"tb_name"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// AddThreadResponse 发帖响应
type AddThreadResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	No    int64  `json:"no"`
}

// ReplyInfo 回复信息
type ReplyInfo struct {
	// 回复ID
	Id string `json:"id"`
	// 帖子ID
	Tid string `json:"tid"`
	// 楼层ID
	Pid string `json:"pid"`
	// 回复内容
	Content string `json:"content"`
	// 回复时间
	Time string `json:"time"`
	// 回复者昵称
	FromName string `json:"from_name"`
	// 回复者头像
	FromAvatar string `json:"from_avatar"`
	// 被回复者昵称
	ToName string `json:"to_name"`
	// 帖子标题
	Title string `json:"title"`
	// 贴吧名称
	TbName string `json:"tb_name"`
}

// GetMsgRequest 查询艾特/回复信息请求
//
// type reply or at
// pn 页码
type GetMsgRequest struct {
	Type string `json:"type"`
	Pn   int64  `json:"pn"`
}

// GetMsgResponse 查询艾特/回复信息响应
type GetMsgResponse struct {
	// one of reply_list, at_list
	ReplyList []ReplyInfo `json:"reply_list"`
	AtList    []ReplyInfo `json:"at_list"`
}
