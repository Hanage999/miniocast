package miniocast

// Index は、差分比較のためのキー
type Index struct {
	FileLink string
	Updated  string
}

// Indexes は、Indexのスライス
type Indexes []Index

// Len は、Indexesの長さを返す
func (idxs Indexes) Len() int {
	return len(idxs)
}

// Swap は、Indexesの要素を入れ替える
func (idxs Indexes) Swap(i, j int) {
	idxs[i], idxs[j] = idxs[j], idxs[i]
}

// Less は、Indexesの小さい方を判定する
func (idxs Indexes) Less(i, j int) bool {
	less := idxs[i].FileLink < idxs[j].FileLink
	if idxs[i].FileLink == idxs[j].FileLink {
		less = idxs[i].Updated < idxs[j].Updated
	}
	return less
}
