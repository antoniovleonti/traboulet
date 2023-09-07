"use strict";

let boardDisplay = new BoardDisplay(document.getElementById("board"));
let statusDisplay = new StatusDisplay(document.getElementById("status"));
let topPlayer = {
  clock: document.getElementById("other-player-clock"),
  color: document.getElementById("other-player-color"),
  active: document.getElementById("other-player-activity"),
  score: document.getElementById("other-player-score"),
  you: null,
};
let bottomPlayer = {
  clock: document.getElementById("your-clock"),
  color: document.getElementById("your-color"),
  active: document.getElementById("your-activity"),
  score: document.getElementById("your-score"),
  you: document.getElementById("you-indicator"),
};
let playerDisplayManager = new PlayerDisplayManager(topPlayer, bottomPlayer);

function getURLBase() {
  let urlParts = window.location.href.split("/");
  let gameID = urlParts[urlParts.length-1];
  return '/api/games/' + gameID;
}

const stream = new EventSource(getURLBase() + "/event-stream", {
  withCredentials: true,
});
stream.addEventListener('state-push', function(e) {
  update(JSON.parse(e.data));
});
stream.onerror = function(e) {
  console.log(e);
};

function getStateAndUpdate() {
  fetch(getURLBase() + '/state')
      .then(response => {
        if (response.ok) {
          return response.json();
        }
      })
      .then(state => {
        update(state);
      });
}

function update(state) {
  if (state == null) {
    return
  }
  document.getElementById("error").hidden = true
  document.getElementById("content").hidden = false
  boardDisplay.update(state.board);
  statusDisplay.update(state.status);
  playerDisplayManager.update(state.idToPlayer, state.colorToPlayer);
}

let moveForm = document.getElementById("move-form");

moveForm.addEventListener("submit", e => {
  e.preventDefault();
  let formRaw = Object.fromEntries(new FormData(moveForm));
  let data = JSON.stringify({
    x: parseInt(formRaw.x),
    y: parseInt(formRaw.y),
    d: formRaw.d,
  });
  fetch(getURLBase() + '/move',
        { method: 'POST', body: data })
      .then(response => {
        if (!response.ok) {
          response.text().then(txt => {
            console.log(`${response.status} ${txt}`);
          });
        }
      });
});

getStateAndUpdate();
