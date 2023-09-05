"use strict";

getChallenge();

function hasCookieWithName(name) {
  let cookies = document.cookie.split(";");
  for (let i = 0; i < cookies.length; i++) {
    let nameval = cookies[i].split("=");
    if (nameval[0].trim() == name) {
      return true
    }
  }
  return false
}

function getChallenge() {
  let urlParts = window.location.href.split("/");
  let challengeID = urlParts[urlParts.length-1];
  fetch('/api/challenges/' + challengeID)
      .then(response => {
        if (response.ok) {
          return response.json();
        }
      })
      .then(challenge => {
        if (challenge == null) {
          return;
        }
        let createdByMe = hasCookieWithName(challenge.creatorID);
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
  let urlParts = window.location.href.split("/");
  let challengeID = urlParts[urlParts.length-1];
  fetch('/api/challenges/' + challengeID + '/accept',
        { method: 'POST', redirect: 'follow' })
      .then(response => {
        if (response.redirected) {
          window.location.href = response.url;
        }
        if (!response.ok) {
          response.text().then(txt => {
          });
        }
      })
      .catch(err => {
      });
});

function getUpdate() {
  let urlParts = window.location.href.split("/");
  let challengeID = urlParts[urlParts.length-1];
  fetch('/api/challenges/' + challengeID + '/update',
        { method: 'GET', redirect: 'follow' })
      .then(response => {
        if (response.redirected) {
          window.location.href = response.url;
        }
        response.text().then(txt => {
          console.log(`${response.status} ${txt}`);
        });
      });
}

getUpdate();
