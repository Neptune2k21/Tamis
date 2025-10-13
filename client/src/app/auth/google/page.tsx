"use client";
import { useEffect, useState } from "react";
import Link from "next/link";

export default function GoogleOAuthPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const initiateGoogleOAuth = async () => {
      try {
        const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
        const response = await fetch(`${API_URL}/api/auth/google`);
        const data = await response.json();
        
        if (data.success && data.data?.auth_url) {
          // Rediriger vers Google OAuth
          window.location.href = data.data.auth_url;
        } else {
          setError(data.message || "Erreur lors de l'initialisation OAuth");
          setLoading(false);
        }
      } catch  {
        setError("Erreur de connexion au serveur");
        setLoading(false);
      }
    };

    initiateGoogleOAuth();
  }, []);

  if (loading) {
    return (
      <div className="page-container">
        <div className="card form-container">
          <div className="text-center space-y-4">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
            <h1 className="text-2xl font-bold text-gradient">Redirection vers Google...</h1>
            <p className="text-neutral">Vous allez être redirigé vers Google pour vous authentifier.</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="card form-container">
        <div className="text-center space-y-4">
          <h1 className="text-2xl font-bold text-red-600">Erreur d&apos;authentification</h1>
          <Link 
            href="/"
            className="btn-primary inline-block"
          >
            Retour à l&apos;accueil
          </Link>
        </div>
      </div>
    </div>
  );
}