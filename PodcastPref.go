package miniocast

// PodcastPref は、設定ファイルから読み込んだ各Podcastの情報を格納する
type PodcastPref struct {
	Title         string
	Subtitle      string
	Author        string
	Email         string
	Description   string
	Link          string
	Bucket        string
	Folder        string
	Serial        int
	Active        bool
	SavePlayState bool
}
