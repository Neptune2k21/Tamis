"use client";
import { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";

export default function OAuthCallbackPage() {
  const params = useSearchParams();
  const router = useRouter();
  const [message, setMessage] = useState("Traitement en cours...");
  const [isError, setIsError] = useState(false);

  useEffect(() => {
    const email = params.get("email");
    const name = params.get("name");
    const provider = params.get("provider");

    if (!email || !provider) {
      setMessage("Paramètres manquants dans l'URL. Veuillez réessayer la connexion.");
      setIsError(true);
      return;
    }

    // Procéder à la connexion/inscription automatique
    handleOAuthLogin(email, name || "Utilisateur", provider);
  }, [params, router]);

  const handleOAuthLogin = async (email: string, name: string, provider: string) => {
    try {
      setMessage(`Connexion en cours pour ${email}...`);
      
      // Appel à l'API pour finaliser la connexion OAuth
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const response = await fetch(`${API_URL}/api/auth/google/callback-process`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, name, provider }),
      });

      const data = await response.json();

      if (data.success && data.data?.token) {
        // Stocker le token JWT
        localStorage.setItem("jwt", data.data.token);
        localStorage.setItem("userEmail", data.data.user.email);
        localStorage.setItem("userName", data.data.user.username);
        
        setMessage(`Bienvenue ${data.data.user.username} ! Redirection en cours...`);
        
        // Rediriger vers le dashboard après 2 secondes
        setTimeout(() => {
          router.push("/dashboard");
        }, 2000);
      } else {
        setMessage(data.message || "Erreur lors de la connexion OAuth");
        setIsError(true);
      }
    } catch (error) {
      console.error("OAuth callback error:", error);
      setMessage("Erreur de connexion au serveur");
      setIsError(true);
    }
  };

  return (
    <div className="page-container">
      <div className="card form-container">
        <div className="text-center space-y-6">
          {!isError ? (
            <>
              <div className="text-6xl animate-bounce">✅</div>
              <h1 className="text-2xl font-bold text-gradient">Connexion OAuth</h1>
              <p className="text-neutral">{message}</p>
              <div className="flex justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              </div>
            </>
          ) : (
            <>
              <div className="text-6xl">❌</div>
              <h1 className="text-2xl font-bold text-red-600">Erreur OAuth</h1>
              <p className="text-neutral bg-red-50 p-4 rounded-lg border border-red-200">
                {message}
              </p>
              <div className="flex flex-col space-y-3">
                <Link href="/auth/google" className="btn-google">
                  Réessayer avec Google
                </Link>
                <Link href="/auth/login" className="btn-secondary">
                  Connexion classique
                </Link>
                <Link href="/" className="text-primary hover:underline">
                  Retour à l&apos;accueil
                </Link>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}