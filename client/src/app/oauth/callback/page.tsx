"use client";
import { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { completeGoogleOAuth } from "@/lib/api/auth";

export default function OAuthCallbackPage() {
  const params = useSearchParams();
  const router = useRouter();
  const [message, setMessage] = useState("Traitement en cours...");

  useEffect(() => {
    const code = params.get("code");
    if (!code) {
      setMessage("Code OAuth manquant dans l'URL.");
      return;
    }
    const jwt = localStorage.getItem("jwt");
    if (!jwt) {
      setMessage("Veuillez d'abord vous connecter.");
      return;
    }
    completeGoogleOAuth(jwt, code).then((data) => {
      if (data.success) {
        setMessage("Compte Google ajouté avec succès !");
        setTimeout(() => router.push("/"), 2000);
      } else {
        setMessage(data.message || data.error || "Erreur lors de l'ajout du compte Google.");
      }
    });
  }, [params, router]);

  return (
    <main className="max-w-md mx-auto mt-16 p-6 bg-white rounded shadow">
      <h1 className="text-2xl font-bold mb-4">Callback OAuth</h1>
      <p>{message}</p>
    </main>
  );
}