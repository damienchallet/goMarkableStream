let height;
let width;
let eventURL;
let portrait;
let draw;
let latestX;
let latestY;
let maxXValue;
let maxYValue;

onmessage = (event) => {
	const data = event.data;

	switch (data.type) {
		case 'init':
			height = event.data.height;
			width = event.data.width;
			eventURL = event.data.eventURL;
			portrait = event.data.portrait;
            maxXValue = event.data.maxXValue;
            maxYValue = event.data.maxYValue;
			initiateEventsListener();
			break;
		case 'portrait':
			portrait = event.data.portrait;
			break;
		case 'terminate':
			console.log("terminating worker");
			close();
			break;
	}
};


async function initiateEventsListener() {
	const eventSource = new EventSource(eventURL);
	draw = true;
	eventSource.onmessage = (event) => {
		const message = JSON.parse(event.data);
		if (message.Type === 3) {
			if (message.Code === 24) { // ABS_PRESSURE
				if (message.Value > 0) {
					draw = false;
					postMessage({ type: 'clear' });
				} else {
					draw = true; // Pen lifted, resume laser
				}
			} else if (message.Code === 25) { // ABS_DISTANCE
				draw = true;
			}
		}
		if (message.Type === 3) {
			// Update laser pointer position
			if (portrait) {
				if (message.Code === 1) { // X-axis input
					latestX = scaleValue(message.Value, maxXValue, width);
				} else if (message.Code === 0) { // Y-axis input
					latestY = height - scaleValue(message.Value, maxYValue, height);
				}
			} else {
				if (message.Code === 1) {
					latestY = scaleValue(message.Value, maxYValue, height);
				} else if (message.Code === 0) {
					latestX = scaleValue(message.Value, maxXValue, width);
				}
			}
			if (draw) {
				postMessage({ type: 'update', X: latestX, Y: latestY });
			}
		}
	}

	eventSource.onerror = () => {
		postMessage({
			type: 'error',
			message: "EventSource error",
		});
		console.error('EventSource error occurred.');
	};

	eventSource.onclose = () => {
		postMessage({
			type: 'error',
			message: 'Connection closed'
		});
		console.log('EventSource connection closed.');
	};
}

// Function to scale the incoming value to the canvas size
function scaleValue(value, maxValue, canvasSize) {
	return (value / maxValue) * canvasSize;
}
