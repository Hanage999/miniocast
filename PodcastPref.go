package miniocast

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/minio/minio-go"
)

// PodcastPref は、設定ファイルから読み込んだ各Podcastの情報を格納する
type PodcastPref struct {
	Title       string
	Subtitle    string
	Author      string
	Email       string
	Description string
	Link        string
	Bucket      string
	Folder      string
	Serial      bool
	Active      bool
}

// Update は、フィードを作成あるいは更新する
func (pref *PodcastPref) Update(ct *minio.Client) {
	pc := pref.newCast()

	items, err := pref.fetchRSSItems(ct)
	if err != nil {
		log.Printf("info: %s のfeed.rssが読み込めませんでした。：%s", pref.Folder, err)
	}

	newInfo, err := pref.fetchNewPodcastFilesInfo(ct, items)
	if err != nil {
		log.Printf("info: %s の中の新規音声ファイルリストが取得できませんでした：%s", pref.Folder, err)
		return
	}

	if len(newInfo) > 0 {
		newItems, err := pref.itemsFromInfo(newInfo, items)
		if err != nil {
			log.Printf("info: %s の新規アイテムの作成に失敗しました：%s", pref.Folder, err)
			return
		}

		for _, item := range newItems {
			_, _ = pc.AddItem(item)
		}

		pc.Items = append(pc.Items, items...)
		now := time.Now()
		pc.AddPubDate(&now)
		pc.AddLastBuildDate(&now)

		// log.Printf("info: %s", pc)
		if err := pref.upload(ct, pc); err != nil {
			log.Printf("info: feed.rssのアップロードに失敗しました：%s", err)
		}
	}

	return
}

// newCast は、Podcast構造体を初期化する
func (pref *PodcastPref) newCast() (pc *podcast.Podcast) {
	now := time.Now()
	pcr := podcast.New(pref.Title, pref.Link, pref.Description, &now, &now)
	pcr.AddAtomLink(pref.Link + "/feed.rss")
	if pref.Subtitle != "" {
		pcr.AddSubTitle(pref.Subtitle)
	}
	pcr.AddAuthor(pref.Author, pref.Email)
	pcr.AddCategory("Personal Journals", nil)
	pcr.AddImage(pref.Link + "/image.jpg")
	pcr.Language = "ja"
	pc = &pcr
	return
}

// fetchRSSItems は、feed.rssに含まれるRSSアイテムを返す
// xmlのデコード：https://qiita.com/chanmitsu55/items/8268f559efa694bd1cfd
func (pref *PodcastPref) fetchRSSItems(ct *minio.Client) (items []*podcast.Item, err error) {
	reader, err := ct.GetObject(pref.Bucket, pref.Folder+"/feed.rss", minio.GetObjectOptions{})
	if err != nil {
		log.Printf("info: %s のRSSファイルが取得できません：%s", pref.Folder, err)
		return
	}
	defer reader.Close()

	decoder := xml.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("info: %s のxmlのデコードに失敗しました：%s", pref.Folder, err)
			return items, err
		}
		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "item" {
				var item podcast.Item
				decoder.DecodeElement(&item, &se)
				items = append(items, &item)
			}
		}
	}

	return
}

// fetchNewPodcastFilesInfo は、ストレージに新規に追加された音声ファイルのObjectInfoを返す
func (pref *PodcastPref) fetchNewPodcastFilesInfo(ct *minio.Client, oldItems []*podcast.Item) (fInfos FileInfos, err error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	lastUpd := time.Time{}
	if len(oldItems) > 0 {
		layout := "Mon, 02 Jan 2006 15:04:05 -0700"
		lastUpd, _ = time.Parse(layout, oldItems[0].PubDateFormatted)
	}

	for object := range ct.ListObjectsV2(pref.Bucket, pref.Folder+"/", true, doneCh) {
		if object.Err != nil {
			log.Printf("alert: %s のファイルリストの取得に失敗しました：%s", pref.Folder, object.Err)
			err = fmt.Errorf("%s", object.Err)
			return
		}
		k := object.Key
		if strings.HasSuffix(k, "mp3") || strings.HasSuffix(k, "m4a") || strings.HasSuffix(k, "m4b") {
			newDate := object.LastModified.Truncate(time.Second)
			if lastUpd.Before(newDate) && !lastUpd.Equal(newDate) {
				fInfos = append(fInfos, object)
			}
		}
	}

	sort.Sort(fInfos)

	return
}

// itemsFromInfo は、音声ファイルのObjectInfoをもとに新規RSSアイテムの構造体を生成する
func (pref *PodcastPref) itemsFromInfo(fInfo FileInfos, existingItems []*podcast.Item) (newItems []podcast.Item, err error) {
	lastID := len(existingItems)
	if lastID > 0 {
		id, _, _ := getDetailsFromName(existingItems[0].Title)
		if id != 0 {
			lastID = id
		}
	}

	for i, info := range fInfo {
		item := podcast.Item{}
		fn := strings.TrimLeft(info.Key, pref.Folder+"/")
		id, title, des := getDetailsFromName(fn)
		idst := ""
		item.Description = des
		if id != 0 {
			idst = " 第" + strconv.Itoa(id) + "回"
		} else if pref.Serial == true {
			idst = " 第" + strconv.Itoa(lastID+len(fInfo)-i) + "回"
		}
		item.Title = title + idst
		url := pref.Link + strings.TrimLeft(info.Key, pref.Folder)
		tp := getType(info)
		item.AddEnclosure(url, tp, info.Size)

		upd := info.LastModified
		// 「for rangeのrangeの返り値には同じ参照先が使用されている。」
		// https://qiita.com/RunEagler/items/008e2b304f27b7fb168a
		// だから、&info.LastModifiedを引数に指定しても、それは最終的に全て同じ値になってしまう
		item.AddPubDate(&upd)
		newItems = append(newItems, item)
	}

	return
}

// upload は、クラウドストレージにfeed.rssをアップロードする
func (pref *PodcastPref) upload(ct *minio.Client, pc *podcast.Podcast) (err error) {
	bts := pc.Bytes()
	reader := bytes.NewReader(bts)
	_, err = ct.PutObject(pref.Bucket, pref.Folder+"/feed.rss", reader, int64(binary.Size(bts)), minio.PutObjectOptions{ContentType: "application/rss+xml"})
	if err != nil {
		log.Printf("alert: %s のrssファイルのアップロードに失敗しました：%s", pref.Folder, err)
	}

	return
}
