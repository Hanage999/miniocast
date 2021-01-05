package miniocast

import (
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/eduncan911/podcast"

	"github.com/comail/colog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	cred := conf.GetStringMapString("Storage")
	conf.UnmarshalKey("Podcasts", &casts)
	uselcl := conf.GetString("SavePlayState")

	s := ""
	if cred["https"] == "true" {
		s = "s"
	}

	vh := cred["bucketasvirtualhost"] == "true"

	for i := range casts {
		if vh {
			casts[i].Link = "http" + s + "://" + casts[i].Bucket + "." + cred["server"] + "/" + casts[i].Folder
		} else {
			casts[i].Link = "http" + s + "://" + cred["server"] + "/" + casts[i].Bucket + "/" + casts[i].Folder
		}
		if uselcl == "true" {
			casts[i].SavePlayState = true
		}
	}

	// クラウドストレージクライアントの生成
	ct, err = minio.New(cred["server"], &minio.Options{
		Creds:  credentials.NewStaticV4(cred["accesskey"], cred["secretkey"], ""),
		Secure: true,
	})
	if err != nil {
		log.Printf("alert: クラウドストレージクライアントが生成できませんでした：%s", err)
	}

	return
}

// getDetailsFromName は、オブジェクト名から各種情報を抽出する
func getDetailsFromName(key string) (id int, title, des string) {
	fullName := strings.Trim(key, " ")
	fullName = strings.TrimSuffix(fullName, path.Ext(fullName))
	des = " "

	ss := strings.SplitN(fullName, " ", 2)
	if fullName == ss[0] {
		log.Printf("trace: no space in string: %s", fullName)
		title = fullName
		return
	}

	title = ss[0]
	des = strings.Trim(ss[1], " ")

	dai := strings.IndexRune(ss[1], '第')
	kai := strings.IndexRune(ss[1][dai+3:], '回')
	if kai != -1 {
		kai += dai + 3
		ids := ss[1][dai+3 : kai]
		var err error
		id, err = strconv.Atoi(ids)
		if kai+3 < len(ss[1]) && err == nil {
			des = strings.Trim(ss[1][kai+3:], " ")
		}
	}

	return
}

// getType は、ストレージオブジェクトのタイプを返す
func getType(info minio.ObjectInfo) (tp podcast.EnclosureType) {
	k := strings.ToLower(info.Key)
	switch {
	case strings.HasSuffix(k, "mp3"):
		tp = podcast.MP3
	default:
		tp = podcast.M4A
	}
	return
}
