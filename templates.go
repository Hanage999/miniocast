package miniocast

func webtmp() string {

	const str string = `
<!DOCTYPE html>
<html lang=ja>

<head>
    <meta charset=utf-8>
    <meta name=viewport content="width=device-width,initial-scale=1">
    <title>{{.Title}}</title>
    <link rel="alternate" type="application/rss+xml" title="{{.Title}}" href="{{.Link}}/feed.rss" />
    <meta name="twitter:card" content="summary" />
    <meta name="twitter:image" content="{{.Link}}/image.jpg" />
    <meta property="og:site_name" content="{{.Title}}" />
    <meta property="og:image" content="{{.Link}}/image.jpg" />
    <meta property="og:url" content="{{.Link}}/index.html" />
    <meta property="og:type" content="blog" />
    <meta property="og:title" content="{{.Title}}" />
    <meta property="og:description" content="{{.Description}}" />
</head>

<body>

<script>
{{template "script" .}}
</script>

<style>
{{template "styles"}}
</style>

<header>
    <div id="header-inner">
        <a id="banner" href="{{.Link}}/index.html">{{.Title}}</a>
        <div class="site-description">{{.Description}}</div>
        <div class="toppic"><img src="{{.Link}}/image.jpg"/></div>
    </div>
</header>

<main>
{{- range $item := .Items -}}
<div class=episode data-timestamp="{{$item.PubDateFormatted}}">
    <a class=title href="{{.FileURL}}">{{$item.Title}}</a>
    <div class=description>{{$item.Description}}</div>
</div>
{{- end -}}
</main>
</body>
</html>	
	`

	return str
}

func csstmp() string {
	const str string = `
{{- define "styles" -}}
body {
  font-family: Verdana, "Droid Sans", Meiryo, Arial;
  color: #222;
  font-size: 16px;
  line-height: 1.6;
  padding: 0;
  margin: 0;
}

header {
  width: 100%;
  background: #f5f5f5;
  border-bottom: 1px solid #e0e0e0;
}

#header-inner, main, footer {
  max-width: 630px;
  padding: 0 15px 30px;
  margin: 0 auto;
}

#header-inner {
  padding: 20px 13px;
  display: grid;
  grid-template-rows: minmax(auto, 100%) minmax(auto, 100%);
  grid-template-columns: 85% minmax(auto, auto);
  grid-template-areas:
	"A C"
	"B C";
}

a {
  color: #1a0dab;
  text-decoration: none;
}

a:visited {
  color: #609;
}

a:hover {
  text-decoration: underline;
}

ul {
  padding-left: 30px;
}

li {
  margin: .3em 0;
}

#banner {
  color: inherit;
  line-height: 1.2;
  font-size: 7.7vw;
  grid-area: A;
}

.site-description {
  font-size: small;
  margin-top: .3em;
  padding: 0 20px 0 0;
  grid-area: B;
}

.toppic {
  grid-area: C;
  display: flex;
  align-items: center;
}

img {
  max-width: 100%;
  max-height: 100%;
}

.episode {
  margin-top: 20px;
  margin-bottom: 30px;
}

.episode-border {
  outline-width: thin;
  outline-color: #c0c0c0;
  outline-offset: 10px;
  outline-style: solid;
}

.title {
  display: inline-block;
  font-size: 18px;
}

.time-saved {
  color: #339900;
}

.description, .donate {
  color: #3C4043;
  font-size: 14px;
  margin-top: .3em;
}

audio {
  width: 90%;
}

.player-container, .seek {
  user-select: none;
  -moz-user-select: none;
  -webkit-user-select: none;
  -ms-user-select: none;
}

.player-container {
  margin: 20px auto 45px auto;
  margin-top: 20px;
  padding: 10px 0 5px 0;
  text-align: center;
  font-size: small;
  border-bottom: 1px solid white;

  position: sticky;
  position: -webkit-sticky;
  top: 0;
  background-color: white;
  transition: border-bottom-color .3s;
}

.player-control {
  text-decoration: underline;
  cursor: pointer;
}

.player-hidden {
  display: none
}

.seek {
  border-bottom: 1px dotted;
  white-space: nowrap;
}

.seek:hover {
  cursor: pointer;
}

.chapters {
  margin: 30px 0px;
  border-spacing: 0;
  border: none;
}

.chap-time {
  padding: .3em .8em .3em 1em;
  vertical-align: top;
  text-align: right;
}

.chap-text {
  padding: .3em .5em .3em 0;
  width: 100%;
}

.chapters tr:nth-child(even) {
  background-color: #f8f8f8;
}

.focus {
  background-color: cornsilk !important;
}

.nobr {
  white-space: nowrap;
  line-height: 2;
}

.date {
  margin-top: .6em;
  padding: 0 .3em 1em 0;
  float: right;
}

footer {
  font-size: small;
  margin-top: 0 auto;
}

@media (max-width: 349px) {
  #banner {
    font-size: 7.6vw;
  }
}

@media (min-width: 500px) {
  #banner {
    display: inline;
    font-size: 32px;
  }
  audio {
    width: 80%;
  }
}
{{- end -}}	
	`

	return str
}

func jstmp() string {
	pre := `
{{- define "script" -}}
let players = {};
{{if .SavePlayState -}}
let store = {};
{{end}}
function saveProgress() {
  for (k in players) {
    if (players[k].currentTime > 0) {
      let prop = {};
      prop.rate = players[k].playbackRate.toFixed(1);
      prop.time = players[k].currentTime;
      store[k] = prop;
    } else {
      delete store[k];
    }

    let title = players[k].parentNode.parentNode.querySelector('.title');
    if (store[k]) {
      title.classList.add('time-saved');
    } else {
      title.classList.remove('time-saved');
    }
  }
  window.localStorage.setItem('players', JSON.stringify(store));
}

function togglePlayer(e) {
  e.preventDefault();
  let episode = e.target.parentElement
  let contn = episode.querySelector('.player-container');
  let player;

  episode.classList.toggle('episode-border');

  if (contn !== null) {
    contn.classList.toggle('player-hidden');
    player = contn.querySelector('.player')
  } else {
    contn = newPlayer(e.target);
    episode.appendChild(contn);
    player = contn.querySelector('.player');
    {{if .SavePlayState -}} player.addEventListener('canplay', loadCurrentTime, {once: true}); {{- end}}
    // Keep the playback speed indicator in sync with the actual speed.
    player.addEventListener('ratechange', () => {
      contn.getElementsByClassName('current-rate')[0].textContent = player.playbackRate.toFixed(1);
      });
    player.addEventListener('play', pauseOtherPlayers);
  }

  if (!episode.classList.contains('episode-border')) {
    player.pause();
  }
}

function loadCurrentTime(e) {
  let player = e.target;
  let epID = e.target.parentElement.parentElement.getAttribute("data-epid");
  let time = getStartTime(epID);
  if (time > 0) {
    players[epID].currentTime = time;
  }  
}

function pauseOtherPlayers(e) {
  for (let k in players) {
    if (!players[k].paused && players[k] !== e.target) {
      players[k].pause();
    }
  }
}

function newPlayer(title) {
  let episode = title.parentElement;
  let epID = episode.getAttribute("data-epid");

  let contn = document.createElement("div");
  contn.classList.add("player-container");
  const tags = 
	`
	post := `;
  contn.insertAdjacentHTML("afterbegin", tags);

  // Set audio file URL.
  let timehash = '';
  {{- if .SavePlayState -}}
  let time = getStartTime(epID);
  if (time > 0) {
    timehash = '#t=' + time;
  }
  {{- end}}
  players[epID] = contn.querySelector('.player');
  let url = episode.querySelector('a').getAttribute('href');
  let ext = url.split('.').pop().toLowerCase();
  let elem = document.createElement('source');
  elem.src = url + timehash;
  if (ext == 'mp3') {
    elem.type = 'audio/mpeg';
  } else {
    elem.type = 'audio/mp4';
  }
  players[epID].appendChild(elem);
  {{if .SavePlayState}}
  let rate = getPlaybackRate(epID);
  if (rate !== '') {
    players[epID].playbackRate = rate;
  }
  {{- end}}
  return contn;
}

// https://stackoverflow.com/a/8076436
String.prototype.hashCode = function () {
  var hash = 0, i, chr;
  if (this.length === 0) return hash;
  for (i = 0; i < this.length; i++) {
    chr = this.charCodeAt(i);
    hash = ((hash << 5) - hash) + chr;
    hash |= 0; // Convert to 32bit integer
  }
  return hash;
};

function getStartTime(epID) {
  if (store[epID]) {
    let t = store[epID].time;
    return t ? parseInt(t) : 0;
  }
  return 0;
}

function getPlaybackRate(epID) {
  if (store[epID]) {
    let rate = store[epID].rate;
    return rate ? rate : '';
  }
  return '';
}

document.addEventListener('DOMContentLoaded', () => {
  let episodes = document.querySelectorAll('.episode');
  {{- if .SavePlayState -}}
  let st;
  if (st = JSON.parse(window.localStorage.getItem('players'))) {
    store = st;
  }
  {{- end -}}
  // Set episode state
  for (let episode of episodes) {
    let url = episode.querySelector('a').getAttribute('href');
    let epid = url.hashCode() + '';
    episode.setAttribute('data-epid', epid);

    let title = episode.querySelector('.title');
    if (store[epid]) {
      title.classList.add('time-saved');
    }
    title.addEventListener('click', togglePlayer);
  }
  {{if .SavePlayState -}}
  window.onunload = saveProgress;
  window.setInterval(saveProgress, 10000);
  {{- end}}
});
{{- end -}}
	`
	div := `
 <audio class="player" preload=metadata controls></audio>
  <br>
  <div>
    <span class=nobr>
      速度 x<span class="current-rate">1.0</span>
      (<span class="player-control" onclick="players['${epID}'].playbackRate-=0.1; return false">遅く</span> /
      <span class="player-control" onclick="players['${epID}'].playbackRate+=0.1; return false">速く</span>)
    </span>
    <span class=nobr style="margin: 0 1em">
      <span class="player-control" onclick="players['${epID}'].currentTime-=15; return false">-15秒</span> /
     <span class="player-control" onclick="players['${epID}'].currentTime-=5; return false">-5秒</span> /
      <span class="player-control" onclick="players['${epID}'].currentTime+=5; return false">+5秒</span> /
      <span class="player-control" onclick="players['${epID}'].currentTime+=15; return false">+15秒</span>
    </span>
  </div>
	`
	jsstr := pre + "`" + div + "`" + post

	return jsstr
}
