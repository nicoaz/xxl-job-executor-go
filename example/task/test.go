package task

import (
	"context"
	"fmt"

	xxl "github.com/nicoaz/xxl-job-executor-go"
)

func Test(cxt context.Context, param *xxl.RunReq) (int64, string) {
	fmt.Println("test one task" + param.ExecutorHandler + " param：" + param.ExecutorParams + " log_id:" + xxl.Int64ToStr(param.LogID))
	return xxl.SuccessCode, "test done"
}
