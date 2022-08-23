package v1

import (
	"net/http"

	"github.com/allanpk716/ChineseSubFinder/pkg/types/backend"

	"github.com/allanpk716/ChineseSubFinder/pkg/lock"
	"github.com/allanpk716/ChineseSubFinder/pkg/logic/cron_helper"
	"github.com/allanpk716/ChineseSubFinder/pkg/sub_formatter"
	"github.com/allanpk716/ChineseSubFinder/pkg/video_scan_and_refresh_helper"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ControllerBase struct {
	log                                 *logrus.Logger
	cronHelper                          *cron_helper.CronHelper
	pathUrlMap                          map[string]string
	videoScanAndRefreshHelper           *video_scan_and_refresh_helper.VideoScanAndRefreshHelper
	videoScanAndRefreshHelperIsRunning  bool
	videoScanAndRefreshHelperLocker     lock.Lock
	videoScanAndRefreshHelperErrMessage string
	restartSignal                       chan interface{}
}

func NewControllerBase(log *logrus.Logger, cronHelper *cron_helper.CronHelper, restartSignal chan interface{}) *ControllerBase {
	cb := &ControllerBase{
		log:        log,
		cronHelper: cronHelper,
		pathUrlMap: make(map[string]string),
		// 这里因为不进行任务的添加，仅仅是扫描，所以 downloadQueue 可以为 nil
		videoScanAndRefreshHelper: video_scan_and_refresh_helper.NewVideoScanAndRefreshHelper(
			sub_formatter.GetSubFormatter(log, cronHelper.Settings.AdvancedSettings.SubNameFormatter),
			cronHelper.FileDownloader, nil),
		videoScanAndRefreshHelperLocker: lock.NewLock(),
		restartSignal:                   restartSignal,
	}

	return cb
}

func (cb *ControllerBase) SetPathUrlMapItem(path string, url string) {
	cb.pathUrlMap[path] = url
}

// GetPathUrlMap x://电影 -- /movie_dir_0  or x://电视剧 -- /series_dir_0
func (cb *ControllerBase) GetPathUrlMap() map[string]string {
	return cb.pathUrlMap
}

func (cb *ControllerBase) Close() {
	cb.cronHelper.Stop()
	cb.videoScanAndRefreshHelper.Cancel()
	cb.videoScanAndRefreshHelperLocker.Close()
}

func (cb *ControllerBase) GetVersion() string {
	return "v1"
}

func (cb *ControllerBase) ErrorProcess(c *gin.Context, funcName string, err error) {
	if err != nil {
		cb.log.Errorln(funcName, err.Error())
		c.JSON(http.StatusInternalServerError, backend.ReplyCommon{Message: err.Error()})
	}
}
