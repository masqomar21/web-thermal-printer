import { io } from "socket.io-client";

const socket = io("https://backend-mpp.newus.id", {
  transports: ["websocket"],
  reconnection: true,
});

const printData = {
  instansi: "test",
  layanan: "test",
  nomor_antrean: "test",
  tanggal: "test",
  jam: "test",
  loket: `Loket test`,
};

socket.on("connect", () => {
  console.log("✅ Connected:", socket.id);

  // Kirim event ke server
  socket.emit("antrean_print", printData);
  socket.emit("watch", printData);
  console.log("📤 Data dikirim:", printData);

  // Tutup koneksi setelah 3 detik
});

socket.on("antrean_print", (msg) => {
  console.log("📩 Chat diterima:", msg);
});

socket.on("disconnect", (reason) => {
  console.log("❌ Disconnected", reason);
});
