package task

import (
	"context"
	"fmt"
	"time"

	xxl "github.com/nicoaz/xxl-job-executor-go"
)

func Test2(cxt context.Context, param *xxl.RunReq) (int64, string) {
	num := 1
	for {
		select {
		case <-cxt.Done():
			fmt.Println("task" + param.ExecutorHandler + "被手动终止")
			return xxl.FailureCode, "task canceled"
		default:
			num++
			time.Sleep(10 * time.Second)
			fmt.Println("test one task"+param.ExecutorHandler+" param："+param.ExecutorParams+"执行行", num)
			if num > 10 {
				fmt.Println("test one task" + param.ExecutorHandler + " param：" + param.ExecutorParams + "执行完毕！")
				return xxl.SuccessCode, "task done"
			}
		}
	}
}
