"use strict";

const boardDisplay =
    new BoardDisplay(document.getElementById("board-marble-layer"),
                     document.getElementById("board-input-layer"));
const statusDisplay = new StatusDisplay(document.getElementById("status"));
const topPlayer = {
  clock: document.getElementById("villain-clock"),
  firstMoveIndicator:
      document.getElementById("villain-first-move-indicator"),
  color: document.getElementById("villain-color"),
  active: document.getElementById("villain-activity"),
  score: document.getElementById("villain-score"),
  you: null,
};
const bottomPlayer = {
  clock: document.getElementById("hero-clock"),
  firstMoveIndicator:
      document.getElementById("hero-first-move-indicator"),
  color: document.getElementById("hero-color"),
  active: document.getElementById("hero-activity"),
  score: document.getElementById("hero-score"),
  you: document.getElementById("you-indicator"),
};
const playerDisplayManager = new PlayerDisplayManager(
    topPlayer, bottomPlayer,
    document.getElementById('resign-button'),
    document.getElementById('rematch-button'),
    document.getElementById('rematch-offer-count'));

const historyManager =
    new HistoryManager(boardDisplay, document.getElementById('move-history'));

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
  document.getElementById("error").hidden = true;
  document.getElementById("game-dash").hidden = false;

  const gameOngoing = state.status == 'ONGOING';

  document.getElementById("resign-button").hidden = !gameOngoing;
  document.getElementById("rematch-button").hidden = gameOngoing;

  const lastSnapshot = state.history[state.history.length-1];
  statusDisplay.update(state.status);
  playerDisplayManager.update(
      state.idToPlayer, state.colorToPlayer,
      state.status == "ONGOING" ? lastSnapshot.whoseTurn : null,
      state.timeControl, state.firstMoveDeadline, state.status);

  historyManager.update(state.history, state.validMoves, state.idToPlayer);
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

document.getElementById('rematch-button').addEventListener('click', () => {
  fetch(getURLBase() + '/rematch-offer',
        { method: 'POST', body: null })
      .then(response => {
        if (!response.ok) {
          response.text().then(txt => {
            console.log(`${response.status} ${txt}`);
          });
        }
      });
});

document.getElementById('move-history-prev').addEventListener('click', () => {
  historyManager.prev();
});

document.getElementById('move-history-next').addEventListener('click', () => {
  historyManager.next();
});

document.getElementById('move-history-first').addEventListener('click', () => {
  historyManager.first();
});

document.getElementById('move-history-last').addEventListener('click', () => {
  historyManager.last();
});

getStateAndUpdate();
