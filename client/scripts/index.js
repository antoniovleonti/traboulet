challengeBrowser =
    new ChallengeBrowser(document.getElementById("challenge-browser"));

function updateChallengeBrowser() {
  fetch('/api/challenges')
      .then(response => {
        if (response.ok) {
          return response.json();
        }
        response.text().then(txt => {
          document.getElementById("content").hidden = true;
          document.getElementById("error").hidden = false;
          console.log(`${response.status} ${txt}`);
        });
      })
      .then(challenges => {
        challengeBrowser.update(challenges);
      });
}

updateChallengeBrowser();

const createRoomForm = document.getElementById("new-challenge-form");

createRoomForm.addEventListener("submit", e => {
  e.preventDefault();
  let formRaw = Object.fromEntries(new FormData(createRoomForm));
  let data = JSON.stringify({
    timeControlNs: formRaw.initialTimeMin * 6e10
  });
  fetch('/api/challenges', { method: 'POST', body: data, redirect: 'follow' })
      .then(response => {
        if (response.redirected) {
          window.location.href = response.url;
        }
        if (!response.ok) {
          response.text().then(txt => {
            console.log(`${response.status} ${txt}`);
          });
        }
      })
      .catch(err => {
        createErr.innerHTML = err;
      });
});

