const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Inscription
export async function register({ email, username, password }: { email: string; username: string; password: string }) {
  const res = await fetch(`${API_URL}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, username, password }),
  });
  return res.json();
}

// Connexion classique
export async function login({ email, password }: { email: string; password: string }) {
  const res = await fetch(`${API_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  return res.json();
}

// Rafraîchir le token JWT
export async function refreshToken(token: string) {
  const res = await fetch(`${API_URL}/api/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token }),
  });
  return res.json();
}

// Initier OAuth Google (nécessite le JWT)
export async function initiateGoogleOAuth(jwt: string) {
  const res = await fetch(`${API_URL}/api/oauth/google/initiate`, {
    headers: { Authorization: `Bearer ${jwt}` },
  });
  return res.json();
}

// Finaliser l'ajout d'un compte Google (POST /api/oauth/complete)
export async function completeGoogleOAuth(jwt: string, code: string) {
  const res = await fetch(`${API_URL}/api/oauth/complete`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${jwt}`,
    },
    body: JSON.stringify({ code }),
  });
  return res.json();
}