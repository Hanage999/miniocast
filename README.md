# miniocast

[![en](https://img.shields.io/badge/lang-en-red.svg)](https://github.com/Hanage999/miniocast/blob/master/README.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/Hanage999/miniocast/blob/master/README.ja.md)

A simple tool to turn your MinIO bucket into a Podcast server. Just add audio files, and miniocast will automatically generate and update the RSS feed and web interface.

## Features

- Use folders in your MinIO bucket as Podcast sites.
- Automatically detects audio files (mp3, m4a, m4b) and generates/updates RSS (`feed.rss`) and web page (`index.html`).
- File names are directly used as episode titles.
- A half-width space in file names splits the title and subtitle.
- Automatically detects episode numbers formatted as "第XX回" (only half-width digits).
- Saves playback state (position and speed) in the browser's local storage.

## Setup

1. Create a MinIO bucket and a folder for your podcast, and apply a Read Only policy to the bucket.
2. Place audio files (mp3, m4a) and a podcast image (`image.jpg`) in the folder.
3. Build miniocast by running the following commands inside `cmd/miniocast`:

```bash
cd cmd/miniocast
go get
go build
```

4. Copy `config.yml.example` to `config.yml` and edit the settings according to your environment.

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
  - Title: Your Podcast Title
    Subtitle: Optional subtitle
    Author: Author Name
    Email: author@example.com
    Description: Description of your podcast
    Bucket: bucket-name
    Folder: folder-name
    Serial: 0
    Active: true
```

5. Run `./miniocast` to generate the RSS feed and web interface.

## How to Use Your Podcast

Register the RSS feed URL directly in your podcast app:

- RSS Feed URL: `{MinIO_bucket_URL}/{folder_name}/feed.rss`
- Web Interface URL: `{MinIO_bucket_URL}/{folder_name}/index.html`

## Acknowledgments

The web interface design utilizes JavaScript and CSS generously provided by [Rui Ueyama](https://x.com/rui314) of [Turing Complete FM](https://turingcomplete.fm). We sincerely appreciate his open and generous contribution.