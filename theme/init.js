(function () {
	"use strict";

	var KEY = "docs-go-theme";
	var root = document.documentElement;
	var appearance = root.getAttribute("data-appearance") || "auto";

	function storedTheme() {
		try {
			var value = localStorage.getItem(KEY);
			if (value === "light" || value === "dark" || value === "system") {
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
		if (stored === "system") {
			return "";
		}
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
	}

	function applyThemeColor() {
		if (!document.body) {
			return;
		}

		var meta = document.querySelector('meta[name="theme-color"]');
		if (!meta) {
			meta = document.createElement("meta");
			meta.setAttribute("name", "theme-color");
			document.head.appendChild(meta);
		}

		meta.setAttribute("content", getComputedStyle(document.body).backgroundColor);
	}

	applyTheme();

	window.addEventListener("load", applyThemeColor, { once: true });

	if (window.matchMedia) {
		window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function () {
			if (!preferredTheme()) {
				applyTheme();
				applyThemeColor();
			}
		});
	}

	window.addEventListener("storage", function (event) {
		if (event.key !== null && event.key !== KEY) {
			return;
		}

		applyTheme();
		applyThemeColor();
	});
})();
