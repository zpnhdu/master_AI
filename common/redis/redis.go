package redis

import (
	"GopherAI/config"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	redisCli "github.com/redis/go-redis/v9"
)

var Rdb *redisCli.Client

var ctx = context.Background()

type captchaEntry struct {
	value     string
	expiresAt time.Time
}

var captchaFallback = struct {
	sync.RWMutex
	items map[string]captchaEntry
}{items: make(map[string]captchaEntry)}

func Init() {
	conf := config.GetConfig()
	host := conf.RedisConfig.RedisHost
	port := conf.RedisConfig.RedisPort
	password := conf.RedisConfig.RedisPassword
	db := conf.RedisDb
	addr := host + ":" + strconv.Itoa(port)

	Rdb = redisCli.NewClient(&redisCli.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		Protocol: 2, // 使用 Protocol 2 避免 maint_notifications 警告
	})

}

func SetCaptchaForEmail(email, captcha string) error {
	key := GenerateCaptcha(email)
	expire := 2 * time.Minute
	if Rdb != nil {
		if err := Rdb.Set(ctx, key, captcha, expire).Err(); err == nil {
			return nil
		}
	}
	captchaFallback.Lock()
	captchaFallback.items[key] = captchaEntry{
		value:     captcha,
		expiresAt: time.Now().Add(expire),
	}
	captchaFallback.Unlock()
	fmt.Printf("[dev redis fallback] captcha for %s is %s\n", email, captcha)
	return nil
}

func CheckCaptchaForEmail(email, userInput string) (bool, error) {
	key := GenerateCaptcha(email)

	if Rdb == nil {
		return checkFallbackCaptcha(key, userInput), nil
	}

	storedCaptcha, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redisCli.Nil {

			return false, nil
		}

		return checkFallbackCaptcha(key, userInput), nil
	}

	if strings.EqualFold(storedCaptcha, userInput) {

		// 验证成功后删除 key
		if err := Rdb.Del(ctx, key).Err(); err != nil {

		} else {

		}
		return true, nil
	}

	return false, nil
}

func checkFallbackCaptcha(key, userInput string) bool {
	captchaFallback.Lock()
	defer captchaFallback.Unlock()
	entry, ok := captchaFallback.items[key]
	if !ok {
		return false
	}
	if time.Now().After(entry.expiresAt) {
		delete(captchaFallback.items, key)
		return false
	}
	if strings.EqualFold(entry.value, userInput) {
		delete(captchaFallback.items, key)
		return true
	}
	return false
}

// InitRedisIndex 初始化 Redis 索引，支持按文件名区分
func InitRedisIndex(ctx context.Context, filename string, dimension int) error {
	indexName := GenerateIndexName(filename)

	// 检查索引是否存在
	_, err := Rdb.Do(ctx, "FT.INFO", indexName).Result()
	if err == nil {
		fmt.Println("索引已存在，跳过创建")
		return nil
	}

	// 如果索引不存在，创建新索引
	if !strings.Contains(err.Error(), "Unknown index name") {
		return fmt.Errorf("检查索引失败: %w", err)
	}

	fmt.Println("正在创建 Redis 索引...")

	prefix := GenerateIndexNamePrefix(filename)

	// 创建索引
	createArgs := []interface{}{
		"FT.CREATE", indexName,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",
		"content", "TEXT",
		"metadata", "TEXT",
		"vector", "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	if err := Rdb.Do(ctx, createArgs...).Err(); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	fmt.Println("索引创建成功！")
	return nil
}

// DeleteRedisIndex 删除 Redis 索引，支持按文件名区分
func DeleteRedisIndex(ctx context.Context, filename string) error {
	indexName := GenerateIndexName(filename)

	// 删除索引
	if err := Rdb.Do(ctx, "FT.DROPINDEX", indexName).Err(); err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}

	fmt.Println("索引删除成功！")
	return nil
}
