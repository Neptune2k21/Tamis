import Link from "next/link";

export default function HomePage() {
  return (
    <div className="page-container bg-gradient-to-br from-primary via-secondary to-neutral text-white">
      <div className="text-center space-y-8 max-w-4xl mx-auto">
        {/* Hero Section */}
        <div className="space-y-6">
          <h1 className="text-6xl font-bold font-display animate-fadeIn">
            Tamis 🧹
          </h1>
          <p className="text-xl text-light max-w-2xl mx-auto leading-relaxed">
            Nettoyez vos emails avec élégance et sécurité. Connectez plusieurs comptes, 
            filtrez intelligemment et supprimez en toute confiance.
          </p>
        </div>

        {/* Features */}
        <div className="grid md:grid-cols-3 gap-6 mt-12">
          <div className="card bg-white/10 backdrop-blur-md border-white/20 text-center">
            <div className="text-4xl mb-4">🔗</div>
            <h3 className="text-lg font-semibold mb-2">Multi-comptes</h3>
            <p className="text-light text-sm">Gmail, Yahoo, Outlook et plus</p>
          </div>
          <div className="card bg-white/10 backdrop-blur-md border-white/20 text-center">
            <div className="text-4xl mb-4">🎯</div>
            <h3 className="text-lg font-semibold mb-2">Filtres intelligents</h3>
            <p className="text-light text-sm">Détection automatique des doublons</p>
          </div>
          <div className="card bg-white/10 backdrop-blur-md border-white/20 text-center">
            <div className="text-4xl mb-4">🛡️</div>
            <h3 className="text-lg font-semibold mb-2">Sécurité maximale</h3>
            <p className="text-light text-sm">Chiffrement et logs détaillés</p>
          </div>
        </div>

        {/* CTA Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center mt-12">
          <Link href="/auth/register" className="btn-primary">
            Commencer gratuitement
          </Link>
          <Link href="/auth/login" className="btn-secondary">
            Se connecter
          </Link>
        </div>

        {/* OAuth Login */}
        <div className="mt-8">
          <p className="text-light mb-4">Ou connectez-vous directement avec :</p>
          <Link href="/auth/google" className="btn-google mx-auto max-w-xs">
            <svg className="w-5 h-5" viewBox="0 0 24 24">
              <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
            Continuer avec Google
          </Link>
        </div>
      </div>
    </div>
  );
}