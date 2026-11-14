(function () {
	"use strict";

	function readBootstrap() {
		try {
			return JSON.parse(document.getElementById("docs-bootstrap").textContent);
		} catch (error) {
			return null;
		}
	}

	var data = readBootstrap() || {};
	var ui = data.ui || {};

	var THEME_KEY = "docs-go-theme";
	var root = document.documentElement;
	var appearance = root.getAttribute("data-appearance") || "auto";

	function systemTheme() {
		if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
			return "dark";
		}

		return "light";
	}

	function storedTheme() {
		try {
			var value = localStorage.getItem(THEME_KEY);
			if (value === "light" || value === "dark") {
				return value;
			}
		} catch (error) {
			/* ignore */
		}

		return "";
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

	function currentTheme() {
		return preferredTheme() || systemTheme();
	}

	function updateThemeColor(theme) {
		var meta = document.querySelector('meta[name="theme-color"]');
		if (meta) {
			meta.setAttribute("content", theme === "dark" ? "#101010" : "#ffffff");
		}
	}

	function applyTheme() {
		var theme = preferredTheme();

		if (theme) {
			root.dataset.theme = theme;
		} else {
			delete root.dataset.theme;
		}

		updateThemeColor(currentTheme());
	}

	var themeButton = document.getElementById("theme-button");
	if (themeButton) {
		themeButton.addEventListener("click", function () {
			var next = currentTheme() === "dark" ? "light" : "dark";

			try {
				if (appearance === "auto" && next === systemTheme()) {
					localStorage.removeItem(THEME_KEY);
				} else {
					localStorage.setItem(THEME_KEY, next);
				}
			} catch (error) {
				/* ignore */
			}

			applyTheme();
		});
	}

	if (window.matchMedia) {
		window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function () {
			if (!preferredTheme()) {
				applyTheme();
			}
		});
	}

	window.addEventListener("storage", function (event) {
		if (event.key !== null && event.key !== THEME_KEY) {
			return;
		}

		applyTheme();
	});

	applyTheme();

	/* mobile sidebar */

	var menuButton = document.getElementById("menu-btn");
	var sidebar = document.getElementById("sidebar");
	var backdrop = document.getElementById("sidebar-backdrop");

	function setSidebar(open) {
		if (!sidebar) {
			return;
		}

		document.body.classList.toggle("sidebar-open", open);

		if (menuButton) {
			menuButton.setAttribute("aria-expanded", open ? "true" : "false");
		}

		if (backdrop) {
			backdrop.hidden = !open;
		}
	}

	if (menuButton) {
		menuButton.addEventListener("click", function () {
			setSidebar(!document.body.classList.contains("sidebar-open"));
		});
	}

	if (backdrop) {
		backdrop.addEventListener("click", function () {
			setSidebar(false);
		});
	}

	if (sidebar) {
		sidebar.addEventListener("click", function (event) {
			var link = event.target.closest ? event.target.closest("a") : null;
			if (link && window.innerWidth <= 960) {
				setSidebar(false);
			}
		});
	}

	window.addEventListener("resize", function () {
		if (window.innerWidth > 960 && document.body.classList.contains("sidebar-open")) {
			setSidebar(false);
		}
	});

	/* collapsible sidebar groups */

	Array.prototype.forEach.call(document.querySelectorAll(".sidebar-group-row.is-toggle"), function (row) {
		function toggle() {
			var group = row.closest(".sidebar-group");
			if (!group) {
				return;
			}

			var open = !group.classList.contains("is-open");
			group.classList.toggle("is-open", open);
			group.setAttribute("data-open", open ? "true" : "false");
			row.setAttribute("aria-expanded", open ? "true" : "false");
		}

		row.addEventListener("click", toggle);
		row.addEventListener("keydown", function (event) {
			if (event.key === "Enter" || event.key === " ") {
				event.preventDefault();
				toggle();
			}
		});
	});

	/* responsive page menu */

	var pageMenuButton = document.getElementById("page-menu-btn");
	var pageMenu = document.getElementById("page-menu");
	var pageMenuBackdrop = document.getElementById("page-menu-backdrop");

	function setPageMenu(open) {
		if (!pageMenu || !pageMenuButton) {
			return;
		}

		pageMenu.hidden = !open;
		if (pageMenuBackdrop) {
			pageMenuBackdrop.hidden = !open;
		}
		pageMenuButton.setAttribute("aria-expanded", open ? "true" : "false");
	}

	if (pageMenuButton && pageMenu) {
		pageMenuButton.addEventListener("click", function () {
			setPageMenu(pageMenu.hidden);
		});

		if (pageMenuBackdrop) {
			pageMenuBackdrop.addEventListener("click", function () {
				setPageMenu(false);
			});
		}

		pageMenu.addEventListener("click", function (event) {
			if (event.target.closest && event.target.closest("a")) {
				setPageMenu(false);
			}
		});

		document.addEventListener("click", function (event) {
			if (pageMenu.hidden || pageMenu.contains(event.target) || pageMenuButton.contains(event.target)) {
				return;
			}
			setPageMenu(false);
		});

		document.addEventListener("keydown", function (event) {
			if (event.key === "Escape" && !pageMenu.hidden) {
				setPageMenu(false);
			}
		});

		window.addEventListener("resize", function () {
			if (window.innerWidth > 1240 && !pageMenu.hidden) {
				setPageMenu(false);
			}
		});
	}

	/* copy code */

	var COPY_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V6a2 2 0 0 1 2-2h9"/></svg>';
	var CHECK_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m5 13 4 4L19 7"/></svg>';

	function codeText(pre) {
		return pre.innerText;
	}

	function copyText(text) {
		if (navigator.clipboard && navigator.clipboard.writeText) {
			return navigator.clipboard.writeText(text);
		}

		return new Promise(function (resolve, reject) {
			var textarea = document.createElement("textarea");
			textarea.value = text;
			textarea.setAttribute("readonly", "");
			textarea.style.position = "fixed";
			textarea.style.opacity = "0";
			document.body.appendChild(textarea);
			textarea.select();

			try {
				var ok = document.execCommand("copy");
				document.body.removeChild(textarea);

				if (ok) {
					resolve();
					return;
				}

				reject(new Error("copy failed"));
			} catch (error) {
				document.body.removeChild(textarea);
				reject(error);
			}
		});
	}

	Array.prototype.forEach.call(document.querySelectorAll(".code-block"), function (block) {
		var pre = block.querySelector("pre");
		if (!pre) {
			return;
		}

		var button = document.createElement("button");
		button.type = "button";
		button.className = "copy-btn";
		button.setAttribute("aria-label", ui.copy || "Copy");
		button.title = ui.copy || "Copy";
		button.innerHTML = COPY_ICON;

		button.addEventListener("click", function () {
			copyText(codeText(pre)).then(function () {
				button.classList.add("copied");
				button.innerHTML = CHECK_ICON;
				button.title = ui.copied || "Copied";

				window.setTimeout(function () {
					button.classList.remove("copied");
					button.innerHTML = COPY_ICON;
					button.title = ui.copy || "Copy";
				}, 1600);
			}).catch(function () {
				/* ignore */
			});
		});

		block.appendChild(button);
	});

	/* external links in the document */

	Array.prototype.forEach.call(document.querySelectorAll(".doc a[href]"), function (link) {
		var href = link.getAttribute("href");
		if (!href || href.charAt(0) === "#" || href.indexOf("mailto:") === 0 || href.indexOf("tel:") === 0) {
			return;
		}

		var url;
		try {
			url = new URL(link.href, window.location.href);
		} catch (error) {
			return;
		}

		if ((url.protocol === "http:" || url.protocol === "https:") && url.origin !== window.location.origin) {
			link.target = "_blank";
			link.rel = "noopener noreferrer";
		}
	});

	/* outline scrollspy */

	var outlineLinks = Array.prototype.slice.call(document.querySelectorAll(".outline-list a"));

	if (outlineLinks.length > 0) {
		var entries = [];

		Array.prototype.forEach.call(outlineLinks, function (link) {
			var id = link.getAttribute("href") || "";
			if (id.charAt(0) !== "#") {
				return;
			}

			var decoded = id.slice(1);
			try {
				decoded = decodeURIComponent(decoded);
			} catch (error) {
				/* ignore */
			}

			var heading = document.getElementById(decoded);
			if (heading) {
				entries.push({ link: link, heading: heading });
			}
		});

		var ticking = false;

		function updateOutline() {
			ticking = false;
			var active = null;

			for (var index = 0; index < entries.length; index += 1) {
				if (entries[index].heading.getBoundingClientRect().top <= 132) {
					active = entries[index].link;
				} else {
					break;
				}
			}

			if (!active && entries.length > 0) {
				active = entries[0].link;
			}

			Array.prototype.forEach.call(outlineLinks, function (link) {
				link.classList.toggle("active", link === active);
			});
		}

		function requestOutlineUpdate() {
			if (!ticking) {
				ticking = true;
				window.requestAnimationFrame(updateOutline);
			}
		}

		window.addEventListener("scroll", requestOutlineUpdate, { passive: true });
		window.addEventListener("resize", requestOutlineUpdate);
		updateOutline();
	}

	document.addEventListener("keydown", function (event) {
		if (event.key === "Escape" && document.body.classList.contains("sidebar-open")) {
			setSidebar(false);
		}
	});

})();
