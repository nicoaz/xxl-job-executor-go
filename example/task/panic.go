package task

import (
	"context"

	xxl "github.com/nicoaz/xxl-job-executor-go"
)

func Panic(cxt context.Context, param *xxl.RunReq) (int64, string) {
	panic("test panic")
}
