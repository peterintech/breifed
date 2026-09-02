(() => {
	const modalID = "onboarding-dialog";
	const dismissedKey = "briefed:onboarding-dismissed";
	let scrollListener;
	let onboardingTimer;

	function modal() {
		return document.getElementById(modalID);
	}

	function openOnboarding() {
		const dialog = modal();
		if (!dialog || dialog.open) return;
		dialog.showModal();
		document.body.classList.add("overlay-open");
	}

	function dismissOnboarding() {
		sessionStorage.setItem(dismissedKey, "true");
		const dialog = modal();
		if (dialog?.open) dialog.close();
		document.body.classList.remove("overlay-open");
	}

	function armOnboarding() {
		const body = document.body;
		if (body.dataset.authenticated === "true" || sessionStorage.getItem(dismissedKey)) return;

		const show = () => {
			window.clearTimeout(onboardingTimer);
			window.removeEventListener("scroll", scrollListener);
			openOnboarding();
		};

		onboardingTimer = window.setTimeout(show, 15000);
		scrollListener = () => {
			const scrollable = document.documentElement.scrollHeight - window.innerHeight;
			if (scrollable > 0 && window.scrollY / scrollable >= 0.3) show();
		};
		window.addEventListener("scroll", scrollListener, { passive: true });
	}

	function closeDrawer() {
		const overlay = document.getElementById("preferences-overlay");
		if (!overlay) return;
		const panel = overlay.querySelector(".drawer-panel");
		const backdrop = overlay.querySelector(".drawer-backdrop");
		panel?.setAttribute("data-open", "false");
		backdrop?.setAttribute("data-open", "false");
		document.body.classList.remove("overlay-open");
		window.setTimeout(() => overlay.remove(), 250);
	}

	function activateDrawer(root = document) {
		const overlay = root.querySelector?.("#preferences-overlay");
		if (!overlay) return;
		document.body.classList.add("overlay-open");
		window.requestAnimationFrame(() => {
			overlay.querySelector(".drawer-panel")?.setAttribute("data-open", "true");
			overlay.querySelector(".drawer-backdrop")?.setAttribute("data-open", "true");
			overlay.querySelector("[data-drawer-focus]")?.focus();
		});
	}

	document.addEventListener("click", (event) => {
		const target = event.target.closest("[data-action]");
		if (!target) return;
		switch (target.dataset.action) {
		case "open-onboarding":
			openOnboarding();
			break;
		case "dismiss-onboarding":
			dismissOnboarding();
			break;
		case "close-drawer":
			closeDrawer();
			break;
		}
	});

	document.addEventListener("cancel", (event) => {
		if (event.target.id === modalID) {
			event.preventDefault();
			dismissOnboarding();
		}
	});

	document.addEventListener("keydown", (event) => {
		if (event.key === "Escape" && document.getElementById("preferences-overlay")) {
			event.preventDefault();
			closeDrawer();
		}
	});

	document.addEventListener("htmx:afterSwap", (event) => {
		activateDrawer(event.target);
	});

	document.body.addEventListener("preferencesSaved", () => {
		closeDrawer();
		const form = document.getElementById("timeline-filters");
		if (form) htmx.trigger(form, "submit");
	});

	window.addEventListener("error", (event) => {
		const image = event.target;
		if (image instanceof HTMLImageElement && image.dataset.articleImage !== undefined) {
			image.hidden = true;
			image.closest("[data-image-wrap]")?.setAttribute("hidden", "");
		}
	}, true);

	window.Briefed = { openOnboarding, dismissOnboarding, closeDrawer };
	armOnboarding();
})();
