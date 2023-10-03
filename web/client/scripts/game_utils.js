function getMyID() {
  let cookies = document.cookie.split(";");
  if (cookies.length == 0) {
    return null
  }
  let nameval = cookies[0].split("=");
  return nameval[0].trim()
}
