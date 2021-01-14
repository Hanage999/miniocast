package miniocast

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"golang.org/x/net/html"
)

// Web は、index.htmlに含めるデータを格納する
type Web struct {
	Title         string
	Subtitle      string
	Author        string
	Description   string
	Link          string
	Items         []*WebItem
	SavePlayState bool
}

// WebItem は、index.htmlに含める各エピソードのデータを格納する
type WebItem struct {
	FileURL          string
	PubDateFormatted string
	Title            string
	Subtitle         string
	Description      string
}

// UpdateWeb は、フィードを作成あるいは更新する
func (pref *PodcastPref) UpdateWeb(ct *minio.Client) {
	items, err := pref.fetchWebItems(ct)
	if err != nil {
		log.Printf("info: %s のindex.htmlが読み込めませんでした。：%s", pref.Folder, err)
	}

	newInfo, err := pref.fetchNewWebItemsInfo(ct, items)
	if err != nil {
		log.Printf("info: %s の中の新規音声ファイルリストが取得できませんでした：%s", pref.Folder, err)
		return
	}

	if len(newInfo) > 0 {
		web := pref.newWeb()

		newItems, err := pref.webItemsFromInfo(newInfo, items)
		if err != nil {
			log.Printf("info: %s の新規Webアイテムの作成に失敗しました：%s", pref.Folder, err)
			return
		}

		web.Items = append(newItems, items...)

		// log.Printf("info: %v", web)
		if err := pref.uploadWeb(ct, &web); err != nil {
			log.Printf("info: index.htmlのアップロードに失敗しました：%s", err)
		}
	}

	return
}

// newWeb は、webデータを初期化する
func (pref *PodcastPref) newWeb() (web Web) {
	web.Title = pref.Title
	web.Subtitle = pref.Subtitle
	web.Author = pref.Author
	web.Description = pref.Description
	web.Link = pref.Link
	web.SavePlayState = pref.SavePlayState
	return
}

// fetchWebItems は、index.htmlに含まれるアイテムを返す
// xmlのデコード：https://qiita.com/chanmitsu55/items/8268f559efa694bd1cfd
func (pref *PodcastPref) fetchWebItems(ct *minio.Client) (items []*WebItem, err error) {
	ctx := context.Background()
	reader, err := ct.GetObject(ctx, pref.Bucket, pref.Folder+"/index.html", minio.GetObjectOptions{})
	if err != nil {
		log.Printf("info: %s のindex.htmlが取得できません：%s", pref.Folder, err)
		return
	}
	defer reader.Close()

	root, err := html.Parse(reader)
	if err != nil {
		log.Printf("info: %s のindex.htmlがパースできません：%s", pref.Folder, err)
		return
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "episode" {
					item := itemFromNode(n)
					items = append(items, &item)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	return
}

func itemFromNode(n *html.Node) (item WebItem) {
	for _, a := range n.Attr {
		if a.Key == "data-timestamp" {
			item.PubDateFormatted = a.Val
		}
	}
	t := n.FirstChild.NextSibling
	for _, a := range t.Attr {
		if a.Key == "href" {
			item.FileURL = a.Val
		}
	}
	item.Title = t.FirstChild.Data
	d := t.NextSibling.NextSibling
	item.Subtitle = d.FirstChild.Data

	return
}

// fetchNewWebItemsInfo は、ストレージに新規に追加された音声ファイルのObjectInfoを返す
func (pref *PodcastPref) fetchNewWebItemsInfo(ct *minio.Client, oldItems []*WebItem) (fInfos FileInfos, err error) {
	ctx := context.Background()

	lastUpd := time.Time{}
	if len(oldItems) > 0 {
		layout := "Mon, 02 Jan 2006 15:04:05 -0700"
		lastUpd, _ = time.Parse(layout, oldItems[0].PubDateFormatted)
	}

	objectCh := ct.ListObjects(ctx, pref.Bucket, minio.ListObjectsOptions{
		Prefix:    pref.Folder,
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
			newDate := object.LastModified.Truncate(time.Second)
			if lastUpd.Before(newDate) && !lastUpd.Equal(newDate) {
				fInfos = append(fInfos, object)
			}
		}
	}

	sort.Sort(fInfos)

	return
}

// webItemsFromInfo は、音声ファイルのObjectInfoをもとに新規アイテムの構造体を生成する
func (pref *PodcastPref) webItemsFromInfo(fInfo FileInfos, existingItems []*WebItem) (newItems []*WebItem, err error) {
	lastID := len(existingItems)
	if lastID > 0 {
		id, _, _ := getDetailsFromName(existingItems[0].Title)
		if id != 0 {
			lastID = id
		}
	}

	for i, info := range fInfo {
		item := WebItem{}
		fn := strings.TrimLeft(info.Key, pref.Folder+"/")
		id, title, sub := getDetailsFromName(fn)
		idst := ""
		item.Subtitle = sub
		if id != 0 {
			idst = " 第" + strconv.Itoa(id) + "回"
		} else if pref.Serial == true {
			idst = " 第" + strconv.Itoa(lastID+len(fInfo)-i) + "回"
		}
		item.Title = title + idst
		item.FileURL = pref.Link + strings.TrimLeft(info.Key, pref.Folder)

		item.PubDateFormatted = parseDate(info.LastModified)

		newItems = append(newItems, &item)
	}

	return
}

func parseDate(t time.Time) (upd string) {
	if !t.IsZero() {
		return t.Format(time.RFC1123Z)
	}
	return time.Now().UTC().Format(time.RFC1123Z)
}

// uploadWeb は、クラウドストレージにindex.htmlをアップロードする
func (pref *PodcastPref) uploadWeb(ct *minio.Client, web *Web) (err error) {
	ctx := context.Background()
	tmpstr := webtmp() + csstmp() + jstmp()
	wbt := template.Must(template.New("web").Parse(tmpstr))

	buf := new(bytes.Buffer)

	if err = wbt.Execute(buf, *web); err != nil {
		log.Printf("alert: index.htmlのテンプレート展開に失敗しました：%s", err)
		return
	}

	l := int64(buf.Len())

	_, err = ct.PutObject(ctx, pref.Bucket, pref.Folder+"/index.html", buf, l, minio.PutObjectOptions{ContentType: "text/html"})
	if err != nil {
		log.Printf("alert: %s のindex.htmlのアップロードに失敗しました：%s", pref.Folder, err)
	}

	return
}
