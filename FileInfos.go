package miniocast

import "github.com/minio/minio-go"

// FileInfos は、クラウドストレージオブジェクトの情報のスライス
type FileInfos []minio.ObjectInfo

// Len は、FileInfosの長さを返す
func (fInfo FileInfos) Len() int {
	return len(fInfo)
}

// Swap は、FileInfosの要素を入れ替える
func (fInfo FileInfos) Swap(i, j int) {
	fInfo[i], fInfo[j] = fInfo[j], fInfo[i]
}

// Less は、FileInfosの小さい方を判定する
func (fInfo FileInfos) Less(i, j int) bool {
	less := fInfo[i].LastModified.After(fInfo[j].LastModified)
	idi, _, _ := getDetailsFromName(fInfo[i].Key)
	idj, _, _ := getDetailsFromName(fInfo[j].Key)
	if idi != 0 && idj != 0 {
		less = idi > idj
	}
	return less
}
