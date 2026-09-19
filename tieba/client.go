package tieba

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
)

// API URLs
const (
	urlTbs            = "https://tieba.baidu.com/dc/common/tbs"
	urlLike           = "https://c.tieba.baidu.com/c/f/forum/like"
	urlSign           = "https://c.tieba.baidu.com/c/c/forum/sign"
	urlFollow         = "https://tieba.baidu.com/c/f/follow/add"
	urlUnfollow       = "https://tieba.baidu.com/c/f/follow/remove"
	urlLikeTieba      = "https://tieba.baidu.com/c/f/forum/likeForum"
	urlAddTieba       = "https://tieba.baidu.com/f/like/add"
	urlUnfavoTiebaNew = "https://tieba.baidu.com/c/f/forum/unfavo"
	urlTiebaFid       = "https://tieba.baidu.com/f/commit/share/fnameShareApi?ie=utf-8&fname=%s"
	urlReplyPost      = "https://c.tieba.baidu.com/c/c/post/add"
	urlFlorPid        = "https://c.tieba.baidu.com/c/p/floor"
	urlTbatPost       = "https://tieba.baidu.com/c/f/message/"
	urlThreadAdd      = "https://c.tieba.baidu.com/c/c/thread/add"
)

type Client struct {
	// 百度使用的 token BDUSS
	bduss string

	// 锁
	mu sync.RWMutex

	// http 客户端
	client *http.Client

	// 日志
	log *slog.Logger

	// TBS缓存相关字段
	tbs           string
	tbsExpireTime time.Time

	// 重试相关字段
	maxRetries int
	retryDelay time.Duration

	// TODO 邮件通知等？
}

func WithLog(log *slog.Logger) Option {
	return func(c *Client) {
		c.log = log
	}
}

func WithMaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = max(n, 1)
	}
}

func WithRetryDelay(delay time.Duration) Option {
	return func(c *Client) {
		c.retryDelay = max(delay, 0)
	}
}

type Option func(c *Client)

func NewClient(bduss string, opts ...Option) (*Client, error) {
	if bduss == "" {
		return nil, errors.New("bduss is required")
	}

	client := &Client{
		bduss: bduss,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: slog.Default(),

		maxRetries: 3,
		retryDelay: time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// GetMsg 查询艾特/回复 信息
func (c *Client) GetMsg(ctx context.Context, request *GetMsgRequest) (*GetMsgResponse, error) {
	// 构建请求参数，只包含encode方法没有的参数
	data := map[string]string{
		"pn": Itoa(request.Pn),
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, fmt.Sprintf("%s%sme", urlTbatPost, request.Type), data)
	if err != nil {
		return nil, err
	}

	var response GetMsgResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	// 根据请求类型获取对应的列表
	switch request.Type {
	case "reply":
		if len(response.ReplyList) == 0 {
			return nil, errors.New("no reply list")
		}
	case "at":
		if len(response.AtList) == 0 {
			return nil, errors.New("no at list")
		}
	default:
		return nil, errors.New("invalid type, must be reply or at")
	}

	// 如果没有数据，返回空列表
	return &response, nil
}

// Reply 回帖-客户端
func (c *Client) Reply(ctx context.Context, request *ReplyRequest) (*ReplyResponse, error) {
	// 获取fid
	fidResp, err := c.GetFid(ctx, &FidRequest{
		TbName: request.TbName,
	})
	if err != nil {
		return nil, err
	}

	// 构建请求参数，只包含encode方法没有的或需要覆盖的参数
	data := map[string]string{
		"anonymous": "1",
		"content":   request.Content,
		"fid":       Itoa(fidResp.Data.Fid),
		"from":      "1008621x",
		"is_ad":     "0",
		"kw":        request.TbName,
		"new_vcode": "1",
		"tbs":       request.Tbs,
		"tid":       request.Tid,
		"vcode_tag": "11",
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlReplyPost, data)
	if err != nil {
		return nil, err
	}

	var response ReplyResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ReplyFloor 楼中楼回复-客户端
func (c *Client) ReplyFloor(ctx context.Context, request *ReplyFloorRequest) (*ReplyFloorResponse, error) {
	// 获取fid
	fidResp, err := c.GetFid(ctx, &FidRequest{
		TbName: request.TbName,
	})
	if err != nil {
		return nil, err
	}

	// 构建请求参数，只包含encode方法没有的或需要覆盖的参数
	data := map[string]string{
		"anonymous": "1",
		"content":   request.Content,
		"fid":       Itoa(fidResp.Data.Fid),
		"kw":        request.TbName,
		"new_vcode": "1",
		"post_from": "3",
		"quote_id":  request.Pid,
		"tbs":       request.Tbs,
		"tid":       request.Tid,
		"vcode_tag": "12",
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlReplyPost, data)
	if err != nil {
		return nil, err
	}

	var response ReplyFloorResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// AddThread 发帖-客户端
func (c *Client) AddThread(ctx context.Context, request *AddThreadRequest) (*AddThreadResponse, error) {
	// 获取fid
	fidResp, err := c.GetFid(ctx, &FidRequest{
		TbName: request.TbName,
	})
	if err != nil {
		return nil, err
	}

	// 构建请求参数，只包含encode方法没有的或需要覆盖的参数
	data := map[string]string{
		"anonymous":      "1",
		"call_from":      "2",
		"can_no_forum":   "0",
		"content":        request.Content,
		"entrance_type":  "1",
		"fid":            Itoa(fidResp.Data.Fid),
		"from":           "1001128p",
		"is_feedback":    "0",
		"is_hide":        "1",
		"is_link_thread": "0",
		"st_type":        "notitle",
		"is_ntitle": func() string {
			if request.Title == "" {
				return "1"
			} else {
				return "0"
			}
		}(),
		"kw":            request.TbName,
		"new_vcode":     "1",
		"reply_uid":     "null",
		"takephoto_num": "0",
		"tbs":           request.Tbs,
		"stErrorNums":   "0",
		"title":         request.Title,
		"vcode_tag":     "12",
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlThreadAdd, data)
	if err != nil {
		return nil, err
	}

	var response AddThreadResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetFid 获取贴吧fid
func (c *Client) GetFid(ctx context.Context, request *FidRequest) (*FidResponse, error) {
	_url := fmt.Sprintf(urlTiebaFid, url.QueryEscape(request.TbName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, _url, nil)
	if err != nil {
		return nil, err
	}

	var response FidResponse

	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Focus 关注一个贴吧
func (c *Client) Focus(ctx context.Context, request *FocusRequest) (*FocusResponse, error) {
	// 获取fid
	fidResp, err := c.GetFid(ctx, &FidRequest{
		TbName: request.KW,
	})
	if err != nil {
		return nil, err
	}

	// 构建请求参数，只包含encode方法没有的参数
	data := map[string]string{
		"fid": Itoa(fidResp.Data.Fid),
		"kw":  request.KW,
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlLikeTieba, data)
	if err != nil {
		return nil, err
	}

	var response FocusResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Focus2 关注一个贴吧(网页接口)
func (c *Client) Focus2(ctx context.Context, fid, stoken string) (*FocusResponse, error) {
	// 构建请求参数
	data := map[string]string{
		"fid": fid,
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlAddTieba, data)
	if err != nil {
		return nil, err
	}

	var response FocusResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Unfocus 取消关注一个贴吧
func (c *Client) Unfocus(ctx context.Context, request *UnfocusRequest) (*UnfocusResponse, error) {
	// 获取fid
	fidResp, err := c.GetFid(ctx, &FidRequest{
		TbName: request.KW,
	})
	if err != nil {
		return nil, err
	}

	// 构建请求参数，只包含encode方法没有的参数
	data := map[string]string{
		"fid": Itoa(fidResp.Data.Fid),
		"kw":  request.KW,
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlUnfavoTiebaNew, data)
	if err != nil {
		return nil, err
	}

	var response UnfocusResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Follow 关注一个人
func (c *Client) Follow(ctx context.Context, request *FollowRequest) (*FollowResponse, error) {
	// 构建请求参数，只包含encode方法没有的参数
	data := map[string]string{
		"portrait": request.Portrait,
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlFollow, data)
	if err != nil {
		return nil, err
	}

	var response FollowResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Unfollow 取消关注一个人
func (c *Client) Unfollow(ctx context.Context, request *UnfollowRequest) (*UnfollowResponse, error) {
	// 构建请求参数，只包含encode方法没有的参数
	data := map[string]string{
		"portrait": request.Portrait,
	}

	// 生成签名并构建请求
	req, err := c.encode(ctx, http.MethodPost, urlUnfollow, data)
	if err != nil {
		return nil, err
	}

	var response UnfollowResponse
	err = c.doWithJSON(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Favorite
// 获取关注的贴吧列表.
func (c *Client) Favorite(ctx context.Context, request *FavoriteRequest) (*FavoriteResponse, error) {
	c.log.InfoContext(ctx, "Favorite Start")
	defer c.log.InfoContext(ctx, "Favorite Finished")

	request.pageNo = 0

	response := FavoriteResponse{
		HasMore: "1",
	}

	// 处理分页
	for response.HasMore == "1" {
		time.Sleep(time.Second * 3)

		request.pageNo++

		res, err := c.favorite(ctx, request)
		if err != nil {
			c.log.ErrorContext(ctx, "Favorite", "HasMore Request", request, "err", err)

			break
		}

		c.log.DebugContext(ctx, "Favorite", "HasMore res", res)

		// 合并
		res.ForumList.NonGconForum = append(response.ForumList.NonGconForum, res.ForumList.NonGconForum...)
		res.ForumList.GconForum = append(response.ForumList.GconForum, res.ForumList.GconForum...)

		response = *res
	}

	return &response, nil
}

func (c *Client) favorite(ctx context.Context, request *FavoriteRequest) (*FavoriteResponse, error) {
	data := map[string]string{
		"page_size": Itoa(request.PageSize),
		"page_no":   Itoa(request.pageNo),
	}

	// 生成签名并构建请求
	r, err := c.encode(ctx, http.MethodPost, urlLike, data)
	if err != nil {
		return nil, err
	}

	var response FavoriteResponse

	err = c.doWithJSON(r, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Sign
// 贴吧签到.
func (c *Client) Sign(ctx context.Context, request *SignRequest) (*SignResponse, error) {
	data := map[string]string{
		"fid": request.Fid,
		"kw":  request.KW,
	}

	// 生成签名并构建请求
	r, err := c.encode(ctx, http.MethodPost, urlSign, data)
	if err != nil {
		return nil, err
	}

	var response SignResponse

	err = c.doWithJSON(r, &response)
	if err != nil {
		var check *Error
		if errors.As(err, &check) && check.Code() == CodeSignRepeat {
			return &response, nil
		}

		return nil, err
	}

	c.log.DebugContext(ctx, "Sign", "request", request)
	c.log.DebugContext(ctx, "Sign", "response", response)

	return &response, nil
}

// getTbs 获取TBS（Token Bucket System）值，用于后续操作，带有10分钟缓存
func (c *Client) getTbs(ctx context.Context) (string, error) {
	// 检查缓存是否有效
	c.mu.RLock()
	if c.tbs != "" && time.Now().Before(c.tbsExpireTime) {
		c.mu.RUnlock()
		return c.tbs, nil
	}
	c.mu.RUnlock()

	// 缓存无效，重新获取
	c.log.InfoContext(ctx, "getTbs Start")
	defer c.log.InfoContext(ctx, "getTbs Finished")

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, urlTbs, nil)
	if err != nil {
		return "", err
	}

	var response struct {
		Tbs string `json:"tbs"`

		IsLogin int64 `json:"is_login"`
	}

	err = c.doWithJSON(r, &response)
	if err != nil {
		return "", err
	}

	// 更新缓存，设置10分钟过期时间
	c.mu.Lock()
	c.tbs = response.Tbs
	c.tbsExpireTime = time.Now().Add(10 * time.Minute)
	c.mu.Unlock()

	return c.tbs, nil
}

// 根据规则，编码数据并构建请求
func (c *Client) encode(ctx context.Context, method, urlStr string, src map[string]string) (*http.Request, error) {
	// 获取tbs值
	tbs, err := c.getTbs(ctx)
	if err != nil {
		// 如果获取tbs失败，使用空字符串，让百度API返回错误
		tbs = ""
		c.log.ErrorContext(ctx, "Failed to get tbs", "error", err)
	}

	data := map[string]string{
		// const (
		//	// iPhone 苹果客户端
		//	iPhone ClientType = 1
		//	// Android 安卓客户端
		//	Android ClientType = 2
		//	// WP WindowsPhone客户端
		//	WP ClientType = 3
		//	// W8 Windows 8客户端
		//	W8 ClientType = 4
		// )
		"_client_type":    "2",
		"_client_id":      "wappc_1534235498291_488",
		"_client_version": "9.7.8.0",
		"_phone_imei":     "000000000000000",
		"from":            "1008621y",
		"model":           "MI+5",
		"net_type":        "1",
		"vcode_tag":       "11",

		"BDUSS": c.bduss,
		"tbs":   tbs,

		"timestamp": Itoa(time.Now().Unix()),
	}

	maps.Insert(data, maps.All(src))

	values := encode(data)

	req, err := http.NewRequestWithContext(ctx, method, urlStr, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func encode(data map[string]string) url.Values {
	values := url.Values{}
	buf := &strings.Builder{}
	for _, k := range slices.Sorted(maps.Keys(data)) {
		v := data[k]
		buf.WriteString(k + "=" + v)
		values.Add(k, v)
	}

	hash := md5.Sum([]byte(buf.String() + "tiebaclient!!!"))
	sign := strings.ToUpper(hex.EncodeToString(hash[:]))

	values.Add("sign", sign)

	return values
}

func (c *Client) header() http.Header {
	h := http.Header{}

	h.Add("Host", "tieba.baidu.com")
	h.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")
	h.Add("Content-Type", "application/x-www-form-urlencoded")

	cookie := http.Cookie{
		Name:  "BDUSS",
		Value: c.bduss,
	}

	h.Add("Cookie", cookie.String())

	return h
}

func (c *Client) doWithJSON(req *http.Request, point any) (err error) {
	for k, values := range c.header() {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	var body []byte
	err = c.retry(func() (err error) {
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read body error: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http status code: %d", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 错误检查
	var check Error

	err = json.Unmarshal(body, &check)
	if err != nil {
		return fmt.Errorf("json decode check %s , err: %w", req.Host, err)
	}

	if check.Msg() != "" || check.Code() != "" {
		return &check
	}

	err = json.Unmarshal(body, point)
	if err != nil {
		return fmt.Errorf("json decode point %s , err: %w", req.Host, err)
	}

	return nil
}

func (c *Client) retry(f func() error) error {
	var err error

	for i := 0; i < c.maxRetries; i++ {
		err = f()
		if err == nil {
			return nil
		}

		time.Sleep(c.retryDelay)
	}

	return nil
}
