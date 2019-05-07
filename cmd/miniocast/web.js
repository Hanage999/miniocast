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
    // Keep the playback speed indicator in sync with the actual speed.
    player = contn.querySelector('.player');
    player.addEventListener('ratechange', () => {
      contn.getElementsByClassName('current-rate')[0].textContent = player.playbackRate.toFixed(1);
    });
    player.addEventListener('play', pauseOtherPlayers);
  }

  if (!episode.classList.contains('episode-border')) {
    player.pause();
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

  const tags = `
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
  `;

  contn.insertAdjacentHTML("afterbegin", tags);

  // Set audio file URL.
  players[epID] = contn.querySelector('.player');
  let elem = document.createElement('source');
  elem.src = episode.querySelector('a').getAttribute('href');
  let ext = elem.src.split('.').pop().toLowerCase();
  if (ext == 'mp3') {
    elem.type = 'audio/mpeg';
  } else {
    elem.type = 'audio/mp4';
  }
  {{if .SavePlayState}}
  // Restore the previous player state.
  let rate;
  let time = getStartTime(epID);
  if (store[epID]) {
    rate = store[epID].rate;
    if (rate !== '') {
      players[epID].playbackRate = rate;
    }
    if (time > 0) {
      players[epID].currentTime = time;
    }
  }
  {{end}}
  players[epID].appendChild(elem);

  return contn;
}

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
  if (store && store[epID]) {
    let t = store[epID].time;
    return t ? parseInt(t) : 0;
  }
  return 0;
}

document.addEventListener('DOMContentLoaded', () => {
  let episodes = document.querySelectorAll('.episode');

  {{if .SavePlayState -}}
  if (JSON.parse(window.localStorage.getItem('players'))) {
    store = JSON.parse(window.localStorage.getItem('players'));
  }
  {{end}}

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
