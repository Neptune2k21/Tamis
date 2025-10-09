"use client";
import { useState } from "react";
import { login, initiateGoogleOAuth } from "@/lib/api/auth";

export default function LoginPage() {
  const [form, setForm] = useState({ email: "", password: "" });
  const [message, setMessage] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setMessage("");
    const data = await login(form);
    setMessage(data.message || data.error);
    if (data.success && data.data?.token) {
      localStorage.setItem("jwt", data.data.token);
    }
  }

  async function handleGoogleOAuth() {
    const jwt = localStorage.getItem("jwt");
    if (!jwt) {
      setMessage("Veuillez d'abord vous connecter pour obtenir un JWT.");
      return;
    }
    const data = await initiateGoogleOAuth(jwt);
    if (data.data?.auth_url) {
      window.location.href = data.data.auth_url;
    } else {
      setMessage(data.message || data.error || "Erreur OAuth");
    }
  }

  return (
    <main className="max-w-md mx-auto mt-16 p-6 bg-white rounded shadow">
      <h1 className="text-2xl font-bold mb-4">Connexion</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          className="w-full border p-2 rounded"
          placeholder="Email"
          type="email"
          value={form.email}
          onChange={e => setForm({ ...form, email: e.target.value })}
          required
        />
        <input
          className="w-full border p-2 rounded"
          placeholder="Mot de passe"
          type="password"
          value={form.password}
          onChange={e => setForm({ ...form, password: e.target.value })}
          required
        />
        <button className="w-full bg-primary text-white p-2 rounded" type="submit">
          Se connecter
        </button>
      </form>
      <button
        className="w-full mt-4 bg-red-500 text-white p-2 rounded"
        onClick={handleGoogleOAuth}
        type="button"
      >
        Se connecter avec Google
      </button>
      {message && <p className="mt-4 text-center">{message}</p>}
    </main>
  );
}