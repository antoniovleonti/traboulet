"use strict";

let boardDisplay = new BoardDisplay(document.getElementById("board"));
let statusDisplay = new StatusDisplay(document.getElementById("status"));
let clocksDisplay = new ClocksDisplay(
    document.getElementById("your-clock"),
    document.getElementById("other-player-clock"));
let scoreDisplay = new ScoreDisplay(
    document.getElementById("your-score"),
    document.getElementById("other-player-score"));
let colorDisplay = new ColorDisplay(
    document.getElementById("your-color"),
    document.getElementById("other-player-color"));
let activityDisplay = new ActivityDisplay(
    document.getElementById("your-activity"),
    document.getElementById("other-player-activity"));

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
        if (state != null) {
          update(state);
        }
      });
}

function update(state) {
  boardDisplay.update(state.board);
  statusDisplay.update(state.status);
  clocksDisplay.update(state.idToPlayer);
  scoreDisplay.update(state.idToPlayer);
  colorDisplay.update(state.idToPlayer);
  activityDisplay.update(state.whoseTurn, state.colorToPlayer);
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
