package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/subosito/twilio"
	credis "gitlab.99safe.org/rrp/rrp-backend/redis"
)

type TwilioSmsService struct {
}

var twilioSmsService *TwilioSmsService

//https://www.twilio.com/console
const accountSid = "AC7b859488303df3de07a1b38126f3d983"
const authToken = "7ca7e00ce564147baf2eda6f0af7ee99"
const formTel = "+12566935185"

func NewTwilioSmsService() *TwilioSmsService {
	if twilioSmsService == nil {
		l.Lock()
		if twilioSmsService == nil {
			twilioSmsService = &TwilioSmsService{}
		}
		l.Unlock()
	}
	return twilioSmsService
}

//VerifySMSCode 验证短信验证码
func (twilioSmsService TwilioSmsService) VerifySMSCode(smsCode string, receiveTel string) bool {
	//根据手机号码和验证码去缓存里取之前存入的数据，取到验证通过，否则验证失败。
	key := fmt.Sprintf("%s:Sms:%s:", receiveTel, smsCode)
	conn := credis.RedisClientManagerInstance().Client().Get()
	defer conn.Close()
	//	labels, err := redis.Strings(conn.Do("HVALS", key))
	keys, err := redis.Strings(conn.Do(KEYS, key)) //根据手机号码和短信验证码去取缓存
	if err != nil {
		Log.Error(err)
		return false
	}
	Log.Debug("redis with tel and smscode data:", keys)
	if len(keys) == 0 {
		return false
	}
	return true
}

func (twilioSmsService TwilioSmsService) SendMessageCode(code string, receiveTel string) (err error) {
	conn := credis.RedisClientManagerInstance().Client().Get()
	defer conn.Close()
	//首先判断redis 是否存在该电话号码
	keys, err := redis.Strings(conn.Do(KEYS, fmt.Sprintf("%s:Sms:*:", receiveTel)))
	if err != nil {
		Log.Error(err)
		return err
	}

	if len(keys) > 0 {
		err = SendMessageCodeError{
			error: errors.New("SendMessageCodeError"),
		}
		Log.Error(err)
		return err
	}

	c := twilio.NewTwilio(accountSid, authToken)

	//生成短信验证码，并且记录当前时间以及接收短信的付客信息，存入缓存
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	smsCode := fmt.Sprintf("%06v", rnd.Int31n(1000000))

	//redis key
	key := fmt.Sprintf("%s:Sms:%s:", receiveTel, smsCode)

	_, err = conn.Do("HMSET", key, "smsCode", smsCode, "createTime", time.Now().Unix())
	if err != nil {
		Log.Error(err)
		return err
	}
	//设置过期时间
	_, err = conn.Do("EXPIRE", key, 120)
	if err != nil {
		Log.Error(err)
		return err
	}

	//发送短信url:https://sms.server.matocloud.com/sms/is/api/sms/simple/sendSms
	//API接口user：cqwqk
	//API接口userKey：eOLqF7gJ6dgDDxj4

	//网宿云发送短信请求参数：
	//1: auth-user=cqwqk

	// timestamp := time.Now()
	// authTimeStamp := timestamp.Format("20060102150405")
	//2:auth-timeStamp=authTimeStamp

	//3:auth-signature

	//发送短信
	_, err = c.SendSMS(formTel, fmt.Sprintf("+%s%s", code, receiveTel), smsCode, twilio.SMSParams{})
	if err != nil {
		Log.Error(err)
		return err
	}

	// //校验接收验证码的手机号码上一次发送验证码的时间与当前时间必须要大于等于1分钟
	// smsCodeTimeWithRedis, err := redis.String(conn.Do("HGET", key, "createTime"))

	// //获取存入缓存的时间戳
	// labeljson2, err := redis.String(conn.Do("HGET", rece, "createTime"))
	// Log.Debug("SMSCode with redis of the time:", labeljson2)
	// //获取存入缓存的短信验证码
	// labeljson3, err := redis.String(conn.Do("HGET", rece, "smsCode"))
	// Log.Debug("SMSCode with redis of the smsCode:", labeljson3)

	// go NewSysLogService().CreateSysLog(ctx.SysLog, payer, receiveTel)
	// Log.Debug("SMSCode with redis of the time:", smsCodeTimeWithRedis)

	// smsCodeTimeWithRedisInt64, err := strconv.ParseInt(smsCodeTimeWithRedis, 10, 64)
	// if err != nil {
	// 	return err
	// }

	// if time.Now().Unix()-smsCodeTimeWithRedisInt64 < 60 {
	// 	return IntervalIsOneMinuteError{
	// 		error: errors.New("Interval is one minute."),
	// 	}
	// }

	// payerService := NewPayerService()
	// payer := model.Payer{}
	//除注册付客操作外，其他操作的发送验证码均需要去查询数据库是否存在接收验证码的手机号码
	// if payerID > 0 {
	// 	payer, err = payerService.GetPayerByID(payerID)
	// 	if err != nil {
	// 		return FindPayersError{
	// 			error: errors.New("find payer by id  error."),
	// 		}
	// 	}
	// 	//获取该付客数据库存入的手机号码与填写的手机号码进行对比
	// 	payerTel := payer.Tel
	// 	if payerTel != receiveTel {
	// 		return TelNotMatchError{
	// 			error: errors.New("TelNotMatchError."),
	// 		}
	// 	}
	// }

	// if err != nil {
	// 	//发送失败则要删除redis
	// 	_, err = conn.Do("HDEL", rece)
	// 	if err != nil {
	// 		Log.Error(err)
	// 		return err
	// 	}
	// 	return SendMessageCodeError{
	// 		error: errors.New("send message code error."),
	// 	}
	// }

	// keys, err := redis.Strings(conn.Do(KEYS, rece))
	// Log.Debug("SMSCode with redis of the key:", keys)

	// //获取存入缓存的时间戳
	// labeljson2, err := redis.String(conn.Do("HGET", rece, "createTime"))
	// Log.Debug("SMSCode with redis of the time:", labeljson2)
	// //获取存入缓存的短信验证码
	// labeljson3, err := redis.String(conn.Do("HGET", rece, "smsCode"))
	// Log.Debug("SMSCode with redis of the smsCode:", labeljson3)

	return nil
}
