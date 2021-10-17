package miniocast

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/xml"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/eduncan911/podcast"
	"github.com/minio/minio-go/v7"
)

// RSS はfeed.rss のデータ全体。
type RSS struct {
	XMLName xml.Name        `xml:"rss"`
	Channel podcast.Podcast `xml:"channel"`
}

// UpdateRSS は、フィードを作成あるいは更新する
func (pref *PodcastPref) UpdateRSS(infos FileInfos, ct *minio.Client) {
	rss := pref.newRSS()

	items, err := pref.itemsFromInfo(infos)
	if err != nil {
		log.Printf("info: %s の新規アイテムの作成に失敗しました：%s", pref.Folder, err)
		return
	}

	for _, item := range items {
		_, _ = rss.AddItem(item)
	}

	now := time.Now()
	rss.AddPubDate(&now)
	rss.AddLastBuildDate(&now)

	// log.Printf("info: %s", rss)
	if err := pref.uploadRSS(ct, rss); err != nil {
		log.Printf("info: feed.rssのアップロードに失敗しました：%s", err)
	}
}

// newRSS は、Podcast構造体を初期化する
func (pref *PodcastPref) newRSS() (rss *podcast.Podcast) {
	now := time.Now()
	rssr := podcast.New(pref.Title, pref.Link+"/index.html", pref.Description, &now, &now)
	rssr.AddAtomLink(pref.Link + "/feed.rss")
	if pref.Subtitle != "" {
		rssr.AddSubTitle(pref.Subtitle)
	}
	rssr.AddAuthor(pref.Author, pref.Email)
	rssr.AddCategory("Personal Journals", nil)
	rssr.AddImage(pref.Link + "/image.jpg")
	rssr.Language = "ja"
	rss = &rssr
	return
}

// fetchExistingIndexes は、既存のアイテムのインデックスを返す
func (pref *PodcastPref) fetchExistingIndexes(ct *minio.Client) (olds Indexes, err error) {
	ctx := context.Background()
	reader, err := ct.GetObject(ctx, pref.Bucket, pref.Folder+"/feed.rss", minio.GetObjectOptions{})
	if err != nil {
		log.Printf("info: %s のRSSファイルが取得できません：%s", pref.Folder, err)
		return
	}
	defer reader.Close()

	var items []*podcast.Item
	decoder := xml.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("info: %s のxmlのデコードに失敗しました：%s", pref.Folder, err)
			return Indexes{}, err
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

	for _, item := range items {
		idx := Index{}
		idx.FileLink = item.Link
		idx.Updated = item.PubDateFormatted
		olds = append(olds, idx)
	}

	return
}

// itemsFromInfo は、音声ファイルのObjectInfoをもとにRSSアイテムの構造体を生成する
func (pref *PodcastPref) itemsFromInfo(fInfos FileInfos) (items []podcast.Item, err error) {
	lastID := len(fInfos)
	if lastID == 0 {
		return
	}

	for i, info := range fInfos {
		item := podcast.Item{}
		fn := strings.TrimLeft(info.Key, pref.Folder+"/")
		id, title, sub := getDetailsFromName(fn)
		idst := ""
		// Descriptionは空にできない。
		item.Description = "　"
		item.ISubtitle = sub
		if id != 0 {
			idst = " 第" + strconv.Itoa(id) + "回"
		} else if pref.Serial > 0 {
			idst = " 第" + strconv.Itoa(lastID+pref.Serial-1-i) + "回"
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
		items = append(items, item)
	}

	return
}

// uploadRSS は、クラウドストレージにfeed.rssをアップロードする
func (pref *PodcastPref) uploadRSS(ct *minio.Client, rss *podcast.Podcast) (err error) {
	ctx := context.Background()
	bts := rss.Bytes()
	reader := bytes.NewReader(bts)
	_, err = ct.PutObject(ctx, pref.Bucket, pref.Folder+"/feed.rss", reader, int64(binary.Size(bts)), minio.PutObjectOptions{ContentType: "application/rss+xml"})
	if err != nil {
		log.Printf("alert: %s のrssファイルのアップロードに失敗しました：%s", pref.Folder, err)
	}

	return
}
