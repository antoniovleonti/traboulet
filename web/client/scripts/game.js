"use strict";

const boardDisplay =
    new BoardDisplay(document.getElementById("marble-container"));
const statusDisplay = new StatusDisplay(document.getElementById("status"));
const topPlayer = {
  clock: document.getElementById("other-player-clock"),
  firstMoveIndicator:
      document.getElementById("other-player-first-move-indicator"),
  color: document.getElementById("other-player-color"),
  active: document.getElementById("other-player-activity"),
  score: document.getElementById("other-player-score"),
  you: null,
};
const bottomPlayer = {
  clock: document.getElementById("your-clock"),
  firstMoveIndicator:
      document.getElementById("your-first-move-indicator"),
  color: document.getElementById("your-color"),
  active: document.getElementById("your-activity"),
  score: document.getElementById("your-score"),
  you: document.getElementById("you-indicator"),
};
const playerDisplayManager = new PlayerDisplayManager(topPlayer, bottomPlayer);
const moveHistoryDisplay =
    new MoveHistoryDisplay(document.getElementById('move-history'));

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
  document.getElementById("game-dash").hidden = false

  const lastSnapshot = state.history[state.history.length-1];
  statusDisplay.update(state.status);
  playerDisplayManager.update(
      state.idToPlayer, state.colorToPlayer,
      state.status == "ONGOING" ? lastSnapshot.whoseTurn : null,
      state.timeControl, state.firstMoveDeadline);

  const myID = PlayerDisplayManager.getMyID(Object.keys(state.idToPlayer));
  const isYourTurn = ((myID != null) &&
                      (state.idToPlayer[myID].color == lastSnapshot.whoseTurn));
  boardDisplay.update(lastSnapshot.board, state.validMoves, isYourTurn);

  moveHistoryDisplay.update(state.history);
}

document.getElementById('resign-button').addEventListener('click', () => {
  fetch(getURLBase() + '/resignation',
        { method: 'POST', body: null })
      .then(response => {
        if (!response.ok) {
          response.text().then(txt => {
            console.log(`${response.status} ${txt}`);
          });
        }
      });
});

getStateAndUpdate();
