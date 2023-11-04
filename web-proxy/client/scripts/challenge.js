"use strict";

function getChallenge() {
  fetch(getAPIBase(), { method: 'GET', redirect: 'follow' })
      .then(response => {
        if (response.redirected) {
          window.location.href = response.url;
        }
        else if (response.ok) {
          return response.json();
        }
      })
      .then(challenge => {
        if (challenge == null) {
          return;
        }
        let createdByMe = (getMyID() == challenge.creatorID);
        document.getElementById("not-found").hidden = true;
        document.getElementById("found").hidden = createdByMe;
        document.getElementById("waiting").hidden = !createdByMe;
        updateChallengeDisplay(
            challenge, document.getElementsByClassName("time-control-span"));
      });
}

function updateChallengeDisplay(challenge, timeControlSpans) {
  for (let i = 0; i < timeControlSpans.length; i++) {
    let span = timeControlSpans[i];
    // clear
    while (span.lastChild) {
      span.removeChild(span.lastChild);
    }
    let timeNs = challenge.config.timeControlNs;

    console.log(timeNs);
    let timeTxt = document.createTextNode(ChallengeBrowser.formatNs(timeNs));
    span.appendChild(timeTxt);
  }
}

document.getElementById('accept-button').addEventListener("click", e => {
  e.preventDefault();
  fetch(getAPIBase() + '/accept', { method: 'POST', redirect: 'follow' })
      .then(response => {
        if (response.redirected) {
          window.location.href = response.url;
        }
      });
});

const eventSource =
    new EventSource(getAPIBase() + "/event-source", { withCredentials: true, });

eventSource.addEventListener("game-created", function(e) {
  window.location.href = e.data;
});

getChallenge();
