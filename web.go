package miniocast

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"strconv"
	"strings"
	"time"

	// for templates
	_ "embed"

	"github.com/minio/minio-go/v7"
	"golang.org/x/net/html"
)

//go:embed templates/web.gtpl
var webtmp string

//go:embed templates/css.gtpl
var csstmp string

//go:embed templates/js.gtpl
var jstmp string

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

// UpdateWeb は、index.htmlを作成あるいは更新する
func (pref *PodcastPref) UpdateWeb(infos FileInfos, ct *minio.Client) {
	web := pref.newWeb()

	newItems, err := pref.webItemsFromInfo(infos)
	if err != nil {
		log.Printf("info: %s の新規Webアイテムの作成に失敗しました：%s", pref.Folder, err)
		return
	}

	web.Items = newItems

	if err := pref.uploadWeb(ct, &web); err != nil {
		log.Printf("info: index.htmlのアップロードに失敗しました：%s", err)
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

// webItemsFromInfo は、音声ファイルのObjectInfoをもとに新規アイテムの構造体を生成する
func (pref *PodcastPref) webItemsFromInfo(fInfos FileInfos) (newItems []*WebItem, err error) {
	lastID := len(fInfos)
	if lastID == 0 {
		return
	}

	for i, info := range fInfos {
		item := WebItem{}
		fn := strings.TrimLeft(info.Key, pref.Folder+"/")
		id, title, sub := getDetailsFromName(fn)
		idst := ""
		item.Subtitle = sub
		if id != 0 {
			idst = " 第" + strconv.Itoa(id) + "回"
		} else if pref.Serial == true {
			idst = " 第" + strconv.Itoa(lastID+len(fInfos)-i) + "回"
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

	tmpstr := webtmp + csstmp + jstmp
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
