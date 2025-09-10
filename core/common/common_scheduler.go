package common

import (
	"BaseGoUni/core/utils"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func InitScheduler() {
	// 初始化定时任务
	c := cron.New(cron.WithSeconds())
	// 是否运行调度器的检查
	if !utils.CsConfig.RunScheduler {
		c.Start()
		return
	}

	// 公共的任务添加函数，减少重复代码
	addScheduledTask := func(schedule string, lockKey string, lockDuration time.Duration, task func(), logMessage string) {
		_, err := c.AddFunc(schedule, func() {
			if logMessage != "" {
				log.Printf("开始任务: %s", logMessage)
			}
			lock, _ := utils.AcquireLock(lockKey, lockDuration)
			if !lock {
				if logMessage != "" {
					log.Printf("任务获取锁失败: %s", logMessage)
				}
				return
			}
			startTime := time.Now()
			defer func() {
				utils.ReleaseLock(lockKey)
				if logMessage != "" {
					log.Printf("任务执行完毕: %s;time=%.2fs", logMessage, time.Now().Sub(startTime).Seconds())
				}
			}()
			task()
		})
		if err != nil {
			log.Printf("添加任务失败: %s, error: %v", logMessage, err)
		} else {
			log.Printf("添加任务成功: %s", logMessage)
		}
	}

	addScheduledTask("*/10 * * * * *", "host_info_get", 1*time.Minute, func() {
		utils.FlushTempHostInfo()
	}, "")
	c.Start()
	log.Println("Scheduler started successfully")
}
