(function () {
	"use strict";

	var KEY = "docs-go-theme";
	var root = document.documentElement;
	var appearance = root.getAttribute("data-appearance") || "auto";

	function storedTheme() {
		try {
			var value = localStorage.getItem(KEY);
			if (value === "light" || value === "dark") {
				return value;
			}
		} catch (error) {
			/* ignore */
		}

		return "";
	}

	function systemTheme() {
		if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
			return "dark";
		}

		return "light";
	}

	function preferredTheme() {
		var stored = storedTheme();
		if (stored) {
			return stored;
		}

		if (appearance === "light" || appearance === "dark") {
			return appearance;
		}

		return "";
	}

	function applyTheme() {
		var theme = preferredTheme();

		if (theme) {
			root.dataset.theme = theme;
		} else {
			delete root.dataset.theme;
		}

		var meta = document.querySelector('meta[name="theme-color"]');
		if (meta) {
			meta.setAttribute("content", (theme || systemTheme()) === "dark" ? "#101010" : "#ffffff");
		}
	}

	applyTheme();

	if (window.matchMedia) {
		window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function () {
			if (!preferredTheme()) {
				applyTheme();
			}
		});
	}

	window.addEventListener("storage", function (event) {
		if (event.key !== null && event.key !== KEY) {
			return;
		}

		applyTheme();
	});
})();
