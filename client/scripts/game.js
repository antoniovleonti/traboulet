"use strict";

let boardDisplay =
    new BoardDisplay(document.getElementById("marble-container"));
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
  statusDisplay.update(state.status);
  playerDisplayManager.update(
      state.idToPlayer, state.colorToPlayer,
      state.status == "ONGOING" ? state.whoseTurn : null, state.timeControl,
      state.firstMoveDeadline);

  const myID = PlayerDisplayManager.getMyID(Object.keys(state.idToPlayer));
  const isYourTurn = ((myID != null) &&
                      (state.idToPlayer[myID].color == state.whoseTurn));
  boardDisplay.update(state.board, state.validMoves, isYourTurn);
}

getStateAndUpdate();
