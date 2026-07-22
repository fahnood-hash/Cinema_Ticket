<script setup>
import { onMounted, ref } from "vue";
import { auth } from "../firebase";

const emit = defineEmits(["back"]);

const bookings = ref([]);
const userFilter = ref("");
const seatFilter = ref("");
const error = ref("");
const loading = ref(false);

async function loadBookings() {
  loading.value = true;
  error.value = "";

  try {
    const token = await auth.currentUser.getIdToken();

    const params = new URLSearchParams();

    if (userFilter.value) {
      params.set("user_id", userFilter.value);
    }

    if (seatFilter.value) {
      params.set("seat_id", seatFilter.value);
    }

    const query = params.toString();
    const url = `http://localhost:8080/admin/bookings${query ? `?${query}` : ""}`;

    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || "Could not load bookings");
    }

    bookings.value = data;
  } catch (err) {
    error.value = err.message;
  } finally {
    loading.value = false;
  }
}

onMounted(loadBookings);
</script>

<template>
  <main class="admin-page">
    <header>
      <div>
        <p class="eyebrow">ADMIN</p>
        <h1>Booking Dashboard</h1>
      </div>

      <button class="back-button" @click="emit('back')">
        Back
      </button>
    </header>

    <section class="filters">
      <input v-model="userFilter" placeholder="Filter by user ID" />
      <input v-model="seatFilter" placeholder="Filter by seat ID, e.g. A1" />

      <button @click="loadBookings">
        Search
      </button>
    </section>

    <p v-if="loading">Loading bookings...</p>
    <p v-if="error" class="error">{{ error }}</p>

    <section v-if="!loading && !error" class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Booking ID</th>
            <th>User ID</th>
            <th>Seat</th>
            <th>Status</th>
            <th>Created at</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="booking in bookings" :key="booking.id">
            <td>{{ booking.id }}</td>
            <td>{{ booking.user_id }}</td>
            <td>{{ booking.seat_id }}</td>
            <td><span class="status">{{ booking.status }}</span></td>
            <td>{{ new Date(booking.created_at).toLocaleString() }}</td>
          </tr>

          <tr v-if="bookings.length === 0">
            <td colspan="5">No bookings found.</td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<style scoped>

.admin-page {
  min-height: 100vh;
  box-sizing: border-box;
  /*padding: 32px;*/
  background: #0f0f0f;
  color: #e0e0e0;
  font-family: Arial, sans-serif;
}

header,
.filters,
.table-wrap {
  width: 100%;
  max-width: none;
  margin: 0;
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
  padding-bottom: 18px;
  border-bottom: 1px solid #333;
}

h1 {
  margin: 0;
  color: #4fc3f7;
}

.eyebrow {
  margin: 0 0 6px;
  color: #4fc3f7;
  font-size: 12px;
  font-weight: bold;
  letter-spacing: 2px;
}

.filters {
  display: flex;
  gap: 10px;
  margin-bottom: 22px;
  flex-wrap: wrap;
}

input {
  min-width: 220px;
  padding: 11px;
  border: 1px solid #444;
  border-radius: 6px;
  background: #1a1a1a;
  color: white;
}

button {
  padding: 11px 16px;
  border: 0;
  border-radius: 6px;
  background: #2563eb;
  color: white;
  cursor: pointer;
}

.back-button {
  background: #333;
}

.table-wrap {
  height: calc(100vh - 230px);
  overflow: auto;
  border: 1px solid #333;
  border-radius: 8px;
}

.table-wrap {
  width: 100%;
  height: calc(100vh - 230px);
  overflow-y: auto;
  overflow-x: hidden;
  border: 1px solid #333;
  border-radius: 8px;
}

table {
  width: 100%;
  min-width: 0;
  table-layout: fixed;
  border-collapse: collapse;
}

th,
td {
  overflow-wrap: anywhere;
  word-break: break-word;
}

th {
  color: #4fc3f7;
  background: #1a1a1a;
}

.status {
  color: #27ae60;
  font-weight: bold;
}

.error {
  max-width: 1100px;
  margin: 20px auto;
  color: #e74c3c;
}
</style>