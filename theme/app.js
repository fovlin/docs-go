(function () {
	"use strict";

	function readBootstrap() {
		try {
			return JSON.parse(document.getElementById("docs-bootstrap").textContent);
		} catch (error) {
			return null;
		}
	}

	var ui = readBootstrap() || {};

	var THEME_KEY = "docs-go-theme";
	var root = document.documentElement;
	var appearance = root.getAttribute("data-appearance") || "auto";

	function systemTheme() {
		if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
			return "dark";
		}

		return "light";
	}

	var themeButton = document.getElementById("theme-button");
	if (themeButton) {
		themeButton.addEventListener("click", function () {
			var system = systemTheme();
			var next = (root.dataset.theme || system) === "dark" ? "light" : "dark";
			var followSystem = appearance === "auto" && next === system;

			try {
				if (followSystem) {
					localStorage.removeItem(THEME_KEY);
				} else {
					localStorage.setItem(THEME_KEY, next);
				}
			} catch (error) {
				/* ignore */
			}

			if (followSystem) {
				delete root.dataset.theme;
			} else {
				root.dataset.theme = next;
			}

			var meta = document.querySelector('meta[name="theme-color"]');
			if (meta) {
				meta.setAttribute("content", next === "dark" ? "#101010" : "#ffffff");
			}
		});
	}

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

		if (!open && sidebar.contains(document.activeElement) && menuButton) {
			menuButton.focus();
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

	document.querySelectorAll(".sidebar-group-row.is-toggle").forEach(function (row) {
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

		if (!open && pageMenu.contains(document.activeElement)) {
			pageMenuButton.focus();
		}
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

		window.addEventListener("resize", function () {
			if (window.innerWidth > 1240 && !pageMenu.hidden) {
				setPageMenu(false);
			}
		});
	}

	/* copy code */

	var COPY_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15V6a2 2 0 0 1 2-2h9"/></svg>';
	var CHECK_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m5 13 4 4L19 7"/></svg>';

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

	var copyLabel = ui.copy || "Copy";
	var copiedLabel = ui.copied || "Copied";

	function setCopyState(button, copied) {
		button.classList.toggle("copied", copied);
		button.innerHTML = copied ? CHECK_ICON : COPY_ICON;

		var label = copied ? copiedLabel : copyLabel;
		button.title = label;
		button.setAttribute("aria-label", label);
	}

	document.querySelectorAll(".code-block").forEach(function (block) {
		var pre = block.querySelector("pre");
		if (!pre) {
			return;
		}

		var button = document.createElement("button");
		button.type = "button";
		button.className = "copy-btn";
		setCopyState(button, false);

		button.addEventListener("click", function () {
			copyText(pre.innerText).then(function () {
				setCopyState(button, true);

				window.setTimeout(function () {
					setCopyState(button, false);
				}, 1600);
			}).catch(function () {
				/* ignore */
			});
		});

		block.appendChild(button);
	});

	/* external links in the document */

	document.querySelectorAll(".doc a[href]").forEach(function (link) {
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

	var outlineLinks = Array.from(document.querySelectorAll(".outline-list a"));
	var outlineEntries = [];
	var outlineByID = {};

	outlineLinks.forEach(function (link) {
		var href = link.getAttribute("href") || "";
		if (href.charAt(0) !== "#") {
			return;
		}

		var decoded = href.slice(1);
		try {
			decoded = decodeURIComponent(decoded);
		} catch (error) {
			/* ignore */
		}

		var heading = document.getElementById(decoded);
		if (!heading) {
			return;
		}

		var entry = outlineByID[decoded];
		if (!entry) {
			entry = { heading: heading, links: [] };
			outlineByID[decoded] = entry;
			outlineEntries.push(entry);
		}
		entry.links.push(link);
	});

	if (outlineEntries.length > 0) {
		var topbar = document.querySelector(".topbar");
		var outlineCutoff = topbar ? topbar.getBoundingClientRect().height + 24 : 80;
		var ticking = false;

		function updateOutline() {
			ticking = false;
			var active = outlineEntries[0];

			for (var index = 0; index < outlineEntries.length; index += 1) {
				if (outlineEntries[index].heading.getBoundingClientRect().top <= outlineCutoff) {
					active = outlineEntries[index];
				} else {
					break;
				}
			}

			if (window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 2) {
				active = outlineEntries[outlineEntries.length - 1];
			}

			outlineLinks.forEach(function (link) {
				link.classList.toggle("active", active.links.indexOf(link) !== -1);
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
		if (event.key !== "Escape") {
			return;
		}

		if (pageMenu && !pageMenu.hidden) {
			event.preventDefault();
			setPageMenu(false);
			return;
		}

		if (document.body.classList.contains("sidebar-open")) {
			setSidebar(false);
		}
	});

})();
