package miniocast

import (
	"bytes"
	"context"
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
	"github.com/minio/minio-go/v7"
)

// RSS はfeed.rss のデータ全体。
type RSS struct {
	XMLName xml.Name        `xml:"rss"`
	Channel podcast.Podcast `xml:"channel"`
}

// UpdateRSS は、フィードを作成あるいは更新する
func (pref *PodcastPref) UpdateRSS(ct *minio.Client) {
	lastupdate, err := pref.fetchRSSLastupdate(ct)
	if err != nil {
		log.Printf("info: %s のfeed.rssが読み込めませんでした。：%s", pref.Folder, err)
	}

	infos, err := pref.renewedList(ct, lastupdate)
	if err != nil {
		log.Printf("info: %s の中の新規音声ファイルリストが取得できませんでした：%s", pref.Folder, err)
		return
	}

	if len(infos) > 0 {
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

	return
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

// fetchRSSLastupdate は、feed.rssの最新アイテムの更新日時を返す
func (pref *PodcastPref) fetchRSSLastupdate(ct *minio.Client) (lastupdate time.Time, err error) {
	ctx := context.Background()
	reader, err := ct.GetObject(ctx, pref.Bucket, pref.Folder+"/feed.rss", minio.GetObjectOptions{})
	if err != nil {
		log.Printf("info: %s のRSSファイルが取得できません：%s", pref.Folder, err)
		return
	}
	defer reader.Close()

	/*	var rss RSS
		if err = xml.NewDecoder(reader).Decode(&rss); err != nil {
			log.Printf("info: xmlデータを構造体に読み込めませんでした：%v", err)
			return
		}
		updstr := rss.Channel.LastBuildDate

		lastupdate, _ = time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", updstr)
	*/
	var items []*podcast.Item
	decoder := xml.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("info: %s のxmlのデコードに失敗しました：%s", pref.Folder, err)
			return time.Time{}, err
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

	if len(items) > 0 {
		updstr := items[0].PubDateFormatted
		lastupdate, _ = time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", updstr)
	}

	return
}

// renewedList は、ストレージに新しくファイルが追加されたら全ファイルリストを返す
func (pref *PodcastPref) renewedList(ct *minio.Client, lastupdate time.Time) (fInfos FileInfos, err error) {
	ctx := context.Background()

	new := false
	var infos FileInfos

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
			infos = append(infos, object)
			newDate := object.LastModified.Truncate(time.Second)
			if lastupdate.Before(newDate) && !lastupdate.Equal(newDate) {
				new = true
			}
		}
	}

	if new {
		sort.Sort(infos)
		fInfos = infos
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
		} else if pref.Serial == true {
			idst = " 第" + strconv.Itoa(lastID-i) + "回"
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
