document.body.addEventListener("timerEnded", () => {
	let plays = 0;

	let audio = new Audio("/static/sound/timer-ended-short.mp3");

	audio.playbackRate = 0.75;

	audio.addEventListener("ended", () => {
		plays += 1;
		if (plays < 5) {
			setTimeout(() => {
				audio.play();
			}, 200);
		}
	});

	audio.play();
})
