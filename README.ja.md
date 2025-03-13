# miniocast

[![en](https://img.shields.io/badge/lang-en-red.svg)](https://github.com/Hanage999/miniocast/blob/master/README.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/Hanage999/miniocast/blob/master/README.ja.md)

MinIOバケットをPodcastサーバーとして活用できるシンプルなツールです。音声ファイルを置くだけでPodcast用のRSSフィードやWebインターフェイスを自動生成・更新します。

## 主な特徴

- MinIOバケット内のフォルダをPodcastサイトとして利用可能
- フォルダ内の音声ファイル（mp3, m4a, m4b）を検出して自動的にRSS (`feed.rss`) とWebページ (`index.html`) を生成・更新
- ファイル名がそのままエピソードタイトルに使用されます
- ファイル名に半角スペースがある場合、それを境にタイトルとサブタイトルに分割します
- 「第XX回」（半角数字のみ対応）の形式でエピソード番号を自動認識可能
- ブラウザのローカルストレージを使用して再生位置・再生速度を記憶可能

## セットアップ方法

1. MinIOバケットを作成し、Podcast用のフォルダを準備。バケットにはRead Onlyポリシーを適用してください。
2. フォルダ内に音声ファイル（mp3, m4a）を配置し、`image.jpg` をPodcastのイメージとして置きます。
3. `cmd/miniocast` 内で以下のコマンドを実行してビルドします：

```bash
cd cmd/miniocast
go get
go build
```

4. `config.yml.example` をコピーして `config.yml` を作成し、環境に応じて設定を編集します。

```yaml
Storage:
  Server: minio.example.com
  Endpoint: minio.example.com
  AccessKey: YOUR_ACCESS_KEY
  SecretKey: YOUR_SECRET_KEY
  HTTPS: true
  SecureEndpoint: true
  BucketAsVirtualHost: false

SavePlayState: true

Podcasts:
  - Title: Podcastのタイトル
    Subtitle: サブタイトル（任意）
    Author: 作者名
    Email: メールアドレス
    Description: Podcastの説明
    Bucket: バケット名
    Folder: フォルダ名
    Serial: 0
    Active: true
```

5. `./miniocast` を実行し、RSSフィードとWebインターフェイスを生成します。

## Podcastの利用方法

生成されたPodcastはRSSフィードURLを直接Podcastアプリに登録して利用します。

- RSSフィードURL: `{MinIOバケットのURL}/{フォルダ名}/feed.rss`
- WebインターフェイスURL: `{MinIOバケットのURL}/{フォルダ名}/index.html`

## 謝辞

Webインターフェイスのデザインは、[Turing Complete FM](https://turingcomplete.fm)の[Rui Ueyama](https://x.com/rui314)さんが気前よく公開してくださっている JavaScript と CSS を使わせていただきました。