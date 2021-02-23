{{- define "script" -}}
let players = {};

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
  <audio class="player" data-epid="${epID}" preload=metadata controls></audio>
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
        elem.type = 'audio/x-m4a';
    }
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

document.addEventListener('DOMContentLoaded', () => {
    // Initialize the global variables.
    let episodes = document.querySelectorAll('.episode');
    for (let episode of episodes) {
        let url = episode.querySelector('a').getAttribute('href');
        let epid = url.hashCode() + '';
        episode.setAttribute('data-epid', epid);

        episode.querySelector('.title').addEventListener('click', togglePlayer);
    }
});
{{- end -}}