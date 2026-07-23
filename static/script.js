// Live search: the client-server "event" feature for this project.
// Every keystroke (debounced) triggers a request to the Go backend's
// /api/search endpoint; the server searches artists, members, locations,
// first albums, and creation dates, and responds with JSON that we render
// here without ever reloading the page.
(function () {
  var input = document.getElementById("search-input");
  var resultsBox = document.getElementById("search-results");
  if (!input || !resultsBox) return;

  var debounceTimer = null;
  var currentController = null;

  function closeResults() {
    resultsBox.classList.remove("open");
    resultsBox.innerHTML = "";
  }

  function renderResults(items) {
    resultsBox.innerHTML = "";
    if (!items || items.length === 0) {
      var empty = document.createElement("div");
      empty.className = "search-empty";
      empty.textContent = "No matches found.";
      resultsBox.appendChild(empty);
      resultsBox.classList.add("open");
      return;
    }

    items.forEach(function (item) {
      var row = document.createElement("div");
      row.className = "search-result";
      row.setAttribute("role", "button");
      row.tabIndex = 0;

      var label = document.createElement("span");
      label.className = "label";
      label.textContent = item.label;

      var type = document.createElement("span");
      type.className = "type";
      type.textContent = item.type;

      row.appendChild(label);
      row.appendChild(type);

      function go() {
        window.location.href = "/artist?id=" + encodeURIComponent(item.artistId);
      }
      row.addEventListener("click", go);
      row.addEventListener("keypress", function (e) {
        if (e.key === "Enter") go();
      });

      resultsBox.appendChild(row);
    });

    resultsBox.classList.add("open");
  }

  function runSearch(query) {
    if (currentController) currentController.abort();
    currentController = new AbortController();

    fetch("/api/search?q=" + encodeURIComponent(query), { signal: currentController.signal })
      .then(function (res) {
        if (!res.ok) throw new Error("search request failed: " + res.status);
        return res.json();
      })
      .then(renderResults)
      .catch(function (err) {
        if (err.name === "AbortError") return; // superseded by a newer keystroke
        resultsBox.innerHTML = '<div class="search-status">Search is unavailable right now.</div>';
        resultsBox.classList.add("open");
        console.error(err);
      });
  }

  input.addEventListener("input", function () {
    var query = input.value.trim();
    window.clearTimeout(debounceTimer);

    if (query.length === 0) {
      closeResults();
      return;
    }

    debounceTimer = window.setTimeout(function () {
      runSearch(query);
    }, 200);
  });

  document.addEventListener("click", function (e) {
    if (!resultsBox.contains(e.target) && e.target !== input) {
      closeResults();
    }
  });

  input.addEventListener("keydown", function (e) {
    if (e.key === "Escape") closeResults();
  });
})();