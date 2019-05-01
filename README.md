# miniocast

Amazon S3互換のクラウドストレージ、MinIOのバケットをPodcastサーバーにしてしまおうという寸法です。

## 動作環境
+ macOS 10.14.4
+ Go 1.12.4
+ MinIO 2019-04-23T23:50:36Z

## 機能
+ バケット内のフォルダを１つのPodcastサイトとし、そこに存在する音声ファイルをもとに、同一フォルダ内に自動的にfeed.rssを生成します。複数のPodcastサイトを設定することも可能です。
+ エピソードタイトルは、「ファイル名 第〜回」となります。エピソードの数に応じて、自動的に連番をつけます。
+ ファイル名に、半角スペースの後で「第<半角数字>回」（第023回、第5回など）という文字列がすでに存在する場合は、それを尊重します。数字は全角ダメ、絶対。
+ さらに、「第~回」の後に内容紹介の文章が含まれている場合、それをDescriptionとして使います。
+ ブラウザからアクセスできるインターフェイスは生成しません。各種クライアントアプリケーション（iOSのPodcastアプリ, Overcastなど）から直接feed.rssを読み込んで活用してください。
+ 生成されたfeed.rssは、なぜかmacOSのiTunesからは読み込めません。

## 使い方
0. 下準備1：MinIOのバケットとPodcastごとのフォルダを作成し、バケットにはRead Onlyポリシーを適用しておく。
1. 下準備2：Podcastにしたいフォルダ直下に、image.jpgという名前でタイトル画像を設置しておく。
2. cmd/miniocast フォルダで go get、go build すると、フォルダに miniocast コマンドができる。MinIOサーバーと同じマシンに置く必要はない。（ただ後述の通り、同一マシンに置いておくと、より便利に使える場合もあるかも）
3. config.yml.example を config.yml にリネームまたはコピーし、バケットとフォルダなどの設定に応じて書き換えるあるいは追記する。
4. ./miniocast で起動。

## より便利に

たとえばmacOS上でMinIOを運用しているなら、MinIOがサーバーとして使っているローカルフォルダにフォルダアクションを適用できます。

そこで、同一マシンにminiocastと設定ファイルを置き、そのフォルダにファイルが追加されたらminiocastを自動的に起動するようなAutomatorワークフローを組んでおくと便利です（フォルダやコマンドのアクセス権にご注意）。

## Amazon S3 では

……試しておりません。