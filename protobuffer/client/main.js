const API_URL = "http://localhost:8080/fetch";
const REFRESH_MS = 15000;

const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
const TWEEN_MS = reduceMotion ? 0 : 1200;

const OCCUPANCY = {
  EMPTY: { tone: "roomy", label: "Empty" },
  MANY_SEATS_AVAILABLE: { tone: "roomy", label: "Many seats free" },
  FEW_SEATS_AVAILABLE: { tone: "some", label: "Few seats free" },
  STANDING_ROOM_ONLY: { tone: "busy", label: "Standing room only" },
  CRUSHED_STANDING_ROOM_ONLY: { tone: "busy", label: "Crowded" },
  FULL: { tone: "busy", label: "Full" },
  NOT_ACCEPTING_PASSENGERS: { tone: "busy", label: "Not accepting passengers" },
};
const NO_DATA = { tone: "unknown", label: "No seat data" };

const STATUS = {
  STOPPED_AT: "At a stop",
  INCOMING_AT: "Arriving at a stop",
  IN_TRANSIT_TO: "On the move",
};

const ARROW_SVG =
  '<svg class="arrow" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 2 L20 21 L12 16.5 L4 21 Z"/></svg>';

/* Map */

const map = L.map("map", { zoomControl: false }).setView([48.4284, -123.3656], 12);
L.control.zoom({ position: window.innerWidth <= 640 ? "topright" : "bottomright" }).addTo(map);

L.tileLayer("https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png?key=cb1_5528_1_41c6030ab831eebf5b285ad8", {
  maxZoom: 19,
  subdomains: "abcd",
  attribution: "&copy; OpenStreetMap contributors &copy; CARTO",
}).addTo(map);

const layer = L.featureGroup().addTo(map);

/* State */

const buses = new Map(); // vehicle id -> { marker, data, tween }
const selectedRoutes = new Set();
let routeSignature = "";
let firstLoad = true;
let animating = false;
let lastFeedTime = null;

const els = {
  summary: document.getElementById("summary"),
  routes: document.getElementById("routes"),
  clear: document.getElementById("clear"),
  status: document.getElementById("status"),
};

/* Data */

function normalize(entity) {
  const v = entity.vehicle;
  if (!v || !v.position || !v.trip) return null;
  return {
    id: v.vehicle?.id ?? entity.id,
    route: String(v.trip.routeId ?? "?").replace(/-VIC$/, ""),
    lat: v.position.latitude,
    lng: v.position.longitude,
    bearing: v.position.bearing,
    speedKmh: Math.round((v.position.speed ?? 0) * 3.6),
    status: v.currentStatus,
    occupancy: v.occupancyStatus,
    startTime: v.trip.startTime,
  };
}

async function load() {
  try {
    const res = await fetch(API_URL);
    if (!res.ok) throw new Error(`Server replied ${res.status}`);
    const feed = await res.json();

    const list = (feed.entity ?? []).map(normalize).filter(Boolean);
    lastFeedTime = feed.header?.timestamp ? new Date(Number(feed.header.timestamp) * 1000) : new Date();

    applyFeed(list);
    els.status.textContent = "";
  } catch (err) {
    console.error(err);
    els.status.textContent =
      "Can't load buses from localhost:8080. Check that the server is running and allows requests from this page. Showing the last positions received.";
  }
}

/* Rendering */

function tone(d) {
  return (OCCUPANCY[d.occupancy] ?? NO_DATA).tone;
}

function isMoving(d) {
  return d.bearing !== undefined && d.speedKmh > 1;
}

function popupHtml(d) {
  const occ = OCCUPANCY[d.occupancy] ?? NO_DATA;
  const speed = d.speedKmh > 0 ? ` at ${d.speedKmh} km/h` : "";
  return `
    <div class="popup">
      <strong class="route">Route ${d.route}</strong>
      <p>${STATUS[d.status] ?? "Position reported"}${speed}</p>
      <p><span class="tone ${occ.tone}"></span>${occ.label}</p>
      <p class="small">Trip started ${d.startTime.slice(0, 5)}</p>
    </div>`;
}

function paint(entry) {
  const d = entry.data;
  const el = entry.marker.getElement()?.querySelector(".bus");
  if (!el) return;
  el.className = `bus ${tone(d)}${isMoving(d) ? " moving" : ""}`;
  el.style.setProperty("--tone", `var(--${tone(d)})`);
  el.style.setProperty("--bearing", `${d.bearing ?? 0}deg`);
  entry.marker.setPopupContent(popupHtml(d));
}

function applyFeed(list) {
  const seen = new Set();

  for (const d of list) {
    seen.add(d.id);
    let entry = buses.get(d.id);

    if (!entry) {
      const marker = L.marker([d.lat, d.lng], {
        icon: L.divIcon({
          className: "bus-icon",
          html: `<span class="bus">${ARROW_SVG}<span class="dot"></span></span>`,
          iconSize: [28, 28],
          iconAnchor: [14, 14],
          popupAnchor: [0, -12],
        }),
        title: `Route ${d.route}`,
        keyboard: true,
      }).bindPopup("");
      entry = { marker, data: d, tween: null };
      buses.set(d.id, entry);
      marker.addTo(layer);
    } else {
      const from = entry.marker.getLatLng();
      entry.tween = { from, to: L.latLng(d.lat, d.lng), start: performance.now() };
      entry.data = d;
    }
  }

  for (const [id, entry] of buses) {
    if (!seen.has(id)) {
      layer.removeLayer(entry.marker);
      buses.delete(id);
    }
  }

  for (const entry of buses.values()) paint(entry);

  buildRouteChips();
  applyFilter();
  startTween();

  if (firstLoad && buses.size) {
    map.fitBounds(layer.getBounds().pad(0.08));
    firstLoad = false;
  }
}

/* Smooth movement between refreshes */

function startTween() {
  if (animating || TWEEN_MS === 0) {
    if (TWEEN_MS === 0) {
      for (const e of buses.values()) {
        if (e.tween) e.marker.setLatLng(e.tween.to);
        e.tween = null;
      }
    }
    return;
  }
  animating = true;
  requestAnimationFrame(tick);
}

function tick(now) {
  let active = false;
  for (const e of buses.values()) {
    if (!e.tween) continue;
    const t = Math.min(1, (now - e.tween.start) / TWEEN_MS);
    const k = t * (2 - t);
    e.marker.setLatLng([
      e.tween.from.lat + (e.tween.to.lat - e.tween.from.lat) * k,
      e.tween.from.lng + (e.tween.to.lng - e.tween.from.lng) * k,
    ]);
    if (t < 1) active = true;
    else e.tween = null;
  }
  if (active) requestAnimationFrame(tick);
  else animating = false;
}

/* Route filter */

function buildRouteChips() {
  const routes = [...new Set([...buses.values()].map((e) => e.data.route))].sort((a, b) => {
    const na = parseInt(a, 10);
    const nb = parseInt(b, 10);
    return na - nb || a.localeCompare(b);
  });

  const signature = routes.join(",");
  if (signature === routeSignature) return;
  routeSignature = signature;

  els.routes.replaceChildren(
    ...routes.map((route) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "chip";
      btn.textContent = route;
      btn.dataset.route = route;
      btn.setAttribute("aria-pressed", String(selectedRoutes.has(route)));
      btn.setAttribute("aria-label", `Route ${route}`);
      return btn;
    })
  );
}

els.routes.addEventListener("click", (event) => {
  const btn = event.target.closest(".chip");
  if (!btn) return;
  const route = btn.dataset.route;
  if (selectedRoutes.has(route)) selectedRoutes.delete(route);
  else selectedRoutes.add(route);
  btn.setAttribute("aria-pressed", String(selectedRoutes.has(route)));
  applyFilter();
});

els.clear.addEventListener("click", () => {
  selectedRoutes.clear();
  for (const chip of els.routes.children) chip.setAttribute("aria-pressed", "false");
  applyFilter();
});

function applyFilter() {
  const filtering = selectedRoutes.size > 0;
  let shown = 0;

  for (const entry of buses.values()) {
    const visible = !filtering || selectedRoutes.has(entry.data.route);
    const onMap = layer.hasLayer(entry.marker);
    if (visible && !onMap) layer.addLayer(entry.marker);
    if (!visible && onMap) layer.removeLayer(entry.marker);
    if (visible) shown++;
  }

  els.clear.hidden = !filtering;

  const time = lastFeedTime
    ? lastFeedTime.toLocaleTimeString("en-CA", { hour: "numeric", minute: "2-digit", second: "2-digit" })
    : "";
  const count = filtering
    ? `${shown} of ${buses.size} buses`
    : `${buses.size} ${buses.size === 1 ? "bus" : "buses"} on the road`;
  els.summary.textContent = time ? `${count}. Updated ${time}` : count;
}

/* Go */

load();
setInterval(load, REFRESH_MS);