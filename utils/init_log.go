package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	nested_formatter "github.com/antonfisher/nested-logrus-formatter"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	DEFAULT_ROTATE_LOGFILE = 7
	DEFAULT_ROTATE_MBYTES  = 10
	MAX_ROTATE_LOGFILE     = 70
	MAX_ROTATE_MBYTES      = 100
)

// logfile is log filename such as myserver.log
// default rotate 7 files with 10M per file.
func InitLog(logfile string, isdebug bool) error {
	return InitLogRotate(logfile, isdebug, DEFAULT_ROTATE_LOGFILE, DEFAULT_ROTATE_MBYTES)
}

// logfile is logfilename such as myserver.log
// default rotate 7 files with 10M per file. rotate_mbytes is MBytes.
func InitLogRotate(logfile string, isdebug bool, rotate_files, rotate_mbytes uint) error {
	logrus.SetFormatter(&nested_formatter.Formatter{
		HideKeys:        true,
		TimestampFormat: "01-02 15:04:05", //time.DateTime, time.RFC3339,
		// FieldsOrder:     []string{"model", "file"},
		CallerFirst: true,
		CustomCallerFormatter: func(f *runtime.Frame) string {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return fmt.Sprintf(" [%s:%d %s()]", path.Base(f.File), f.Line, funcName)
		},
	})
	//logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true, FullTimestamp: true})
	//logrus.SetFormatter(&logrus.JSONFormatter{})

	if isdebug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetReportCaller(true) // 设置在输出日志中添加文件名和方法信息

	// create subdir 'log'
	if _, err := os.Stat("log"); err != nil {
		if err = os.Mkdir("log", 0755); err != nil {
			logrus.Errorf("create subdir 'log' failed: %s", err)
			return err
		}
	}

	// set log file how to rotate
	if rotate_files > MAX_ROTATE_LOGFILE {
		logrus.Warnf("rotate_files %d is bigger than %d, set to %d", rotate_files, MAX_ROTATE_LOGFILE, MAX_ROTATE_LOGFILE)
		rotate_files = MAX_ROTATE_LOGFILE
	}
	if rotate_mbytes > MAX_ROTATE_MBYTES {
		logrus.Warnf("rotate_mbytes %dM is bigger than %dM, set to %dM", rotate_mbytes, MAX_ROTATE_MBYTES, MAX_ROTATE_MBYTES)
		rotate_mbytes = MAX_ROTATE_MBYTES
	}
	logf, err := rotatelogs.New(
		"log/"+logfile+".%Y%m%d",
		//rotatelogs.WithLinkName(prgname),
		//rotatelogs.WithMaxAge(-1),
		//rotatelogs.WithRotationCount(10),
		rotatelogs.WithRotationCount(rotate_files),                  // max number log files
		rotatelogs.WithRotationSize(int64(rotate_mbytes*1024*1024)), // bytes per log file
	)
	if err != nil {
		logrus.Errorf("failed to create rotatelogs: %s", err)
		return err
	}
	logrus.SetOutput(io.MultiWriter(os.Stdout, logf))

	return nil
}
