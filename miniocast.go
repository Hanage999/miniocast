package miniocast

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/eduncan911/podcast"

	"github.com/comail/colog"
	"github.com/minio/minio-go"
	"github.com/spf13/viper"
)

var (
	version       = "1"
	revision      = "0"
	maxRetry      = 5
	retryInterval = time.Duration(5) * time.Second
)

// Initialize は、設定ファイルを読み込んで初期化する
func Initialize() (casts []*PodcastPref, ct *minio.Client, err error) {
	// colog 設定
	if version == "" {
		colog.SetDefaultLevel(colog.LDebug)
		colog.SetMinLevel(colog.LTrace)
		colog.SetFormatter(&colog.StdFormatter{
			Colors: true,
			Flag:   log.Ldate | log.Ltime | log.Lshortfile,
		})
	} else {
		colog.SetDefaultLevel(colog.LDebug)
		colog.SetMinLevel(colog.LInfo)
		colog.SetFormatter(&colog.StdFormatter{
			Colors: true,
			Flag:   log.Ldate | log.Ltime,
		})
	}
	colog.Register()

	// 設定ファイル読み込み
	conf := viper.New()
	conf.SetConfigName("config")
	conf.AddConfigPath(".")
	conf.SetConfigType("yaml")
	if err := conf.ReadInConfig(); err != nil {
		log.Printf("alert: 設定ファイルが読み込めませんでした")
		return nil, nil, err
	}
	cred := conf.GetStringMapString("StorageCredentials")
	conf.UnmarshalKey("Podcasts", &casts)
	for i := range casts {
		casts[i].Link = "https://" + cred["server"] + "/" + casts[i].Bucket + "/" + casts[i].Folder
	}

	// クラウドストレージクライアントの生成
	ct, err = minio.New(cred["server"], cred["accesskey"], cred["secretkey"], true)
	if err != nil {
		log.Printf("alert: クラウドストレージクライアントが生成できませんでした：%s", err)
	}

	return
}

// getDetailsFromName は、オブジェクト名から各種情報を抽出する
func getDetailsFromName(key string) (id int, date, des string, err error) {
	fullName := strings.Trim(key, " ")
	fullName = strings.TrimSuffix(fullName, path.Ext(fullName))
	ss := strings.SplitN(fullName, " ", 2)
	if fullName == ss[0] {
		err = fmt.Errorf("no space in string: %s", fullName)
		return
	}

	date = ss[0]

	dai := strings.IndexRune(ss[1], '第')
	kai := strings.IndexRune(ss[1], '回')
	ids := ss[1][dai+3 : kai]
	id, _ = strconv.Atoi(ids)

	des = " "
	if kai+3 < len(ss[1]) {
		des = strings.Trim(ss[1][kai+3:], " ")
	}

	return
}

// getType は、ストレージオブジェクトのタイプを返す
func getType(info minio.ObjectInfo) (tp podcast.EnclosureType) {
	k := info.Key
	switch {
	case strings.HasSuffix(k, "mp3"):
		tp = podcast.MP3
	case strings.HasSuffix(k, "m4a"):
		tp = podcast.M4A
	case strings.HasSuffix(k, "m4b"):
		tp = podcast.M4A
	}
	return
}
