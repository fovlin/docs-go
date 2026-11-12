(function () {
	"use strict";

	var KEY = "docs-go-theme";
	var root = document.documentElement;
	var appearance = root.getAttribute("data-appearance") || "auto";
	var stored = null;

	try {
		stored = localStorage.getItem(KEY);
	} catch (error) {
		stored = null;
	}

	function systemTheme() {
		if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
			return "dark";
		}

		return "light";
	}

	function preferredTheme() {
		if (stored === "light" || stored === "dark") {
			return stored;
		}

		if (appearance === "light" || appearance === "dark") {
			return appearance;
		}

		return "";
	}

	function currentTheme() {
		return preferredTheme() || systemTheme();
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
			meta.setAttribute("content", currentTheme() === "dark" ? "#101010" : "#ffffff");
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

		try {
			stored = localStorage.getItem(KEY);
		} catch (error) {
			stored = null;
		}

		applyTheme();
	});
})();
