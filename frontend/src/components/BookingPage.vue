<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";

const props = defineProps({
  userId: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["back"]);

const API_URL = "http://localhost:8080";

const seats = ref([]);
const activeBooking = ref(null);
const error = ref("");
const now = ref(Date.now());
const confirmedBooking = ref(null);

let pollTimer;
let clockTimer;

const seatRows = computed(() => {
  const rows = {};

  for (const seat of seats.value) {
    const row = seat.id[0];

    if (!rows[row]) {
      rows[row] = [];
    }

    rows[row].push(seat);
  }

  return rows;
});

const remainingTime = computed(() => {
  if (!activeBooking.value?.expires_at) return "";

  const seconds = Math.max(
    0,
    Math.floor((new Date(activeBooking.value.expires_at).getTime() - now.value) / 1000),
  );

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  return `${String(minutes).padStart(2, "0")}:${String(remainingSeconds).padStart(2, "0")}`;
});

async function loadSeats() {
  try {
    const response = await fetch(`${API_URL}/seats`);

    if (!response.ok) {
      throw new Error("Could not load seats");
    }

    seats.value = await response.json();
  } catch (err) {
    error.value = err.message;
  }
}

async function lockSeat(seat) {
  if (seat.status !== "AVAILABLE" || activeBooking.value) return;

  error.value = "";

  try {
    const response = await fetch(`${API_URL}/seats/${seat.id}/lock`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: props.userId }),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || "Could not lock seat");
    }

    activeBooking.value = data;
    await loadSeats();
  } catch (err) {
    error.value = err.message;
  }
}

async function confirmBooking() {
  if (!activeBooking.value) return;

  try {
    const response = await fetch(
      `${API_URL}/bookings/${activeBooking.value.id}/confirm`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: props.userId }),
      },
    );

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || "Could not confirm booking");
    }
    confirmedBooking.value = data;
    activeBooking.value = null;
    await loadSeats();
  } catch (err) {
    error.value = err.message;
  }
}

async function releaseBooking() {
  if (!activeBooking.value) return;

  try {
    const response = await fetch(
      `${API_URL}/bookings/${activeBooking.value.id}`,
      {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: props.userId }),
      },
    );

    if (!response.ok) {
      throw new Error("Could not release seat");
    }

    activeBooking.value = null;
    await loadSeats();
  } catch (err) {
    error.value = err.message;
  }
}

function seatClass(seat) {
  if (seat.status === "BOOKED") return "seat--booked";
  if (activeBooking.value?.seat_id === seat.id) return "seat--mine";
  if (seat.status === "LOCKED") return "seat--locked";
  return "";
}

onMounted(() => {
  loadSeats();
  pollTimer = setInterval(loadSeats, 2000);
  clockTimer = setInterval(async () => {
  now.value = Date.now();

  if (
    activeBooking.value?.expires_at &&
    new Date(activeBooking.value.expires_at).getTime() <= now.value
  ) {
    activeBooking.value = null;
    error.value = "Your 5-minute seat hold expired.";
    await loadSeats();
  }
}, 1000);
});

onUnmounted(() => {
  clearInterval(pollTimer);
  clearInterval(clockTimer);
});
</script>

<template>
  <main class="booking-page">
    <header>
    <p class="user-id">
        User: {{ props.userId }}
    </p>

    <button class="back-button" @click="emit('back')">
        Back
    </button>
    </header>

    <h2 class="movie-title">The Odyssey</h2>

    <section class="content">
      <div class="screen-area">
        <p class="screen-label">Screen</p>
        <div class="screen-bar"></div>

        <div class="seat-grid">
          <div v-for="(row, label) in seatRows" :key="label" class="seat-row">
            <span class="row-label">{{ label }}</span>

            <button
              v-for="seat in row"
              :key="seat.id"
              class="seat"
              :class="seatClass(seat)"
              :disabled="seat.status !== 'AVAILABLE' || !!activeBooking"
              @click="lockSeat(seat)"
            >
              {{ seat.id }}
            </button>
          </div>
        </div>

        <div class="legend">
          <span><i class="available"></i> Available</span>
          <span><i class="mine"></i> Your hold</span>
          <span><i class="locked"></i> Locked</span>
          <span><i class="booked"></i> Booked</span>
        </div>
      </div>

      <aside class="checkout">
        <h2>Checkout</h2> 

        <template v-if="activeBooking">
          <p>Seat: <strong>{{ activeBooking.seat_id }}</strong></p>
          <p>Time remaining</p>
          <p class="timer">{{ remainingTime }}</p>

          <button class="confirm" @click="confirmBooking">Confirm</button>
          <button class="release" @click="releaseBooking">Release</button>
        </template>

        <p v-else>Select an available seat.</p>
      </aside>
    </section>

    <section v-if="confirmedBooking" class="confirmation">
    <h2>Booking confirmed</h2>
    <p>Booking ID: {{ confirmedBooking.id }}</p>
    <p>Seat: {{ confirmedBooking.seat_id }}</p>
    <p>User: {{ confirmedBooking.user_id }}</p>
    <p>Status: {{ confirmedBooking.status }}</p>
    </section>


    <p v-if="error" class="error">{{ error }}</p>
  </main>
</template>

<style scoped>
.booking-page {
  min-height: 100vh;
  padding: 32px;
  background: #ffffff;
  color: #e0e0e0;
  font-family: Arial, sans-serif;
}

.user-id {
  margin: 0;
  padding: 8px 12px;
  border-radius: 6px;
  background: #1a1a1a;
  color: #ffffff;
  font-family: monospace;
  font-size: 13px;
}

header,
.content {
  max-width: 1000px;
  margin: 0 auto;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #333333;
  padding-bottom: 18px;
}

h1,
h2 {
  margin: 0;
}

.movie-title {
  max-width: 1000px;
  margin: 24px auto;
  color: #4fc3f7;
}

.content {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}

.screen-area {
  flex: 1;
  min-width: 320px;
}

.screen-label {
  text-align: center;
  color: #888;
  text-transform: uppercase;
  font-size: 12px;
}

.screen-bar {
  height: 4px;
  margin: 12px 30px 28px;
  background: linear-gradient(90deg, transparent, #4fc3f7, transparent);
}

.seat-grid {
  display: grid;
  gap: 10px;
  justify-content: center;
}

.seat-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.row-label {
  width: 22px;
  color: #888;
}

.seat {
  width: 42px;
  height: 35px;
  border: 0;
  border-radius: 6px 6px 3px 3px;
  color: #ddd;
  background: #2a2a2a;
  cursor: pointer;
}

.seat:hover:not(:disabled) {
  background: #3a3a3a;
}

.seat:disabled {
  cursor: not-allowed;
}

.seat--mine {
  background: #d4a017;
  color: #111;
}

.seat--locked {
  background: #e06c00;
}

.seat--booked {
  background: #c0392b;
}

.legend {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
  flex-wrap: wrap;
  color: #aaa;
  font-size: 13px;
}

.legend i {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 3px;
}

.available { background: #2a2a2a; }
.mine { background: #d4a017; }
.locked { background: #e06c00; }
.booked { background: #c0392b; }

.checkout {
  width: 250px;
  padding: 22px;
  border: 1px solid #333;
  border-radius: 10px;
  background: #1a1a1a;
  margin: 30px auto 0;
}

.checkout h2 {
  margin-top: 0;
  color: #4fc3f7;
}

.confirmation h2 {
  margin-top: 0;
  color: #4fc3f7;
}

.confirmation {
  width: 100%;
  margin-top: 24px;
  padding: 20px;
  border: 1px solid #27ae60;
  border-radius: 10px;
  background: #163323;
  color: white;
  text-align: center;
}

.timer {
  color: #d4a017;
  font-size: 32px;
  font-weight: bold;
  text-align: center;
}

button {
  padding: 10px 14px;
  border: 0;
  border-radius: 6px;
  color: white;
  cursor: pointer;
}

.back-button { background: #333; }
.confirm { width: 100%; margin-bottom: 8px; background: #27ae60; }
.release { width: 100%; background: #e74c3c; }

.error {
  max-width: 1000px;
  margin: 20px auto;
  color: #e74c3c;
}
</style>