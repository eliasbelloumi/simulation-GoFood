const { DeckGL, ScatterplotLayer, MapboxOverlay, TripsLayer, PathLayer, IconLayer, SimpleMeshLayer } = deck;


// ----------------- MAPBOX -----------------------------------

const MAPTILER_KEY = 'r1kBZMtyx7YNJku6LYQO';
const compiegneBFRondPoint = [2.819475, 49.415672];

const map = new maplibregl.Map({
    style: `https://api.maptiler.com/maps/basic-v2/style.json?key=${MAPTILER_KEY}`,
    center: compiegneBFRondPoint,
    zoom: 15.5,
    pitch: 45,
    bearing: -17.6,
    container: 'map',
    antialias: true
});



// 3D building extrusion
map.on('load', () => {
    // Insert the layer beneath any symbol layer.
    const layers = map.getStyle().layers;

    let labelLayerId;
    for (let i = 0; i < layers.length; i++) {
        if (layers[i].type === 'symbol' && layers[i].layout['text-field']) {
            labelLayerId = layers[i].id;
            break;
        }
    }

    map.addSource('openmaptiles', {
        url: `https://api.maptiler.com/tiles/v3/tiles.json?key=${MAPTILER_KEY}`,
        type: 'vector',
    });

    map.addLayer(
        {
            'id': '3d-buildings',
            'source': 'openmaptiles',
            'source-layer': 'building',
            'type': 'fill-extrusion',
            'minzoom': 15,
            'filter': ['!=', ['get', 'hide_3d'], true],
            'paint': {
                'fill-extrusion-color': [
                    'interpolate',
                    ['linear'],
                    ['get', 'render_height'], 0, 'lightgray', 200, 'royalblue', 400, 'lightblue'
                ],
                'fill-extrusion-height': [
                    'interpolate',
                    ['linear'],
                    ['zoom'],
                    15,
                    0,
                    16,
                    ['get', 'render_height']
                ],
                'fill-extrusion-base': ['case',
                    ['>=', ['get', 'zoom'], 16],
                    ['get', 'render_min_height'], 0
                ]
            }
        },
        labelLayerId
    );
});

function flyToLocation(coordinates) {
    map.flyTo({
        center: coordinates, // [longitude, latitude]
        pitch: 45,
        bearing: -17.6,
        zoom: 15.5,
        speed: 1.2,
        curve: 1,
        essential: true
    });
}


// ----------------- CHART Setup-----------------------------------

const ctx = document.getElementById('myChart').getContext('2d');
const chart = new Chart(ctx, {
    type: 'line',
    data: {
        labels: [10.8], // x-axis labels for time
        datasets: [
            {
                label: 'Run Price Ratio',
                data: [0], // y-axis data
                borderColor: 'rgba(255, 99, 71, 1)',
                borderWidth: 2,
                fill: false
            },
            {
                label: 'Money Made',
                data: [0],
                borderColor: 'rgba(54, 162, 235, 1)',
                borderWidth: 2,
                unit: '€',
                fill: false
            },
            {
                label: 'Average Money Made By Order',
                data: [0],
                borderColor: 'rgba(255, 206, 86, 1)',
                borderWidth: 2,
                fill: false
            },
            {
                label: 'Number of Orders',
                data: [0],
                borderColor: 'rgba(75, 192, 192, 1)',
                borderWidth: 2,
                fill: false
            }
        ]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            x: {
                type: 'linear',
                min: 10,
                title: {
                    display: true,
                    text: 'Time (hours)',
                    font: {
                        size: 14
                    }
                },
                ticks: {
                    callback: function (value) {
                        return `${Math.floor(value)}:00`; // Hours format
                    },
                    stepSize: 1 // Each tick represents 1 hour
                }
            },
            y: {
                title: {
                    display: true,
                    text: 'Values',
                    font: {
                        size: 14
                    }
                },
                beginAtZero: true
            }
        },
        plugins: {
            legend: {
                display: true,
                position: 'top',
                labels: {
                    font: {
                        size: 12
                    }
                }
            },
            tooltip: {
                mode: 'index',
                intersect: false
            }
        },
        animation: {
            duration: 0
        },
        elements: {
            point: {
                radius: 3
            }
        },
        tooltips: {
            callbacks: {
                label: function (tooltipItem, data) {
                    let label = data.datasets[tooltipItem.datasetIndex].label || '';
                    let value = tooltipItem.yLabel;
                    if (label === 'Money Made' || label === 'Average Money Made By Order') {
                        return `${label}: €${value}`;
                    } else if (label === 'Number of Orders') {
                        return `${label}: ${value}%`;
                    }
                    return `${label}: ${value}`;
                }
            }
        },
    }
});

const updateChart = (statistics) => {
    chart.data.labels.push(`${statistics.time}`);
    chart.data.datasets[0].data.push(statistics.runPriceRatio);
    chart.data.datasets[1].data.push(statistics.moneyMade);
    chart.data.datasets[2].data.push(statistics.averageMoneyMadeByOrder);
    chart.data.datasets[3].data.push(statistics.numOrder);

    chart.update();
}


// ----------------- WEBSOCKET -----------------------------------

var ws = null;
const connectedElem = document.getElementById('connected')


document.getElementById("connect").onclick = function (evt) {
    if (ws) {
        return false;
    }

    const host = "localhost";
    const port = `4444`

    try {
        ws = new WebSocket("ws://" + host + ":" + port + "/ws");
    }
    catch (err) {
        console.error("Error connecting to websocket", err)
        return false;
    }

    ws.onopen = function (evt) {
        console.log("Connection open ...");
        connectedElem.style.backgroundColor = "greenyellow";
    };

    ws.onclose = function (evt) {
        connectedElem.style.backgroundColor = "antiquewhite";
        hasLaunched = false;
        highlightedPaths = []
        destinations = []
        console.log("Connection closed");
        ws = null;
    }

    ws.onmessage = function (evt) {
        gameState = JSON.parse(evt.data)
        updateGame(gameState)
    }

    ws.onerror = function (evt) {
        console.log("Connection error ...");
    }

    return false;
}

// ----------------- Const & global variable -----------------------------------

const delivererLineColor = [97, 27, 31] // Dark Red
const delivererColor = [255, 99, 71, 255] // Red
const replacementPathColor = [211, 84, 0, 204] // Orange
const toRestauPathColor = [41, 128, 255, 255] // Blue 
const toClientPathColor = [39, 174, 96, 204] // Green
const customerColor = [46, 204, 113, 255] // Dark Green
const notHungryCustomerColor = [96, 230, 125] // ligth green
const customerLineColor = [0, 255, 0] // Green
const restaurantColor = [30, 144, 255, 255] // Blue
const lineRestaurantColor = [0, 0, 255]

const delivererHighlight = [255, 140, 100, 255]; // Rouge-orange vif
const replacementPathHighlight = [255, 75, 25, 150]; // Orange vif
const toRestauPathHighlight = [51, 153, 200, 200]; // Bleu ciel plus éclatant
const toClientPathHighlight = [50, 220, 130, 150]; // Vert vif
const customerHighlight = [60, 240, 140, 255]; // Vert lime éclatant
const restaurantHighlight = [60, 180, 255, 255]; // Bleu clair lumineux

let restaurantOverlay = null;
let delivererOverlay = null;
let customerOverlay = null;
let restaurantLayer;
let customerLayer;
let currentTime;
let accelerationTime;
let deliverers;
let numDeliverers;
let trailLength = 1700;
let hasLaunched = false;
const desiredFPS = 30;
const frameDuration = 1000 / desiredFPS; // Duration of one frame in milliseconds

const updateGame = (gameState) => {
    const msgType = gameState.msgType

    if (msgType == "init") {
        updateSimulationInfos(gameState.initInfos, gameState.actualTime)
        currentTime = gameState.currentTimestamp
        accelerationTime = gameState.accelerationTime
        if (accelerationTime <= 0.01) {
            trailLength = 500;
        }
    }

    if (msgType == "totalStatistics") {
        updateStatistics(gameState.statistics)
    }

    if (msgType == "statistics") {
        updateChart(gameState.statistics)
        updateRunPriceRatio(gameState.statistics.runPriceRatio)
    }

    if (msgType == "restaurants") {
        createRestaurantLayer(gameState.restaurants)
    }

    if (msgType == "customers") {
        createCustomerLayer(gameState.customers)
    }

    if (msgType == "deliverers") {
        updateActualTime(gameState.actualTime)
        deliverers = gameState.deliverers

        if (!hasLaunched) {
            currentTime = gameState.currentTimestamp
            numDeliverers = deliverers.length
            hasLaunched = true;
            createDelivererLayer()
            lastRenderTime = Date.now();
            updateAnimationFrame();
        }
    }

}

// ----------------- Update text infos on HTML -----------------------------------

const updateSimulationInfos = (initInfos, actualTime) => {
    document.getElementById('actualTime').textContent = actualTime || '-';
    document.getElementById('duration').textContent = initInfos.duration || '-';
    document.getElementById('accelerationTime').textContent = initInfos.accelerationTime || '-';
    document.getElementById('numDeliverer').textContent = initInfos.numDeliverer || '-';
    document.getElementById('numRestau').textContent = initInfos.numRestau || '-';
    document.getElementById('numCustomer').textContent = initInfos.numCustomer || '-';
}

const updateActualTime = (actualTime) => {
    document.getElementById('actualTime').textContent = actualTime || '-';
}

const updateRunPriceRatio = (runPriceRatio) => {
    document.getElementById('runPriceRatio').textContent = (runPriceRatio * 100).toFixed(2) + ' %' || '-';
}

function updateStatistics(statistics) {
    document.getElementById('moneyMade').textContent = statistics.moneyMade.toFixed(2) + ' €' || '-';
    document.getElementById('averageMoneyMadeByOrder').textContent = statistics.averageMoneyMadeByOrder.toFixed(2) + ' €' || '-';
    document.getElementById('numOrder').textContent = statistics.numOrder || '-';
}

// ----------------- RESTAURANT -----------------------------------

const getRestaurantTooltip = ({ object }) => {
    return object && {
        html: `
            <div><strong>Name:</strong> ${object.name}</div>
            ${object.amenity ? `<div><strong>Amenity:</strong> ${object.amenity}</div > ` : ''}
            <div><strong>Food Type:</strong> ${object.foodType}</div>
            <div><strong>Schedule:</strong> ${object.schedule}</div>
            <div><strong>Price:</strong> ${object.price}</div>
            <div><strong>Preparation Time:</strong> ${object.preparationTime}</div>
            <div><strong>Stock:</strong> ${object.stock}</ div>
        `,
        style: {
            border: '4px solid blue',
            borderRadius: '5px',
        }
    };
}


const createRestaurantLayer = (restaurantsJson) => {

    restaurantLayer = new ScatterplotLayer({
        id: `restaurants`,
        data: restaurantsJson,
        getPosition: d => d.coordinates,
        getFillColor: restaurantColor,
        stroked: true,
        getLineWidth: 1,
        lineWidthMinPixels: 1,
        lineWidthMaxPixels: 7,
        getLineColor: lineRestaurantColor,
        getRadius: 1.75,
        radiusMaxPixels: 50,
        radiusMinPixels: 1,
        radiusScale: 3,
        highlightColor: restaurantHighlight,
        autoHighlight: true,
        pickable: true,
    });

    if (restaurantOverlay) {
        // Update existing overlay
        restaurantOverlay.setProps({ layers: restaurantLayer });
    } else {
        restaurantOverlay = new MapboxOverlay({
            interleaved: false,
            layers: restaurantLayer,

            getTooltip: getRestaurantTooltip
        });
        map.addControl(restaurantOverlay);
    }
}

// ----------------- CUSTOMER -----------------------------------

const getCustomerTooltip = ({ object }) => {
    return object && {
        html: `
            <div><strong>Name:</strong> ${object.name}</div>
            <div><strong>Hungry Level:</strong> ${object.hungryLevel}</div>
            <div><strong>Food Preferences:</strong> ${object.foodPreferences}</div>
            <div><strong>Looking for food:</strong> ${object.wantsToOrder}</div>
        `,
        style: {
            border: '4px solid green',
            borderRadius: '5px',
        }
    };
}
const createCustomerLayer = (customerJson) => {

    customerLayer = new ScatterplotLayer({
        id: `customers`,
        data: customerJson,
        getPosition: d => d.coordinates,
        getFillColor: d => d.wantsToOrder ? customerColor : notHungryCustomerColor,
        stroked: true,
        getLineWidth: 1,
        lineWidthMinPixels: 1,
        lineWidthMaxPixels: 7,
        getLineColor: customerLineColor,
        getRadius: 1.75,
        radiusMaxPixels: 50,
        radiusMinPixels: 1,
        radiusScale: 3,
        highlightColor: customerHighlight,
        autoHighlight: true,
        pickable: true,
    });

    if (customerOverlay) {
        // Update existing overlay
        customerOverlay.setProps({ layers: customerLayer });
    } else {

        customerOverlay = new MapboxOverlay({
            interleaved: false,
            layers: customerLayer,

            getTooltip: getCustomerTooltip
        });
        map.addControl(customerOverlay);
    }
}

// ----------------- DELIVERER -----------------------------------

const moveToDelivererButton = document.getElementById('moveToDeliverer')
let delivererIndex = 0
moveToDelivererButton.onclick = () => {
    const deliverer = deliverers[delivererIndex]
    flyToLocation(deliverer.position)
    displayDelivererPaths(deliverer)
    delivererIndex += 1
    if (delivererIndex >= numDeliverers) {
        delivererIndex = 0
    }
}

const delivererOnClick = (info) => {
    highlightedPaths = []
    destinations = []
}

let lastHoveredDelivererName = null;
let highlightedPaths = [];  // global variable for Delivere paths
let destinations = []; // global variable for Deliverer icons

const displayDelivererPaths = (deliverer) => {
    console.log("Hovered deliverer:", deliverer);
    lastHoveredDelivererName = deliverer.name;
    const delivererData = { name: deliverer.name, position: deliverer.position, currentPathType: deliverer.currentPathType, isMoving: deliverer.isMoving, dailyGoal: deliverer.dailyGoal, moneyMadeToday: deliverer.moneyMadeToday, currentOrder: deliverer.currentOrder, rating: deliverer.rating, numOrder: deliverer.numOrder, state: deliverer.state };
    switch (deliverer.currentPathType) {
        case "replacement":
            highlightedPaths = [{
                color: replacementPathHighlight,
                path: deliverer.replacementPath.pathNode,
                ...delivererData
            }];
            destinations = [{ coordinates: deliverer.replacementPath.destination, color: replacementPathColor, ...delivererData }]
            break;
        case "toRestau":
            highlightedPaths = [{
                ...delivererData,
                color: replacementPathHighlight,
                path: deliverer.replacementPath.pathNode
            }, {
                ...delivererData,
                color: toRestauPathHighlight,
                path: deliverer.toRestauPath.pathNode
            }];
            destinations = [{ coordinates: deliverer.currentOrder.restauCoordinates, color: restaurantColor, ...delivererData }, { coordinates: deliverer.currentOrder.customerCoordinates, color: customerColor, ...delivererData }]
            break;
        case "toClient":
            highlightedPaths = [{
                ...delivererData,
                color: replacementPathHighlight,
                path: deliverer.replacementPath.pathNode
            }, {
                ...delivererData,
                color: toRestauPathHighlight,
                path: deliverer.toRestauPath.pathNode
            },
            {
                ...delivererData,
                color: toClientPathHighlight,
                path: deliverer.toClientPath.pathNode
            }];
            destinations = [{ coordinates: deliverer.currentOrder.customerCoordinates, color: customerColor, ...delivererData }, { coordinates: deliverer.currentOrder.restauCoordinates, color: restaurantColor, ...delivererData }]
            break;
    }
}

const delivererOnHover = (info) => {
    if (info.object && info.object.name !== lastHoveredDelivererName) {
        const deliverer = info.object;
        displayDelivererPaths(deliverer);
    }
};

const returnDelivererLayer = () => {
    const delivererLayer = new TripsLayer({
        id: 'TripsLayer',
        data: deliverers, // Updated on updateGame

        getPath: d => {
            switch (d.currentPathType) {
                case "replacement":
                    return d.replacementPath.pathNode;
                case "toRestau":
                    return d.toRestauPath.pathNode;
                case "toClient":
                    return d.toClientPath.pathNode;
            }
        },
        getTimestamps: d => {
            switch (d.currentPathType) {
                case "replacement":
                    return d.replacementPath.pathTimestamp;
                case "toRestau":
                    return d.toRestauPath.pathTimestamp;
                case "toClient":
                    return d.toClientPath.pathTimestamp;
            }
        },
        getColor: d => {
            switch (d.currentPathType) {
                case "replacement":
                    return replacementPathColor;
                case "toRestau":
                    return toRestauPathColor;
                case "toClient":
                    return toClientPathColor;
            }
        },
        currentTime,
        widthMinPixels: 5,
        trailLength,
        opacity: 1,

        capRounded: true,
        jointRounded: true,
        pickable: true,
        onHover: delivererOnHover,
        onClick: delivererOnClick,
        autoHighlight: true,
    })
    const onHoverPathLayer = new PathLayer({
        id: 'HighlightPaths',
        data: highlightedPaths, // Updated on displayDelivererPaths
        getPath: d => d.path,
        getColor: d => d.color,
        widthScale: 1,
        widthMinPixels: 2,
        getWidth: 20,
        capRounded: true,
        jointRounded: true,
        pickable: true,
        onClick: delivererOnClick,
        opacity: 0.6,
    });
    const destinationLayer = new IconLayer({
        id: 'destinationLayer',
        data: destinations, // Updated on displayDelivererPaths
        getPosition: d => d.coordinates,
        getIcon: d => 'marker',
        getColor: d => d.color,
        getSize: 30,
        iconAtlas: './pin.png',
        iconMapping: {
            marker: {
                x: 0,
                y: 0,
                width: 512,
                height: 512,
                anchorY: 512,
                mask: true
            }
        },
        pickable: true,
        autoHighlight: true,
        onClick: delivererOnClick,
    });
    const idleDeliverers = new ScatterplotLayer({
        id: 'idleDeliverers',
        data: deliverers,
        getPosition: d => d.isMoving ? false : d.position,
        getFillColor: delivererColor,
        stroked: true,
        getLineWidth: 1,
        lineWidthMinPixels: 1,
        lineWidthMaxPixels: 7,
        getLineColor: delivererLineColor,
        getRadius: 1.75,
        radiusMaxPixels: 50,
        radiusMinPixels: 1,
        radiusScale: 3,
        highlightColor: delivererHighlight,
        autoHighlight: true,
        onHover: delivererOnHover,
        pickable: true,
    });
    return { delivererLayer, destinationLayer, idleDeliverers, onHoverPathLayer };
}


const getDelivererTooltip = ({ object }) => {
    return object && {
        html: `
    <div><strong>Name:</strong> ${object.name}</div>
    <div><strong>Rating:</strong> ${object.rating}</div>
    <div><strong>State:</strong> ${object.state}</div>
    ${object.currentPathType ? `<div><strong>Last Move Action:</strong> ${object.currentPathType}</div>` : ''}
    <div><strong>Money Made Today:</strong> ${object.moneyMadeToday}</div>
    <div><strong>Daily Goal:</strong> ${object.dailyGoal}</div>
    <div><strong>Orders treated:</strong> ${object.numOrder}</div>
${(object.currentOrder && object.currentOrder.nbPlate != 0) ? `
    <div><strong>Current Order:</strong></div>
    <div class="order">
        <div>Number of Plates: ${object.currentOrder.nbPlate}</div>
        <div>Total Price: ${object.currentOrder.price}</div>
        <div>Run Price (for deliverer): ${object.currentOrder.runPrice}</div>
        <div><span style="color: blue;">Restaurant:</span>
            ${object.currentOrder.restauName}
        </div>
        <div><span style="color: green;">Customer:</span> 
            ${object.currentOrder.customerName}
        </div>
    </div>
`   : ''}`,
        style: {
            border: '4px solid orange',
            borderRadius: '5px',
        }
    };
}

const createDelivererLayer = () => {
    const { delivererLayer, destinationLayer, idleDeliverers, onHoverPathLayer } = returnDelivererLayer();

    delivererOverlay = new MapboxOverlay({
        interleaved: false,
        layers: [delivererLayer, destinationLayer, idleDeliverers, onHoverPathLayer],
        getTooltip: getDelivererTooltip
    });
    map.addControl(delivererOverlay);

}

let lastRenderTime;
// Animation loop that use global variable currentTime
function updateAnimationFrame() {
    const now = Date.now();
    const deltaTime = now - lastRenderTime; // Time difference in milliseconds


    if (deltaTime >= frameDuration) {
        currentTime += deltaTime; // Increment time locally
        lastRenderTime = now;

        const { delivererLayer, destinationLayer, idleDeliverers, onHoverPathLayer } = returnDelivererLayer();
        delivererOverlay.setProps({
            layers: [delivererLayer, destinationLayer, idleDeliverers, onHoverPathLayer],
        }); // Update the layer with the new time
    }
    requestAnimationFrame(updateAnimationFrame); // Loop the animation
}