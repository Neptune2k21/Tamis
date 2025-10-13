"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

interface User {
  id: number;
  email: string;
  username: string;
  auth_provider: string;
}

export default function DashboardPage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem("jwt");
    if (!token) {
      router.push("/auth/login");
      return;
    }

    // Récupérer les infos utilisateur
    fetchUserInfo(token);
  }, [router]);

  const fetchUserInfo = async (token: string) => {
    try {
      const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const response = await fetch(`${API_URL}/api/user/me`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const data = await response.json();
      
      if (data.success) {
        setUser(data.data);
      } else {
        console.error("Failed to fetch user:", data.message);
        localStorage.removeItem("jwt");
        router.push("/auth/login");
      }
    } catch (error) {
      console.error("Error fetching user:", error);
      localStorage.removeItem("jwt");
      router.push("/auth/login");
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem("jwt");
    localStorage.removeItem("userEmail");
    router.push("/");
  };

  if (loading) {
    return (
      <div className="page-container">
        <div className="card">
          <div className="flex items-center justify-center space-x-2">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            <span>Chargement...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <h1 className="text-2xl font-bold text-gradient">Tamis 🧹</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">Bonjour, {user?.username}</span>
              <button 
                onClick={handleLogout}
                className="btn-secondary text-sm py-2 px-4"
              >
                Déconnexion
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="space-y-8">
          {/* Welcome Card */}
          <div className="card">
            <div className="text-center space-y-4">
              <h2 className="text-3xl font-bold text-gradient">
                Bienvenue sur votre dashboard !
              </h2>
              <p className="text-neutral max-w-2xl mx-auto">
                Connecté en tant que <strong>{user?.email}</strong> via {" "}
                <span className="capitalize bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-sm">
                  {user?.auth_provider}
                </span>
              </p>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="grid md:grid-cols-3 gap-6">
            <div className="card text-center hover:shadow-xl transition-shadow cursor-pointer">
              <div className="text-4xl mb-4">📧</div>
              <h3 className="text-lg font-semibold mb-2">Mes Emails</h3>
              <p className="text-neutral text-sm">Voir tous vos emails consolidés</p>
            </div>
            
            <div className="card text-center hover:shadow-xl transition-shadow cursor-pointer">
              <div className="text-4xl mb-4">⚙️</div>
              <h3 className="text-lg font-semibold mb-2">Comptes</h3>
              <p className="text-neutral text-sm">Gérer vos comptes emails</p>
            </div>
            
            <div className="card text-center hover:shadow-xl transition-shadow cursor-pointer">
              <div className="text-4xl mb-4">🔄</div>
              <h3 className="text-lg font-semibold mb-2">Synchroniser</h3>
              <p className="text-neutral text-sm">Mettre à jour vos emails</p>
            </div>
          </div>

          {/* Stats */}
          <div className="card">
            <h3 className="text-xl font-semibold mb-4">Statistiques</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center p-4 bg-blue-50 rounded-lg">
                <div className="text-2xl font-bold text-blue-600">0</div>
                <div className="text-sm text-blue-600">Comptes connectés</div>
              </div>
              <div className="text-center p-4 bg-green-50 rounded-lg">
                <div className="text-2xl font-bold text-green-600">0</div>
                <div className="text-sm text-green-600">Emails traités</div>
              </div>
              <div className="text-center p-4 bg-yellow-50 rounded-lg">
                <div className="text-2xl font-bold text-yellow-600">0</div>
                <div className="text-sm text-yellow-600">Emails supprimés</div>
              </div>
              <div className="text-center p-4 bg-purple-50 rounded-lg">
                <div className="text-2xl font-bold text-purple-600">0 Mo</div>
                <div className="text-sm text-purple-600">Espace libéré</div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}