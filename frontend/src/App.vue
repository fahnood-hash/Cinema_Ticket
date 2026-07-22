<script setup>
import { ref } from "vue";
import { signInWithPopup, signOut } from "firebase/auth";
import { auth, googleProvider } from "./firebase";
import BookingPage from "./components/BookingPage.vue";
import AdminPage from "./components/AdminPage.vue";

const user = ref(null);
const error = ref("");
const role = ref("USER");
const currentPage = ref("home");

function openAdminPage() {
  currentPage.value = "admin";
}

async function login() {
  error.value = "";

  try {
    const result = await signInWithPopup(auth, googleProvider);
    user.value = result.user;
    const token = await result.user.getIdToken();

    const response = await fetch("http://localhost:8080/me", {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error("Could not load user profile");
    }

    const profile = await response.json();
    role.value = profile.role;
  } catch (err) {
    error.value = err.message;
  }
}

async function logout() {
  await signOut(auth);
  user.value = null;
  role.value = "USER";
  currentPage.value = "home";
}

function openBookingPage() {
  currentPage.value = "bookings";
}

function goHome() {
  currentPage.value = "home";
}
</script>

<template>
  <BookingPage
    v-if="user && currentPage === 'bookings'"
    :user-id="user.uid"
    @back="goHome"
  />

  <AdminPage
  v-else-if="user && currentPage === 'admin'"
  @back="goHome"
  />

  <main v-else>
    <h1>Cinema Ticket Booking</h1>
    
    <button v-if="!user" @click="login">
      Sign in with Google
    </button>

    <section v-else>
      <p>Signed in as: {{ user.email }}</p>
      <p>Role: {{ role }}</p>

      <button @click="openBookingPage">
        My bookings
      </button>

      <button v-if="role === 'ADMIN'"
        @click="openAdminPage"
      >
        Admin Dashboard
      </button>

      <button @click="logout">
        Sign out
      </button>

      <p class="success">Firebase login successful</p>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
  </main>
</template>

<style scoped>
main {
  max-width: 640px;
  margin: 80px auto;
  padding: 24px;
  text-align: center;
  font-family: Arial, sans-serif;
}

button {
  margin: 4px;
  padding: 12px 18px;
  border: 0;
  border-radius: 8px;
  background: #2563eb;
  color: white;
  cursor: pointer;
}

.success {
  color: #15803d;
}

.error {
  color: #dc2626;
}
</style>