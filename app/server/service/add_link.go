package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unsafe"

	"github.com/1911860538/short_link/app/component"
	"github.com/1911860538/short_link/config"
)

type AddLinkSvc struct {
	Database component.DatabaseItf
}

type AddLinkParams struct {
	UserId   string
	LongUrl  string
	Deadline time.Time
}

type AddLinkRes struct {
	StatusCode int
	Msg        string

	Code string
}

const (
	msgUrlInvalid = "不是一个合法的http或https链接"
	msgUrlRespErr = "链接请求未能正常响应"
)

var (
	confLongUrlConnTimeout = time.Duration(config.Conf.Core.LongUrlConnTimeout) * time.Second
	confExpiredKeepHours   = time.Duration(config.Conf.Core.ExpiredKeepDays*24) * time.Hour
)

func (s *AddLinkSvc) Do(ctx context.Context, params AddLinkParams) (AddLinkRes, error) {
	// 检查url合法性
	u, err := url.ParseRequestURI(params.LongUrl)
	if err != nil {
		return s.badRequest(msgUrlInvalid)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return s.badRequest(msgUrlInvalid)
	}
	client := http.Client{
		Timeout: confLongUrlConnTimeout,
	}
	// 优先使用HEAD请求，如果服务器不支持HEAD请求，再使用GET请求
	headResp, err := client.Head(u.String())
	if err != nil {
		return s.badRequest(msgUrlRespErr)
	}
	if headResp.Body != nil {
		defer headResp.Body.Close()
	}
	respOk := headResp.StatusCode == http.StatusOK
	if !respOk && headResp.StatusCode == http.StatusMethodNotAllowed {
		getResp, err := client.Get(u.String())
		if err != nil {
			return s.badRequest(msgUrlRespErr)
		}
		if getResp.Body != nil {
			defer getResp.Body.Close()
		}
		respOk = getResp.StatusCode == http.StatusOK
	}
	if !respOk {
		return s.badRequest(msgUrlRespErr)
	}

	// 检查这个userId是不是已经生成了此longUrl的code
	filter := map[string]any{
		"user_id":  params.UserId,
		"long_url": params.LongUrl,
	}
	oldLink, err := s.Database.Get(ctx, filter)
	if err != nil {
		return s.internalErr(err)
	}
	if oldLink != nil && !oldLink.Expired() {
		return s.codeConflicted(oldLink.Code)
	}

	// 生成longUrl对应的code
	code, err := GenCode(params.UserId, params.LongUrl, "")
	if err != nil {
		return s.internalErr(err)
	}

	// 保存到数据库，这里要注意可能和数据库code冲突
	var ttlTime time.Time
	if params.Deadline.IsZero() {
		ttlTime = time.Time{}
	} else {
		ttlTime = params.Deadline.Add(confExpiredKeepHours)
	}
	link := &component.Link{
		UserId:    params.UserId,
		Code:      code,
		Salt:      "",
		LongUrl:   params.LongUrl,
		Deadline:  params.Deadline,
		TtlTime:   ttlTime,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.trySaveLink(ctx, link); err != nil {
		return s.internalErr(err)
	}

	return s.ok(link.Code)
}

func (s *AddLinkSvc) ok(code string) (AddLinkRes, error) {
	return AddLinkRes{
		StatusCode: http.StatusCreated,
		Code:       code,
	}, nil
}

func (s *AddLinkSvc) badRequest(errMsg string) (AddLinkRes, error) {
	return AddLinkRes{
		StatusCode: http.StatusBadRequest,
		Msg:        errMsg,
	}, nil
}

func (s *AddLinkSvc) codeConflicted(code string) (AddLinkRes, error) {
	return AddLinkRes{
		StatusCode: http.StatusConflict,
		Msg:        fmt.Sprintf("你已对该链接已生成了对应的短链接，短链接code为：%s", code),
	}, nil
}

func (s *AddLinkSvc) internalErr(err error) (AddLinkRes, error) {
	return AddLinkRes{
		StatusCode: http.StatusInternalServerError,
	}, err
}

func (s *AddLinkSvc) trySaveLink(ctx context.Context, link *component.Link) error {
	_, existed, err := s.Database.Create(ctx, link)
	if err != nil {
		return err
	}
	if !existed {
		return nil
	}

	nowTimestampStr := strconv.FormatInt(time.Now().UnixMilli(), 10)
	link.Salt = nowTimestampStr
	link.Code, err = GenCode(link.UserId, link.Code, nowTimestampStr)
	if err != nil {
		return err
	}

	return s.trySaveLink(ctx, link)
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenCode
/*
下面一通计算，
和随机生成字母数字code的区别是，
尽量保证同样的userId+longUrl每次生成的code一样，
如果userId+longUrl生成了数据库已有的code，
则加上当前时间戳字符串作为盐salt，
递归，直到生成的code数据库中没有
*/
func GenCode(userId string, longUlr string, salt string) (string, error) {
	// 首先对userId+longUrl+salt md5 主要为了防止longUrl包含汉字等字符串
	hasher := md5.New()
	if _, err := io.WriteString(hasher, userId+longUlr+salt); err != nil {
		return "", err
	}
	hashStr := hex.EncodeToString(hasher.Sum(nil))

	stepLen := len(hashStr) / confCodeLen
	remain := len(hashStr) % confCodeLen
	if remain > 0 {
		stepLen += 1
	}
	lettersLen := uint32(len(letters))
	b := make([]byte, confCodeLen)

	for i := 0; i < confCodeLen; i++ {
		// 根据要生成的code长度，切分md5字符串
		var piece string
		if remain > 0 && i == confCodeLen-1 {
			piece = hashStr[i*stepLen : i*stepLen+remain]
		} else {
			piece = hashStr[i*stepLen : i*stepLen+stepLen]
		}

		// 为切片元素生成对应的整形数值
		h := fnv.New32a()
		pieceBytes := unsafe.Slice(unsafe.StringData(piece), len(piece))
		if _, err := h.Write(pieceBytes); err != nil {
			return "", err
		}
		pieceHash32 := h.Sum32()

		// 切片字符的整形，取len(letters)余数，并取letters索引为该余数的letter
		letterIdx := pieceHash32 % lettersLen
		b[i] = letters[letterIdx]
	}

	return unsafe.String(unsafe.SliceData(b), len(b)), nil
}
