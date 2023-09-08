"use strict";

let boardDisplay = new BoardDisplay(document.getElementById("board"));
let statusDisplay = new StatusDisplay(document.getElementById("status"));
let topPlayer = {
  clock: document.getElementById("other-player-clock"),
  firstMoveIndicator:
      document.getElementById("other-player-first-move-indicator"),
  color: document.getElementById("other-player-color"),
  active: document.getElementById("other-player-activity"),
  score: document.getElementById("other-player-score"),
  you: null,
};
let bottomPlayer = {
  clock: document.getElementById("your-clock"),
  firstMoveIndicator:
      document.getElementById("your-first-move-indicator"),
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

var validMoves = null;
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
  console.log("state: ", state);
  document.getElementById("error").hidden = true
  document.getElementById("content").hidden = false
  boardDisplay.update(state.board, state.validMoves);
  statusDisplay.update(state.status);
  playerDisplayManager.update(state.idToPlayer, state.colorToPlayer,
                              state.whoseTurn, state.firstMoveDeadline);
  validMoves = state.validMoves;
}

let moveForm = document.getElementById("move-form");
moveForm.addEventListener("submit", e => {
  e.preventDefault();
  let formRaw = Object.fromEntries(new FormData(moveForm));
  let move = {
    x: parseInt(formRaw.x),
    y: parseInt(formRaw.y),
    d: formRaw.d,
  };
  let isValid = false;
  for (const m of validMoves) {
    if (move.x == m.x && move.y == m.y && move.d == m.d) {
      isValid = true;
      break;
    }
  }
  if (validMoves == null || !isValid) {
    console.error("move ", move, "was not in list of valid moves");
    return;
  }
  let data = JSON.stringify(move);
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
