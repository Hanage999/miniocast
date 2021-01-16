package miniocast

import (
	"context"
	"fmt"
	"log"
	"path"
	"sort"
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

// UpdatePodcast はフォルダに入っている音声ファイルに応じてポッドキャストを更新する
func (pref *PodcastPref) UpdatePodcast(ct *minio.Client) (err error) {
	infos, err := pref.UpdatedInfos(ct)
	if err != nil {
		log.Printf("alert: %s の更新が必要か確認できませんでした：%s", pref.Title, err)
		return
	}
	if len(infos) > 0 {
		pref.UpdateRSS(infos, ct)
		pref.UpdateWeb(infos, ct)
	}

	return
}

// UpdatedInfos は、更新が必要なアイテムを返す
func (pref *PodcastPref) UpdatedInfos(ct *minio.Client) (updatedInfos FileInfos, err error) {
	olds, err := pref.fetchExistingIndexes(ct)
	if err != nil {
		log.Printf("info: %s のfeed.rssから既存アイテムを取り出せませんでした。：%s", pref.Folder, err)
	}

	infos, err := pref.fileList(ct)
	if err != nil {
		log.Printf("info: %s の中の音声ファイルリストが取得できませんでした：%s", pref.Folder, err)
		return
	}

	var news Indexes
	for _, info := range infos {
		var idx Index
		idx.FileLink = pref.Link + strings.TrimLeft(info.Key, pref.Folder)
		idx.Updated = info.LastModified.Format("Mon, 02 Jan 2006 15:04:05 -0700")
		news = append(news, idx)
	}

	needsUpdate := false
	if len(olds) != len(news) {
		needsUpdate = true
	} else {
		sort.Sort(olds)
		sort.Sort(news)
		for i, idx := range olds {
			if idx.FileLink != news[i].FileLink || idx.Updated != news[i].Updated {
				needsUpdate = true
				break
			}
		}
	}

	if needsUpdate {
		updatedInfos = infos
	}

	return
}

// fileList は、フォルダにある全音声ファイルリストを返す
func (pref *PodcastPref) fileList(ct *minio.Client) (fInfos FileInfos, err error) {
	ctx := context.Background()

	objectCh := ct.ListObjects(ctx, pref.Bucket, minio.ListObjectsOptions{
		Prefix:    pref.Folder + "/",
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			log.Printf("alert: %s のファイルリストの取得に失敗しました：%s", pref.Folder, object.Err)
			err = fmt.Errorf("%s", object.Err)
			return
		}
		k := strings.ToLower(object.Key)
		if strings.HasSuffix(k, "mp3") || strings.HasSuffix(k, "m4a") || strings.HasSuffix(k, "m4b") {
			fInfos = append(fInfos, object)
		}
	}

	sort.Sort(fInfos)

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
